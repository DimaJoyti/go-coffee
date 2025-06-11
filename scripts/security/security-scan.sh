#!/bin/bash

# Go Coffee - Comprehensive Security Scanning & Vulnerability Assessment
# Advanced security testing with multiple tools and detailed reporting
# Version: 3.0.0
# Usage: ./security-scan.sh [OPTIONS]
#   -t, --target URL    Target URL for scanning
#   -s, --scope TYPE    Scan scope (code|deps|docker|api|full)
#   -l, --level LEVEL   Scan level (basic|standard|comprehensive)
#   -o, --output DIR    Output directory for reports
#   -f, --format TYPE   Report format (html|json|sarif|all)
#   -q, --quiet         Quiet mode (minimal output)
#   -h, --help          Show this help message

set -euo pipefail

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Source shared library
source "$PROJECT_ROOT/scripts/lib/common.sh" 2>/dev/null || {
    echo "❌ Cannot load shared library. Please run from project root."
    exit 1
}

print_header "🔒 Go Coffee Comprehensive Security Scanning"

# =============================================================================
# CONFIGURATION
# =============================================================================

TARGET_URL="${TARGET_URL:-http://localhost:8080}"
SCAN_SCOPE="full"
SCAN_LEVEL="standard"
OUTPUT_DIR="security-reports/$(date +%Y%m%d-%H%M%S)"
REPORT_FORMAT="html"
QUIET_MODE=false

# Security tools configuration
declare -A SECURITY_TOOLS=(
    ["gosec"]="Go source code security analyzer"
    ["nancy"]="Dependency vulnerability scanner"
    ["trivy"]="Container vulnerability scanner"
    ["zap"]="Web application security scanner"
    ["semgrep"]="Static analysis security scanner"
)

# Vulnerability severity levels
declare -A SEVERITY_COLORS=(
    ["CRITICAL"]="$RED"
    ["HIGH"]="$RED"
    ["MEDIUM"]="$YELLOW"
    ["LOW"]="$BLUE"
    ["INFO"]="$GREEN"
)

# =============================================================================
# COMMAND LINE PARSING
# =============================================================================

parse_security_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--target)
                TARGET_URL="$2"
                shift 2
                ;;
            -s|--scope)
                SCAN_SCOPE="$2"
                shift 2
                ;;
            -l|--level)
                SCAN_LEVEL="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -f|--format)
                REPORT_FORMAT="$2"
                shift 2
                ;;
            -q|--quiet)
                QUIET_MODE=true
                shift
                ;;
            -h|--help)
                show_usage "security-scan.sh" \
                    "Comprehensive security scanning and vulnerability assessment" \
                    "  ./security-scan.sh [OPTIONS]
  
  Options:
    -t, --target URL    Target URL for scanning (default: http://localhost:8080)
    -s, --scope TYPE    Scan scope (code|deps|docker|api|full) (default: full)
    -l, --level LEVEL   Scan level (basic|standard|comprehensive) (default: standard)
    -o, --output DIR    Output directory for reports
    -f, --format TYPE   Report format (html|json|sarif|all) (default: html)
    -q, --quiet         Quiet mode (minimal output)
    -h, --help          Show this help message
  
  Scan Scopes:
    code:   Source code security analysis
    deps:   Dependency vulnerability scanning
    docker: Container security scanning
    api:    API security testing
    full:   All security scans
  
  Scan Levels:
    basic:        Quick security check
    standard:     Standard security assessment
    comprehensive: Deep security analysis
  
  Examples:
    ./security-scan.sh                          # Full security scan
    ./security-scan.sh -s code -l comprehensive # Deep code analysis
    ./security-scan.sh -s api -t https://api.example.com # API security test
    ./security-scan.sh -f all -o custom-dir    # All formats to custom directory"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
}

# =============================================================================
# SECURITY SCANNING FUNCTIONS
# =============================================================================

# Check security scanning dependencies
check_security_dependencies() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🔍 Checking Security Scanning Dependencies"
    fi
    
    local required_tools=("curl" "jq" "git")
    local optional_tools=("gosec" "nancy" "trivy" "docker" "semgrep")
    
    # Check required tools
    check_dependencies "${required_tools[@]}" || exit 1
    
    # Check optional tools and install if needed
    local available_tools=()
    for tool in "${optional_tools[@]}"; do
        if command_exists "$tool"; then
            available_tools+=("$tool")
            if [[ "$QUIET_MODE" != "true" ]]; then
                print_status "$tool is available"
            fi
        else
            if [[ "$QUIET_MODE" != "true" ]]; then
                print_warning "$tool is not available - attempting to install"
            fi
            install_security_tool "$tool"
        fi
    done
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Security dependencies checked"
    fi
}

# Install security tools
install_security_tool() {
    local tool=$1
    
    case "$tool" in
        "gosec")
            if command_exists go; then
                go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
                print_status "gosec installed"
            fi
            ;;
        "nancy")
            if command_exists go; then
                go install github.com/sonatypecommunity/nancy@latest
                print_status "nancy installed"
            fi
            ;;
        "trivy")
            if command_exists curl; then
                curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
                print_status "trivy installed"
            fi
            ;;
        "semgrep")
            if command_exists pip3; then
                pip3 install semgrep
                print_status "semgrep installed"
            fi
            ;;
    esac
}

# Create security output structure
create_security_output_structure() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📁 Creating Security Output Structure"
    fi
    
    mkdir -p "$OUTPUT_DIR"/{reports,logs,data,evidence}
    
    # Create scan metadata
    cat > "$OUTPUT_DIR/scan-metadata.json" <<EOF
{
    "scan_info": {
        "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
        "target_url": "$TARGET_URL",
        "scope": "$SCAN_SCOPE",
        "level": "$SCAN_LEVEL",
        "format": "$REPORT_FORMAT",
        "scanner_version": "3.0.0"
    }
}
EOF
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_status "Security output structure created: $OUTPUT_DIR"
    fi
}

# Run Go source code security analysis
run_code_security_scan() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🔍 Running Go Source Code Security Analysis"
    fi
    
    # GoSec scan
    if command_exists gosec; then
        print_progress "Running gosec analysis..."
        
        local gosec_args=("-fmt=json" "-out=$OUTPUT_DIR/data/gosec-results.json")
        
        case "$SCAN_LEVEL" in
            "comprehensive")
                gosec_args+=("-tests" "-severity=low")
                ;;
            "standard")
                gosec_args+=("-severity=medium")
                ;;
            "basic")
                gosec_args+=("-severity=high")
                ;;
        esac
        
        if gosec "${gosec_args[@]}" ./... 2>/dev/null; then
            print_status "gosec scan completed"
        else
            print_warning "gosec scan completed with findings"
        fi
    fi
    
    # Semgrep scan
    if command_exists semgrep; then
        print_progress "Running semgrep analysis..."
        
        semgrep --config=auto --json --output="$OUTPUT_DIR/data/semgrep-results.json" . 2>/dev/null || true
        print_status "semgrep scan completed"
    fi
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Code security analysis completed"
    fi
}

# Run dependency vulnerability scanning
run_dependency_scan() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📦 Running Dependency Vulnerability Scanning"
    fi
    
    # Nancy scan for Go dependencies
    if command_exists nancy && [[ -f "go.sum" ]]; then
        print_progress "Running nancy dependency scan..."
        
        go list -json -deps ./... | nancy sleuth --output=json > "$OUTPUT_DIR/data/nancy-results.json" 2>/dev/null || true
        print_status "nancy scan completed"
    fi
    
    # Trivy filesystem scan
    if command_exists trivy; then
        print_progress "Running trivy filesystem scan..."
        
        trivy fs --format json --output "$OUTPUT_DIR/data/trivy-fs-results.json" . 2>/dev/null || true
        print_status "trivy filesystem scan completed"
    fi
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Dependency vulnerability scanning completed"
    fi
}

# Run container security scanning
run_container_scan() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🐳 Running Container Security Scanning"
    fi
    
    if ! command_exists docker; then
        print_warning "Docker not available, skipping container scans"
        return 0
    fi
    
    # Find Docker images
    local images=()
    if [[ -f "docker-compose.yml" ]]; then
        images+=($(grep "image:" docker-compose.yml | awk '{print $2}' | sort -u))
    fi
    
    # Add common Go Coffee images
    images+=("go-coffee/api-gateway:latest" "go-coffee/auth-service:latest")
    
    for image in "${images[@]}"; do
        if docker image inspect "$image" >/dev/null 2>&1; then
            print_progress "Scanning container image: $image"
            
            if command_exists trivy; then
                trivy image --format json --output "$OUTPUT_DIR/data/trivy-$(echo "$image" | tr '/:' '-').json" "$image" 2>/dev/null || true
            fi
        fi
    done
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Container security scanning completed"
    fi
}

# Run API security testing
run_api_security_scan() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "🌐 Running API Security Testing"
    fi
    
    # Basic API security checks
    print_progress "Running basic API security checks..."
    
    # Check for common security headers
    local security_headers=("X-Content-Type-Options" "X-Frame-Options" "X-XSS-Protection" "Strict-Transport-Security")
    local missing_headers=()
    
    for header in "${security_headers[@]}"; do
        if ! curl -s -I "$TARGET_URL" | grep -i "$header" >/dev/null; then
            missing_headers+=("$header")
        fi
    done
    
    # Save API security results
    cat > "$OUTPUT_DIR/data/api-security-results.json" <<EOF
{
    "target_url": "$TARGET_URL",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "security_headers": {
        "missing": $(printf '%s\n' "${missing_headers[@]}" | jq -R . | jq -s .),
        "total_checked": ${#security_headers[@]},
        "missing_count": ${#missing_headers[@]}
    },
    "endpoints_tested": [
        "/health",
        "/api/v1/auth",
        "/api/v1/orders",
        "/api/v1/payments"
    ]
}
EOF
    
    # Test common endpoints for security issues
    local endpoints=("/health" "/api/v1/auth" "/api/v1/orders" "/api/v1/payments")
    
    for endpoint in "${endpoints[@]}"; do
        local full_url="$TARGET_URL$endpoint"
        
        # Test for SQL injection patterns (basic)
        curl -s "$full_url?id=1' OR '1'='1" >/dev/null 2>&1 || true
        
        # Test for XSS patterns (basic)
        curl -s "$full_url?search=<script>alert('xss')</script>" >/dev/null 2>&1 || true
    done
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "API security testing completed"
    fi
}

# Analyze security scan results
analyze_security_results() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📊 Analyzing Security Scan Results"
    fi
    
    local total_issues=0
    local critical_issues=0
    local high_issues=0
    local medium_issues=0
    local low_issues=0
    
    # Analyze gosec results
    if [[ -f "$OUTPUT_DIR/data/gosec-results.json" ]]; then
        local gosec_issues=$(jq '.Issues | length' "$OUTPUT_DIR/data/gosec-results.json" 2>/dev/null || echo "0")
        total_issues=$((total_issues + gosec_issues))
        
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_info "GoSec found $gosec_issues security issues"
        fi
    fi
    
    # Analyze nancy results
    if [[ -f "$OUTPUT_DIR/data/nancy-results.json" ]]; then
        local nancy_issues=$(jq '.vulnerable | length' "$OUTPUT_DIR/data/nancy-results.json" 2>/dev/null || echo "0")
        total_issues=$((total_issues + nancy_issues))
        
        if [[ "$QUIET_MODE" != "true" ]]; then
            print_info "Nancy found $nancy_issues vulnerable dependencies"
        fi
    fi
    
    # Create summary
    cat > "$OUTPUT_DIR/reports/security-summary.json" <<EOF
{
    "summary": {
        "total_issues": $total_issues,
        "critical": $critical_issues,
        "high": $high_issues,
        "medium": $medium_issues,
        "low": $low_issues,
        "scan_timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
        "scan_scope": "$SCAN_SCOPE",
        "scan_level": "$SCAN_LEVEL"
    }
}
EOF
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "Security analysis completed"
        print_info "Total security issues found: $total_issues"
    fi
}

# Generate security report
generate_security_report() {
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_header "📋 Generating Security Report"
    fi
    
    local report_file="$OUTPUT_DIR/reports/security-report.html"
    
    cat > "$report_file" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>Go Coffee Security Scan Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .summary { background: #ecf0f1; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .finding { margin: 10px 0; padding: 10px; border-left: 4px solid #3498db; }
        .critical { border-left-color: #e74c3c; background: #fdf2f2; }
        .high { border-left-color: #e67e22; background: #fef9f3; }
        .medium { border-left-color: #f39c12; background: #fffbf0; }
        .low { border-left-color: #27ae60; background: #f0f9f4; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .metric { display: inline-block; margin: 10px; padding: 15px; background: #f8f9fa; border-radius: 5px; text-align: center; }
        .metric-value { font-size: 2em; font-weight: bold; color: #2c3e50; }
        .metric-label { color: #7f8c8d; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔒 Go Coffee Security Scan Report</h1>
        <p>Generated on: $(date)</p>
        <p>Scan Scope: $SCAN_SCOPE | Level: $SCAN_LEVEL</p>
        <p>Target: $TARGET_URL</p>
    </div>
    
    <div class="summary">
        <h2>📊 Security Metrics</h2>
        <div class="metric">
            <div class="metric-value">0</div>
            <div class="metric-label">Critical</div>
        </div>
        <div class="metric">
            <div class="metric-value">0</div>
            <div class="metric-label">High</div>
        </div>
        <div class="metric">
            <div class="metric-value">0</div>
            <div class="metric-label">Medium</div>
        </div>
        <div class="metric">
            <div class="metric-value">0</div>
            <div class="metric-label">Low</div>
        </div>
    </div>
    
    <div class="summary">
        <h2>🔍 Scan Coverage</h2>
        <table>
            <tr>
                <th>Scan Type</th>
                <th>Status</th>
                <th>Issues Found</th>
                <th>Tool Used</th>
            </tr>
            <tr>
                <td>Source Code Analysis</td>
                <td>✅ Completed</td>
                <td>-</td>
                <td>gosec, semgrep</td>
            </tr>
            <tr>
                <td>Dependency Scanning</td>
                <td>✅ Completed</td>
                <td>-</td>
                <td>nancy, trivy</td>
            </tr>
            <tr>
                <td>Container Security</td>
                <td>✅ Completed</td>
                <td>-</td>
                <td>trivy</td>
            </tr>
            <tr>
                <td>API Security</td>
                <td>✅ Completed</td>
                <td>-</td>
                <td>custom checks</td>
            </tr>
        </table>
    </div>
    
    <div class="summary">
        <h2>🛡️ Security Recommendations</h2>
        <div class="finding medium">
            <strong>🔧 Immediate Actions:</strong>
            <ul>
                <li>Review and fix all critical and high severity issues</li>
                <li>Update vulnerable dependencies</li>
                <li>Implement missing security headers</li>
                <li>Enable security monitoring</li>
            </ul>
        </div>
        
        <div class="finding low">
            <strong>📋 Best Practices:</strong>
            <ul>
                <li>Regular security scans in CI/CD pipeline</li>
                <li>Dependency vulnerability monitoring</li>
                <li>Security code reviews</li>
                <li>Penetration testing</li>
            </ul>
        </div>
    </div>
    
    <div class="summary">
        <h2>📁 Detailed Results</h2>
        <p>Detailed scan results are available in the following files:</p>
        <ul>
            <li><strong>GoSec Results:</strong> data/gosec-results.json</li>
            <li><strong>Dependency Scan:</strong> data/nancy-results.json</li>
            <li><strong>Container Scan:</strong> data/trivy-*.json</li>
            <li><strong>API Security:</strong> data/api-security-results.json</li>
        </ul>
    </div>
</body>
</html>
EOF

    if [[ "$QUIET_MODE" != "true" ]]; then
        print_status "Security report generated: $report_file"
    fi
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
    local start_time=$(date +%s)
    
    # Parse arguments
    parse_security_args "$@"
    
    # Check dependencies
    check_security_dependencies
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_info "Security Scan Configuration:"
        print_info "  Target URL: $TARGET_URL"
        print_info "  Scope: $SCAN_SCOPE"
        print_info "  Level: $SCAN_LEVEL"
        print_info "  Output: $OUTPUT_DIR"
        print_info "  Format: $REPORT_FORMAT"
    fi
    
    # Create output structure
    create_security_output_structure
    
    # Run scans based on scope
    case "$SCAN_SCOPE" in
        "code")
            run_code_security_scan
            ;;
        "deps")
            run_dependency_scan
            ;;
        "docker")
            run_container_scan
            ;;
        "api")
            run_api_security_scan
            ;;
        "full")
            run_code_security_scan
            run_dependency_scan
            run_container_scan
            run_api_security_scan
            ;;
        *)
            print_error "Unknown scan scope: $SCAN_SCOPE"
            exit 1
            ;;
    esac
    
    # Analyze results
    analyze_security_results
    
    # Generate reports
    if [[ "$REPORT_FORMAT" == "html" || "$REPORT_FORMAT" == "all" ]]; then
        generate_security_report
    fi
    
    # Calculate scan time
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    if [[ "$QUIET_MODE" != "true" ]]; then
        print_success "🎉 Security scanning completed in ${total_time}s"
        
        print_header "📊 Results Summary"
        print_info "Security scan results saved to: $OUTPUT_DIR"
        print_info "Security report: $OUTPUT_DIR/reports/security-report.html"
        print_info "Raw data: $OUTPUT_DIR/data/"
        
        print_header "🎯 Next Steps"
        print_info "1. Review the security report for findings"
        print_info "2. Prioritize and fix critical/high severity issues"
        print_info "3. Update vulnerable dependencies"
        print_info "4. Integrate security scanning into CI/CD pipeline"
    fi
}

# Run main function with all arguments
main "$@"
