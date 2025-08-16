package blockchain

import (
	"frizo-blockchain/common"
	"math/big"
	"time"
)

// ChainConfig Chain Config
type ChainConfig struct {
	ChainID *big.Int // Chain ID, for sign txn

	// Genesis
	GenesisHash   common.Hash
	ConsensusType string // "pos" or "poa"

	// Block
	BlockTime    time.Duration
	MaxBlockSize uint64
	MaxBlockGas  uint64

	// Transaction
	MinGasPrice *big.Int
	MaxTxSize   uint64

	// PoS setup
	MinStake          *big.Int // min stake value (32 or 128)
	ValidatorSetSize  uint64   // Validator Set Size
	BlockPeriod       uint64   // block creation interval（secs）
	EpochLength       uint64   // epoch length（block count）
	FinalizationDelay uint64   // finalized confirmation delay

	// Gas setup
	InitialGasLimit uint64 // init gas limit
}

// CacheConfig cache
type CacheConfig struct {
	HeaderCacheSize  int
	BlockCacheSize   int
	ReceiptCacheSize int
	StateCacheSize   int

	Pruning        bool
	PruningTimeout time.Duration
}

// DefaultCacheConfig default cache setup
var DefaultCacheConfig = &CacheConfig{
	HeaderCacheSize:  512,
	BlockCacheSize:   256,
	ReceiptCacheSize: 256,
	StateCacheSize:   128,
	Pruning:          false,
}
