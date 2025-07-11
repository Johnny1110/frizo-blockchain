package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"frizo-blockchain/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// CreateWallet create a new walletAddress and private-key
func CreateWallet() (common.Address, string, error) {
	// 1. create private
	privKey, err := GenerateKey()
	if err != nil {
		return common.Address{}, "", err
	}

	pubKey := privKey.PublicKey

	// 2. generate address
	addr := PubkeyToAddress(&pubKey)
	privKeyStr := ExportPrivateKey(privKey)

	return addr, privKeyStr, nil
}

// GenerateKey generates a new private key.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

// PubkeyToAddress converts a public key to an Ethereum address.
func PubkeyToAddress(pubkey *ecdsa.PublicKey) common.Address {
	// concat pub-key x,y
	pubBytes := FromECDSAPub(pubkey)
	pubBytesWithoutPrefix := pubBytes[1:]
	// remove prefix (04) and hash
	hashPubKey := Keccak256(pubBytesWithoutPrefix)
	// using last 20 bytes as address
	return common.BytesToAddress(hashPubKey[12:])
}

// FromECDSAPub exports a public key into a binary format.
func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	// concat public-key x,y
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

func ImportPrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func ExportPrivateKey(privateKey *ecdsa.PrivateKey) string {
	return hex.EncodeToString(crypto.FromECDSA(privateKey))
}
