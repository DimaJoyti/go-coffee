# Test script for LLM Orchestrator Simple (PowerShell)
param(
    [string]$BaseUrl = "http://localhost:8080",
    [int]$Port = 8080
)

# Global variables
$OrchestratorProcess = $null
$WorkloadId = ""

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

# Function to start orchestrator
function Start-Orchestrator {
    Write-Status "Starting LLM Orchestrator..."
    
    $processArgs = @(
        "--config=config/llm-orchestrator-simple.yaml"
        "--port=$Port"
        "--log-level=info"
    )
    
    try {
        $global:OrchestratorProcess = Start-Process -FilePath ".\bin\llm-orchestrator-simple.exe" -ArgumentList $processArgs -PassThru -NoNewWindow
        Start-Sleep -Seconds 3
        
        if ($global:OrchestratorProcess -and !$global:OrchestratorProcess.HasExited) {
            Write-Success "LLM Orchestrator started with PID: $($global:OrchestratorProcess.Id)"
            return $true
        } else {
            Write-Error "Failed to start LLM Orchestrator"
            return $false
        }
    } catch {
        Write-Error "Error starting orchestrator: $($_.Exception.Message)"
        return $false
    }
}

# Function to stop orchestrator
function Stop-Orchestrator {
    if ($global:OrchestratorProcess -and !$global:OrchestratorProcess.HasExited) {
        Write-Status "Stopping LLM Orchestrator..."
        try {
            $global:OrchestratorProcess.Kill()
            $global:OrchestratorProcess.WaitForExit(5000)
            Write-Success "LLM Orchestrator stopped"
        } catch {
            Write-Warning "Error stopping orchestrator: $($_.Exception.Message)"
        }
    }
}

# Function to make HTTP request
function Invoke-ApiRequest {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [string]$Body = $null,
        [hashtable]$Headers = @{}
    )
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
            UseBasicParsing = $true
        }
        
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }
        
        $response = Invoke-WebRequest @params
        return @{
            StatusCode = $response.StatusCode
            Content = $response.Content
            Success = $true
        }
    } catch {
        return @{
            StatusCode = $_.Exception.Response.StatusCode.value__
            Content = $_.Exception.Message
            Success = $false
        }
    }
}

# Function to test health endpoint
function Test-Health {
    Write-Status "Testing health endpoint..."
    $response = Invoke-ApiRequest -Url "$BaseUrl/health"
    
    if ($response.Success -and $response.StatusCode -eq 200) {
        Write-Success "Health check passed"
        Write-Host $response.Content
        return $true
    } else {
        Write-Error "Health check failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to test metrics endpoint
function Test-Metrics {
    Write-Status "Testing metrics endpoint..."
    $response = Invoke-ApiRequest -Url "$BaseUrl/metrics"
    
    if ($response.Success -and $response.StatusCode -eq 200) {
        Write-Success "Metrics endpoint working"
        Write-Host $response.Content
        return $true
    } else {
        Write-Error "Metrics endpoint failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to test workload creation
function Test-CreateWorkload {
    Write-Status "Testing workload creation..."
    
    $workloadData = @{
        name = "test-llama2"
        modelName = "llama2"
        modelType = "text-generation"
        resources = @{
            cpu = "2000m"
            memory = "8Gi"
            gpu = 1
        }
        labels = @{
            environment = "test"
            team = "ai-research"
        }
    } | ConvertTo-Json -Depth 3
    
    $response = Invoke-ApiRequest -Url "$BaseUrl/workloads" -Method "POST" -Body $workloadData
    
    if ($response.Success -and $response.StatusCode -eq 201) {
        Write-Success "Workload created successfully"
        Write-Host $response.Content
        
        # Extract workload ID
        try {
            $workload = $response.Content | ConvertFrom-Json
            $global:WorkloadId = $workload.id
            Write-Success "Workload ID: $global:WorkloadId"
        } catch {
            Write-Warning "Could not extract workload ID"
        }
        return $true
    } else {
        Write-Error "Workload creation failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to test workload listing
function Test-ListWorkloads {
    Write-Status "Testing workload listing..."
    $response = Invoke-ApiRequest -Url "$BaseUrl/workloads"
    
    if ($response.Success -and $response.StatusCode -eq 200) {
        Write-Success "Workload listing successful"
        Write-Host $response.Content
        return $true
    } else {
        Write-Error "Workload listing failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to test workload retrieval
function Test-GetWorkload {
    if (-not $global:WorkloadId) {
        Write-Warning "No workload ID available, skipping get test"
        return $true
    }
    
    Write-Status "Testing workload retrieval for ID: $global:WorkloadId"
    $response = Invoke-ApiRequest -Url "$BaseUrl/workloads/$global:WorkloadId"
    
    if ($response.Success -and $response.StatusCode -eq 200) {
        Write-Success "Workload retrieval successful"
        Write-Host $response.Content
        return $true
    } else {
        Write-Error "Workload retrieval failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to test scheduling
function Test-ScheduleWorkload {
    if (-not $global:WorkloadId) {
        Write-Warning "No workload ID available, skipping schedule test"
        return $true
    }
    
    Write-Status "Testing workload scheduling for ID: $global:WorkloadId"
    
    $scheduleData = @{
        workloadId = $global:WorkloadId
    } | ConvertTo-Json
    
    $response = Invoke-ApiRequest -Url "$BaseUrl/schedule" -Method "POST" -Body $scheduleData
    
    if ($response.Success -and $response.StatusCode -eq 200) {
        Write-Success "Workload scheduling successful"
        Write-Host $response.Content
        return $true
    } else {
        Write-Error "Workload scheduling failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to test status endpoint
function Test-Status {
    Write-Status "Testing status endpoint..."
    $response = Invoke-ApiRequest -Url "$BaseUrl/status"
    
    if ($response.Success -and $response.StatusCode -eq 200) {
        Write-Success "Status endpoint working"
        Write-Host $response.Content
        return $true
    } else {
        Write-Error "Status endpoint failed with status: $($response.StatusCode)"
        Write-Host $response.Content
        return $false
    }
}

# Function to cleanup
function Invoke-Cleanup {
    Write-Status "Cleaning up..."
    Stop-Orchestrator
    Write-Success "Cleanup completed"
}

# Main test execution
function Invoke-MainTests {
    Write-Status "Starting LLM Orchestrator API Tests"
    
    # Check if binary exists
    if (-not (Test-Path ".\bin\llm-orchestrator-simple.exe")) {
        Write-Error "LLM Orchestrator binary not found. Please build it first:"
        Write-Error "go build -o bin/llm-orchestrator-simple.exe ./cmd/llm-orchestrator-simple"
        return $false
    }
    
    # Check if config exists
    if (-not (Test-Path "config/llm-orchestrator-simple.yaml")) {
        Write-Error "Configuration file not found: config/llm-orchestrator-simple.yaml"
        return $false
    }
    
    try {
        # Start orchestrator
        if (-not (Start-Orchestrator)) {
            return $false
        }
        
        # Wait for startup
        Start-Sleep -Seconds 2
        
        # Run tests
        $testResults = @()
        $testResults += Test-Health
        $testResults += Test-Metrics
        $testResults += Test-Status
        $testResults += Test-ListWorkloads
        $testResults += Test-CreateWorkload
        $testResults += Test-GetWorkload
        $testResults += Test-ScheduleWorkload
        
        # Wait for metrics to update
        Start-Sleep -Seconds 5
        
        # Test metrics again
        Write-Status "Testing updated metrics..."
        $testResults += Test-Metrics
        
        # Check if all tests passed
        $failedTests = $testResults | Where-Object { $_ -eq $false }
        if ($failedTests.Count -eq 0) {
            Write-Success "All tests passed! 🎉"
            return $true
        } else {
            Write-Error "Some tests failed"
            return $false
        }
    } finally {
        Invoke-Cleanup
    }
}

# Set up error handling
$ErrorActionPreference = "Continue"

# Run main tests
try {
    $result = Invoke-MainTests
    if ($result) {
        exit 0
    } else {
        exit 1
    }
} catch {
    Write-Error "Unexpected error: $($_.Exception.Message)"
    Invoke-Cleanup
    exit 1
}
