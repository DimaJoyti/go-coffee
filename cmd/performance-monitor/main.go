package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type PerformanceMonitor struct {
	services  map[string]*ServiceMetrics
	mu        sync.RWMutex
	startTime time.Time
}

type ServiceMetrics struct {
	Name           string                 `json:"name"`
	Port           int                    `json:"port"`
	Status         string                 `json:"status"`
	ResponseTime   time.Duration          `json:"response_time"`
	LastCheck      time.Time              `json:"last_check"`
	HealthEndpoint string                 `json:"health_endpoint"`
	RequestCount   int64                  `json:"request_count"`
	ErrorCount     int64                  `json:"error_count"`
	CPUUsage       float64                `json:"cpu_usage"`
	MemoryUsage    uint64                 `json:"memory_usage"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
}

type SystemMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	MemoryTotal uint64    `json:"memory_total"`
	MemoryUsed  uint64    `json:"memory_used"`
	GoRoutines  int       `json:"goroutines"`
	Uptime      string    `json:"uptime"`
	Timestamp   time.Time `json:"timestamp"`
}

type PerformanceReport struct {
	SystemMetrics SystemMetrics              `json:"system_metrics"`
	Services      map[string]*ServiceMetrics `json:"services"`
	Summary       PerformanceSummary         `json:"summary"`
	Timestamp     time.Time                  `json:"timestamp"`
}

type PerformanceSummary struct {
	TotalServices       int           `json:"total_services"`
	HealthyServices     int           `json:"healthy_services"`
	UnhealthyServices   int           `json:"unhealthy_services"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TotalRequests       int64         `json:"total_requests"`
	TotalErrors         int64         `json:"total_errors"`
	ErrorRate           float64       `json:"error_rate"`
}

func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		services:  make(map[string]*ServiceMetrics),
		startTime: time.Now(),
	}
}

func (pm *PerformanceMonitor) RegisterService(name string, port int, healthEndpoint string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.services[name] = &ServiceMetrics{
		Name:           name,
		Port:           port,
		HealthEndpoint: healthEndpoint,
		Status:         "unknown",
		CustomMetrics:  make(map[string]interface{}),
	}
}

func (pm *PerformanceMonitor) checkServiceHealth(service *ServiceMetrics) {
	start := time.Now()

	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://localhost:%d%s", service.Port, service.HealthEndpoint)

	resp, err := client.Get(url)
	service.ResponseTime = time.Since(start)
	service.LastCheck = time.Now()
	service.RequestCount++

	if err != nil {
		service.Status = "unhealthy"
		service.ErrorCount++
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		service.Status = "healthy"
	} else {
		service.Status = "unhealthy"
		service.ErrorCount++
	}
}

func (pm *PerformanceMonitor) getSystemMetrics() SystemMetrics {
	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()

	var avgCPU float64
	if len(cpuPercent) > 0 {
		avgCPU = cpuPercent[0]
	}

	return SystemMetrics{
		CPUUsage:    avgCPU,
		MemoryUsage: memInfo.UsedPercent,
		MemoryTotal: memInfo.Total,
		MemoryUsed:  memInfo.Used,
		GoRoutines:  runtime.NumGoroutine(),
		Uptime:      time.Since(pm.startTime).String(),
		Timestamp:   time.Now(),
	}
}

func (pm *PerformanceMonitor) generateReport() PerformanceReport {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var totalResponseTime time.Duration
	var totalRequests, totalErrors int64
	healthyCount := 0

	for _, service := range pm.services {
		totalResponseTime += service.ResponseTime
		totalRequests += service.RequestCount
		totalErrors += service.ErrorCount
		if service.Status == "healthy" {
			healthyCount++
		}
	}

	avgResponseTime := time.Duration(0)
	if len(pm.services) > 0 {
		avgResponseTime = totalResponseTime / time.Duration(len(pm.services))
	}

	errorRate := float64(0)
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	return PerformanceReport{
		SystemMetrics: pm.getSystemMetrics(),
		Services:      pm.services,
		Summary: PerformanceSummary{
			TotalServices:       len(pm.services),
			HealthyServices:     healthyCount,
			UnhealthyServices:   len(pm.services) - healthyCount,
			AverageResponseTime: avgResponseTime,
			TotalRequests:       totalRequests,
			TotalErrors:         totalErrors,
			ErrorRate:           errorRate,
		},
		Timestamp: time.Now(),
	}
}

func (pm *PerformanceMonitor) startMonitoring(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.mu.RLock()
			services := make([]*ServiceMetrics, 0, len(pm.services))
			for _, service := range pm.services {
				services = append(services, service)
			}
			pm.mu.RUnlock()

			for _, service := range services {
				go pm.checkServiceHealth(service)
			}
		}
	}
}

func main() {
	fmt.Println("🚀 Go Coffee Performance Monitor Starting...")
	fmt.Println("📊 Real-time performance monitoring and optimization")

	monitor := NewPerformanceMonitor()

	// Register Go Coffee services
	monitor.RegisterService("api-gateway", 8080, "/health")
	monitor.RegisterService("redis-mcp-server", 8108, "/api/v1/redis-mcp/health")
	monitor.RegisterService("web-ui-mcp-server", 3001, "/health")
	monitor.RegisterService("web-ui-backend", 8090, "/health")
	monitor.RegisterService("ai-search", 8092, "/api/v1/ai-search/health")
	monitor.RegisterService("frontend", 3002, "/")

	// Start monitoring in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitor.startMonitoring(ctx)

	// Setup HTTP server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Performance monitoring endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "performance-monitor",
			"timestamp": time.Now(),
			"uptime":    time.Since(monitor.startTime).String(),
		})
	})

	r.GET("/metrics", func(c *gin.Context) {
		report := monitor.generateReport()
		c.JSON(200, report)
	})

	r.GET("/metrics/prometheus", func(c *gin.Context) {
		report := monitor.generateReport()

		// Generate Prometheus format metrics
		var metrics []string
		metrics = append(metrics, fmt.Sprintf("# HELP go_coffee_system_cpu_usage System CPU usage percentage"))
		metrics = append(metrics, fmt.Sprintf("# TYPE go_coffee_system_cpu_usage gauge"))
		metrics = append(metrics, fmt.Sprintf("go_coffee_system_cpu_usage %.2f", report.SystemMetrics.CPUUsage))

		metrics = append(metrics, fmt.Sprintf("# HELP go_coffee_system_memory_usage System memory usage percentage"))
		metrics = append(metrics, fmt.Sprintf("# TYPE go_coffee_system_memory_usage gauge"))
		metrics = append(metrics, fmt.Sprintf("go_coffee_system_memory_usage %.2f", report.SystemMetrics.MemoryUsage))

		for name, service := range report.Services {
			metrics = append(metrics, fmt.Sprintf("# HELP go_coffee_service_response_time_seconds Service response time in seconds"))
			metrics = append(metrics, fmt.Sprintf("# TYPE go_coffee_service_response_time_seconds gauge"))
			metrics = append(metrics, fmt.Sprintf("go_coffee_service_response_time_seconds{service=\"%s\"} %.6f", name, service.ResponseTime.Seconds()))

			status := 0.0
			if service.Status == "healthy" {
				status = 1.0
			}
			metrics = append(metrics, fmt.Sprintf("# HELP go_coffee_service_up Service availability"))
			metrics = append(metrics, fmt.Sprintf("# TYPE go_coffee_service_up gauge"))
			metrics = append(metrics, fmt.Sprintf("go_coffee_service_up{service=\"%s\"} %.0f", name, status))
		}

		c.Header("Content-Type", "text/plain")
		for _, metric := range metrics {
			c.String(200, metric+"\n")
		}
	})

	r.GET("/dashboard", func(c *gin.Context) {
		report := monitor.generateReport()

		html := generateDashboardHTML(report)
		c.Header("Content-Type", "text/html")
		c.String(200, html)
	})

	port := os.Getenv("MONITOR_PORT")
	if port == "" {
		port = "9999"
	}

	fmt.Printf("🌐 Performance Monitor running on http://localhost:%s\n", port)
	fmt.Printf("📊 Dashboard: http://localhost:%s/dashboard\n", port)
	fmt.Printf("📈 Metrics: http://localhost:%s/metrics\n", port)
	fmt.Printf("🔍 Prometheus: http://localhost:%s/metrics/prometheus\n", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func generateDashboardHTML(report PerformanceReport) string {

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Go Coffee Performance Monitor</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 20px; border-radius: 10px; margin-bottom: 20px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .metric-value { font-size: 2em; font-weight: bold; color: #667eea; }
        .metric-label { color: #666; margin-top: 5px; }
        .service-status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 0.8em; font-weight: bold; }
        .healthy { background: #d4edda; color: #155724; }
        .unhealthy { background: #f8d7da; color: #721c24; }
        .services-table { width: 100%%; border-collapse: collapse; background: white; border-radius: 10px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .services-table th, .services-table td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        .services-table th { background: #f8f9fa; font-weight: 600; }
        .refresh-btn { background: #667eea; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; margin-bottom: 20px; }
        .refresh-btn:hover { background: #5a6fd8; }
    </style>
    <script>
        function refreshData() {
            location.reload();
        }
        setInterval(refreshData, 30000); // Auto-refresh every 30 seconds
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>☕ Go Coffee Performance Monitor</h1>
            <p>Real-time system and service performance monitoring</p>
            <p>Last updated: %s</p>
        </div>
        
        <button class="refresh-btn" onclick="refreshData()">🔄 Refresh Data</button>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-value">%.1f%%</div>
                <div class="metric-label">CPU Usage</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">%.1f%%</div>
                <div class="metric-label">Memory Usage</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">%d</div>
                <div class="metric-label">Total Services</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">%d</div>
                <div class="metric-label">Healthy Services</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">%.2fms</div>
                <div class="metric-label">Avg Response Time</div>
            </div>
            <div class="metric-card">
                <div class="metric-value">%.2f%%</div>
                <div class="metric-label">Error Rate</div>
            </div>
        </div>
        
        <div class="metric-card">
            <h3>Service Status</h3>
            <table class="services-table">
                <thead>
                    <tr>
                        <th>Service</th>
                        <th>Status</th>
                        <th>Port</th>
                        <th>Response Time</th>
                        <th>Requests</th>
                        <th>Errors</th>
                        <th>Last Check</th>
                    </tr>
                </thead>
                <tbody>`,
		report.Timestamp.Format("2006-01-02 15:04:05"),
		report.SystemMetrics.CPUUsage,
		report.SystemMetrics.MemoryUsage,
		report.Summary.TotalServices,
		report.Summary.HealthyServices,
		float64(report.Summary.AverageResponseTime.Nanoseconds())/1000000,
		report.Summary.ErrorRate) + generateServiceRows(report.Services) + `
                </tbody>
            </table>
        </div>
    </div>
</body>
</html>`
}

func generateServiceRows(services map[string]*ServiceMetrics) string {
	var rows string
	for name, service := range services {
		statusClass := "unhealthy"
		if service.Status == "healthy" {
			statusClass = "healthy"
		}

		rows += fmt.Sprintf(`
                    <tr>
                        <td>%s</td>
                        <td><span class="service-status %s">%s</span></td>
                        <td>%d</td>
                        <td>%.2fms</td>
                        <td>%d</td>
                        <td>%d</td>
                        <td>%s</td>
                    </tr>`,
			name,
			statusClass,
			service.Status,
			service.Port,
			float64(service.ResponseTime.Nanoseconds())/1000000,
			service.RequestCount,
			service.ErrorCount,
			service.LastCheck.Format("15:04:05"))
	}
	return rows
}
