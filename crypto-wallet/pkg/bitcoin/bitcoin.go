// Package bitcoin provides a comprehensive implementation of Bitcoin cryptography and transaction handling.
// This package includes elliptic curve cryptography, SEC format encoding, Base58Check encoding,
// transaction creation and validation, Bitcoin Script, and address generation.
package bitcoin

import (
	"fmt"
	"math/big"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/address"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/base58"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/script"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/sec"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/transaction"
)

// Wallet represents a Bitcoin wallet with key management capabilities
type Wallet struct {
	privateKey *big.Int
	publicKey  *ecc.Point
	testnet    bool
}

// NewWallet creates a new Bitcoin wallet with a randomly generated private key
func NewWallet(testnet bool) (*Wallet, error) {
	curve := ecc.GetSecp256k1()
	
	privateKey, publicKey, err := curve.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Wallet{
		privateKey: privateKey,
		publicKey:  publicKey,
		testnet:    testnet,
	}, nil
}

// NewWalletFromPrivateKey creates a wallet from an existing private key
func NewWalletFromPrivateKey(privateKey *big.Int, testnet bool) (*Wallet, error) {
	curve := ecc.GetSecp256k1()
	
	if !curve.IsValidPrivateKey(privateKey) {
		return nil, fmt.Errorf("invalid private key")
	}

	publicKey, err := curve.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	return &Wallet{
		privateKey: privateKey,
		publicKey:  publicKey,
		testnet:    testnet,
	}, nil
}

// NewWalletFromWIF creates a wallet from a WIF (Wallet Import Format) private key
func NewWalletFromWIF(wif string) (*Wallet, error) {
	privateKeyBytes, _, testnet, err := base58.DecodeWIF(wif)
	if err != nil {
		return nil, fmt.Errorf("invalid WIF: %w", err)
	}

	privateKey, err := sec.DecodePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	return NewWalletFromPrivateKey(privateKey, testnet)
}

// GetPrivateKey returns the private key
func (w *Wallet) GetPrivateKey() *big.Int {
	return new(big.Int).Set(w.privateKey)
}

// GetPublicKey returns the public key
func (w *Wallet) GetPublicKey() *ecc.Point {
	return &ecc.Point{
		X: new(big.Int).Set(w.publicKey.X),
		Y: new(big.Int).Set(w.publicKey.Y),
		A: new(big.Int).Set(w.publicKey.A),
		B: new(big.Int).Set(w.publicKey.B),
	}
}

// GetPrivateKeyWIF returns the private key in WIF format
func (w *Wallet) GetPrivateKeyWIF(compressed bool) string {
	privateKeyBytes := sec.EncodePrivateKey(w.privateKey)
	return base58.EncodeWIF(privateKeyBytes, compressed, w.testnet)
}

// GetAddress returns the P2PKH address
func (w *Wallet) GetAddress() string {
	addr := address.NewP2PKHAddress(w.publicKey, w.testnet)
	return addr.String()
}

// GetPublicKeyCompressed returns the compressed public key in SEC format
func (w *Wallet) GetPublicKeyCompressed() []byte {
	return sec.EncodePublicKeyCompressed(w.publicKey)
}

// GetPublicKeyUncompressed returns the uncompressed public key in SEC format
func (w *Wallet) GetPublicKeyUncompressed() []byte {
	return sec.EncodePublicKeyUncompressed(w.publicKey)
}

// SignMessage signs a message with the wallet's private key
func (w *Wallet) SignMessage(message []byte) (*ecc.Signature, error) {
	return ecc.SignMessage(w.privateKey, message)
}

// VerifyMessage verifies a message signature
func (w *Wallet) VerifyMessage(message []byte, signature *ecc.Signature) bool {
	return signature.VerifyMessage(w.publicKey, message)
}

// CreateTransaction creates a new transaction
func (w *Wallet) CreateTransaction(utxos []*transaction.UTXO, toAddress string, amount uint64, feePerByte uint64) (*transaction.Transaction, error) {
	return transaction.CreateSimpleTransaction(
		w.privateKey,
		utxos,
		toAddress,
		amount,
		feePerByte,
		w.testnet,
	)
}

// BitcoinUtils provides utility functions for Bitcoin operations
type BitcoinUtils struct{}

// NewBitcoinUtils creates a new BitcoinUtils instance
func NewBitcoinUtils() *BitcoinUtils {
	return &BitcoinUtils{}
}

// GenerateKeyPair generates a new private/public key pair
func (bu *BitcoinUtils) GenerateKeyPair() (*big.Int, *ecc.Point, error) {
	curve := ecc.GetSecp256k1()
	return curve.GenerateKeyPair()
}

// PrivateKeyToWIF converts a private key to WIF format
func (bu *BitcoinUtils) PrivateKeyToWIF(privateKey *big.Int, compressed, testnet bool) string {
	privateKeyBytes := sec.EncodePrivateKey(privateKey)
	return base58.EncodeWIF(privateKeyBytes, compressed, testnet)
}

// WIFToPrivateKey converts a WIF to private key
func (bu *BitcoinUtils) WIFToPrivateKey(wif string) (*big.Int, bool, bool, error) {
	privateKeyBytes, compressed, testnet, err := base58.DecodeWIF(wif)
	if err != nil {
		return nil, false, false, err
	}

	privateKey, err := sec.DecodePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, false, false, err
	}

	return privateKey, compressed, testnet, nil
}

// PublicKeyToAddress converts a public key to a Bitcoin address
func (bu *BitcoinUtils) PublicKeyToAddress(publicKey *ecc.Point, testnet bool) string {
	return address.PublicKeyToP2PKHAddress(publicKey, testnet)
}

// ValidateAddress validates a Bitcoin address
func (bu *BitcoinUtils) ValidateAddress(addressStr string) bool {
	return address.IsValid(addressStr)
}

// GetAddressInfo returns information about a Bitcoin address
func (bu *BitcoinUtils) GetAddressInfo(addressStr string) (string, string, error) {
	addr, err := address.ParseAddress(addressStr)
	if err != nil {
		return "", "", err
	}
	return addr.Type, addr.Network, nil
}

// CreateMultisigAddress creates a multisig address
func (bu *BitcoinUtils) CreateMultisigAddress(publicKeys []*ecc.Point, threshold int, testnet bool) (string, error) {
	addr, err := address.CreateMultisigAddress(publicKeys, threshold, testnet)
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

// Hash160 calculates RIPEMD160(SHA256(data))
func (bu *BitcoinUtils) Hash160(data []byte) []byte {
	return script.Hash160(data)
}

// Hash256 calculates SHA256(SHA256(data))
func (bu *BitcoinUtils) Hash256(data []byte) []byte {
	return script.Hash256(data)
}

// Example demonstrates basic Bitcoin operations
func Example() {
	fmt.Println("=== Bitcoin Cryptography Demo ===")

	// 1. Create a new wallet
	fmt.Println("\n1. Creating a new wallet...")
	wallet, err := NewWallet(false) // mainnet
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}

	fmt.Printf("Address: %s\n", wallet.GetAddress())
	fmt.Printf("Private Key (WIF): %s\n", wallet.GetPrivateKeyWIF(true))

	// 2. Sign and verify a message
	fmt.Println("\n2. Signing and verifying a message...")
	message := []byte("Hello, Bitcoin!")
	signature, err := wallet.SignMessage(message)
	if err != nil {
		fmt.Printf("Error signing message: %v\n", err)
		return
	}

	valid := wallet.VerifyMessage(message, signature)
	fmt.Printf("Message: %s\n", string(message))
	fmt.Printf("Signature valid: %t\n", valid)

	// 3. Create a multisig address
	fmt.Println("\n3. Creating a 2-of-3 multisig address...")
	utils := NewBitcoinUtils()
	
	var publicKeys []*ecc.Point
	for i := 0; i < 3; i++ {
		_, pubKey, err := utils.GenerateKeyPair()
		if err != nil {
			fmt.Printf("Error generating key pair: %v\n", err)
			return
		}
		publicKeys = append(publicKeys, pubKey)
	}

	multisigAddr, err := utils.CreateMultisigAddress(publicKeys, 2, false)
	if err != nil {
		fmt.Printf("Error creating multisig address: %v\n", err)
		return
	}
	fmt.Printf("Multisig address: %s\n", multisigAddr)

	// 4. Address validation
	fmt.Println("\n4. Address validation...")
	fmt.Printf("Wallet address valid: %t\n", utils.ValidateAddress(wallet.GetAddress()))
	fmt.Printf("Multisig address valid: %t\n", utils.ValidateAddress(multisigAddr))
	fmt.Printf("Invalid address valid: %t\n", utils.ValidateAddress("invalid_address"))

	// 5. Key format conversions
	fmt.Println("\n5. Key format conversions...")
	privateKey := wallet.GetPrivateKey()
	wif := utils.PrivateKeyToWIF(privateKey, true, false)
	fmt.Printf("WIF: %s\n", wif)

	recoveredKey, compressed, testnet, err := utils.WIFToPrivateKey(wif)
	if err != nil {
		fmt.Printf("Error recovering private key: %v\n", err)
		return
	}
	fmt.Printf("Recovered key matches: %t\n", privateKey.Cmp(recoveredKey) == 0)
	fmt.Printf("Compressed: %t, Testnet: %t\n", compressed, testnet)

	fmt.Println("\n=== Demo completed successfully! ===")
}

// GetVersion returns the version of the Bitcoin package
func GetVersion() string {
	return "1.0.0"
}

// GetSupportedFeatures returns a list of supported Bitcoin features
func GetSupportedFeatures() []string {
	return []string{
		"Elliptic Curve Cryptography (secp256k1)",
		"ECDSA Signatures",
		"SEC1 Format (compressed/uncompressed public keys)",
		"Base58Check Encoding",
		"WIF (Wallet Import Format)",
		"P2PKH (Pay-to-Public-Key-Hash) addresses",
		"P2SH (Pay-to-Script-Hash) addresses",
		"P2PK (Pay-to-Public-Key) transactions",
		"Bitcoin Script",
		"Transaction creation and validation",
		"Multisig addresses",
		"Mainnet and Testnet support",
		"SIGHASH types",
		"Transaction fee calculation",
	}
}
