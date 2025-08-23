package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
	"math/big"
)

// IChainReader define blockchain reader
type IChainReader interface {
	// Config
	Config() *ChainConfig

	// Current State
	CurrentBlock() *types.Block
	CurrentHeader() *types.Header
	CurrentState() state.StateDB

	// Block retrieval by hash
	GetBlockByHash(hash common.Hash) *types.Block
	GetHeaderByHash(hash common.Hash) *types.Header

	// Block retrieval by block number
	GetBlockByNumber(number uint64) *types.Block
	GetHeaderByNumber(number uint64) *types.Header

	// Get block components
	GetBody(hash common.Hash) *types.Body
	GetReceipts(hash common.Hash) types.Receipts

	// Transaction retrieval
	GetTransaction(txHash common.Hash) (*types.Transaction, common.Hash, uint64, uint64)

	// Canonical chain
	GetCanonicalHash(number uint64) common.Hash
	HasBlock(hash common.Hash, number uint64) bool
}

// IChainWriter Define how to write into blockchain
type IChainWriter interface {
	// Block insertion
	InsertBlock(block *types.Block) error
	InsertBlocks(blocks []*types.Block) (int, error)
	InsertChain(blocks []*types.Block) (int, error)

	// Write components
	WriteBlock(block *types.Block) error
	WriteHeader(header *types.Header) error
	WriteReceipts(hash common.Hash, receipts types.Receipts) error

	// State operations
	CommitState(root common.Hash) error
	Rollback(target common.Hash) error

	// Set head
	SetHead(number uint64) error
}

// IChainProcessor process block and txn
type IChainProcessor interface {
	// Process block
	Process(block *types.Block, sdb state.StateDB) (types.Receipts, []*types.Log, uint64, error)

	// Validate
	ValidateBody(block *types.Block) error
	ValidateState(block *types.Block, sdb state.StateDB, receipts types.Receipts, usedGas uint64) error
}

// IChainStateReader chain state reader
type IChainStateReader interface {
	GetBalance(addr common.Address) *big.Int
	GetNonce(addr common.Address) uint64
	GetCode(addr common.Address) []byte
	GetState(addr common.Address, key common.Hash) common.Hash
	HasAccount(addr common.Address) bool
}

// TxPoolReader Txn Pool reader
type TxPoolReader interface {
	IChainStateReader
	CurrentBlock() *types.Block
	GetBlock(hash common.Hash, number uint64) *types.Block
}
