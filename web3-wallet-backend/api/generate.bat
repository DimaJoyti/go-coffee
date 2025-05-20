@echo off

REM Create output directories
mkdir proto\wallet 2>nul
mkdir proto\transaction 2>nul
mkdir proto\contract 2>nul
mkdir proto\security 2>nul

REM Generate Go code from proto files
protoc -I=proto ^
  --go_out=. ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=. ^
  --go-grpc_opt=paths=source_relative ^
  proto/wallet.proto

protoc -I=proto ^
  --go_out=. ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=. ^
  --go-grpc_opt=paths=source_relative ^
  proto/transaction.proto

protoc -I=proto ^
  --go_out=. ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=. ^
  --go-grpc_opt=paths=source_relative ^
  proto/contract.proto

protoc -I=proto ^
  --go_out=. ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=. ^
  --go-grpc_opt=paths=source_relative ^
  proto/security.proto

echo Code generation completed successfully!
