package bitcoin_test

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/address"
	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/base58"
	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/script"
	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/sec"
	"github.com/DimaJoyti/go-coffee/pkg/bitcoin/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecp256k1(t *testing.T) {
	curve := ecc.GetSecp256k1()

	t.Run("Generator point is on curve", func(t *testing.T) {
		generator, err := curve.Generator()
		require.NoError(t, err)
		assert.True(t, generator.IsOnCurve())
	})

	t.Run("Generate key pair", func(t *testing.T) {
		privateKey, publicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		assert.True(t, curve.IsValidPrivateKey(privateKey))
		assert.True(t, publicKey.IsOnCurve())
	})

	t.Run("Private key to public key", func(t *testing.T) {
		// Test with known private key
		privateKeyHex := "18e14a7b6a307f426a94f8114701e7c8e774e7f9a47e2c2035db29a206321725"
		privateKey, _ := new(big.Int).SetString(privateKeyHex, 16)
		
		publicKey, err := curve.PrivateKeyToPublicKey(privateKey)
		require.NoError(t, err)
		
		// Expected public key coordinates
		expectedX := "50863ad64a87ae8a2fe83c1af1a8403cb53f53e486d8511dad8a04887e5b2352"
		expectedY := "2cd470243453a299fa9e77237716103abc11a1df38855ed6f2ee187e9c582ba6"
		
		assert.Equal(t, expectedX, hex.EncodeToString(publicKey.X.Bytes()))
		assert.Equal(t, expectedY, hex.EncodeToString(publicKey.Y.Bytes()))
	})
}

func TestECDSASignature(t *testing.T) {
	curve := ecc.GetSecp256k1()
	
	t.Run("Sign and verify", func(t *testing.T) {
		// Generate key pair
		privateKey, publicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		// Message to sign
		message := []byte("Hello, Bitcoin!")
		
		// Sign message
		signature, err := ecc.SignMessage(privateKey, message)
		require.NoError(t, err)
		
		// Verify signature
		valid := signature.VerifyMessage(publicKey, message)
		assert.True(t, valid)
		
		// Verify with wrong message should fail
		wrongMessage := []byte("Wrong message")
		valid = signature.VerifyMessage(publicKey, wrongMessage)
		assert.False(t, valid)
	})

	t.Run("DER encoding", func(t *testing.T) {
		privateKey, _, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		message := []byte("Test message")
		signature, err := ecc.SignMessage(privateKey, message)
		require.NoError(t, err)
		
		der := signature.DER()
		assert.True(t, len(der) > 0)
		assert.Equal(t, byte(0x30), der[0]) // SEQUENCE tag
	})
}

func TestSECEncoding(t *testing.T) {
	curve := ecc.GetSecp256k1()
	
	t.Run("Compressed and uncompressed encoding", func(t *testing.T) {
		privateKey, publicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		// Test compressed encoding
		compressed := sec.EncodePublicKeyCompressed(publicKey)
		assert.Equal(t, 33, len(compressed))
		assert.True(t, compressed[0] == 0x02 || compressed[0] == 0x03)
		
		// Test uncompressed encoding
		uncompressed := sec.EncodePublicKeyUncompressed(publicKey)
		assert.Equal(t, 65, len(uncompressed))
		assert.Equal(t, byte(0x04), uncompressed[0])
		
		// Test decoding
		decodedCompressed, err := sec.DecodePublicKey(compressed)
		require.NoError(t, err)
		assert.True(t, publicKey.Equal(decodedCompressed))
		
		decodedUncompressed, err := sec.DecodePublicKey(uncompressed)
		require.NoError(t, err)
		assert.True(t, publicKey.Equal(decodedUncompressed))
		
		// Test private key encoding
		privateKeyBytes := sec.EncodePrivateKey(privateKey)
		assert.Equal(t, 32, len(privateKeyBytes))
		
		decodedPrivateKey, err := sec.DecodePrivateKey(privateKeyBytes)
		require.NoError(t, err)
		assert.Equal(t, 0, privateKey.Cmp(decodedPrivateKey))
	})
}

func TestBase58(t *testing.T) {
	t.Run("Basic encoding and decoding", func(t *testing.T) {
		testData := []byte("Hello, World!")
		
		encoded := base58.Encode(testData)
		assert.True(t, len(encoded) > 0)
		
		decoded, err := base58.Decode(encoded)
		require.NoError(t, err)
		assert.Equal(t, testData, decoded)
	})

	t.Run("Base58Check encoding", func(t *testing.T) {
		testData := []byte("Test data")
		
		encoded := base58.EncodeCheck(testData)
		assert.True(t, len(encoded) > 0)
		
		decoded, err := base58.DecodeCheck(encoded)
		require.NoError(t, err)
		assert.Equal(t, testData, decoded)
		
		// Test invalid checksum
		invalidEncoded := encoded[:len(encoded)-1] + "X"
		_, err = base58.DecodeCheck(invalidEncoded)
		assert.Error(t, err)
	})

	t.Run("WIF encoding and decoding", func(t *testing.T) {
		privateKeyBytes := make([]byte, 32)
		for i := range privateKeyBytes {
			privateKeyBytes[i] = byte(i + 1)
		}
		
		// Test mainnet uncompressed
		wif := base58.EncodeWIF(privateKeyBytes, false, false)
		assert.True(t, len(wif) > 0)
		
		decoded, compressed, testnet, err := base58.DecodeWIF(wif)
		require.NoError(t, err)
		assert.Equal(t, privateKeyBytes, decoded)
		assert.False(t, compressed)
		assert.False(t, testnet)
		
		// Test testnet compressed
		wifTestnet := base58.EncodeWIF(privateKeyBytes, true, true)
		decoded, compressed, testnet, err = base58.DecodeWIF(wifTestnet)
		require.NoError(t, err)
		assert.Equal(t, privateKeyBytes, decoded)
		assert.True(t, compressed)
		assert.True(t, testnet)
	})
}

func TestBitcoinAddress(t *testing.T) {
	curve := ecc.GetSecp256k1()
	
	t.Run("P2PKH address creation", func(t *testing.T) {
		privateKey, publicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		// Test mainnet address
		addr := address.NewP2PKHAddress(publicKey, false)
		addrStr := addr.String()
		assert.True(t, len(addrStr) > 0)
		assert.True(t, addrStr[0] == '1') // Mainnet P2PKH starts with '1'
		
		// Test testnet address
		addrTestnet := address.NewP2PKHAddress(publicKey, true)
		addrTestnetStr := addrTestnet.String()
		assert.True(t, len(addrTestnetStr) > 0)
		assert.True(t, addrTestnetStr[0] == 'm' || addrTestnetStr[0] == 'n') // Testnet P2PKH
		
		// Test address parsing
		parsedAddr, err := address.ParseAddress(addrStr)
		require.NoError(t, err)
		assert.Equal(t, "P2PKH", parsedAddr.Type)
		assert.Equal(t, "mainnet", parsedAddr.Network)
		
		// Test WIF to address conversion
		privateKeyBytes := sec.EncodePrivateKey(privateKey)
		wif := base58.EncodeWIF(privateKeyBytes, true, false)
		addrFromWIF, err := address.WIFToP2PKHAddress(wif)
		require.NoError(t, err)
		
		// Should match the original address (both use compressed public key)
		compressedAddr := address.PublicKeyToP2PKHAddress(publicKey, false)
		assert.Equal(t, compressedAddr, addrFromWIF)
	})

	t.Run("Multisig address", func(t *testing.T) {
		// Generate 3 key pairs
		var publicKeys []*ecc.Point
		for i := 0; i < 3; i++ {
			_, publicKey, err := curve.GenerateKeyPair()
			require.NoError(t, err)
			publicKeys = append(publicKeys, publicKey)
		}
		
		// Create 2-of-3 multisig address
		multisigAddr, err := address.CreateMultisigAddress(publicKeys, 2, false)
		require.NoError(t, err)
		
		addrStr := multisigAddr.String()
		assert.True(t, len(addrStr) > 0)
		assert.True(t, addrStr[0] == '3') // P2SH addresses start with '3'
		
		assert.Equal(t, "P2SH", multisigAddr.Type)
		assert.Equal(t, "mainnet", multisigAddr.Network)
	})
}

func TestBitcoinScript(t *testing.T) {
	t.Run("P2PKH script creation", func(t *testing.T) {
		hash160 := make([]byte, 20)
		for i := range hash160 {
			hash160[i] = byte(i)
		}
		
		scriptObj := script.CreateP2PKHScript(hash160)
		assert.True(t, scriptObj.IsP2PKH())
		
		extractedHash := scriptObj.GetP2PKHHash160()
		assert.Equal(t, hash160, extractedHash)
		
		// Test serialization and parsing
		serialized := scriptObj.Serialize()
		parsed, err := script.Parse(serialized)
		require.NoError(t, err)
		assert.True(t, parsed.IsP2PKH())
	})

	t.Run("Script signature creation and parsing", func(t *testing.T) {
		curve := ecc.GetSecp256k1()
		privateKey, publicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		// Create a dummy signature
		message := []byte("test message")
		signature, err := ecc.SignMessage(privateKey, message)
		require.NoError(t, err)
		
		// Create script signature
		derSig := signature.DER()
		derSig = append(derSig, byte(transaction.SighashAll))
		scriptSig := script.CreateP2PKHScriptSig(derSig, publicKey)
		
		// Parse script signature
		parsedSig, parsedPubKey, sighashType, err := script.ParseP2PKHScriptSig(scriptSig)
		require.NoError(t, err)
		
		assert.Equal(t, signature.R, parsedSig.R)
		assert.Equal(t, signature.S, parsedSig.S)
		assert.True(t, publicKey.Equal(parsedPubKey))
		assert.Equal(t, byte(transaction.SighashAll), sighashType)
	})
}

func TestTransaction(t *testing.T) {
	t.Run("Transaction creation and serialization", func(t *testing.T) {
		// Create a simple transaction
		prevTxHash := make([]byte, 32)
		for i := range prevTxHash {
			prevTxHash[i] = byte(i)
		}
		
		txIn := transaction.NewTxIn(prevTxHash, 0)
		
		scriptPubKey := []byte{0x76, 0xa9, 0x14} // OP_DUP OP_HASH160 <20 bytes>
		scriptPubKey = append(scriptPubKey, make([]byte, 20)...)
		scriptPubKey = append(scriptPubKey, 0x88, 0xac) // OP_EQUALVERIFY OP_CHECKSIG
		
		txOut := transaction.NewTxOut(5000000000, scriptPubKey) // 50 BTC
		
		tx := transaction.NewTransaction(1, []*transaction.TxIn{txIn}, []*transaction.TxOut{txOut}, 0)
		
		// Test serialization
		serialized := tx.Serialize()
		assert.True(t, len(serialized) > 0)
		
		// Test hash calculation
		hash := tx.Hash()
		assert.Equal(t, 32, len(hash))
		
		// Test ID
		id := tx.ID()
		assert.Equal(t, 64, len(id)) // 32 bytes as hex string
	})

	t.Run("Transaction builder", func(t *testing.T) {
		curve := ecc.GetSecp256k1()
		privateKey, publicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		
		// Create a UTXO
		utxo := &transaction.UTXO{
			TxHash:      make([]byte, 32),
			OutputIndex: 0,
			Amount:      100000000, // 1 BTC
			ScriptPubKey: address.NewP2PKHAddress(publicKey, false).ScriptPubKey(),
		}
		
		// Create destination address
		_, destPublicKey, err := curve.GenerateKeyPair()
		require.NoError(t, err)
		destAddr := address.NewP2PKHAddress(destPublicKey, false)
		
		// Build transaction
		builder := transaction.NewTransactionBuilder()
		builder.AddInput(utxo.TxHash, utxo.OutputIndex, utxo)
		err = builder.AddOutput(destAddr.String(), 50000000) // 0.5 BTC
		require.NoError(t, err)
		builder.SetFee(10000) // 0.0001 BTC fee
		
		// Sign transaction
		tx, err := builder.SignTransaction([]*big.Int{privateKey})
		require.NoError(t, err)
		
		assert.Equal(t, 1, len(tx.TxIns))
		assert.Equal(t, 2, len(tx.TxOuts)) // Output + change
		
		// Verify transaction
		err = transaction.ValidateTransaction(tx, []*transaction.UTXO{utxo})
		assert.NoError(t, err)
	})
}
