package types

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// Address represents a blockchain address
type Address struct {
	Value string `json:"value"`
}

// NewAddress creates a new address
func NewAddress(value string) *Address {
	return &Address{
		Value: strings.ToLower(value),
	}
}

// NewAddressFromPublicKey creates an address from a public key
func NewAddressFromPublicKey(publicKey []byte) *Address {
	hash := sha256.Sum256(publicKey)
	addressBytes := hash[len(hash)-20:]
	return &Address{
		Value: "0x" + hex.EncodeToString(addressBytes),
	}
}

// String returns the string representation of the address
func (a *Address) String() string {
	return a.Value
}

// IsValid checks if the address format is valid
func (a *Address) IsValid() bool {
	if a.Value == "" {
		return false
	}
	
	// Check if it's a valid hex address (0x + 40 hex chars)
	pattern := `^0x[a-fA-F0-9]{40}$`
	matched, _ := regexp.MatchString(pattern, a.Value)
	return matched
}

// Bytes returns the address as bytes
func (a *Address) Bytes() []byte {
	if len(a.Value) < 2 {
		return nil
	}
	
	hexStr := a.Value[2:] // Remove "0x" prefix
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil
	}
	
	return bytes
}

// Equal checks if two addresses are equal
func (a *Address) Equal(other *Address) bool {
	return strings.EqualFold(a.Value, other.Value)
}

// IsZero checks if the address is a zero address
func (a *Address) IsZero() bool {
	return a.Value == "0x0000000000000000000000000000000000000000" || a.Value == ""
}