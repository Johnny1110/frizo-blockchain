package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
	"math/big"
)

// StateProcessor state transform
type StateProcessor struct {
	bc *Blockchain
}

// ExecuteTransaction exec txn
func (sp *StateProcessor) ExecuteTransaction(
	stateDB *state.SimpleStateDB,
	tx *types.Transaction) (*big.Int, error) {
	// 1. get from
	from, err := tx.From()
	if err != nil {
		return big.NewInt(0), err
	}

	// 2. check nonce
	if tx.Nonce() != stateDB.GetNonce(from) {
		return big.NewInt(0), common.ErrNonceMismatch
	}

	// 3. check from balance
	totalCost := new(big.Int).Add(tx.Value(), tx.TotalCost())
	if stateDB.GetBalance(from).Cmp(totalCost) < 0 {
		return big.NewInt(0), common.ErrInsufficientFunds
	}

	// 4. create snapshot
	snapshot := stateDB.Snapshot()

	// 5. - gas fee
	if err := stateDB.SubBalance(from, tx.GasCost()); err != nil {
		stateDB.RevertToSnapshot(snapshot)
		return big.NewInt(0), err
	}

	// 6. do transfer value
	if tx.To() != nil {
		if err := stateDB.Transfer(from, *tx.To(), tx.Value()); err != nil {
			stateDB.RevertToSnapshot(snapshot)
			return big.NewInt(0), err
		}
	}

	// 7. increase nonce
	stateDB.SetNonce(from, stateDB.GetNonce(from)+1)

	// return used gas
	return tx.GasCost(), nil
}

func (sp *StateProcessor) CreateReceipt(
	index int,
	tx *types.Transaction,
	gasUsed *big.Int,
	cumulativeGasUsed *big.Int,
	stateRoot common.Hash,
	success bool,
) *types.Receipt {
	status := uint64(1)
	if !success {
		status = 0
	}

	return &types.Receipt{
		PostState:         stateRoot,
		Status:            status,
		CumulativeGasUsed: cumulativeGasUsed,
		TxHash:            tx.Hash(),
		GasUsed:           gasUsed,
		TransactionIndex:  uint(index),
	}
}
