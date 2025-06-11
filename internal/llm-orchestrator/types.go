package llmorchestrator

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// LLMWorkloadSpec defines the desired state of an LLM workload
type LLMWorkloadSpec struct {
	// Model configuration
	ModelName    string            `json:"modelName"`
	ModelVersion string            `json:"modelVersion"`
	ModelSize    string            `json:"modelSize"` // small, medium, large, xlarge
	ModelType    string            `json:"modelType"` // text-generation, embedding, classification
	Parameters   map[string]string `json:"parameters,omitempty"`

	// Resource requirements
	Resources ResourceRequirements `json:"resources"`

	// Scaling configuration
	Scaling ScalingConfig `json:"scaling"`

	// Performance requirements
	Performance PerformanceConfig `json:"performance"`

	// Security configuration
	Security SecurityConfig `json:"security,omitempty"`
}

// ResourceRequirements defines compute resource needs
type ResourceRequirements struct {
	CPU              string `json:"cpu"`               // e.g., "2000m"
	Memory           string `json:"memory"`            // e.g., "8Gi"
	GPU              string `json:"gpu,omitempty"`     // e.g., "1"
	GPUType          string `json:"gpuType,omitempty"` // e.g., "nvidia.com/gpu"
	Storage          string `json:"storage,omitempty"` // e.g., "100Gi"
	NetworkBandwidth string `json:"networkBandwidth,omitempty"`
}

// ScalingConfig defines auto-scaling behavior
type ScalingConfig struct {
	MinReplicas     int32            `json:"minReplicas"`
	MaxReplicas     int32            `json:"maxReplicas"`
	TargetMetrics   []ScalingMetric  `json:"targetMetrics"`
	ScalingBehavior *ScalingBehavior `json:"scalingBehavior,omitempty"`
	Strategy        string           `json:"strategy"` // horizontal, vertical, hybrid
}

// ScalingMetric defines metrics for auto-scaling
type ScalingMetric struct {
	Type   string  `json:"type"`   // cpu, memory, requests_per_second, queue_length, latency
	Target float64 `json:"target"` // target value for the metric
}

// ScalingBehavior defines scaling policies
type ScalingBehavior struct {
	ScaleUp   *ScalingPolicy `json:"scaleUp,omitempty"`
	ScaleDown *ScalingPolicy `json:"scaleDown,omitempty"`
}

// ScalingPolicy defines scaling rate limits
type ScalingPolicy struct {
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds"`
	MaxChangePercent           int32 `json:"maxChangePercent"`
	MaxChangePods              int32 `json:"maxChangePods"`
}

// PerformanceConfig defines performance requirements
type PerformanceConfig struct {
	MaxLatency      time.Duration `json:"maxLatency"`      // e.g., 500ms
	MinThroughput   float64       `json:"minThroughput"`   // requests per second
	MaxTokensPerSec int32         `json:"maxTokensPerSec"` // tokens per second
	BatchSize       int32         `json:"batchSize"`       // optimal batch size
	ConcurrentUsers int32         `json:"concurrentUsers"` // expected concurrent users
	SLARequirements SLAConfig     `json:"slaRequirements"`
}

// SLAConfig defines service level agreements
type SLAConfig struct {
	Availability    float64 `json:"availability"`    // e.g., 99.9
	ResponseTimeP95 int32   `json:"responseTimeP95"` // 95th percentile response time in ms
	ResponseTimeP99 int32   `json:"responseTimeP99"` // 99th percentile response time in ms
	ErrorRate       float64 `json:"errorRate"`       // maximum acceptable error rate
}

// SecurityConfig defines security requirements
type SecurityConfig struct {
	Encryption         bool     `json:"encryption"`
	AccessControl      []string `json:"accessControl,omitempty"`
	NetworkPolicies    []string `json:"networkPolicies,omitempty"`
	SecretRefs         []string `json:"secretRefs,omitempty"`
	ComplianceLevel    string   `json:"complianceLevel,omitempty"`    // basic, strict, enterprise
	DataClassification string   `json:"dataClassification,omitempty"` // public, internal, confidential
}

// LLMWorkloadStatus defines the observed state
type LLMWorkloadStatus struct {
	Phase              string              `json:"phase"`
	Conditions         []WorkloadCondition `json:"conditions,omitempty"`
	CurrentReplicas    int32               `json:"currentReplicas"`
	ReadyReplicas      int32               `json:"readyReplicas"`
	LastScaleTime      *metav1.Time        `json:"lastScaleTime,omitempty"`
	CurrentMetrics     map[string]float64  `json:"currentMetrics,omitempty"`
	ResourceUsage      ResourceUsage       `json:"resourceUsage,omitempty"`
	PerformanceMetrics PerformanceMetrics  `json:"performanceMetrics,omitempty"`
	Endpoints          []string            `json:"endpoints,omitempty"`
}

// WorkloadCondition describes the state of a workload
type WorkloadCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
}

// ResourceUsage tracks actual resource consumption
type ResourceUsage struct {
	CPU     float64 `json:"cpu"`     // CPU usage in cores
	Memory  float64 `json:"memory"`  // Memory usage in bytes
	GPU     float64 `json:"gpu"`     // GPU utilization percentage
	Network float64 `json:"network"` // Network I/O in bytes/sec
	Storage float64 `json:"storage"` // Storage I/O in bytes/sec
}

// PerformanceMetrics tracks performance indicators
type PerformanceMetrics struct {
	RequestsPerSecond float64       `json:"requestsPerSecond"`
	AverageLatency    time.Duration `json:"averageLatency"`
	P95Latency        time.Duration `json:"p95Latency"`
	P99Latency        time.Duration `json:"p99Latency"`
	TokensPerSecond   float64       `json:"tokensPerSecond"`
	ErrorRate         float64       `json:"errorRate"`
	QueueLength       int32         `json:"queueLength"`
	ActiveConnections int32         `json:"activeConnections"`
	ThroughputMBps    float64       `json:"throughputMBps"`
	ModelAccuracy     float64       `json:"modelAccuracy,omitempty"`
	LastUpdated       metav1.Time   `json:"lastUpdated"`
}

// LLMWorkload represents a complete LLM workload resource
type LLMWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMWorkloadSpec   `json:"spec,omitempty"`
	Status LLMWorkloadStatus `json:"status,omitempty"`
}

// LLMWorkloadList contains a list of LLMWorkload
type LLMWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLMWorkload `json:"items"`
}

// DeepCopyObject returns a generically typed copy of an object
func (in *LLMWorkload) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy returns a deep copy of the LLMWorkload
func (in *LLMWorkload) DeepCopy() *LLMWorkload {
	if in == nil {
		return nil
	}
	out := new(LLMWorkload)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *LLMWorkload) DeepCopyInto(out *LLMWorkload) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopyInto copies all properties of LLMWorkloadSpec
func (in *LLMWorkloadSpec) DeepCopyInto(out *LLMWorkloadSpec) {
	*out = *in
	if in.Parameters != nil {
		in, out := &in.Parameters, &out.Parameters
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	in.Scaling.DeepCopyInto(&out.Scaling)
	in.Performance.DeepCopyInto(&out.Performance)
	in.Security.DeepCopyInto(&out.Security)
}

// DeepCopyInto copies all properties of ResourceRequirements
func (in *ResourceRequirements) DeepCopyInto(out *ResourceRequirements) {
	*out = *in
}

// DeepCopyInto copies all properties of ScalingConfig
func (in *ScalingConfig) DeepCopyInto(out *ScalingConfig) {
	*out = *in
	if in.TargetMetrics != nil {
		in, out := &in.TargetMetrics, &out.TargetMetrics
		*out = make([]ScalingMetric, len(*in))
		copy(*out, *in)
	}
	if in.ScalingBehavior != nil {
		in, out := &in.ScalingBehavior, &out.ScalingBehavior
		*out = new(ScalingBehavior)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopyInto copies all properties of ScalingBehavior
func (in *ScalingBehavior) DeepCopyInto(out *ScalingBehavior) {
	*out = *in
	if in.ScaleUp != nil {
		in, out := &in.ScaleUp, &out.ScaleUp
		*out = new(ScalingPolicy)
		**out = **in
	}
	if in.ScaleDown != nil {
		in, out := &in.ScaleDown, &out.ScaleDown
		*out = new(ScalingPolicy)
		**out = **in
	}
}

// DeepCopyInto copies all properties of PerformanceConfig
func (in *PerformanceConfig) DeepCopyInto(out *PerformanceConfig) {
	*out = *in
	out.SLARequirements = in.SLARequirements
}

// DeepCopyInto copies all properties of SecurityConfig
func (in *SecurityConfig) DeepCopyInto(out *SecurityConfig) {
	*out = *in
	if in.AccessControl != nil {
		in, out := &in.AccessControl, &out.AccessControl
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.NetworkPolicies != nil {
		in, out := &in.NetworkPolicies, &out.NetworkPolicies
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SecretRefs != nil {
		in, out := &in.SecretRefs, &out.SecretRefs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopyInto copies all properties of LLMWorkloadStatus
func (in *LLMWorkloadStatus) DeepCopyInto(out *LLMWorkloadStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]WorkloadCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.LastScaleTime != nil {
		in, out := &in.LastScaleTime, &out.LastScaleTime
		*out = (*in).DeepCopy()
	}
	if in.CurrentMetrics != nil {
		in, out := &in.CurrentMetrics, &out.CurrentMetrics
		*out = make(map[string]float64, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.ResourceUsage = in.ResourceUsage
	out.PerformanceMetrics = in.PerformanceMetrics
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopyInto copies all properties of WorkloadCondition
func (in *WorkloadCondition) DeepCopyInto(out *WorkloadCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// WorkloadController interface defines the contract for LLM workload management
type WorkloadController interface {
	// Lifecycle management
	CreateWorkload(ctx context.Context, workload *LLMWorkload) error
	UpdateWorkload(ctx context.Context, workload *LLMWorkload) error
	DeleteWorkload(ctx context.Context, name, namespace string) error
	GetWorkload(ctx context.Context, name, namespace string) (*LLMWorkload, error)
	ListWorkloads(ctx context.Context, namespace string) (*LLMWorkloadList, error)

	// Scaling operations
	ScaleWorkload(ctx context.Context, name, namespace string, replicas int32) error
	GetWorkloadMetrics(ctx context.Context, name, namespace string) (*PerformanceMetrics, error)

	// Health and status
	GetWorkloadStatus(ctx context.Context, name, namespace string) (*LLMWorkloadStatus, error)
	UpdateWorkloadStatus(ctx context.Context, workload *LLMWorkload) error
}

// GroupVersion is group version used to register these objects
var GroupVersion = schema.GroupVersion{Group: "llm.orchestrator.io", Version: "v1"}

// SchemeBuilder is used to add go types to the GroupVersionKind scheme
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme adds the types in this group-version to the given scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&LLMWorkload{},
		&LLMWorkloadList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
