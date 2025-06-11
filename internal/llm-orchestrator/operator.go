package llmorchestrator

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// LLMWorkloadOperator manages LLM workloads in Kubernetes
type LLMWorkloadOperator struct {
	client.Client
	logger          *zap.Logger
	scheme          *runtime.Scheme
	recorder        record.EventRecorder
	kubeClient      kubernetes.Interface
	scheduler       *LLMScheduler
	resourceManager *ResourceManager
	modelRegistry   *ModelRegistry
	config          *OperatorConfig
}

// OperatorConfig defines operator configuration
type OperatorConfig struct {
	// Reconciliation settings
	ReconcileInterval       time.Duration `yaml:"reconcileInterval"`
	MaxConcurrentReconciles int           `yaml:"maxConcurrentReconciles"`

	// Workload management
	DefaultNamespace      string `yaml:"defaultNamespace"`
	DefaultImage          string `yaml:"defaultImage"`
	DefaultServiceAccount string `yaml:"defaultServiceAccount"`

	// Resource defaults
	DefaultCPURequest    string `yaml:"defaultCPURequest"`
	DefaultMemoryRequest string `yaml:"defaultMemoryRequest"`
	DefaultCPULimit      string `yaml:"defaultCPULimit"`
	DefaultMemoryLimit   string `yaml:"defaultMemoryLimit"`

	// Monitoring
	MetricsEnabled     bool `yaml:"metricsEnabled"`
	HealthCheckEnabled bool `yaml:"healthCheckEnabled"`

	// Security
	PodSecurityContext *corev1.PodSecurityContext `yaml:"podSecurityContext"`
	SecurityContext    *corev1.SecurityContext    `yaml:"securityContext"`
}

// NewLLMWorkloadOperator creates a new LLM workload operator
func NewLLMWorkloadOperator(
	client client.Client,
	logger *zap.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
	kubeClient kubernetes.Interface,
	scheduler *LLMScheduler,
	resourceManager *ResourceManager,
	modelRegistry *ModelRegistry,
	config *OperatorConfig,
) *LLMWorkloadOperator {
	return &LLMWorkloadOperator{
		Client:          client,
		logger:          logger,
		scheme:          scheme,
		recorder:        recorder,
		kubeClient:      kubeClient,
		scheduler:       scheduler,
		resourceManager: resourceManager,
		modelRegistry:   modelRegistry,
		config:          config,
	}
}

// SetupWithManager sets up the operator with the controller manager
func (r *LLMWorkloadOperator) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&LLMWorkload{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.config.MaxConcurrentReconciles,
		}).
		Complete(r)
}

// Reconcile handles LLM workload reconciliation
func (r *LLMWorkloadOperator) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger.With(
		zap.String("namespace", req.Namespace),
		zap.String("name", req.Name),
	)

	log.Info("Reconciling LLM workload")

	// Fetch the LLMWorkload instance
	var workload LLMWorkload
	if err := r.Get(ctx, req.NamespacedName, &workload); err != nil {
		if errors.IsNotFound(err) {
			log.Info("LLM workload not found, probably deleted")
			return ctrl.Result{}, nil
		}
		log.Error("Failed to get LLM workload", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Handle deletion
	if !workload.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &workload)
	}

	// Add finalizer if not present
	if !containsFinalizer(workload.Finalizers, "llm-orchestrator.io/finalizer") {
		workload.Finalizers = append(workload.Finalizers, "llm-orchestrator.io/finalizer")
		if err := r.Update(ctx, &workload); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile the workload
	result, err := r.reconcileWorkload(ctx, &workload)
	if err != nil {
		log.Error("Failed to reconcile workload", zap.Error(err))
		r.recorder.Event(&workload, corev1.EventTypeWarning, "ReconcileError", err.Error())

		// Update status with error
		workload.Status.Phase = "Failed"
		workload.Status.Conditions = append(workload.Status.Conditions, WorkloadCondition{
			Type:               "Ready",
			Status:             "False",
			LastTransitionTime: metav1.Now(),
			Reason:             "ReconcileError",
			Message:            err.Error(),
		})

		if updateErr := r.Status().Update(ctx, &workload); updateErr != nil {
			log.Error("Failed to update status", zap.Error(updateErr))
		}

		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	log.Info("Successfully reconciled LLM workload")
	return result, nil
}

// reconcileWorkload handles the main reconciliation logic
func (r *LLMWorkloadOperator) reconcileWorkload(ctx context.Context, workload *LLMWorkload) (ctrl.Result, error) {
	log := r.logger.With(
		zap.String("workload", workload.Name),
		zap.String("namespace", workload.Namespace),
	)

	// Validate model exists in registry
	if err := r.validateModel(ctx, workload); err != nil {
		return ctrl.Result{}, fmt.Errorf("model validation failed: %w", err)
	}

	// Allocate resources
	allocation, err := r.resourceManager.AllocateResources(ctx, workload)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("resource allocation failed: %w", err)
	}

	// Create or update ConfigMap for model configuration
	if err := r.reconcileConfigMap(ctx, workload); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile ConfigMap: %w", err)
	}

	// Create or update Deployment
	if err := r.reconcileDeployment(ctx, workload, allocation); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile Deployment: %w", err)
	}

	// Create or update Service
	if err := r.reconcileService(ctx, workload); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile Service: %w", err)
	}

	// Update workload status
	if err := r.updateWorkloadStatus(ctx, workload); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	log.Info("Workload reconciliation completed successfully")
	return ctrl.Result{RequeueAfter: r.config.ReconcileInterval}, nil
}

// validateModel validates that the specified model exists in the registry
func (r *LLMWorkloadOperator) validateModel(ctx context.Context, workload *LLMWorkload) error {
	modelName := workload.Spec.ModelName
	modelVersion := workload.Spec.ModelVersion

	if modelVersion == "" {
		modelVersion = "latest"
	}

	_, err := r.modelRegistry.GetModel(ctx, modelName, modelVersion)
	if err != nil {
		return fmt.Errorf("model %s:%s not found in registry: %w", modelName, modelVersion, err)
	}

	return nil
}

// reconcileConfigMap creates or updates the ConfigMap for model configuration
func (r *LLMWorkloadOperator) reconcileConfigMap(ctx context.Context, workload *LLMWorkload) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workload.Name + "-config",
			Namespace: workload.Namespace,
		},
		Data: map[string]string{
			"model_name":    workload.Spec.ModelName,
			"model_version": workload.Spec.ModelVersion,
			"model_type":    workload.Spec.ModelType,
		},
	}

	// Add model parameters
	for key, value := range workload.Spec.Parameters {
		configMap.Data[key] = value
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(workload, configMap, r.scheme); err != nil {
		return err
	}

	// Create or update ConfigMap
	var existingConfigMap corev1.ConfigMap
	err := r.Get(ctx, client.ObjectKey{Name: configMap.Name, Namespace: configMap.Namespace}, &existingConfigMap)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, configMap)
		}
		return err
	}

	// Update existing ConfigMap
	existingConfigMap.Data = configMap.Data
	return r.Update(ctx, &existingConfigMap)
}

// reconcileDeployment creates or updates the Deployment for the LLM workload
func (r *LLMWorkloadOperator) reconcileDeployment(ctx context.Context, workload *LLMWorkload, allocation *ResourceAllocation) error {
	deployment := r.buildDeployment(workload, allocation)

	// Set owner reference
	if err := ctrl.SetControllerReference(workload, deployment, r.scheme); err != nil {
		return err
	}

	// Create or update Deployment
	var existingDeployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, &existingDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, deployment)
		}
		return err
	}

	// Update existing Deployment
	existingDeployment.Spec = deployment.Spec
	return r.Update(ctx, &existingDeployment)
}

// buildDeployment constructs a Deployment for the LLM workload
func (r *LLMWorkloadOperator) buildDeployment(workload *LLMWorkload, allocation *ResourceAllocation) *appsv1.Deployment {
	labels := map[string]string{
		"app":                          workload.Name,
		"llm-orchestrator.io/workload": workload.Name,
		"llm-orchestrator.io/model":    workload.Spec.ModelName,
		"llm-orchestrator.io/version":  workload.Spec.ModelVersion,
	}

	// Build container
	container := corev1.Container{
		Name:  "llm-server",
		Image: r.getModelImage(workload),
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 8080,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "grpc",
				ContainerPort: 9090,
				Protocol:      corev1.ProtocolTCP,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "MODEL_NAME",
				Value: workload.Spec.ModelName,
			},
			{
				Name:  "MODEL_VERSION",
				Value: workload.Spec.ModelVersion,
			},
			{
				Name:  "MODEL_TYPE",
				Value: workload.Spec.ModelType,
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    parseQuantity(allocation.AllocatedResources.CPU, r.config.DefaultCPURequest),
				corev1.ResourceMemory: parseQuantity(allocation.AllocatedResources.Memory, r.config.DefaultMemoryRequest),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    parseQuantity(allocation.AllocatedResources.CPU*1.5, r.config.DefaultCPULimit),
				corev1.ResourceMemory: parseQuantity(allocation.AllocatedResources.Memory*1.5, r.config.DefaultMemoryLimit),
			},
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/health",
					Port: intstr.FromString("http"),
				},
			},
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      5,
			FailureThreshold:    3,
		},
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/ready",
					Port: intstr.FromString("http"),
				},
			},
			InitialDelaySeconds: 10,
			PeriodSeconds:       5,
			TimeoutSeconds:      3,
			FailureThreshold:    3,
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config",
				MountPath: "/etc/config",
				ReadOnly:  true,
			},
		},
	}

	// Add GPU resources if specified
	if allocation.AllocatedResources.GPU > 0 {
		container.Resources.Requests["nvidia.com/gpu"] = parseQuantity(float64(allocation.AllocatedResources.GPU), "0")
		container.Resources.Limits["nvidia.com/gpu"] = parseQuantity(float64(allocation.AllocatedResources.GPU), "0")
	}

	// Build pod template
	podTemplate := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: r.config.DefaultServiceAccount,
			SecurityContext:    r.config.PodSecurityContext,
			Containers:         []corev1.Container{container},
			Volumes: []corev1.Volume{
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: workload.Name + "-config",
							},
						},
					},
				},
			},
			NodeSelector: map[string]string{
				"kubernetes.io/arch": "amd64",
			},
			Tolerations: []corev1.Toleration{
				{
					Key:      "nvidia.com/gpu",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	// Apply security context
	if r.config.SecurityContext != nil {
		container.SecurityContext = r.config.SecurityContext
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workload.Name,
			Namespace: workload.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &workload.Spec.Scaling.MinReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: podTemplate,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
					MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: "25%"},
				},
			},
		},
	}
}

// reconcileService creates or updates the Service for the LLM workload
func (r *LLMWorkloadOperator) reconcileService(ctx context.Context, workload *LLMWorkload) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workload.Name,
			Namespace: workload.Namespace,
			Labels: map[string]string{
				"app":                          workload.Name,
				"llm-orchestrator.io/workload": workload.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": workload.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http"),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "grpc",
					Port:       9090,
					TargetPort: intstr.FromString("grpc"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(workload, service, r.scheme); err != nil {
		return err
	}

	// Create or update Service
	var existingService corev1.Service
	err := r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, &existingService)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.Create(ctx, service)
		}
		return err
	}

	// Update existing Service
	existingService.Spec.Ports = service.Spec.Ports
	existingService.Spec.Selector = service.Spec.Selector
	return r.Update(ctx, &existingService)
}

// updateWorkloadStatus updates the status of the LLM workload
func (r *LLMWorkloadOperator) updateWorkloadStatus(ctx context.Context, workload *LLMWorkload) error {
	// Get deployment status
	var deployment appsv1.Deployment
	err := r.Get(ctx, client.ObjectKey{Name: workload.Name, Namespace: workload.Namespace}, &deployment)
	if err != nil {
		return err
	}

	// Update workload status based on deployment status
	workload.Status.CurrentReplicas = deployment.Status.Replicas
	workload.Status.ReadyReplicas = deployment.Status.ReadyReplicas

	if deployment.Status.ReadyReplicas == deployment.Status.Replicas && deployment.Status.Replicas > 0 {
		workload.Status.Phase = "Running"
		workload.Status.Conditions = []WorkloadCondition{
			{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: metav1.Now(),
				Reason:             "DeploymentReady",
				Message:            "All replicas are ready",
			},
		}
	} else {
		workload.Status.Phase = "Pending"
		workload.Status.Conditions = []WorkloadCondition{
			{
				Type:               "Ready",
				Status:             "False",
				LastTransitionTime: metav1.Now(),
				Reason:             "DeploymentNotReady",
				Message:            "Waiting for replicas to be ready",
			},
		}
	}

	// Update endpoints
	var service corev1.Service
	err = r.Get(ctx, client.ObjectKey{Name: workload.Name, Namespace: workload.Namespace}, &service)
	if err == nil {
		workload.Status.Endpoints = []string{
			fmt.Sprintf("http://%s.%s.svc.cluster.local", service.Name, service.Namespace),
			fmt.Sprintf("grpc://%s.%s.svc.cluster.local:9090", service.Name, service.Namespace),
		}
	}

	return r.Status().Update(ctx, workload)
}

// handleDeletion handles workload deletion
func (r *LLMWorkloadOperator) handleDeletion(ctx context.Context, workload *LLMWorkload) (ctrl.Result, error) {
	r.logger.Info("Handling workload deletion", zap.String("workload", workload.Name))

	// Perform cleanup tasks here
	// For example, deallocate resources, clean up external resources, etc.

	// Remove finalizer
	workload.Finalizers = removeFinalizer(workload.Finalizers, "llm-orchestrator.io/finalizer")
	if err := r.Update(ctx, workload); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Helper functions
func (r *LLMWorkloadOperator) getModelImage(workload *LLMWorkload) string {
	// This would typically look up the image from the model registry
	// For now, return a default image
	if r.config.DefaultImage != "" {
		return r.config.DefaultImage
	}
	return "llm-server:latest"
}

func parseQuantity(value interface{}, defaultValue string) resource.Quantity {
	var quantityStr string

	switch v := value.(type) {
	case string:
		quantityStr = v
	case float64:
		quantityStr = strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		quantityStr = strconv.Itoa(v)
	case int32:
		quantityStr = strconv.FormatInt(int64(v), 10)
	case int64:
		quantityStr = strconv.FormatInt(v, 10)
	default:
		quantityStr = defaultValue
	}

	if quantityStr == "" {
		quantityStr = defaultValue
	}

	quantity, err := resource.ParseQuantity(quantityStr)
	if err != nil {
		// Fall back to default value
		quantity, _ = resource.ParseQuantity(defaultValue)
	}

	return quantity
}

func containsFinalizer(finalizers []string, finalizer string) bool {
	for _, f := range finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}

func removeFinalizer(finalizers []string, finalizer string) []string {
	var result []string
	for _, f := range finalizers {
		if f != finalizer {
			result = append(result, f)
		}
	}
	return result
}
