package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin/script"
)

// SIGHASH types for transaction signing
const (
	SighashAll          = 0x01
	SighashNone         = 0x02
	SighashSingle       = 0x03
	SighashAnyoneCanPay = 0x80
)

// TxIn represents a transaction input
type TxIn struct {
	PrevTx        []byte // Previous transaction hash (32 bytes)
	PrevIndex     uint32 // Previous output index
	ScriptSig     []byte // Script signature
	Sequence      uint32 // Sequence number
}

// NewTxIn creates a new transaction input
func NewTxIn(prevTx []byte, prevIndex uint32) *TxIn {
	return &TxIn{
		PrevTx:    prevTx,
		PrevIndex: prevIndex,
		ScriptSig: []byte{},
		Sequence:  0xffffffff, // Default sequence
	}
}

// Serialize serializes the transaction input
func (txIn *TxIn) Serialize() []byte {
	var buf bytes.Buffer

	// Previous transaction hash (32 bytes, little-endian)
	buf.Write(reverseBytes(txIn.PrevTx))

	// Previous output index (4 bytes, little-endian)
	binary.Write(&buf, binary.LittleEndian, txIn.PrevIndex)

	// Script signature length (varint)
	buf.Write(encodeVarint(uint64(len(txIn.ScriptSig))))

	// Script signature
	buf.Write(txIn.ScriptSig)

	// Sequence (4 bytes, little-endian)
	binary.Write(&buf, binary.LittleEndian, txIn.Sequence)

	return buf.Bytes()
}

// TxOut represents a transaction output
type TxOut struct {
	Amount        uint64 // Amount in satoshis
	ScriptPubKey  []byte // Script public key
}

// NewTxOut creates a new transaction output
func NewTxOut(amount uint64, scriptPubKey []byte) *TxOut {
	return &TxOut{
		Amount:       amount,
		ScriptPubKey: scriptPubKey,
	}
}

// Serialize serializes the transaction output
func (txOut *TxOut) Serialize() []byte {
	var buf bytes.Buffer

	// Amount (8 bytes, little-endian)
	binary.Write(&buf, binary.LittleEndian, txOut.Amount)

	// Script public key length (varint)
	buf.Write(encodeVarint(uint64(len(txOut.ScriptPubKey))))

	// Script public key
	buf.Write(txOut.ScriptPubKey)

	return buf.Bytes()
}

// Transaction represents a Bitcoin transaction
type Transaction struct {
	Version  uint32   // Transaction version
	TxIns    []*TxIn  // Transaction inputs
	TxOuts   []*TxOut // Transaction outputs
	Locktime uint32   // Lock time
}

// NewTransaction creates a new transaction
func NewTransaction(version uint32, txIns []*TxIn, txOuts []*TxOut, locktime uint32) *Transaction {
	return &Transaction{
		Version:  version,
		TxIns:    txIns,
		TxOuts:   txOuts,
		Locktime: locktime,
	}
}

// Serialize serializes the transaction
func (tx *Transaction) Serialize() []byte {
	var buf bytes.Buffer

	// Version (4 bytes, little-endian)
	binary.Write(&buf, binary.LittleEndian, tx.Version)

	// Input count (varint)
	buf.Write(encodeVarint(uint64(len(tx.TxIns))))

	// Inputs
	for _, txIn := range tx.TxIns {
		buf.Write(txIn.Serialize())
	}

	// Output count (varint)
	buf.Write(encodeVarint(uint64(len(tx.TxOuts))))

	// Outputs
	for _, txOut := range tx.TxOuts {
		buf.Write(txOut.Serialize())
	}

	// Locktime (4 bytes, little-endian)
	binary.Write(&buf, binary.LittleEndian, tx.Locktime)

	return buf.Bytes()
}

// Hash calculates the transaction hash (double SHA256)
func (tx *Transaction) Hash() []byte {
	serialized := tx.Serialize()
	first := sha256.Sum256(serialized)
	second := sha256.Sum256(first[:])
	return reverseBytes(second[:]) // Bitcoin uses little-endian for hashes
}

// ID returns the transaction ID (hex string of hash)
func (tx *Transaction) ID() string {
	hash := tx.Hash()
	return fmt.Sprintf("%x", hash)
}

// SignatureHash calculates the signature hash for a specific input
func (tx *Transaction) SignatureHash(inputIndex int, scriptCode []byte, sighashType int) ([]byte, error) {
	if inputIndex >= len(tx.TxIns) {
		return nil, fmt.Errorf("input index out of range")
	}

	// Create a copy of the transaction for signing
	txCopy := &Transaction{
		Version:  tx.Version,
		TxIns:    make([]*TxIn, len(tx.TxIns)),
		TxOuts:   make([]*TxOut, len(tx.TxOuts)),
		Locktime: tx.Locktime,
	}

	// Copy inputs
	for i, txIn := range tx.TxIns {
		txCopy.TxIns[i] = &TxIn{
			PrevTx:    txIn.PrevTx,
			PrevIndex: txIn.PrevIndex,
			ScriptSig: []byte{}, // Clear all script signatures
			Sequence:  txIn.Sequence,
		}
	}

	// Copy outputs
	for i, txOut := range tx.TxOuts {
		txCopy.TxOuts[i] = &TxOut{
			Amount:       txOut.Amount,
			ScriptPubKey: txOut.ScriptPubKey,
		}
	}

	// Set the script code for the input being signed
	txCopy.TxIns[inputIndex].ScriptSig = scriptCode

	// Handle different SIGHASH types
	switch sighashType & 0x1f {
	case SighashAll:
		// Default behavior - sign all inputs and outputs
	case SighashNone:
		// Sign all inputs but no outputs
		txCopy.TxOuts = []*TxOut{}
		// Set sequence to 0 for all inputs except the one being signed
		for i := range txCopy.TxIns {
			if i != inputIndex {
				txCopy.TxIns[i].Sequence = 0
			}
		}
	case SighashSingle:
		// Sign all inputs and only the output with the same index
		if inputIndex >= len(txCopy.TxOuts) {
			return nil, fmt.Errorf("SIGHASH_SINGLE: input index exceeds output count")
		}
		// Keep only the output at the same index
		txCopy.TxOuts = txCopy.TxOuts[:inputIndex+1]
		// Set sequence to 0 for all inputs except the one being signed
		for i := range txCopy.TxIns {
			if i != inputIndex {
				txCopy.TxIns[i].Sequence = 0
			}
		}
	}

	// Handle SIGHASH_ANYONECANPAY
	if sighashType&SighashAnyoneCanPay != 0 {
		// Keep only the input being signed
		txCopy.TxIns = []*TxIn{txCopy.TxIns[inputIndex]}
	}

	// Serialize the transaction copy
	serialized := txCopy.Serialize()

	// Append the sighash type (4 bytes, little-endian)
	var buf bytes.Buffer
	buf.Write(serialized)
	binary.Write(&buf, binary.LittleEndian, uint32(sighashType))

	// Double SHA256
	first := sha256.Sum256(buf.Bytes())
	second := sha256.Sum256(first[:])

	return second[:], nil
}

// SignInput signs a transaction input with a private key
func (tx *Transaction) SignInput(inputIndex int, privateKey *big.Int, scriptPubKey []byte, sighashType int) error {
	if inputIndex >= len(tx.TxIns) {
		return fmt.Errorf("input index out of range")
	}

	// Calculate signature hash
	sigHash, err := tx.SignatureHash(inputIndex, scriptPubKey, sighashType)
	if err != nil {
		return fmt.Errorf("failed to calculate signature hash: %w", err)
	}

	// Sign the hash
	signature, err := ecc.Sign(privateKey, sigHash)
	if err != nil {
		return fmt.Errorf("failed to sign hash: %w", err)
	}

	// Encode signature in DER format and append sighash type
	derSig := signature.DER()
	derSig = append(derSig, byte(sighashType))

	// Get public key
	curve := ecc.GetSecp256k1()
	publicKey, err := curve.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	// Create script signature (signature + public key for P2PKH)
	scriptSig := script.CreateP2PKHScriptSig(derSig, publicKey)

	// Set the script signature
	tx.TxIns[inputIndex].ScriptSig = scriptSig

	return nil
}

// VerifyInput verifies a transaction input signature
func (tx *Transaction) VerifyInput(inputIndex int, scriptPubKey []byte) (bool, error) {
	if inputIndex >= len(tx.TxIns) {
		return false, fmt.Errorf("input index out of range")
	}

	// Parse the script signature to extract signature and public key
	scriptSig := tx.TxIns[inputIndex].ScriptSig
	signature, publicKey, sighashType, err := script.ParseP2PKHScriptSig(scriptSig)
	if err != nil {
		return false, fmt.Errorf("failed to parse script signature: %w", err)
	}

	// Calculate signature hash
	sigHash, err := tx.SignatureHash(inputIndex, scriptPubKey, int(sighashType))
	if err != nil {
		return false, fmt.Errorf("failed to calculate signature hash: %w", err)
	}

	// Verify signature
	return signature.Verify(publicKey, sigHash), nil
}

// encodeVarint encodes an integer as a variable-length integer
func encodeVarint(value uint64) []byte {
	if value < 0xfd {
		return []byte{byte(value)}
	} else if value <= 0xffff {
		var buf bytes.Buffer
		buf.WriteByte(0xfd)
		binary.Write(&buf, binary.LittleEndian, uint16(value))
		return buf.Bytes()
	} else if value <= 0xffffffff {
		var buf bytes.Buffer
		buf.WriteByte(0xfe)
		binary.Write(&buf, binary.LittleEndian, uint32(value))
		return buf.Bytes()
	} else {
		var buf bytes.Buffer
		buf.WriteByte(0xff)
		binary.Write(&buf, binary.LittleEndian, value)
		return buf.Bytes()
	}
}

// reverseBytes reverses a byte slice
func reverseBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[len(data)-1-i] = b
	}
	return result
}

// Fee calculates the transaction fee (total inputs - total outputs)
func (tx *Transaction) Fee(inputAmounts []uint64) uint64 {
	if len(inputAmounts) != len(tx.TxIns) {
		return 0
	}

	var totalInput, totalOutput uint64

	for _, amount := range inputAmounts {
		totalInput += amount
	}

	for _, txOut := range tx.TxOuts {
		totalOutput += txOut.Amount
	}

	if totalInput > totalOutput {
		return totalInput - totalOutput
	}

	return 0
}

// Size returns the serialized size of the transaction in bytes
func (tx *Transaction) Size() int {
	return len(tx.Serialize())
}

// VSize returns the virtual size of the transaction (for fee calculation)
func (tx *Transaction) VSize() int {
	// For non-SegWit transactions, vsize equals size
	return tx.Size()
}
