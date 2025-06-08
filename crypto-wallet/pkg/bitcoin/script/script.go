package script

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin/ecc"
	"github.com/DimaJoyti/go-coffee/web3-wallet-backend/pkg/bitcoin/sec"
	"golang.org/x/crypto/ripemd160"
)

// Bitcoin Script opcodes
const (
	// Constants
	OP_0         = 0x00
	OP_FALSE     = OP_0
	OP_PUSHDATA1 = 0x4c
	OP_PUSHDATA2 = 0x4d
	OP_PUSHDATA4 = 0x4e
	OP_1NEGATE   = 0x4f
	OP_1         = 0x51
	OP_TRUE      = OP_1
	OP_2         = 0x52
	OP_3         = 0x53
	OP_4         = 0x54
	OP_5         = 0x55
	OP_6         = 0x56
	OP_7         = 0x57
	OP_8         = 0x58
	OP_9         = 0x59
	OP_10        = 0x5a
	OP_11        = 0x5b
	OP_12        = 0x5c
	OP_13        = 0x5d
	OP_14        = 0x5e
	OP_15        = 0x5f
	OP_16        = 0x60

	// Flow control
	OP_NOP      = 0x61
	OP_IF       = 0x63
	OP_NOTIF    = 0x64
	OP_ELSE     = 0x67
	OP_ENDIF    = 0x68
	OP_VERIFY   = 0x69
	OP_RETURN   = 0x6a

	// Stack
	OP_TOALTSTACK   = 0x6b
	OP_FROMALTSTACK = 0x6c
	OP_IFDUP        = 0x73
	OP_DEPTH        = 0x74
	OP_DROP         = 0x75
	OP_DUP          = 0x76
	OP_NIP          = 0x77
	OP_OVER         = 0x78
	OP_PICK         = 0x79
	OP_ROLL         = 0x7a
	OP_ROT          = 0x7b
	OP_SWAP         = 0x7c
	OP_TUCK         = 0x7d
	OP_2DROP        = 0x6d
	OP_2DUP         = 0x6e
	OP_3DUP         = 0x6f
	OP_2OVER        = 0x70
	OP_2ROT         = 0x71
	OP_2SWAP        = 0x72

	// Splice
	OP_SIZE = 0x82

	// Bitwise logic
	OP_EQUAL       = 0x87
	OP_EQUALVERIFY = 0x88

	// Arithmetic
	OP_1ADD      = 0x8b
	OP_1SUB      = 0x8c
	OP_NEGATE    = 0x8f
	OP_ABS       = 0x90
	OP_NOT       = 0x91
	OP_0NOTEQUAL = 0x92
	OP_ADD       = 0x93
	OP_SUB       = 0x94
	OP_BOOLAND   = 0x9a
	OP_BOOLOR    = 0x9b
	OP_NUMEQUAL  = 0x9c
	OP_NUMEQUALVERIFY = 0x9d
	OP_NUMNOTEQUAL    = 0x9e
	OP_LESSTHAN       = 0x9f
	OP_GREATERTHAN    = 0xa0
	OP_LESSTHANOREQUAL    = 0xa1
	OP_GREATERTHANOREQUAL = 0xa2
	OP_MIN = 0xa3
	OP_MAX = 0xa4
	OP_WITHIN = 0xa5

	// Crypto
	OP_RIPEMD160           = 0xa6
	OP_SHA1                = 0xa7
	OP_SHA256              = 0xa8
	OP_HASH160             = 0xa9
	OP_HASH256             = 0xaa
	OP_CODESEPARATOR       = 0xab
	OP_CHECKSIG            = 0xac
	OP_CHECKSIGVERIFY      = 0xad
	OP_CHECKMULTISIG       = 0xae
	OP_CHECKMULTISIGVERIFY = 0xaf

	// Locktime
	OP_CHECKLOCKTIMEVERIFY = 0xb1
	OP_CHECKSEQUENCEVERIFY = 0xb2
)

// Script represents a Bitcoin script
type Script struct {
	Commands []interface{} // Can contain opcodes (int) or data ([]byte)
}

// NewScript creates a new script
func NewScript(commands []interface{}) *Script {
	return &Script{Commands: commands}
}

// Serialize serializes the script to bytes
func (s *Script) Serialize() []byte {
	var buf bytes.Buffer

	for _, cmd := range s.Commands {
		switch v := cmd.(type) {
		case int:
			// Opcode
			buf.WriteByte(byte(v))
		case []byte:
			// Data
			length := len(v)
			if length <= 75 {
				// Direct push
				buf.WriteByte(byte(length))
				buf.Write(v)
			} else if length <= 255 {
				// OP_PUSHDATA1
				buf.WriteByte(OP_PUSHDATA1)
				buf.WriteByte(byte(length))
				buf.Write(v)
			} else if length <= 65535 {
				// OP_PUSHDATA2
				buf.WriteByte(OP_PUSHDATA2)
				buf.WriteByte(byte(length & 0xff))
				buf.WriteByte(byte(length >> 8))
				buf.Write(v)
			} else {
				// OP_PUSHDATA4
				buf.WriteByte(OP_PUSHDATA4)
				buf.WriteByte(byte(length & 0xff))
				buf.WriteByte(byte((length >> 8) & 0xff))
				buf.WriteByte(byte((length >> 16) & 0xff))
				buf.WriteByte(byte(length >> 24))
				buf.Write(v)
			}
		}
	}

	return buf.Bytes()
}

// Parse parses a script from bytes
func Parse(scriptBytes []byte) (*Script, error) {
	var commands []interface{}
	i := 0

	for i < len(scriptBytes) {
		opcode := int(scriptBytes[i])
		i++

		if opcode >= 1 && opcode <= 75 {
			// Direct data push
			if i+opcode > len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			commands = append(commands, scriptBytes[i:i+opcode])
			i += opcode
		} else if opcode == OP_PUSHDATA1 {
			if i >= len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			length := int(scriptBytes[i])
			i++
			if i+length > len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			commands = append(commands, scriptBytes[i:i+length])
			i += length
		} else if opcode == OP_PUSHDATA2 {
			if i+2 > len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			length := int(scriptBytes[i]) | (int(scriptBytes[i+1]) << 8)
			i += 2
			if i+length > len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			commands = append(commands, scriptBytes[i:i+length])
			i += length
		} else if opcode == OP_PUSHDATA4 {
			if i+4 > len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			length := int(scriptBytes[i]) | (int(scriptBytes[i+1]) << 8) |
				(int(scriptBytes[i+2]) << 16) | (int(scriptBytes[i+3]) << 24)
			i += 4
			if i+length > len(scriptBytes) {
				return nil, fmt.Errorf("script truncated")
			}
			commands = append(commands, scriptBytes[i:i+length])
			i += length
		} else {
			// Regular opcode
			commands = append(commands, opcode)
		}
	}

	return NewScript(commands), nil
}

// CreateP2PKHScript creates a Pay-to-Public-Key-Hash script
func CreateP2PKHScript(hash160 []byte) *Script {
	if len(hash160) != 20 {
		return nil
	}

	commands := []interface{}{
		OP_DUP,
		OP_HASH160,
		hash160,
		OP_EQUALVERIFY,
		OP_CHECKSIG,
	}

	return NewScript(commands)
}

// CreateP2PKScript creates a Pay-to-Public-Key script
func CreateP2PKScript(publicKey *ecc.Point) *Script {
	pubKeyBytes := sec.EncodePublicKeyCompressed(publicKey)
	
	commands := []interface{}{
		pubKeyBytes,
		OP_CHECKSIG,
	}

	return NewScript(commands)
}

// CreateP2PKHScriptSig creates a script signature for P2PKH
func CreateP2PKHScriptSig(signature []byte, publicKey *ecc.Point) []byte {
	pubKeyBytes := sec.EncodePublicKeyCompressed(publicKey)
	
	commands := []interface{}{
		signature,
		pubKeyBytes,
	}

	script := NewScript(commands)
	return script.Serialize()
}

// ParseP2PKHScriptSig parses a P2PKH script signature
func ParseP2PKHScriptSig(scriptSig []byte) (*ecc.Signature, *ecc.Point, byte, error) {
	script, err := Parse(scriptSig)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to parse script: %w", err)
	}

	if len(script.Commands) != 2 {
		return nil, nil, 0, fmt.Errorf("invalid P2PKH script signature")
	}

	// Extract signature
	sigBytes, ok := script.Commands[0].([]byte)
	if !ok {
		return nil, nil, 0, fmt.Errorf("invalid signature in script")
	}

	if len(sigBytes) == 0 {
		return nil, nil, 0, fmt.Errorf("empty signature")
	}

	// Extract sighash type (last byte)
	sighashType := sigBytes[len(sigBytes)-1]
	derSig := sigBytes[:len(sigBytes)-1]

	// Parse DER signature
	signature, err := parseDERSignature(derSig)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to parse DER signature: %w", err)
	}

	// Extract public key
	pubKeyBytes, ok := script.Commands[1].([]byte)
	if !ok {
		return nil, nil, 0, fmt.Errorf("invalid public key in script")
	}

	publicKey, err := sec.DecodePublicKey(pubKeyBytes)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to decode public key: %w", err)
	}

	return signature, publicKey, sighashType, nil
}

// parseDERSignature parses a DER-encoded signature
func parseDERSignature(der []byte) (*ecc.Signature, error) {
	if len(der) < 6 {
		return nil, fmt.Errorf("DER signature too short")
	}

	if der[0] != 0x30 {
		return nil, fmt.Errorf("invalid DER signature: missing SEQUENCE tag")
	}

	length := int(der[1])
	if length+2 != len(der) {
		return nil, fmt.Errorf("invalid DER signature: incorrect length")
	}

	// Parse r
	if der[2] != 0x02 {
		return nil, fmt.Errorf("invalid DER signature: missing r INTEGER tag")
	}

	rLength := int(der[3])
	if 4+rLength >= len(der) {
		return nil, fmt.Errorf("invalid DER signature: r length exceeds data")
	}

	r := new(big.Int).SetBytes(der[4 : 4+rLength])

	// Parse s
	sOffset := 4 + rLength
	if der[sOffset] != 0x02 {
		return nil, fmt.Errorf("invalid DER signature: missing s INTEGER tag")
	}

	sLength := int(der[sOffset+1])
	if sOffset+2+sLength != len(der) {
		return nil, fmt.Errorf("invalid DER signature: s length incorrect")
	}

	s := new(big.Int).SetBytes(der[sOffset+2 : sOffset+2+sLength])

	return ecc.NewSignature(r, s), nil
}

// Hash160 calculates RIPEMD160(SHA256(data))
func Hash160(data []byte) []byte {
	sha := sha256.Sum256(data)
	ripemd := ripemd160.New()
	ripemd.Write(sha[:])
	return ripemd.Sum(nil)
}

// Hash256 calculates SHA256(SHA256(data))
func Hash256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

// PublicKeyToHash160 converts a public key to its hash160
func PublicKeyToHash160(publicKey *ecc.Point) []byte {
	pubKeyBytes := sec.EncodePublicKeyCompressed(publicKey)
	return Hash160(pubKeyBytes)
}

// IsP2PKH checks if a script is a Pay-to-Public-Key-Hash script
func (s *Script) IsP2PKH() bool {
	return len(s.Commands) == 5 &&
		s.Commands[0] == OP_DUP &&
		s.Commands[1] == OP_HASH160 &&
		len(s.Commands[2].([]byte)) == 20 &&
		s.Commands[3] == OP_EQUALVERIFY &&
		s.Commands[4] == OP_CHECKSIG
}

// IsP2PK checks if a script is a Pay-to-Public-Key script
func (s *Script) IsP2PK() bool {
	return len(s.Commands) == 2 &&
		s.Commands[1] == OP_CHECKSIG
}

// GetP2PKHHash160 extracts the hash160 from a P2PKH script
func (s *Script) GetP2PKHHash160() []byte {
	if !s.IsP2PKH() {
		return nil
	}
	return s.Commands[2].([]byte)
}

// String returns a string representation of the script
func (s *Script) String() string {
	var parts []string
	for _, cmd := range s.Commands {
		switch v := cmd.(type) {
		case int:
			parts = append(parts, fmt.Sprintf("OP_%d", v))
		case []byte:
			parts = append(parts, fmt.Sprintf("%x", v))
		}
	}
	return fmt.Sprintf("Script(%s)", parts)
}
