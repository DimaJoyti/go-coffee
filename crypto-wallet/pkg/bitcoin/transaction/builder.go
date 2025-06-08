package transaction

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin/address"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin/ecc"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash       []byte // Transaction hash
	OutputIndex  uint32 // Output index
	Amount       uint64 // Amount in satoshis
	ScriptPubKey []byte // Script public key
	Address      string // Address (for convenience)
}

// TransactionBuilder helps build Bitcoin transactions
type TransactionBuilder struct {
	version  uint32
	inputs   []*TxIn
	outputs  []*TxOut
	locktime uint32
	utxos    []*UTXO
	fee      uint64
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		version:  1,
		inputs:   []*TxIn{},
		outputs:  []*TxOut{},
		locktime: 0,
		utxos:    []*UTXO{},
		fee:      0,
	}
}

// SetVersion sets the transaction version
func (tb *TransactionBuilder) SetVersion(version uint32) *TransactionBuilder {
	tb.version = version
	return tb
}

// SetLocktime sets the transaction locktime
func (tb *TransactionBuilder) SetLocktime(locktime uint32) *TransactionBuilder {
	tb.locktime = locktime
	return tb
}

// AddInput adds an input to the transaction
func (tb *TransactionBuilder) AddInput(txHash []byte, outputIndex uint32, utxo *UTXO) *TransactionBuilder {
	input := NewTxIn(txHash, outputIndex)
	tb.inputs = append(tb.inputs, input)
	
	if utxo != nil {
		tb.utxos = append(tb.utxos, utxo)
	}
	
	return tb
}

// AddOutput adds an output to the transaction
func (tb *TransactionBuilder) AddOutput(addressStr string, amount uint64) error {
	// Parse address to get script public key
	addr, err := address.ParseAddress(addressStr)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}

	scriptPubKey := addr.ScriptPubKey()
	output := NewTxOut(amount, scriptPubKey)
	tb.outputs = append(tb.outputs, output)

	return nil
}

// AddP2PKHOutput adds a P2PKH output
func (tb *TransactionBuilder) AddP2PKHOutput(publicKey *ecc.Point, amount uint64, testnet bool) *TransactionBuilder {
	addr := address.NewP2PKHAddress(publicKey, testnet)
	scriptPubKey := addr.ScriptPubKey()
	output := NewTxOut(amount, scriptPubKey)
	tb.outputs = append(tb.outputs, output)
	
	return tb
}

// SetFee sets the transaction fee
func (tb *TransactionBuilder) SetFee(fee uint64) *TransactionBuilder {
	tb.fee = fee
	return tb
}

// CalculateFee calculates the transaction fee based on size and fee rate
func (tb *TransactionBuilder) CalculateFee(feePerByte uint64) uint64 {
	// Estimate transaction size
	estimatedSize := tb.EstimateSize()
	return uint64(estimatedSize) * feePerByte
}

// EstimateSize estimates the transaction size in bytes
func (tb *TransactionBuilder) EstimateSize() int {
	// Base size: version (4) + input count (1) + output count (1) + locktime (4)
	size := 4 + 1 + 1 + 4

	// Add input sizes
	for range tb.inputs {
		// Each input: prev_hash (32) + prev_index (4) + script_sig_len (1) + script_sig (~107 for P2PKH) + sequence (4)
		size += 32 + 4 + 1 + 107 + 4 // ~148 bytes per P2PKH input
	}

	// Add output sizes
	for _, output := range tb.outputs {
		// Each output: amount (8) + script_len (1) + script
		size += 8 + 1 + len(output.ScriptPubKey)
	}

	return size
}

// Build builds the transaction
func (tb *TransactionBuilder) Build() (*Transaction, error) {
	if len(tb.inputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one input")
	}

	if len(tb.outputs) == 0 {
		return nil, fmt.Errorf("transaction must have at least one output")
	}

	// Check if we have enough UTXOs for inputs
	if len(tb.utxos) != len(tb.inputs) {
		return nil, fmt.Errorf("number of UTXOs must match number of inputs")
	}

	// Calculate total input and output amounts
	totalInput := uint64(0)
	for _, utxo := range tb.utxos {
		totalInput += utxo.Amount
	}

	totalOutput := uint64(0)
	for _, output := range tb.outputs {
		totalOutput += output.Amount
	}

	// Check if we have enough funds
	if totalInput < totalOutput+tb.fee {
		return nil, fmt.Errorf("insufficient funds: input=%d, output=%d, fee=%d", 
			totalInput, totalOutput, tb.fee)
	}

	// Create transaction
	tx := NewTransaction(tb.version, tb.inputs, tb.outputs, tb.locktime)
	
	return tx, nil
}

// SignTransaction signs all inputs of a transaction
func (tb *TransactionBuilder) SignTransaction(privateKeys []*big.Int) (*Transaction, error) {
	// Check if we need to add change output
	if len(tb.utxos) > 0 && len(privateKeys) > 0 {
		totalInput := uint64(0)
		for _, utxo := range tb.utxos {
			totalInput += utxo.Amount
		}

		totalOutput := uint64(0)
		for _, output := range tb.outputs {
			totalOutput += output.Amount
		}

		// Add change output if needed
		change := totalInput - totalOutput - tb.fee
		if change > 0 {
			// Generate change address from the first private key
			curve := ecc.GetSecp256k1()
			publicKey, err := curve.PrivateKeyToPublicKey(privateKeys[0])
			if err != nil {
				return nil, fmt.Errorf("failed to generate change public key: %w", err)
			}

			changeAddr := address.NewP2PKHAddress(publicKey, false) // Assume mainnet
			err = tb.AddOutput(changeAddr.String(), change)
			if err != nil {
				return nil, fmt.Errorf("failed to add change output: %w", err)
			}
		}
	}

	// Build the transaction
	tx, err := tb.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Check if we have enough private keys
	if len(privateKeys) != len(tx.TxIns) {
		return nil, fmt.Errorf("number of private keys must match number of inputs")
	}

	// Sign each input
	for i, privateKey := range privateKeys {
		if i >= len(tb.utxos) {
			return nil, fmt.Errorf("missing UTXO for input %d", i)
		}

		utxo := tb.utxos[i]
		err := tx.SignInput(i, privateKey, utxo.ScriptPubKey, SighashAll)
		if err != nil {
			return nil, fmt.Errorf("failed to sign input %d: %w", i, err)
		}
	}

	return tx, nil
}

// CreateSimpleTransaction creates a simple P2PKH to P2PKH transaction
func CreateSimpleTransaction(
	fromPrivateKey *big.Int,
	fromUTXOs []*UTXO,
	toAddressStr string,
	amount uint64,
	feePerByte uint64,
	testnet bool,
) (*Transaction, error) {

	builder := NewTransactionBuilder()

	// Add inputs
	totalInput := uint64(0)
	for _, utxo := range fromUTXOs {
		builder.AddInput(utxo.TxHash, utxo.OutputIndex, utxo)
		totalInput += utxo.Amount
	}

	// Add main output
	err := builder.AddOutput(toAddressStr, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to add output: %w", err)
	}

	// Calculate fee
	estimatedFee := builder.CalculateFee(feePerByte)
	builder.SetFee(estimatedFee)

	// Add change output if needed
	change := totalInput - amount - estimatedFee
	if change > 0 {
		// Generate change address from the same private key
		curve := ecc.GetSecp256k1()
		publicKey, err := curve.PrivateKeyToPublicKey(fromPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to generate public key: %w", err)
		}

		changeAddr := address.NewP2PKHAddress(publicKey, testnet)
		err = builder.AddOutput(changeAddr.String(), change)
		if err != nil {
			return nil, fmt.Errorf("failed to add change output: %w", err)
		}
	}

	// Sign transaction
	privateKeys := make([]*big.Int, len(fromUTXOs))
	for i := range privateKeys {
		privateKeys[i] = fromPrivateKey
	}

	return builder.SignTransaction(privateKeys)
}

// CreateMultisigTransaction creates a multisig transaction
func CreateMultisigTransaction(
	privateKeys []*big.Int,
	fromUTXOs []*UTXO,
	toAddressStr string,
	amount uint64,
	feePerByte uint64,
) (*Transaction, error) {

	if len(privateKeys) != len(fromUTXOs) {
		return nil, fmt.Errorf("number of private keys must match number of UTXOs")
	}

	builder := NewTransactionBuilder()

	// Add inputs
	totalInput := uint64(0)
	for _, utxo := range fromUTXOs {
		builder.AddInput(utxo.TxHash, utxo.OutputIndex, utxo)
		totalInput += utxo.Amount
	}

	// Add output
	err := builder.AddOutput(toAddressStr, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to add output: %w", err)
	}

	// Calculate and set fee
	estimatedFee := builder.CalculateFee(feePerByte)
	builder.SetFee(estimatedFee)

	// Check if we have enough funds
	if totalInput < amount+estimatedFee {
		return nil, fmt.Errorf("insufficient funds")
	}

	// Add change output if needed
	change := totalInput - amount - estimatedFee
	if change > 0 {
		// Use the first private key for change address
		curve := ecc.GetSecp256k1()
		publicKey, err := curve.PrivateKeyToPublicKey(privateKeys[0])
		if err != nil {
			return nil, fmt.Errorf("failed to generate change public key: %w", err)
		}

		changeAddr := address.NewP2PKHAddress(publicKey, false) // Assume mainnet
		err = builder.AddOutput(changeAddr.String(), change)
		if err != nil {
			return nil, fmt.Errorf("failed to add change output: %w", err)
		}
	}

	// Sign transaction
	return builder.SignTransaction(privateKeys)
}

// ValidateTransaction validates a transaction
func ValidateTransaction(tx *Transaction, utxos []*UTXO) error {
	if len(tx.TxIns) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	if len(tx.TxOuts) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	if len(utxos) != len(tx.TxIns) {
		return fmt.Errorf("number of UTXOs must match number of inputs")
	}

	// Verify each input
	for i, txIn := range tx.TxIns {
		utxo := utxos[i]
		
		// Verify the input references the correct UTXO
		if !bytes.Equal(txIn.PrevTx, utxo.TxHash) || txIn.PrevIndex != utxo.OutputIndex {
			return fmt.Errorf("input %d references incorrect UTXO", i)
		}

		// Verify the signature
		valid, err := tx.VerifyInput(i, utxo.ScriptPubKey)
		if err != nil {
			return fmt.Errorf("failed to verify input %d: %w", i, err)
		}

		if !valid {
			return fmt.Errorf("invalid signature for input %d", i)
		}
	}

	return nil
}

// CalculateTransactionFee calculates the actual fee of a transaction
func CalculateTransactionFee(tx *Transaction, utxos []*UTXO) uint64 {
	if len(utxos) != len(tx.TxIns) {
		return 0
	}

	totalInput := uint64(0)
	for _, utxo := range utxos {
		totalInput += utxo.Amount
	}

	totalOutput := uint64(0)
	for _, txOut := range tx.TxOuts {
		totalOutput += txOut.Amount
	}

	if totalInput > totalOutput {
		return totalInput - totalOutput
	}

	return 0
}
