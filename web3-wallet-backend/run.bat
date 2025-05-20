@echo off

REM Build the services
echo Building the services...

echo Building the API Gateway...
cd cmd\api-gateway
go build -o api-gateway.exe
cd ..\..

echo Building the Wallet Service...
cd cmd\wallet-service
go build -o wallet-service.exe
cd ..\..

echo Building the Transaction Service...
cd cmd\transaction-service
go build -o transaction-service.exe
cd ..\..

echo Building the Smart Contract Service...
cd cmd\smart-contract-service
go build -o smart-contract-service.exe
cd ..\..

echo Building the Security Service...
cd cmd\security-service
go build -o security-service.exe
cd ..\..

REM Start the services
echo Starting the services...

REM Start the API Gateway
echo Starting the API Gateway...
start "API Gateway" cmd /c "cd cmd\api-gateway && api-gateway.exe"

REM Start the Wallet Service
echo Starting the Wallet Service...
start "Wallet Service" cmd /c "cd cmd\wallet-service && wallet-service.exe"

REM Start the Transaction Service
echo Starting the Transaction Service...
start "Transaction Service" cmd /c "cd cmd\transaction-service && transaction-service.exe"

REM Start the Smart Contract Service
echo Starting the Smart Contract Service...
start "Smart Contract Service" cmd /c "cd cmd\smart-contract-service && smart-contract-service.exe"

REM Start the Security Service
echo Starting the Security Service...
start "Security Service" cmd /c "cd cmd\security-service && security-service.exe"

echo All services started successfully!
