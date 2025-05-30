# Solana Integration Guide

## Overview

This document describes the Solana blockchain integration in the Web3 Wallet Backend system. The integration provides comprehensive support for Solana wallets, DeFi protocols, and token operations.

## Features

### 🔑 Wallet Management
- **Solana Wallet Creation**: Generate new Solana wallets with ed25519 key pairs
- **Private Key Import**: Import existing Solana private keys
- **Mnemonic Support**: Generate and restore wallets from BIP39 mnemonics
- **Balance Queries**: Check SOL and SPL token balances

### 🏦 DeFi Integration
- **Raydium DEX**: Automated market maker (AMM) integration
- **Jupiter Aggregator**: Best price routing across multiple DEXs
- **Liquidity Provision**: Add/remove liquidity from pools
- **Token Swaps**: Execute token swaps with optimal routing

### 🔐 Security Features
- **Key Encryption**: Secure storage of private keys
- **Derivation Paths**: Standard Solana derivation paths (m/44'/501'/0'/0')
- **Transaction Signing**: Secure transaction signing with user wallets

## Architecture

### Core Components

1. **SolanaClient** (`pkg/blockchain/solana.go`)
   - RPC client for Solana network communication
   - Balance queries and transaction operations
   - Account creation and management

2. **KeyManager Extensions** (`pkg/crypto/keys.go`)
   - Solana key pair generation
   - ed25519 cryptographic operations
   - Mnemonic to private key conversion

3. **DeFi Clients**
   - **RaydiumClient** (`internal/defi/raydium.go`)
   - **JupiterClient** (`internal/defi/jupiter.go`)

4. **Wallet Service** (`internal/wallet/service.go`)
   - Multi-chain wallet management
   - Solana-specific operations

## Configuration

### Environment Variables

```yaml
# Solana Configuration
solana:
  network: "mainnet-beta"
  rpc_url: "https://api.mainnet-beta.solana.com"
  ws_url: "wss://api.mainnet-beta.solana.com"
  cluster: "mainnet-beta"
  commitment: "confirmed"
  timeout: "30s"
  max_retries: 3
  confirmation_blocks: 32
```

### Production Environment

```yaml
solana:
  rpc_url: "${SOLANA_RPC_URL}"
  ws_url: "${SOLANA_WS_URL}"
  cluster: "${SOLANA_CLUSTER}"
  commitment: "confirmed"
  timeout: "30s"
  max_retries: 3
  confirmation_blocks: 32
```

## API Usage

### Creating a Solana Wallet

```go
// Create wallet request
req := &models.CreateWalletRequest{
    UserID: "user123",
    Name:   "My Solana Wallet",
    Chain:  models.ChainSolana,
    Type:   models.WalletTypeHD,
}

// Create wallet
resp, err := walletService.CreateWallet(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Wallet Address: %s\n", resp.Wallet.Address)
fmt.Printf("Private Key: %s\n", resp.PrivateKey)
fmt.Printf("Mnemonic: %s\n", resp.Mnemonic)
```

### Checking SOL Balance

```go
// Get balance request
req := &models.GetBalanceRequest{
    WalletID: "wallet123",
}

// Get balance
resp, err := walletService.GetBalance(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("SOL Balance: %s\n", resp.Balance)
```

### Checking SPL Token Balance

```go
// Get token balance request
req := &models.GetBalanceRequest{
    WalletID:     "wallet123",
    TokenAddress: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
}

// Get balance
resp, err := walletService.GetBalance(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("USDC Balance: %s\n", resp.Balance)
```

## DeFi Operations

### Raydium Integration

```go
// Create Raydium client
raydiumClient, err := defi.NewRaydiumClient(rpcURL, logger)
if err != nil {
    log.Fatal(err)
}

// Get available pools
pools, err := raydiumClient.GetPools(ctx)
if err != nil {
    log.Fatal(err)
}

// Get swap quote
quote, err := raydiumClient.GetSwapQuote(ctx,
    "So11111111111111111111111111111111111111112", // SOL
    "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
    decimal.NewFromFloat(1.0))
if err != nil {
    log.Fatal(err)
}

// Execute swap
signature, err := raydiumClient.ExecuteSwap(ctx, quote, userWallet)
if err != nil {
    log.Fatal(err)
}
```

### Jupiter Integration

```go
// Create Jupiter client
jupiterClient := defi.NewJupiterClient(logger)

// Execute swap with best routing
signature, err := jupiterClient.ExecuteSwap(ctx,
    "So11111111111111111111111111111111111111112", // SOL
    "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
    decimal.NewFromFloat(1.0),
    100, // 1% slippage
    userWallet)
if err != nil {
    log.Fatal(err)
}
```

## Token Addresses

### Common Solana Tokens

| Token | Symbol | Address |
|-------|--------|---------|
| Solana | SOL | `So11111111111111111111111111111111111111112` |
| USD Coin | USDC | `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v` |
| Tether | USDT | `Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB` |
| Marinade SOL | mSOL | `mSoLzYCxHdYgdzU16g5QSh3i5K3z3KZK7ytfqcJm7So` |
| Lido Staked SOL | stSOL | `7dHbWXmci3dT8UFYWYZweBLXgycu7Y3iL6trKn1Y7ARj` |

## Error Handling

### Common Errors

1. **Invalid Address Format**
   ```
   Error: invalid address: invalid base58 string
   Solution: Ensure Solana addresses are valid base58 strings
   ```

2. **Insufficient Balance**
   ```
   Error: insufficient funds for transaction
   Solution: Check wallet balance before executing transactions
   ```

3. **Network Connection**
   ```
   Error: failed to connect to Solana RPC
   Solution: Verify RPC URL and network connectivity
   ```

## Testing

### Unit Tests

```bash
# Run Solana-specific tests
go test ./pkg/blockchain -run TestSolana
go test ./pkg/crypto -run TestSolana
go test ./internal/defi -run TestRaydium
go test ./internal/defi -run TestJupiter

# Or use Makefile commands
make solana-test          # Run all Solana tests
make unit-test           # Run all unit tests
make test               # Run all tests
```

### Integration Tests

```bash
# Run integration tests with testnet
SOLANA_CLUSTER=devnet go test ./tests/integration -run TestSolanaIntegration

# Or use Makefile
make integration-test    # Run integration tests
```

### Test Coverage

```bash
# Generate coverage reports
make coverage           # Unit test coverage
make coverage-integration  # Integration test coverage
```

## Monitoring

### Key Metrics

- **Transaction Success Rate**: Monitor successful vs failed transactions
- **Balance Query Latency**: Track RPC response times
- **DeFi Operation Success**: Monitor swap and liquidity operations
- **Error Rates**: Track different types of errors

### Logging

```go
// Enable debug logging for Solana operations
logger := logger.New("solana").WithLevel(logger.DebugLevel)
```

## Security Considerations

1. **Private Key Storage**: Always encrypt private keys before storage
2. **RPC Endpoints**: Use trusted RPC providers
3. **Transaction Validation**: Validate all transaction parameters
4. **Rate Limiting**: Implement rate limiting for RPC calls
5. **Error Handling**: Never expose sensitive information in error messages

## Deployment

### Docker Configuration

```dockerfile
# Add Solana dependencies
RUN go mod download github.com/gagliardetto/solana-go
```

### Environment Setup

```bash
# Set Solana environment variables
export SOLANA_RPC_URL="https://api.mainnet-beta.solana.com"
export SOLANA_WS_URL="wss://api.mainnet-beta.solana.com"
export SOLANA_CLUSTER="mainnet-beta"
```

## Roadmap

### Planned Features

- [ ] Solana NFT support
- [ ] Staking operations
- [ ] Cross-chain bridges
- [ ] Advanced DeFi strategies
- [ ] Governance participation
- [ ] Mobile wallet integration

## Support

For issues and questions related to Solana integration:

1. Check the error logs for detailed information
2. Verify network connectivity and RPC endpoints
3. Ensure proper configuration of environment variables
4. Review the Solana documentation for protocol-specific issues

## References

- [Solana Documentation](https://docs.solana.com/)
- [Solana Go SDK](https://github.com/gagliardetto/solana-go)
- [Raydium Documentation](https://docs.raydium.io/)
- [Jupiter Documentation](https://docs.jup.ag/)
