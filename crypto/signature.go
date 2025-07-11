package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"math/big"
)

// TODO: 公私鑰的產生需要補學術知識，然後回來重寫一次

// Signature represents a digital signature
type Signature struct {
	R, S *big.Int
	V    byte
}

// SignatureLength is the expected length of a signature R(32) + S(32) + V(1)
const SignatureLength = 65

// RecoveryIDOffset is added to the recovery ID to conform to Ethereum's signature format
const RecoveryIDOffset = 27 // 以太坊簽章 V 是 27/28

var (
	// 曲線階 (order)
	secp256k1N = new(big.Int).SetBytes(hexToBytes("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141"))
	// N/2，用於 low‑s 檢查
	secp256k1H = new(big.Int).SetBytes(hexToBytes("7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0"))
)

// GenerateKey generates a new private key.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(S256(), rand.Reader)
}

// Sign calculates an ECDSA signature.
// The produced signature is in the [R || S || V] format where V is 0 or 1.
func Sign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be exactly 32 bytes (%d)", len(hash))
	}

	r, s, err := ecdsa.Sign(rand.Reader, prv, hash)
	if err != nil {
		return nil, err
	}

	// Serialize signature
	sig := make([]byte, SignatureLength)
	copy(sig[0:32], r.Bytes())
	copy(sig[32:64], s.Bytes())

	// Calculate V (recovery ID)
	// This is simplified - in production, you'd need to determine the correct recovery ID
	sig[64] = 0 // or 1, depending on the recovery process

	return sig, nil
}

// VerifySignature checks that the given public key created the signature over hash.
func VerifySignature(pubkey, hash, signature []byte) bool {
	if len(signature) != SignatureLength {
		return false
	}

	// Parse public key
	x, y := elliptic.Unmarshal(S256(), pubkey)
	if x == nil {
		return false
	}

	// Parse signature
	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])

	// Verify
	return ecdsa.Verify(&ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, hash, r, s)
}

// Ecrecover(反推公鑰) returns the uncompressed public key that created the given signature.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	if len(sig) != SignatureLength {
		return nil, errors.New("invalid signature length")
	}

	// This is a simplified implementation
	// In production, you'd implement the full ECDSA recovery algorithm
	// For now, we'll return an error indicating it's not implemented
	return nil, errors.New("ecrecover not implemented in this phase")
}

// PubkeyToAddress converts a public key to an Ethereum address.
func PubkeyToAddress(pubkey ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&pubkey)
	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

// FromECDSA exports a private key into a binary format.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return priv.D.Bytes()
}

// ToECDSA creates a private key from a binary representation.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if 8*len(d) != priv.Params().BitSize {
		return nil, errors.New("invalid private key length")
	}
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

// FromECDSAPub exports a public key into a binary format.
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

// ToECDSAPub creates a public key from a binary representation.
func ToECDSAPub(pub []byte) (*ecdsa.PublicKey, error) {
	if len(pub) == 0 {
		return nil, errors.New("invalid public key")
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, errors.New("invalid public key")
	}
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

// S256 returns an instance of the secp256k1 curve.
func S256() elliptic.Curve {
	// In production, you'd use a proper secp256k1 implementation
	// For now, we'll use P256 as a placeholder
	return elliptic.P256()
}

// ValidateSignatureValues verifies whether the signature values are within the allowed range.
func ValidateSignatureValues(v byte, r, s *big.Int) bool {
	if r.Cmp(big.NewInt(1)) < 0 || r.Cmp(secp256k1N) >= 0 {
		return false
	}
	if s.Cmp(big.NewInt(1)) < 0 || s.Cmp(secp256k1H) > 0 {
		return false
	}
	if v != 0 && v != 1 {
		return false
	}
	return true
}

// hexToBytes converts a hex string to bytes
func hexToBytes(hexStr string) []byte {
	bytes := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		fmt.Sscanf(hexStr[i:i+2], "%x", &bytes[i/2])
	}
	return bytes
}
