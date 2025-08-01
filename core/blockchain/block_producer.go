package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"
)

type BlockProducer struct {
	blockchain *Blockchain
	txPool     *TxPool
	processor  *StateProcessor
}

// ProduceBlock produce new block
func (bp *BlockProducer) ProduceBlock() (*types.Block, error) {
	parent := bp.blockchain.CurrentBlock()

	// 1. create state copy
	stateDB := bp.blockchain.CurrentState().Copy()

	// 2. get pending txns
	pendingTxs := bp.txPool.GetPending()

	newBlockGasLimit := bp.calculateGasLimit(parent)

	// 3. exec txn
	var (
		includedTxs       []*types.Transaction
		gasUsedList       []*big.Int
		cumulativeGasUsed = big.NewInt(0)
	)

	for _, tx := range pendingTxs {
		// check block gas limit
		if new(big.Int).Add(cumulativeGasUsed, tx.GasCost()).Cmp(newBlockGasLimit) > 0 {
			break
		}

		// exc txn
		gasUsed, err := bp.processor.ExecuteTransaction(stateDB, tx)
		if err != nil {
			// skip failed
			log.Warn("Failed to execute transaction", "err", err)
			continue
		}

		includedTxs = append(includedTxs, tx)
		gasUsedList = append(gasUsedList, gasUsed)
		cumulativeGasUsed = new(big.Int).Add(cumulativeGasUsed, tx.GasCost())
	}

	// 4. calculate state root hash
	stateRoot := stateDB.ComputeRoot()

	// 5. create new block header
	header := &types.Header{
		ParentHash:    parent.Hash(),
		Number:        new(big.Int).Add(parent.Number(), big.NewInt(1)),
		Timestamp:     uint64(time.Now().Unix()),
		StateHashRoot: stateRoot,
		GasLimit:      newBlockGasLimit,
		GasUsed:       cumulativeGasUsed,
	}

	// 6. create receipts
	receipts := make([]*types.Receipt, len(includedTxs))
	cumGas := big.NewInt(0)
	for i, tx := range includedTxs {
		cumGas = new(big.Int).Add(cumGas, gasUsedList[i])
		receipts[i] = bp.processor.CreateReceipt(
			i,
			tx,
			gasUsedList[i],
			cumGas,
			stateRoot,
			true,
		)
	}

	// 7. create final new block
	block := types.NewBlock(header, includedTxs, receipts)

	return block, nil
}

func (bp *BlockProducer) calculateGasLimit(parent *types.Block) *big.Int {
	// TODO: simplify: using default max value
	return new(big.Int).SetUint64(common.BlockMaxGasLimit)
}
