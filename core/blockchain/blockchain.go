package blockchain

import (
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
	"frizo-blockchain/storage"
	"math/big"
	"sync"
)

type Blockchain struct {
	mu sync.RWMutex
	db storage.Database

	currentBlock *types.Block
	currentState *state.SimpleStateDB

	// cache:
	blockCache  map[common.Hash]*types.Block
	numberCache map[*big.Int]common.Hash

	genesis *Genesis
}

// Blockchain create func
func NewBlockchain(db storage.Database, genesis *Genesis) (*Blockchain, error) {
	bc := &Blockchain{
		db:          db,
		blockCache:  make(map[common.Hash]*types.Block),
		numberCache: make(map[*big.Int]common.Hash),
		genesis:     genesis,
	}

	// init Genesis
	if err := bc.initGenesis(); err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *Blockchain) initGenesis() error {
	if block := bc.GetBlockByNumber(big.NewInt(0)); block != nil {
		bc.currentBlock = block
		bc.currentState = bc.loadState(block.StateRoot())
		return nil
	}

	// create genesis block
	genesisBlock := bc.genesis.ToBlock()

	// set up pre-alloc accounts
	stateDB := state.NewSimpleStateDB()
	for addr, account := range bc.genesis.Alloc {
		stateDB.SetBalance(addr, account.Balance)
		stateDB.SetNonce(addr, account.Nonce)
	}

	// setup state root
	genesisBlock.Header().StateHashRoot = stateDB.ComputeRoot()

	// store genesis Block
	if err := bc.SaveBlock(genesisBlock); err != nil {
		return err
	}

	bc.currentBlock = genesisBlock
	bc.currentState = stateDB

	return nil

}

// GetBlockByNumber Get Block by block height
func (bc *Blockchain) GetBlockByNumber(number *big.Int) *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	hash, ok := bc.numberCache[number]
	if !ok {
		// load from db
		hash = bc.loadBlockHash(number)
		if hash == (common.Hash{}) {
			return nil
		}
		bc.numberCache[number] = hash
	}

	return bc.GetBlock(hash)
}

// GetBlock get block by Hash
func (bc *Blockchain) GetBlock(hash common.Hash) *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// load from cache
	if block, ok := bc.blockCache[hash]; ok {
		return block
	}

	// load from db
	block, err := bc.loadBlock(hash)
	if err != nil {
		return nil
	}

	// store cache
	bc.blockCache[hash] = block
	return block
}

// CurrentBlock get current block
func (bc *Blockchain) CurrentBlock() *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentBlock
}

// CurrentState get current state
func (bc *Blockchain) CurrentState() *state.SimpleStateDB {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.currentState.Copy()
}

// InsertBlock insert new block into chain
func (bc *Blockchain) InsertBlock(block *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// 1. validate Block
	if err := bc.validateBlock(block); err != nil {
		return err
	}

	// 2. create new state copy
	newState := bc.currentState.Copy()

	// 3. 執行所有交易
	receipts := make([]*types.Receipt, 0, len(block.Transactions()))
	gasUsed := uint64(0)

	for i, tx := range block.Transactions() {
		receipt, err := bc.applyTransaction(newState, tx, block, uint(i), gasUsed)
		if err != nil {
			return fmt.Errorf("failed to apply tx %d: %w", i, err)
		}
		receipts = append(receipts, receipt)
		gasUsed += receipt.GasUsed
	}

	// 4. verify state root
	stateRoot := newState.ComputeRoot()
	if stateRoot != block.StateRoot() {
		return errors.New("state root mismatch")
	}

	// 5. save block and state
	if err := bc.SaveBlock(block); err != nil {
		return err
	}

	// 6. update current state
	bc.currentBlock = block
	bc.currentState = newState

	return nil
}

// validateBlock validate Block
func (bc *Blockchain) validateBlock(block *types.Block) error {
	// basic verify
	if err := block.Validate(); err != nil {
		return err
	}

	// verify parent
	parent := bc.GetBlock(block.ParentHash())
	if parent == nil {
		return common.ErrUnknownParent
	}

	// verify block number
	if block.Number().Uint64() != parent.Number().Uint64()+1 {
		return common.ErrInvalidBlockNumber
	}

	// verify timestamp
	if block.Timestamp() <= parent.Timestamp() {
		return common.ErrInvalidTimestamp
	}

	return nil
}

// loadState load state from db
func (bc *Blockchain) loadState(root common.Hash) *state.SimpleStateDB {
	return bc.db.LoadState(root)
}

// loadBlockHash load from db
func (bc *Blockchain) loadBlockHash(number *big.Int) common.Hash {
	return bc.db.LoadBlockHash(number)
}

// loadBlock load from db
func (bc *Blockchain) loadBlock(hash common.Hash) (*types.Block, error) {
	return bc.db.LoadBlock(hash)
}

func (bc *Blockchain) SaveBlock(block *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	err := bc.db.StoreBlock(block)
	if err != nil {
		return err
	}

	bc.blockCache[block.Hash()] = block
	bc.numberCache[block.Number()] = block.Hash()
	return nil
}

// applyTransaction apply transaction
func (bc *Blockchain) applyTransaction(
	stateDB *state.SimpleStateDB,
	tx *types.Transaction,
	block *types.Block,
	index uint,
	cumulativeGasUsed uint64,
) (*types.Receipt, error) {
	// 1. verify proof
	from, err := tx.From()
	if err != nil {
		return nil, err
	}

	// 2. check nonce
	if tx.Nonce() != stateDB.GetNonce(from) {
		return nil, common.ErrNonceMismatch
	}

	// 3. check balance（include gas fee）
	totalCost := tx.Cost()
	if stateDB.GetBalance(from).Cmp(totalCost) < 0 {
		return nil, common.ErrInsufficientBalance
	}

	snapshot := stateDB.Snapshot()

	// subtract gas fee
	// simplify: using all gas limit as fee
	gasFee := new(big.Int).SetUint64(tx.GasLimit())
	if err = stateDB.SubBalance(from, gasFee); err != nil {
		stateDB.RevertToSnapshot(snapshot)
		return nil, err
	}

	// do value Transfer
	if tx.To() != nil {
		if err = stateDB.Transfer(from, *tx.To(), tx.Value()); err != nil {
			stateDB.RevertToSnapshot(snapshot)
			return nil, err
		}
	}

	// increase account nonce
	stateDB.SetNonce(from, stateDB.GetNonce(from)+1)

	// 5. create receipt
	receipt := &types.Receipt{
		PostState:         stateDB.ComputeRoot().Bytes(),
		Status:            1, // success
		CumulativeGasUsed: cumulativeGasUsed + tx.GasLimit(),
		TxHash:            tx.Hash(),
		GasUsed:           tx.GasLimit(), // simplify: using all gas limit as fee
		BlockHash:         block.Hash(),
		BlockNumber:       block.Number(),
		TransactionIndex:  index,
	}

	return receipt, nil

}
