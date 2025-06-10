package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/address"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/base58"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/script"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/sec"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/transaction"
)

func main() {
	fmt.Println("🚀 Bitcoin Cryptography Implementation Demo")
	fmt.Println("==========================================")

	// Run all examples
	demonstrateEllipticCurveCryptography()
	demonstrateSECFormat()
	demonstrateBase58Encoding()
	demonstrateAddresses()
	demonstrateTransactions()
	demonstrateScripts()
	demonstrateWalletOperations()
	demonstrateAdvancedFeatures()

	fmt.Println("\n✅ All demonstrations completed successfully!")
}

func demonstrateEllipticCurveCryptography() {
	fmt.Println("\n📐 1. Elliptic Curve Cryptography (secp256k1)")
	fmt.Println("---------------------------------------------")

	curve := ecc.GetSecp256k1()

	// Generate a key pair
	privateKey, publicKey, err := curve.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	fmt.Printf("Private Key: %064x\n", privateKey)
	fmt.Printf("Public Key X: %064x\n", publicKey.X)
	fmt.Printf("Public Key Y: %064x\n", publicKey.Y)
	fmt.Printf("Point is on curve: %t\n", publicKey.IsOnCurve())

	// Test point operations
	generator, _ := curve.Generator()
	doubled, _ := generator.Double()
	fmt.Printf("Generator doubled is on curve: %t\n", doubled.IsOnCurve())

	// Test scalar multiplication
	scalar := big.NewInt(12345)
	result, _ := generator.ScalarMult(scalar)
	fmt.Printf("Scalar multiplication result is on curve: %t\n", result.IsOnCurve())
}

func demonstrateSECFormat() {
	fmt.Println("\n🔐 2. SEC Format Encoding")
	fmt.Println("-------------------------")

	curve := ecc.GetSecp256k1()
	_, publicKey, _ := curve.GenerateKeyPair()

	// Compressed format
	compressed := sec.EncodePublicKeyCompressed(publicKey)
	fmt.Printf("Compressed public key (%d bytes): %x\n", len(compressed), compressed)

	// Uncompressed format
	uncompressed := sec.EncodePublicKeyUncompressed(publicKey)
	fmt.Printf("Uncompressed public key (%d bytes): %x\n", len(uncompressed), uncompressed)

	// Test decoding
	decodedCompressed, _ := sec.DecodePublicKey(compressed)
	decodedUncompressed, _ := sec.DecodePublicKey(uncompressed)

	fmt.Printf("Compressed decoding matches: %t\n", publicKey.Equal(decodedCompressed))
	fmt.Printf("Uncompressed decoding matches: %t\n", publicKey.Equal(decodedUncompressed))

	// Test format detection
	fmt.Printf("Compressed format detected: %t\n", sec.IsCompressed(compressed))
	fmt.Printf("Uncompressed format detected: %t\n", sec.IsUncompressed(uncompressed))
}

func demonstrateBase58Encoding() {
	fmt.Println("\n🔤 3. Base58Check Encoding")
	fmt.Println("--------------------------")

	testData := []byte("Hello, Bitcoin!")
	
	// Basic Base58 encoding
	encoded := base58.Encode(testData)
	fmt.Printf("Base58 encoded: %s\n", encoded)

	decoded, _ := base58.Decode(encoded)
	fmt.Printf("Decoded matches: %t\n", string(decoded) == string(testData))

	// Base58Check encoding (with checksum)
	encodedCheck := base58.EncodeCheck(testData)
	fmt.Printf("Base58Check encoded: %s\n", encodedCheck)

	decodedCheck, _ := base58.DecodeCheck(encodedCheck)
	fmt.Printf("Base58Check decoded matches: %t\n", string(decodedCheck) == string(testData))

	// WIF encoding
	privateKeyBytes := make([]byte, 32)
	for i := range privateKeyBytes {
		privateKeyBytes[i] = byte(i + 1)
	}

	wifMainnet := base58.EncodeWIF(privateKeyBytes, true, false)
	wifTestnet := base58.EncodeWIF(privateKeyBytes, true, true)
	
	fmt.Printf("WIF Mainnet: %s\n", wifMainnet)
	fmt.Printf("WIF Testnet: %s\n", wifTestnet)

	// Decode WIF
	decoded, compressed, testnet, _ := base58.DecodeWIF(wifMainnet)
	fmt.Printf("WIF decoded - Compressed: %t, Testnet: %t\n", compressed, testnet)
	fmt.Printf("Private key matches: %t\n", hex.EncodeToString(decoded) == hex.EncodeToString(privateKeyBytes))
}

func demonstrateAddresses() {
	fmt.Println("\n🏠 4. Bitcoin Addresses")
	fmt.Println("-----------------------")

	curve := ecc.GetSecp256k1()
	_, publicKey, _ := curve.GenerateKeyPair()

	// P2PKH addresses
	mainnetAddr := address.NewP2PKHAddress(publicKey, false)
	testnetAddr := address.NewP2PKHAddress(publicKey, true)

	fmt.Printf("Mainnet P2PKH: %s\n", mainnetAddr.String())
	fmt.Printf("Testnet P2PKH: %s\n", testnetAddr.String())

	// Address validation
	fmt.Printf("Mainnet address valid: %t\n", address.IsValid(mainnetAddr.String()))
	fmt.Printf("Address type: %s\n", address.GetAddressType(mainnetAddr.String()))
	fmt.Printf("Address network: %s\n", address.GetAddressNetwork(mainnetAddr.String()))

	// Multisig address
	var publicKeys []*ecc.Point
	for i := 0; i < 3; i++ {
		_, pubKey, _ := curve.GenerateKeyPair()
		publicKeys = append(publicKeys, pubKey)
	}

	multisigAddr, _ := address.CreateMultisigAddress(publicKeys, 2, false)
	fmt.Printf("2-of-3 Multisig: %s\n", multisigAddr.String())
	fmt.Printf("Multisig type: %s\n", multisigAddr.Type)

	// Script public keys
	p2pkhScript := mainnetAddr.ScriptPubKey()
	multisigScript := multisigAddr.ScriptPubKey()
	
	fmt.Printf("P2PKH script length: %d bytes\n", len(p2pkhScript))
	fmt.Printf("P2SH script length: %d bytes\n", len(multisigScript))
}

func demonstrateTransactions() {
	fmt.Println("\n💸 5. Bitcoin Transactions")
	fmt.Println("--------------------------")

	curve := ecc.GetSecp256k1()
	privateKey, publicKey, _ := curve.GenerateKeyPair()

	// Create a mock UTXO
	utxo := &transaction.UTXO{
		TxHash:       make([]byte, 32),
		OutputIndex:  0,
		Amount:       100000000, // 1 BTC in satoshis
		ScriptPubKey: address.NewP2PKHAddress(publicKey, false).ScriptPubKey(),
		Address:      address.PublicKeyToP2PKHAddress(publicKey, false),
	}

	// Fill UTXO hash with test data
	for i := range utxo.TxHash {
		utxo.TxHash[i] = byte(i)
	}

	// Create destination address
	_, destPublicKey, _ := curve.GenerateKeyPair()
	destAddr := address.PublicKeyToP2PKHAddress(destPublicKey, false)

	fmt.Printf("From address: %s\n", utxo.Address)
	fmt.Printf("To address: %s\n", destAddr)
	fmt.Printf("Amount: 0.5 BTC\n")
	fmt.Printf("Fee: 0.0001 BTC\n")

	// Build transaction using the builder
	builder := transaction.NewTransactionBuilder()
	builder.AddInput(utxo.TxHash, utxo.OutputIndex, utxo)
	builder.AddOutput(destAddr, 50000000) // 0.5 BTC
	builder.SetFee(10000) // 0.0001 BTC

	// Sign the transaction
	tx, err := builder.SignTransaction([]*big.Int{privateKey})
	if err != nil {
		log.Printf("Failed to sign transaction: %v", err)
		return
	}

	fmt.Printf("Transaction ID: %s\n", tx.ID())
	fmt.Printf("Transaction size: %d bytes\n", tx.Size())
	fmt.Printf("Number of inputs: %d\n", len(tx.TxIns))
	fmt.Printf("Number of outputs: %d\n", len(tx.TxOuts))

	// Validate transaction
	err = transaction.ValidateTransaction(tx, []*transaction.UTXO{utxo})
	if err != nil {
		fmt.Printf("Transaction validation failed: %v\n", err)
	} else {
		fmt.Println("✅ Transaction validation passed!")
	}

	// Calculate fee
	fee := transaction.CalculateTransactionFee(tx, []*transaction.UTXO{utxo})
	fmt.Printf("Actual fee: %d satoshis\n", fee)
}

func demonstrateScripts() {
	fmt.Println("\n📜 6. Bitcoin Script")
	fmt.Println("--------------------")

	// Create a P2PKH script
	hash160 := make([]byte, 20)
	for i := range hash160 {
		hash160[i] = byte(i)
	}

	p2pkhScript := script.CreateP2PKHScript(hash160)
	fmt.Printf("P2PKH script is valid: %t\n", p2pkhScript.IsP2PKH())

	// Serialize and parse
	serialized := p2pkhScript.Serialize()
	fmt.Printf("Serialized script: %x\n", serialized)

	parsed, _ := script.Parse(serialized)
	fmt.Printf("Parsed script is P2PKH: %t\n", parsed.IsP2PKH())

	// Extract hash160
	extractedHash := parsed.GetP2PKHHash160()
	fmt.Printf("Extracted hash matches: %t\n", hex.EncodeToString(extractedHash) == hex.EncodeToString(hash160))

	// Hash functions
	testData := []byte("test data")
	hash160Result := script.Hash160(testData)
	hash256Result := script.Hash256(testData)

	fmt.Printf("Hash160: %x\n", hash160Result)
	fmt.Printf("Hash256: %x\n", hash256Result)
}

func demonstrateWalletOperations() {
	fmt.Println("\n👛 7. Wallet Operations")
	fmt.Println("-----------------------")

	// Create a new wallet
	wallet, _ := bitcoin.NewWallet(false) // mainnet

	fmt.Printf("Wallet address: %s\n", wallet.GetAddress())
	fmt.Printf("Private key (WIF): %s\n", wallet.GetPrivateKeyWIF(true))

	// Sign and verify a message
	message := []byte("Bitcoin is digital gold!")
	signature, _ := wallet.SignMessage(message)
	valid := wallet.VerifyMessage(message, signature)

	fmt.Printf("Message: %s\n", string(message))
	fmt.Printf("Signature valid: %t\n", valid)

	// Test with wrong message
	wrongMessage := []byte("Wrong message")
	wrongValid := wallet.VerifyMessage(wrongMessage, signature)
	fmt.Printf("Wrong message signature valid: %t\n", wrongValid)

	// Create wallet from WIF
	wif := wallet.GetPrivateKeyWIF(true)
	walletFromWIF, _ := bitcoin.NewWalletFromWIF(wif)
	fmt.Printf("Wallet from WIF address matches: %t\n", 
		wallet.GetAddress() == walletFromWIF.GetAddress())
}

func demonstrateAdvancedFeatures() {
	fmt.Println("\n🔬 8. Advanced Features")
	fmt.Println("-----------------------")

	utils := bitcoin.NewBitcoinUtils()

	// Show supported features
	fmt.Println("Supported features:")
	for i, feature := range bitcoin.GetSupportedFeatures() {
		fmt.Printf("  %d. %s\n", i+1, feature)
	}

	// ECDSA signature recovery
	curve := ecc.GetSecp256k1()
	privateKey, publicKey, _ := curve.GenerateKeyPair()
	
	message := []byte("Recovery test")
	signature, _ := ecc.SignMessage(privateKey, message)

	// Try to recover public key (simplified example)
	fmt.Printf("\nSignature recovery test:\n")
	fmt.Printf("Original public key: %064x%064x\n", publicKey.X, publicKey.Y)
	
	// In a real implementation, you would try different recovery IDs
	for recoveryID := 0; recoveryID < 4; recoveryID++ {
		recovered, err := signature.RecoverPublicKey(script.Hash256(message), recoveryID)
		if err == nil && recovered.Equal(publicKey) {
			fmt.Printf("✅ Public key recovered with recovery ID: %d\n", recoveryID)
			break
		}
	}

	// Address format conversions
	fmt.Printf("\nAddress format conversions:\n")
	testAddr := utils.PublicKeyToAddress(publicKey, false)
	fmt.Printf("P2PKH address: %s\n", testAddr)
	
	addrType, network, _ := utils.GetAddressInfo(testAddr)
	fmt.Printf("Address type: %s, Network: %s\n", addrType, network)

	// Hash functions
	testData := []byte("Advanced test data")
	fmt.Printf("\nHash functions:\n")
	fmt.Printf("Hash160: %x\n", utils.Hash160(testData))
	fmt.Printf("Hash256: %x\n", utils.Hash256(testData))

	fmt.Printf("\nPackage version: %s\n", bitcoin.GetVersion())
}
