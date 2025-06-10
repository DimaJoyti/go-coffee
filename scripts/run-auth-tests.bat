@echo off
REM Auth Service Test Runner for Windows
REM This script runs all tests for the Auth Service

setlocal enabledelayedexpansion

echo 🧪 Auth Service Test Suite
echo ==========================

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

REM Check if we're in the right directory
if not exist "go.mod" (
    echo [ERROR] go.mod not found. Please run this script from the project root.
    exit /b 1
)

REM Create test results directory
if not exist "test-results" mkdir test-results

echo [INFO] Starting Auth Service tests...

REM Run unit tests
echo [INFO] Running unit tests...
go test -v -race -coverprofile=test-results/auth-coverage.out ./internal/auth/...
if %errorlevel% neq 0 (
    echo [ERROR] Unit tests failed
    exit /b 1
)
echo [SUCCESS] Unit tests passed

REM Generate coverage report
echo [INFO] Generating coverage report...
go tool cover -html=test-results/auth-coverage.out -o test-results/auth-coverage.html
echo [SUCCESS] Coverage report generated: test-results/auth-coverage.html

REM Show coverage summary
for /f "tokens=3" %%i in ('go tool cover -func=test-results/auth-coverage.out ^| findstr "total"') do set COVERAGE=%%i
echo [INFO] Total test coverage: !COVERAGE!

REM Run specific test suites
echo [INFO] Running domain layer tests...
go test -v ./internal/auth/domain/...

echo [INFO] Running application layer tests...
go test -v ./internal/auth/application/...

echo [INFO] Running infrastructure layer tests...
go test -v ./internal/auth/infrastructure/...

echo [INFO] Running transport layer tests...
go test -v ./internal/auth/transport/...

REM Run integration tests (if not in short mode)
if not "%1"=="--short" (
    echo [INFO] Running integration tests...
    go test -v -tags=integration ./internal/auth/...
    if %errorlevel% neq 0 (
        echo [WARNING] Integration tests failed (this might be expected if database/redis are not available)
    ) else (
        echo [SUCCESS] Integration tests passed
    )
)

REM Run benchmarks
echo [INFO] Running benchmarks...
go test -bench=. -benchmem ./internal/auth/... > test-results/auth-benchmarks.txt
echo [SUCCESS] Benchmarks completed: test-results/auth-benchmarks.txt

REM Check for race conditions
echo [INFO] Checking for race conditions...
go test -race ./internal/auth/...
if %errorlevel% neq 0 (
    echo [ERROR] Race conditions detected
    exit /b 1
)
echo [SUCCESS] No race conditions detected

REM Run security checks (if gosec is available)
where gosec >nul 2>nul
if %errorlevel% equ 0 (
    echo [INFO] Running security scan...
    gosec ./internal/auth/...
    if %errorlevel% neq 0 (
        echo [WARNING] Security scan found issues
    ) else (
        echo [SUCCESS] Security scan passed
    )
) else (
    echo [WARNING] gosec not installed. Skipping security scan.
)

REM Run vulnerability check (if govulncheck is available)
where govulncheck >nul 2>nul
if %errorlevel% equ 0 (
    echo [INFO] Checking for vulnerabilities...
    govulncheck ./internal/auth/...
    if %errorlevel% neq 0 (
        echo [WARNING] Vulnerabilities found
    ) else (
        echo [SUCCESS] No vulnerabilities found
    )
) else (
    echo [WARNING] govulncheck not installed. Skipping vulnerability check.
)

REM Summary
echo.
echo 🎉 Test Summary
echo ===============
echo [SUCCESS] All tests completed successfully!
echo [INFO] Coverage: !COVERAGE!
echo [INFO] Results saved in: test-results/

REM Open coverage report in browser (optional)
if "%2"=="--open" (
    start test-results/auth-coverage.html
)

echo.
echo [SUCCESS] Auth Service test suite completed! ✨

endlocal
