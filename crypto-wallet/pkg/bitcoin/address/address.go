package address

import (
	"fmt"

	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/base58"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/script"
	"github.com/DimaJoyti/go-coffee/crypto-wallet/pkg/bitcoin/sec"
)

// Address types
const (
	P2PKH = "P2PKH" // Pay-to-Public-Key-Hash
	P2SH  = "P2SH"  // Pay-to-Script-Hash
	P2PK  = "P2PK"  // Pay-to-Public-Key (legacy)
)

// Address represents a Bitcoin address
type Address struct {
	Type    string // Address type (P2PKH, P2SH, P2PK)
	Hash    []byte // Hash160 for P2PKH/P2SH, or public key for P2PK
	Network string // "mainnet" or "testnet"
}

// NewP2PKHAddress creates a new P2PKH address from a public key
func NewP2PKHAddress(publicKey *ecc.Point, testnet bool) *Address {
	hash160 := script.PublicKeyToHash160(publicKey)
	
	network := "mainnet"
	if testnet {
		network = "testnet"
	}

	return &Address{
		Type:    P2PKH,
		Hash:    hash160,
		Network: network,
	}
}

// NewP2SHAddress creates a new P2SH address from a script hash
func NewP2SHAddress(scriptHash []byte, testnet bool) *Address {
	if len(scriptHash) != 20 {
		return nil
	}

	network := "mainnet"
	if testnet {
		network = "testnet"
	}

	return &Address{
		Type:    P2SH,
		Hash:    scriptHash,
		Network: network,
	}
}

// NewP2PKAddress creates a new P2PK address from a public key
func NewP2PKAddress(publicKey *ecc.Point, testnet bool) *Address {
	pubKeyBytes := sec.EncodePublicKeyCompressed(publicKey)
	
	network := "mainnet"
	if testnet {
		network = "testnet"
	}

	return &Address{
		Type:    P2PK,
		Hash:    pubKeyBytes,
		Network: network,
	}
}

// String returns the Base58Check encoded address string
func (addr *Address) String() string {
	switch addr.Type {
	case P2PKH:
		if addr.Network == "testnet" {
			return base58.EncodeAddress(addr.Hash, base58.TestnetP2PKHVersion)
		}
		return base58.EncodeAddress(addr.Hash, base58.MainnetP2PKHVersion)
	case P2SH:
		if addr.Network == "testnet" {
			return base58.EncodeAddress(addr.Hash, base58.TestnetP2SHVersion)
		}
		return base58.EncodeAddress(addr.Hash, base58.MainnetP2SHVersion)
	case P2PK:
		// P2PK doesn't have a standard address format, return hex
		return fmt.Sprintf("P2PK:%x", addr.Hash)
	default:
		return ""
	}
}

// ScriptPubKey returns the script public key for this address
func (addr *Address) ScriptPubKey() []byte {
	switch addr.Type {
	case P2PKH:
		scriptObj := script.CreateP2PKHScript(addr.Hash)
		return scriptObj.Serialize()
	case P2SH:
		// P2SH script: OP_HASH160 <scriptHash> OP_EQUAL
		scriptObj := script.NewScript([]interface{}{
			script.OP_HASH160,
			addr.Hash,
			script.OP_EQUAL,
		})
		return scriptObj.Serialize()
	case P2PK:
		// P2PK script: <publicKey> OP_CHECKSIG
		scriptObj := script.NewScript([]interface{}{
			addr.Hash, // Public key bytes
			script.OP_CHECKSIG,
		})
		return scriptObj.Serialize()
	default:
		return nil
	}
}

// ParseAddress parses a Bitcoin address string
func ParseAddress(addressStr string) (*Address, error) {
	// Try to decode as Base58Check
	hash160, version, err := base58.DecodeAddress(addressStr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	var addrType, network string

	switch version {
	case base58.MainnetP2PKHVersion:
		addrType = P2PKH
		network = "mainnet"
	case base58.MainnetP2SHVersion:
		addrType = P2SH
		network = "mainnet"
	case base58.TestnetP2PKHVersion:
		addrType = P2PKH
		network = "testnet"
	case base58.TestnetP2SHVersion:
		addrType = P2SH
		network = "testnet"
	default:
		return nil, fmt.Errorf("unknown address version: 0x%02x", version)
	}

	return &Address{
		Type:    addrType,
		Hash:    hash160,
		Network: network,
	}, nil
}

// IsValid checks if an address string is valid
func IsValid(addressStr string) bool {
	_, err := ParseAddress(addressStr)
	return err == nil
}

// IsMainnet checks if the address is for mainnet
func (addr *Address) IsMainnet() bool {
	return addr.Network == "mainnet"
}

// IsTestnet checks if the address is for testnet
func (addr *Address) IsTestnet() bool {
	return addr.Network == "testnet"
}

// Equal checks if two addresses are equal
func (addr *Address) Equal(other *Address) bool {
	if addr.Type != other.Type || addr.Network != other.Network {
		return false
	}

	if len(addr.Hash) != len(other.Hash) {
		return false
	}

	for i := range addr.Hash {
		if addr.Hash[i] != other.Hash[i] {
			return false
		}
	}

	return true
}

// PublicKeyToP2PKHAddress converts a public key to a P2PKH address
func PublicKeyToP2PKHAddress(publicKey *ecc.Point, testnet bool) string {
	addr := NewP2PKHAddress(publicKey, testnet)
	return addr.String()
}

// PrivateKeyToP2PKHAddress converts a private key to a P2PKH address
func PrivateKeyToP2PKHAddress(privateKey []byte, testnet bool) (string, error) {
	// Decode private key
	privKey, err := sec.DecodePrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Generate public key
	curve := ecc.GetSecp256k1()
	publicKey, err := curve.PrivateKeyToPublicKey(privKey)
	if err != nil {
		return "", fmt.Errorf("failed to generate public key: %w", err)
	}

	// Create address
	addr := NewP2PKHAddress(publicKey, testnet)
	return addr.String(), nil
}

// WIFToP2PKHAddress converts a WIF private key to a P2PKH address
func WIFToP2PKHAddress(wif string) (string, error) {
	// Decode WIF
	privateKeyBytes, _, testnet, err := base58.DecodeWIF(wif)
	if err != nil {
		return "", fmt.Errorf("invalid WIF: %w", err)
	}

	// Convert to address
	return PrivateKeyToP2PKHAddress(privateKeyBytes, testnet)
}

// CreateMultisigAddress creates a multisig P2SH address
func CreateMultisigAddress(publicKeys []*ecc.Point, threshold int, testnet bool) (*Address, error) {
	if threshold < 1 || threshold > len(publicKeys) {
		return nil, fmt.Errorf("invalid threshold: %d (must be between 1 and %d)", threshold, len(publicKeys))
	}

	if len(publicKeys) > 16 {
		return nil, fmt.Errorf("too many public keys: %d (maximum 16)", len(publicKeys))
	}

	// Create multisig script
	var commands []interface{}

	// Add threshold
	if threshold == 1 {
		commands = append(commands, script.OP_1)
	} else {
		commands = append(commands, script.OP_1+threshold-1)
	}

	// Add public keys
	for _, pubKey := range publicKeys {
		pubKeyBytes := sec.EncodePublicKeyCompressed(pubKey)
		commands = append(commands, pubKeyBytes)
	}

	// Add number of public keys
	numKeys := len(publicKeys)
	if numKeys == 1 {
		commands = append(commands, script.OP_1)
	} else {
		commands = append(commands, script.OP_1+numKeys-1)
	}

	// Add OP_CHECKMULTISIG
	commands = append(commands, script.OP_CHECKMULTISIG)

	// Create script and calculate hash
	multisigScript := script.NewScript(commands)
	scriptBytes := multisigScript.Serialize()
	scriptHash := script.Hash160(scriptBytes)

	return NewP2SHAddress(scriptHash, testnet), nil
}

// GetAddressType returns the type of a Bitcoin address string
func GetAddressType(addressStr string) string {
	addr, err := ParseAddress(addressStr)
	if err != nil {
		return "invalid"
	}
	return addr.Type
}

// GetAddressNetwork returns the network of a Bitcoin address string
func GetAddressNetwork(addressStr string) string {
	addr, err := ParseAddress(addressStr)
	if err != nil {
		return "invalid"
	}
	return addr.Network
}

// ValidateAddressForNetwork checks if an address is valid for a specific network
func ValidateAddressForNetwork(addressStr, network string) bool {
	addr, err := ParseAddress(addressStr)
	if err != nil {
		return false
	}
	return addr.Network == network
}

// ConvertAddressNetwork converts an address between mainnet and testnet
func ConvertAddressNetwork(addressStr string, toTestnet bool) (string, error) {
	addr, err := ParseAddress(addressStr)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}

	// Update network
	if toTestnet {
		addr.Network = "testnet"
	} else {
		addr.Network = "mainnet"
	}

	return addr.String(), nil
}
