package types

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"time"
)

// BlockHeader represents the header of a block
type BlockHeader struct {
	PrevHash    string    `json:"prevHash"`
	MerkleRoot  string    `json:"merkleRoot"`
	StateRoot   string    `json:"stateRoot"`
	Timestamp   time.Time `json:"timestamp"`
	Number      *big.Int  `json:"number"`
	Difficulty  *big.Int  `json:"difficulty"`
	GasLimit    uint64    `json:"gasLimit"`
	GasUsed     uint64    `json:"gasUsed"`
	Validator   string    `json:"validator"`
	Nonce       uint64    `json:"nonce"`
	Hash        string    `json:"hash"`
}

// Block represents a blockchain block
type Block struct {
	Header       *BlockHeader   `json:"header"`
	Transactions []*Transaction `json:"transactions"`
	Size         int            `json:"size"`
	TxCount      int            `json:"txCount"`
}

// NewBlock creates a new block
func NewBlock(prevHash string, transactions []*Transaction, validator string, number *big.Int) *Block {
	header := &BlockHeader{
		PrevHash:   prevHash,
		Timestamp:  time.Now(),
		Number:     number,
		Difficulty: big.NewInt(1),
		GasLimit:   8000000,
		GasUsed:    0,
		Validator:  validator,
		Nonce:      0,
		StateRoot:  "",
	}

	block := &Block{
		Header:       header,
		Transactions: transactions,
		TxCount:      len(transactions),
	}

	block.Header.MerkleRoot = block.ComputeMerkleRoot()
	block.Header.GasUsed = block.ComputeGasUsed()
	block.Header.Hash = block.ComputeHash()
	
	return block
}

// ComputeHash computes the hash of the block
func (b *Block) ComputeHash() string {
	data := b.Header.PrevHash + b.Header.MerkleRoot + b.Header.StateRoot +
		   b.Header.Timestamp.String() + b.Header.Number.String() +
		   b.Header.Difficulty.String() + string(rune(b.Header.GasLimit)) +
		   string(rune(b.Header.GasUsed)) + b.Header.Validator +
		   string(rune(b.Header.Nonce))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ComputeMerkleRoot computes the Merkle root of all transactions
func (b *Block) ComputeMerkleRoot() string {
	if len(b.Transactions) == 0 {
		return ""
	}
	
	var hashes []string
	for _, tx := range b.Transactions {
		hashes = append(hashes, tx.Hash)
	}
	
	return computeMerkleRootFromHashes(hashes)
}

// Helper function to compute Merkle root from transaction hashes
func computeMerkleRootFromHashes(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	
	if len(hashes) == 1 {
		return hashes[0]
	}
	
	var nextLevel []string
	for i := 0; i < len(hashes); i += 2 {
		var combined string
		if i+1 < len(hashes) {
			combined = hashes[i] + hashes[i+1]
		} else {
			combined = hashes[i] + hashes[i]
		}
		hash := sha256.Sum256([]byte(combined))
		nextLevel = append(nextLevel, hex.EncodeToString(hash[:]))
	}
	
	return computeMerkleRootFromHashes(nextLevel)
}

// ComputeGasUsed calculates total gas used by all transactions
func (b *Block) ComputeGasUsed() uint64 {
	var totalGas uint64
	for _, tx := range b.Transactions {
		totalGas += tx.Gas
	}
	return totalGas
}

// IsValid checks if the block is valid
func (b *Block) IsValid() bool {
	if b.Header == nil {
		return false
	}
	
	if b.Header.Hash != b.ComputeHash() {
		return false
	}
	
	if b.Header.MerkleRoot != b.ComputeMerkleRoot() {
		return false
	}
	
	if b.Header.GasUsed != b.ComputeGasUsed() {
		return false
	}
	
	if b.Header.GasUsed > b.Header.GasLimit {
		return false
	}
	
	for _, tx := range b.Transactions {
		if !tx.IsValid() {
			return false
		}
	}
	
	return true
}

// GetSize calculates the size of the block in bytes
func (b *Block) GetSize() int {
	size := 0
	size += len(b.Header.PrevHash)
	size += len(b.Header.MerkleRoot)
	size += len(b.Header.StateRoot)
	size += len(b.Header.Hash)
	size += len(b.Header.Validator)
	size += 8 * 6 // timestamps, numbers, etc.
	
	for _, tx := range b.Transactions {
		size += len(tx.Hash)
		size += len(tx.From)
		size += len(tx.To)
		size += len(tx.Data)
		size += len(tx.Signature)
		size += 8 * 3 // gas, nonce, timestamp
	}
	
	b.Size = size
	return size
}