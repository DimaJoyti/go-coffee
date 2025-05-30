#!/bin/bash

# Fix logger calls in solana_simple.go
sed -i 's/c\.logger\.Debug("Getting Solana transaction status", "tx_hash", txHash)/c.logger.Debug("Getting Solana transaction status", zap.String("tx_hash", txHash))/g' web3-wallet-backend/pkg/blockchain/solana_simple.go

sed -i 's/c\.logger\.Debug("Getting Solana account info", "address", address)/c.logger.Debug("Getting Solana account info", zap.String("address", address))/g' web3-wallet-backend/pkg/blockchain/solana_simple.go

sed -i 's/c\.logger\.Debug("Getting Solana token balance",$/c.logger.Debug("Getting Solana token balance",/g' web3-wallet-backend/pkg/blockchain/solana_simple.go
sed -i 's/"address", address,$/zap.String("address", address),/g' web3-wallet-backend/pkg/blockchain/solana_simple.go
sed -i 's/"token_mint", tokenMint,$/zap.String("token_mint", tokenMint),/g' web3-wallet-backend/pkg/blockchain/solana_simple.go

sed -i 's/c\.logger\.Debug("Getting Solana token accounts", "owner", owner)/c.logger.Debug("Getting Solana token accounts", zap.String("owner", owner))/g' web3-wallet-backend/pkg/blockchain/solana_simple.go

sed -i 's/c\.logger\.Debug("Getting Solana program accounts", "program_id", programID)/c.logger.Debug("Getting Solana program accounts", zap.String("program_id", programID))/g' web3-wallet-backend/pkg/blockchain/solana_simple.go

sed -i 's/c\.logger\.Debug("Getting confirmed Solana transaction", "signature", signature)/c.logger.Debug("Getting confirmed Solana transaction", zap.String("signature", signature))/g' web3-wallet-backend/pkg/blockchain/solana_simple.go

echo "Logger calls fixed!"
