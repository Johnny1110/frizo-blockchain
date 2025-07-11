package types

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	Hash      string    `json:"hash"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Value     *big.Int  `json:"value"`
	Gas       uint64    `json:"gas"`
	GasPrice  *big.Int  `json:"gasPrice"`
	Nonce     uint64    `json:"nonce"`
	Data      []byte    `json:"data"`
	Signature []byte    `json:"signature"`
	Timestamp time.Time `json:"timestamp"`
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, value *big.Int, gas uint64, gasPrice *big.Int, nonce uint64, data []byte) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Value:     value,
		Gas:       gas,
		GasPrice:  gasPrice,
		Nonce:     nonce,
		Data:      data,
		Timestamp: time.Now(),
	}
	tx.Hash = tx.ComputeHash()
	return tx
}

// ComputeHash computes the hash of the transaction
func (tx *Transaction) ComputeHash() string {
	data := tx.From + tx.To + tx.Value.String() + string(rune(tx.Gas)) + 
		   tx.GasPrice.String() + string(rune(tx.Nonce)) + string(tx.Data)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// IsValid checks if the transaction is valid
func (tx *Transaction) IsValid() bool {
	if tx.From == "" || tx.To == "" {
		return false
	}
	if tx.Value == nil || tx.Value.Cmp(big.NewInt(0)) < 0 {
		return false
	}
	if tx.Gas == 0 || tx.GasPrice == nil || tx.GasPrice.Cmp(big.NewInt(0)) <= 0 {
		return false
	}
	return true
}

// GetFee calculates the transaction fee (gas * gasPrice)
func (tx *Transaction) GetFee() *big.Int {
	fee := big.NewInt(0)
	return fee.Mul(big.NewInt(int64(tx.Gas)), tx.GasPrice)
}

// Sign signs the transaction with the given signature
func (tx *Transaction) Sign(signature []byte) {
	tx.Signature = signature
	tx.Hash = tx.ComputeHash()
}