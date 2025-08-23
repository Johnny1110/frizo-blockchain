package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/state"
	"frizo-blockchain/core/types"
	"frizo-blockchain/storage"
	"github.com/ethereum/go-ethereum/common/lru"
	"sync"
	"sync/atomic"
)

type BlockChain struct {
	chainConfig *ChainConfig
	db          storage.IKVStore

	// cache
	hc            *HeaderChain
	bodyCache     *lru.Cache[common.Hash, *types.Body]
	receiptsCache *lru.Cache[common.Hash, types.Receipts]

	// current
	currentBlock     atomic.Pointer[types.Block]
	currentFastBlock atomic.Pointer[types.Block]
	currentHeader    atomic.Pointer[types.Header]

	stateDb state.StateDB // current state

	validator    *BlockValidator // block validator
	processor    *StateProcessor // state processor
	chainManager *ChainManager   // chain manager

	// lock
	procmu  sync.RWMutex // block lock
	chainmu sync.RWMutex // chain lock

	// runtime status
	running       int32 // 0=stop, 1=running
	procInterrupt int32 // interrupt

	quit chan struct{} // quit
}

func NewBlockChain(db storage.IKVStore, chainConfig *ChainConfig, stateDb state.StateDB) (*BlockChain, error) {
	bc := &BlockChain{
		chainConfig:   chainConfig,
		db:            db,
		stateDb:       stateDb,
		bodyCache:     lru.NewCache[common.Hash, *types.Body](100),
		receiptsCache: lru.NewCache[common.Hash, types.Receipts](100),
		quit:          make(chan struct{}),
	}

	bc.validator = NewBlockValidator(chainConfig, bc)
	bc.processor = NewStateProcessor(chainConfig, bc)
	bc.chainManager = NewChainManager(bc)

	// load last state
	if err := bc.loadLastState(); err != nil {
		return nil, err
	}

	// if no block exist, init genesis
	if bc.CurrentBlock() == nil {
		if err := bc.initGenesis(); err != nil {
			return nil, err
		}
	}

	// setup running status
	atomic.StoreInt32(&bc.running, 1)

	return bc, nil
}
