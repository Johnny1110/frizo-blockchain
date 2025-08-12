package blockchain

import (
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
	"frizo-blockchain/db"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

type Blockchain struct {
	db db.Database

	currentBlock *types.Block
	currentState *state.SimpleStateDB

	// cache:
	blockCache  map[common.Hash]*types.Block
	numberCache map[*big.Int]common.Hash

	genesis *Genesis
}

// Blockchain create func
func NewBlockchain(db db.Database, genesis *Genesis) (*Blockchain, error) {
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
	return bc.currentBlock
}

// CurrentState get current state
func (bc *Blockchain) CurrentState() *state.SimpleStateDB {
	return bc.currentState.Copy()
}

// InsertBlock insert new block into chain
func (bc *Blockchain) InsertBlock(block *types.Block) error {
	// 1. validate Block
	if err := bc.validateBlock(block); err != nil {
		return err
	}

	// 2. create new state copy
	newState := bc.currentState.Copy()

	// 3. exec all receipts
	receipts := make([]*types.Receipt, 0, len(block.Transactions()))
	gasUsed := big.NewInt(0)

	for i, tx := range block.Transactions() {
		receipt, err := bc.applyTransaction(newState, tx, uint(i), gasUsed, block)
		if err != nil {
			return fmt.Errorf("failed to apply tx %d: %w", i, err)
		}
		receipts = append(receipts, receipt)
		gasUsed = new(big.Int).Add(gasUsed, receipt.GasUsed)
	}

	// 4. verify state root
	stateRoot := newState.ComputeRoot()
	if stateRoot != block.StateRoot() {
		fmt.Println("stateRoot:", stateRoot)
		fmt.Println("blockRoot:", block.StateRoot())
		log.Error("stateRoot mismatch", "expected", block.StateRoot(), "actual", stateRoot)
		return errors.New("state root mismatch")
	}

	// 5. save block and state
	if err := bc.SaveBlock(block); err != nil {
		return err
	}

	// 6. update current state
	bc.currentBlock = block
	bc.currentState = newState

	bc.db.StoreState(newState)

	return nil
}

// validateBlock validate Block
func (bc *Blockchain) validateBlock(block *types.Block) error {
	// basic verify
	if err := block.Validate(); err != nil {
		log.Error("validate block failed", "err", err)
		return err
	}

	// verify parent
	parent := bc.GetBlock(block.ParentHash())
	if parent == nil {
		return common.ErrUnknownParent
	}

	// verify block number
	if block.Number().Uint64() != parent.Number().Uint64()+1 {
		fmt.Println(block.Number().Uint64(), parent.Number().Uint64())
		log.Error("block number mismatch", "block", block.Number(), "parent", parent.Number())
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
	index uint,
	cumulativeGasUsed *big.Int,
	block *types.Block,
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
	totalCost := tx.TotalCost()
	if stateDB.GetBalance(from).Cmp(totalCost) < 0 {
		return nil, common.ErrInsufficientBalance
	}

	snapshot := stateDB.Snapshot()

	// subtract gas fee
	// simplify: using all gas limit as fee
	if err = stateDB.SubBalance(from, tx.GasCost()); err != nil {
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
		CumulativeGasUsed: new(big.Int).Add(cumulativeGasUsed, tx.GasCost()),
		TxHash:            tx.Hash(),
		GasUsed:           tx.GasLimit(), // simplify: using all gas limit as fee
		BlockHash:         block.Hash(),
		BlockNumber:       block.Number(),
		TransactionIndex:  index,
	}

	return receipt, nil

}
