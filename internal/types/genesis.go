package types

import (
	"math/big"
	"time"
)

// Genesis represents the genesis block configuration
type Genesis struct {
	ChainID     *big.Int           `json:"chainId"`
	Timestamp   time.Time          `json:"timestamp"`
	GasLimit    uint64             `json:"gasLimit"`
	Difficulty  *big.Int           `json:"difficulty"`
	Alloc       map[string]*big.Int `json:"alloc"`
	Validators  []string           `json:"validators"`
	Config      *ChainConfig       `json:"config"`
}

// ChainConfig represents the chain configuration
type ChainConfig struct {
	ChainID        *big.Int `json:"chainId"`
	HomesteadBlock *big.Int `json:"homesteadBlock"`
	EIP150Block    *big.Int `json:"eip150Block"`
	EIP155Block    *big.Int `json:"eip155Block"`
	EIP158Block    *big.Int `json:"eip158Block"`
	ByzantiumBlock *big.Int `json:"byzantiumBlock"`
	PosBlock       *big.Int `json:"posBlock"`
}

// NewGenesis creates a new genesis configuration
func NewGenesis(chainID *big.Int) *Genesis {
	return &Genesis{
		ChainID:    chainID,
		Timestamp:  time.Now(),
		GasLimit:   8000000,
		Difficulty: big.NewInt(1),
		Alloc:      make(map[string]*big.Int),
		Validators: make([]string, 0),
		Config: &ChainConfig{
			ChainID:        chainID,
			HomesteadBlock: big.NewInt(0),
			EIP150Block:    big.NewInt(0),
			EIP155Block:    big.NewInt(0),
			EIP158Block:    big.NewInt(0),
			ByzantiumBlock: big.NewInt(0),
			PosBlock:       big.NewInt(0),
		},
	}
}

// AddAccount adds an account to the genesis allocation
func (g *Genesis) AddAccount(address string, balance *big.Int) {
	g.Alloc[address] = balance
}

// AddValidator adds a validator to the genesis validators
func (g *Genesis) AddValidator(validator string) {
	g.Validators = append(g.Validators, validator)
}

// ToBlock converts the genesis configuration to a genesis block
func (g *Genesis) ToBlock() *Block {
	header := &BlockHeader{
		PrevHash:   "0x0000000000000000000000000000000000000000000000000000000000000000",
		MerkleRoot: "",
		StateRoot:  "",
		Timestamp:  g.Timestamp,
		Number:     big.NewInt(0),
		Difficulty: g.Difficulty,
		GasLimit:   g.GasLimit,
		GasUsed:    0,
		Validator:  "",
		Nonce:      0,
	}

	block := &Block{
		Header:       header,
		Transactions: make([]*Transaction, 0),
		TxCount:      0,
	}

	block.Header.Hash = block.ComputeHash()
	return block
}

// DefaultGenesis returns a default genesis configuration
func DefaultGenesis() *Genesis {
	genesis := NewGenesis(big.NewInt(1337))
	
	// Add some initial accounts with balances
	genesis.AddAccount("0x1000000000000000000000000000000000000001", big.NewInt(1000000000000000000)) // 1 ETH
	genesis.AddAccount("0x2000000000000000000000000000000000000002", big.NewInt(2000000000000000000)) // 2 ETH
	genesis.AddAccount("0x3000000000000000000000000000000000000003", big.NewInt(3000000000000000000)) // 3 ETH
	
	// Add initial validators
	genesis.AddValidator("0x1000000000000000000000000000000000000001")
	genesis.AddValidator("0x2000000000000000000000000000000000000002")
	
	return genesis
}