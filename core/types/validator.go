package types

import "frizo-blockchain/common"

// BlockValidator Block validator interface
type BlockValidator interface {
	// ValidateBody body（txn, etc）
	ValidateBody(block *Block) error

	// ValidateState validate state transform
	ValidateState(block *Block, state common.State, receipts []*Receipt) error
}
