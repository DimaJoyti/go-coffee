@echo off
setlocal enabledelayedexpansion

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed. Please install Go and try again.
    exit /b 1
)

REM Build the services
echo Building the services...

echo Building the producer...
cd producer
go mod tidy
go build -o producer.exe
cd ..

echo Building the consumer...
cd consumer
go mod tidy
go build -o consumer.exe
cd ..

echo Building the streams processor...
cd streams
go mod tidy
go build -o streams.exe
cd ..

echo Building the API Gateway...
cd api-gateway
go mod tidy
go build -o api-gateway.exe
cd ..

REM Start the services
echo Starting the services...

REM Start the streams processor
echo Starting the streams processor...
start "Streams Processor" cmd /c "cd streams && streams.exe"

REM Start the consumer
echo Starting the consumer...
start "Consumer" cmd /c "cd consumer && consumer.exe"

REM Start the producer with gRPC support
echo Starting the producer...
start "Producer" cmd /c "cd producer && producer.exe main_grpc.go"

REM Start the API Gateway
echo Starting the API Gateway...
start "API Gateway" cmd /c "cd api-gateway && api-gateway.exe"

echo All services started.
echo Press any key to stop all services...
pause >nul

REM Stop all services
echo Stopping services...
taskkill /F /FI "WINDOWTITLE eq Streams Processor*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Consumer*" >nul 2>&1
taskkill /F /FI "WINDOWTITLE eq Producer*" >nul 2>&1
echo All services stopped.

endlocal
