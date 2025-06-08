# Bitcoin Cryptography Implementation

A comprehensive Bitcoin cryptography implementation in Go, featuring complete support for Bitcoin protocol fundamentals including elliptic curve cryptography, transaction handling, and wallet operations.

## 🚀 Features

### ✅ Complete Bitcoin Implementation

- **Elliptic Curve Cryptography (secp256k1)**
  - Point operations on elliptic curves
  - Scalar multiplication with binary method
  - Cryptographically secure key generation
  - Point validation with modular arithmetic

- **ECDSA Digital Signatures**
  - Signature creation and verification
  - DER encoding support
  - Public key recovery
  - Low-s canonical form (BIP 62)

- **SEC Format Encoding**
  - Compressed public keys (33 bytes)
  - Uncompressed public keys (65 bytes)
  - Private key encoding (32 bytes)
  - Automatic format detection

- **Base58Check Encoding**
  - Bitcoin Base58 alphabet
  - Checksum validation
  - WIF (Wallet Import Format)
  - Mainnet/Testnet support

- **Bitcoin Addresses**
  - P2PKH (Pay-to-Public-Key-Hash)
  - P2SH (Pay-to-Script-Hash)
  - P2PK (Pay-to-Public-Key)
  - Multisig addresses

- **Bitcoin Script**
  - Stack-based scripting language
  - Core opcodes (OP_DUP, OP_HASH160, OP_CHECKSIG)
  - Script parsing and serialization
  - Hash functions (Hash160, Hash256)

- **Transaction Processing**
  - Transaction creation and validation
  - SIGHASH types support
  - Fee calculation
  - UTXO model
  - Transaction Builder API

- **Wallet Operations**
  - Key generation and management
  - Message signing and verification
  - Address creation
  - Transaction creation

## 📁 Project Structure

```
pkg/bitcoin/
├── bitcoin.go              # Main API
├── bitcoin_test.go         # Comprehensive tests
├── ecc/                    # Elliptic curve cryptography
│   ├── point.go           # Point operations
│   ├── secp256k1.go       # Curve parameters
│   └── signature.go       # ECDSA signatures
├── sec/                    # SEC format encoding
│   └── encoding.go        # Key encoding
├── base58/                 # Base58Check encoding
│   └── base58.go          # Base58 and WIF
├── address/                # Bitcoin addresses
│   └── address.go         # P2PKH, P2SH, multisig
├── script/                 # Bitcoin Script
│   └── script.go          # Scripting language
└── transaction/            # Transaction handling
    ├── transaction.go     # Core transactions
    └── builder.go         # Transaction Builder
```

## 🔧 Installation

```bash
go mod init your-project
go get github.com/DimaJoyti/go-coffee/crypto-wallet
```

## 🎯 Quick Start

### Create a Wallet

```go
package main

import (
    "fmt"
    "github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin"
)

func main() {
    // Create a new wallet
    wallet, err := bitcoin.NewWallet(false) // mainnet
    if err != nil {
        panic(err)
    }

    fmt.Printf("Address: %s\n", wallet.GetAddress())
    fmt.Printf("Private Key (WIF): %s\n", wallet.GetPrivateKeyWIF(true))
}
```

### Sign a Message

```go
// Sign a message
message := []byte("Hello, Bitcoin!")
signature, err := wallet.SignMessage(message)
if err != nil {
    panic(err)
}

// Verify signature
valid := wallet.VerifyMessage(message, signature)
fmt.Printf("Signature valid: %t\n", valid)
```

### Create a Transaction

```go
// Create UTXO
utxo := &transaction.UTXO{
    TxHash:       prevTxHash,
    OutputIndex:  0,
    Amount:       100000000, // 1 BTC in satoshis
    ScriptPubKey: fromAddress.ScriptPubKey(),
}

// Create transaction
tx, err := wallet.CreateTransaction(
    []*transaction.UTXO{utxo},
    "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", // destination
    50000000, // 0.5 BTC
    10,       // 10 sat/byte fee
)
```

### Create Multisig Address

```go
utils := bitcoin.NewBitcoinUtils()

// Generate 3 public keys
var publicKeys []*ecc.Point
for i := 0; i < 3; i++ {
    _, pubKey, _ := utils.GenerateKeyPair()
    publicKeys = append(publicKeys, pubKey)
}

// Create 2-of-3 multisig address
multisigAddr, err := utils.CreateMultisigAddress(publicKeys, 2, false)
```

## 🧪 Testing

Run all tests:

```bash
cd crypto-wallet/pkg/bitcoin
go test -v
```

Run examples:

```bash
cd crypto-wallet/examples
go run bitcoin_example.go
```

## 📚 Documentation

- [Complete Implementation Guide](docs/BITCOIN_IMPLEMENTATION.md)
- [Implementation Summary](BITCOIN_IMPLEMENTATION_SUMMARY.md)
- [API Documentation](docs/BITCOIN_IMPLEMENTATION.md#api-documentation)

## 🔒 Security

### Cryptographic Security
- Uses `crypto/rand` for secure random number generation
- Proper secp256k1 implementation with modular arithmetic
- Input validation for all parameters
- Protection against timing attacks

### Validation
- Point verification on curve membership
- Private key validation (0 < key < n)
- Checksum validation for Base58Check
- Complete transaction and signature validation

## 🎯 Supported Standards

- **Bitcoin Core**: Full compatibility
- **BIP Standards**: Support for relevant BIPs
- **Networks**: Mainnet and Testnet
- **Formats**: WIF, Base58Check, DER, SEC1

## 📈 Performance

- Optimized big number operations
- Efficient scalar multiplication
- Minimal memory allocations
- Parallel test execution

## 🛠️ Extensions

The package is designed for extensibility:

- SegWit transaction support
- Lightning Network integration
- Additional address types (Bech32, P2WPKH, P2WSH)
- Additional SIGHASH types

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Create a Pull Request

## 📝 License

This project is part of the go-coffee ecosystem and is available under the corresponding license.

## 🔗 Related Projects

- [Go Coffee Main Project](../README.md)
- [Web3 Wallet Backend](../web3-wallet-backend/)
- [Order Management Service](../order-service/)

## 📞 Support

For questions and support:
- Create an issue in the repository
- Check the documentation
- Review the examples

---

**Package Version**: 1.0.0  
**Go Version**: 1.22+  
**Completion Date**: January 2025

Built with ❤️ for the Bitcoin community
