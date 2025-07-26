package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"frizo-blockchain/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

// document: https://github.com/Johnny1110/frizo-blockchain/blob/develop/docs/basic/wallet.md

// Signature represents a digital signature
type Signature struct {
	R, S *big.Int
	V    byte
}

func NewSignature(s []byte) (Signature, error) {
	if len(s) < common.SignatureLen {
		return Signature{}, common.ErrInvalidSignature
	}
	return Signature{
		R: new(big.Int).SetBytes(s[:32]),
		S: new(big.Int).SetBytes(s[32:64]),
		V: s[64],
	}, nil
}

// Bytes returns the signature in RSV format: R (32 bytes) | S (32 bytes) | V (1 byte)
func (sign *Signature) HexStr() string {
	return hex.EncodeToString(sign.Bytes())
}

// Bytes returns the signature in RSV format: R (32 bytes) | S (32 bytes) | V (1 byte)
func (sign *Signature) Bytes() []byte {
	rBytes := sign.R.Bytes()
	sBytes := sign.S.Bytes()

	// Pad R and S to 32 bytes
	rPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)

	sPadded := make([]byte, 32)
	copy(sPadded[32-len(sBytes):], sBytes)

	// Combine R + S + V
	result := make([]byte, 65)
	copy(result[:32], rPadded)
	copy(result[32:64], sPadded)
	result[64] = sign.V

	return result
}

func (sign *Signature) Validate() bool {
	return len(sign.Bytes()) == common.SignatureLen
}

// SignMessage calculates an ECDSA signature.
// The produced signature is in the [R || S || V] format.
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) (Signature, error) {
	if len(message) != 32 {
		return Signature{}, common.ErrInvalidMessage
	}

	messageHash := crypto.Keccak256(message)
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		return Signature{}, common.ErrInvalidPrivateKey
	}

	return Signature{
		R: new(big.Int).SetBytes(signature[:32]),
		S: new(big.Int).SetBytes(signature[32:64]),
		V: signature[64],
	}, nil
}

// VerifyAndReturnAddress return address if verify success, otherwise return error
func VerifyAndReturnAddress(message []byte, signature Signature) (common.Address, error) {
	recoveredPubKey, err := Ecrecover(message, signature)
	if err != nil {
		return common.Address{}, err
	}

	ok := VerifySignature(recoveredPubKey, message, signature)
	if !ok {
		return common.Address{}, common.ErrInvalidSignature
	}

	return PubkeyToAddress(recoveredPubKey), nil
}

// VerifySignature checks that the given public key created the signature over hash.
func VerifySignature(publicKey *ecdsa.PublicKey, message []byte, signature Signature) bool {
	signBytes := signature.Bytes()
	if len(signBytes) != common.SignatureLen {
		return false
	}

	messageHash := crypto.Keccak256(message)
	signatureNoRecoveryID := signBytes[:common.SignatureLen-1]

	return crypto.VerifySignature(
		FromECDSAPub(publicKey),
		messageHash,
		signatureNoRecoveryID,
	)
}

// Ecrecover(反推公鑰) returns the uncompressed public key that created the given signature.
func Ecrecover(message []byte, signature Signature) (*ecdsa.PublicKey, error) {
	signBytes := signature.Bytes()
	if len(signBytes) != common.SignatureLen {
		return nil, common.ErrInvalidSignature
	}
	if len(message) > 32 {
		return nil, common.ErrInvalidMessage
	}

	messageHash := crypto.Keccak256(message)

	// recover
	publicKey, err := crypto.SigToPub(messageHash, signBytes)
	if err != nil {
		log.Warn("[crypto][Ecrecover] failed", err)
		return nil, common.ErrInvalidSignature
	}

	return publicKey, nil
}
