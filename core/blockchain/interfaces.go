package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
)

// ChainReader define blockchain reader
type ChainReader interface {
	// Config
	Config() *common.ChainConfig

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
}
