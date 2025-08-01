package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"sync"
)

// TxPool txn pool
type TxPool struct {
	mu         sync.RWMutex
	pending    map[common.Address][]*types.Transaction
	queue      map[common.Address][]*types.Transaction
	blockchain *Blockchain
}

// NewTxPool create txn pool
func NewTxPool(blockchain *Blockchain) *TxPool {
	return &TxPool{
		pending:    make(map[common.Address][]*types.Transaction),
		queue:      make(map[common.Address][]*types.Transaction),
		blockchain: blockchain,
	}
}

// AddTx add txn
func (pool *TxPool) AddTx(tx *types.Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	from, err := tx.From()
	if err != nil {
		return err
	}

	// simplify：add into from address's pending list
	pool.pending[from] = append(pool.pending[from], tx)

	return nil
}

// GetPending get pending txn data
func (pool *TxPool) GetPending() []*types.Transaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	var txs []*types.Transaction
	for _, list := range pool.pending {
		txs = append(txs, list...)
	}
	return txs
}
