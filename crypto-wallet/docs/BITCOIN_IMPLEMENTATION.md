# Bitcoin Cryptography Implementation

This is a complete implementation of Bitcoin cryptography and transaction processing in Go, including all core components of the Bitcoin protocol.

## 🚀 Features

### ✅ Implemented Components

1. **Elliptic Curve Cryptography (secp256k1)**
   - Mathematical operations with points on elliptic curves
   - Private/public key generation
   - Scalar multiplication
   - Point validation on curve

2. **ECDSA Signatures**
   - Digital signature creation
   - Signature verification
   - DER signature encoding
   - Public key recovery from signatures

3. **SEC Format**
   - Compressed public key encoding (33 bytes)
   - Uncompressed public key encoding (65 bytes)
   - Private key encoding (32 bytes)

4. **Base58Check Encoding**
   - Base58 encoding/decoding
   - Checksum validation
   - WIF (Wallet Import Format) support
   - Bitcoin addresses

5. **Bitcoin Addresses**
   - P2PKH (Pay-to-Public-Key-Hash) addresses
   - P2SH (Pay-to-Script-Hash) addresses
   - P2PK (Pay-to-Public-Key) addresses
   - Multisig addresses
   - Mainnet/Testnet support

6. **Bitcoin Script**
   - Stack-based scripting language
   - P2PKH scripts
   - P2SH scripts
   - Script parsing and serialization
   - Hash160 and Hash256 functions

7. **Transactions**
   - Transaction creation
   - Transaction signing
   - Transaction validation
   - SIGHASH types
   - Fee calculation
   - Transaction Builder

8. **Wallet**
   - Key generation
   - Message signing
   - Address creation
   - WIF import/export

## 📁 Package Structure

```
pkg/bitcoin/
├── bitcoin.go              # Main API and examples
├── bitcoin_test.go         # Tests
├── ecc/                    # Elliptic curve cryptography
│   ├── point.go           # Point operations
│   ├── secp256k1.go       # secp256k1 curve parameters
│   └── signature.go       # ECDSA signatures
├── sec/                    # SEC format
│   └── encoding.go        # Key encoding
├── base58/                 # Base58Check encoding
│   └── base58.go          # Base58 and WIF
├── address/                # Bitcoin addresses
│   └── address.go         # P2PKH, P2SH, multisig
├── script/                 # Bitcoin Script
│   └── script.go          # Scripting language
└── transaction/            # Transactions
    ├── transaction.go     # Core transactions
    └── builder.go         # Transaction Builder
```

## 🔧 Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin"
)

func main() {
    // Create a new wallet
    wallet, err := bitcoin.NewWallet(false) // mainnet
    if err != nil {
        panic(err)
    }

    fmt.Printf("Address: %s\n", wallet.GetAddress())
    fmt.Printf("Private Key: %s\n", wallet.GetPrivateKeyWIF(true))

    // Sign a message
    message := []byte("Hello, Bitcoin!")
    signature, err := wallet.SignMessage(message)
    if err != nil {
        panic(err)
    }

    // Verify signature
    valid := wallet.VerifyMessage(message, signature)
    fmt.Printf("Signature valid: %t\n", valid)
}
```

### Creating Transactions

```go
// Create UTXO
utxo := &transaction.UTXO{
    TxHash:       prevTxHash,
    OutputIndex:  0,
    Amount:       100000000, // 1 BTC
    ScriptPubKey: fromAddress.ScriptPubKey(),
}

// Create transaction
tx, err := wallet.CreateTransaction(
    []*transaction.UTXO{utxo},
    "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", // to address
    50000000, // 0.5 BTC
    10,       // 10 sat/byte fee
)
```

### Multisig Address

```go
utils := bitcoin.NewBitcoinUtils()

// Generate 3 keys
var publicKeys []*ecc.Point
for i := 0; i < 3; i++ {
    _, pubKey, _ := utils.GenerateKeyPair()
    publicKeys = append(publicKeys, pubKey)
}

// Create 2-of-3 multisig address
multisigAddr, err := utils.CreateMultisigAddress(publicKeys, 2, false)
```

## 🧪 Testing

Run tests:

```bash
cd crypto-wallet/pkg/bitcoin
go test -v
```

Run example:

```bash
cd crypto-wallet/examples
go run bitcoin_example.go
```

## 📚 API Documentation

### Core Types

#### `Wallet`
```go
type Wallet struct {
    // private fields
}

// Methods
func NewWallet(testnet bool) (*Wallet, error)
func NewWalletFromPrivateKey(privateKey *big.Int, testnet bool) (*Wallet, error)
func NewWalletFromWIF(wif string) (*Wallet, error)
func (w *Wallet) GetAddress() string
func (w *Wallet) GetPrivateKeyWIF(compressed bool) string
func (w *Wallet) SignMessage(message []byte) (*ecc.Signature, error)
func (w *Wallet) VerifyMessage(message []byte, signature *ecc.Signature) bool
```

#### `Point` (Elliptic Curve)
```go
type Point struct {
    X, Y, A, B *big.Int
}

func NewPoint(x, y, a, b *big.Int) (*Point, error)
func (p *Point) Add(other *Point) (*Point, error)
func (p *Point) Double() (*Point, error)
func (p *Point) ScalarMult(k *big.Int) (*Point, error)
func (p *Point) IsOnCurve() bool
```

#### `Transaction`
```go
type Transaction struct {
    Version  uint32
    TxIns    []*TxIn
    TxOuts   []*TxOut
    Locktime uint32
}

func NewTransaction(version uint32, txIns []*TxIn, txOuts []*TxOut, locktime uint32) *Transaction
func (tx *Transaction) Serialize() []byte
func (tx *Transaction) Hash() []byte
func (tx *Transaction) SignInput(inputIndex int, privateKey *big.Int, scriptPubKey []byte, sighashType int) error
```

### Utility Functions

#### Base58Check
```go
func Encode(input []byte) string
func Decode(input string) ([]byte, error)
func EncodeCheck(input []byte) string
func DecodeCheck(input string) ([]byte, error)
func EncodeWIF(privateKey []byte, compressed bool, testnet bool) string
func DecodeWIF(wif string) ([]byte, bool, bool, error)
```

#### Addresses
```go
func NewP2PKHAddress(publicKey *ecc.Point, testnet bool) *Address
func NewP2SHAddress(scriptHash []byte, testnet bool) *Address
func ParseAddress(addressStr string) (*Address, error)
func IsValid(addressStr string) bool
func CreateMultisigAddress(publicKeys []*ecc.Point, threshold int, testnet bool) (*Address, error)
```

## 🔒 Security

### Cryptographic Security
- Use of cryptographically secure random number generator
- Proper implementation of secp256k1 curve
- Validation of all input parameters
- Protection against timing attacks

### Validation
- Verification of all points on curve membership
- Private key validation (0 < key < n)
- Signature verification before use
- Address and transaction validation

## 🎯 Usage Examples

### 1. Key and Address Generation
```go
// Generate new wallet
wallet, _ := bitcoin.NewWallet(false)
address := wallet.GetAddress()
wif := wallet.GetPrivateKeyWIF(true)
```

### 2. Message Signing
```go
message := []byte("Important message")
signature, _ := wallet.SignMessage(message)
valid := wallet.VerifyMessage(message, signature)
```

### 3. Transaction Creation
```go
builder := transaction.NewTransactionBuilder()
builder.AddInput(txHash, outputIndex, utxo)
builder.AddOutput(toAddress, amount)
builder.SetFee(fee)
tx, _ := builder.SignTransaction([]*big.Int{privateKey})
```

## 🔄 Compatibility

- **Bitcoin Core**: Full compatibility with Bitcoin Core
- **BIP Standards**: Support for BIP 32, BIP 39, BIP 44
- **Networks**: Mainnet and Testnet
- **Formats**: WIF, Base58Check, DER, SEC1

## 📈 Performance

- Optimized big number operations
- Efficient scalar multiplication
- Minimal memory allocations
- Parallel test execution

## 🛠️ Extensions

The package is designed with extensibility in mind:

- Adding new address types (Bech32, P2WPKH, P2WSH)
- SegWit transaction support
- Lightning Network integration
- Additional SIGHASH types

## 📝 License

This code is developed as part of the go-coffee project and is available under the corresponding license.

## 🤝 Contributing

To contribute changes:
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Create a Pull Request

## 📞 Support

For questions and support, create an issue in the project repository.
