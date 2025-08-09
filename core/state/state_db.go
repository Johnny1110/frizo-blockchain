package state

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"math/big"
	"sync"
)

// stateDB is the main implementation of StateDB interface
type stateDB struct {
	db   Database
	trie Trie

	// cache
	stateObjects        map[common.Address]*stateObject
	stateObjectsPending map[common.Address]struct{} // stateObj finalized but not yet written to tree
	stateObjectsDirty   map[common.Address]struct{} // stateObj modified in current execution

	// db error
	dbErr error

	// Journal and revisions
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	// transaction context
	txHash    common.Hash
	blockHash common.Hash
	txIndex   int // txn index in block

	// logs
	logs    map[common.Hash][]*types.Log
	logSize uint

	// refund counter
	refund uint64

	// access list
	accessList *accessList

	mu sync.RWMutex
}

// NewStateDB create new stateDB
func NewStateDB(db Database, root common.Hash) (*stateDB, error) {
	trie, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}

	sdb := &stateDB{
		db:                  db,
		trie:                trie,
		stateObjects:        make(map[common.Address]*stateObject),
		stateObjectsPending: make(map[common.Address]struct{}),
		stateObjectsDirty:   make(map[common.Address]struct{}),
		logs:                make(map[common.Hash][]*types.Log),
		journal:             newJournal(),
		accessList:          newAccessList(),
	}

	return sdb, nil
}

func (s *stateDB) Error(addr common.Address) error {
	return s.dbErr
}

// Reset reset the stateDB to given root
func (s *stateDB) Reset(root common.Hash) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tr, err := s.db.OpenTrie(root)
	if err != nil {
		return err
	}

	s.trie = tr
	s.stateObjects = make(map[common.Address]*stateObject)
	s.stateObjectsPending = make(map[common.Address]struct{})
	s.stateObjectsDirty = make(map[common.Address]struct{})
	s.logs = make(map[common.Hash][]*types.Log)
	s.journal = newJournal()
	s.validRevisions = s.validRevisions[:0]
	s.nextRevisionId = 0
	s.accessList = newAccessList()

	return nil
}

func (s *stateDB) Database() Database {
	return s.db
}

func (s *stateDB) CreateAccount(addr common.Address) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newObject, prev := s.createStateObject(addr)
	if prev != nil {
		newObject.SetBalance(prev.data.Balance)
	}
}

func (s *stateDB) SubBalance(addr common.Address, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj := s.getOrNewStateObject(addr)
	if obj != nil {
		return obj.SubBalance(amount)
	}
	return nil
}

func (s *stateDB) AddBalance(addr common.Address, amount *big.Int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj := s.getOrNewStateObject(addr)
	if obj != nil {
		return obj.AddBalance(amount)
	}
	return nil
}

func (s *stateDB) GetBalance(addr common.Address) *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.Balance()
	}
	return big.NewInt(0)
}

func (s *stateDB) GetNonce(addr common.Address) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.Nonce()
	}
	return 0
}

func (s *stateDB) SetNonce(addr common.Address, nonce uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj := s.getOrNewStateObject(addr)
	if obj != nil {
		obj.SetNonce(nonce)
	}
}

func (s *stateDB) GetCodeHash(addr common.Address) common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	if obj != nil {
		return common.BytesToHash(obj.CodeHash())
	}

	return common.Hash{}
}

func (s *stateDB) GetCode(addr common.Address) []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.Code(s)
	}

	return nil
}

func (s *stateDB) SetCode(addr common.Address, code []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj := s.getOrNewStateObject(addr)
	if obj != nil {
		obj.SetCode(common.Keccak256Hash(code), code)
	}
}

func (s *stateDB) GetCodeSize(addr common.Address) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.CodeSize(s.db)
	}
	return 0
}

// GetState retrieves a value from the storage trie
func (s *stateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.GetState(s.db, hash)
	}

	return common.Hash{}
}

// SetState sets a value in the storage trie
func (s *stateDB) SetState(addr common.Address, key, value common.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()

	obj := s.getOrNewStateObject(addr)
	if obj != nil {
		obj.SetState(s.db, key, value)
	}
}

// GetCommittedState retrieves the committed value from the storage trie
func (s *stateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.GetCommittedState(s.db, hash)
	}

	return common.Hash{}
}

// Suicide marks an account for deletion
func (s *stateDB) Suicide(addr common.Address) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	obj := s.getStateObject(addr)
	if obj == nil {
		return false
	}

	s.journal.append(suicideChange{
		account:     &addr,
		prev:        obj.suicided,
		prevBalance: new(big.Int).Set(obj.Balance()),
	})

	obj.markSuicide()
	obj.data.Balance = big.NewInt(0)

	return true
}

// HasSuicided checks if an account has been marked for deletion
func (s *stateDB) HasSuicided(addr common.Address) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	if obj != nil {
		return obj.suicided
	}
	return false
}

// Exist checks if an account exists
func (s *stateDB) Exist(addr common.Address) bool {
	s.mu.RLock()
	s.mu.RUnlock()

	return s.getStateObject(addr) != nil
}

// Empty checks if an account is empty
func (s *stateDB) Empty(addr common.Address) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	obj := s.getStateObject(addr)
	return obj == nil || obj.empty()
}

// TODO -->>>>>

// ===========================================================================================================
// CRUD stateObject ------------------------------------------------------------------------------------------
// ===========================================================================================================

func (s *stateDB) createStateObject(addr common.Address) (new *stateObject, prev *stateObject) {
	prev := s.getStateObject(addr)
	new := newStateObject(s, addr, NewAccount())

	if prev == nil {
		s.journal.append(createObjectChange{account: &addr})
	} else {
		s.journal.append(resetObjectChange{prev: prev.deepCopy(s)})
	}

	s.setStateObject(new)
	return new, prev
}

func (s *stateDB) setStateObject(obj *stateObject) {
	s.stateObjects[obj.address] = obj
}

func (s *stateDB) deleteStateObject(obj *stateObject) {
	addr := obj.address
	err := s.trie.TryDelete(addr.Bytes())
	if err != nil {
		s.setErr(fmt.Errorf("failed to delete stateObject %x: %v", addr, err))
	}
}

func (s *stateDB) updateStateObject(obj *stateObject) {
	addr := obj.address
	data, err := obj.data.EncodeAccount()
	if err != nil {
		s.setErr(fmt.Errorf("updateStateObject encode error: %v", err))
		return
	}

	err = s.trie.TryUpdate(addr.Bytes(), data)
	if err != nil {
		s.setErr(fmt.Errorf("updateStateObject (%x) error: %v", addr.Bytes(), err))
	}
}

func (s *stateDB) getOrNewStateObject(addr common.Address) *stateObject {
	obj := s.getStateObject(addr)
	if obj == nil || obj.deleted {
		obj, _ = s.createStateObject(addr)
	}
	return obj
}

func (s *stateDB) getStateObject(addr common.Address) *stateObject {
	// try load from cache
	if obj, ok := s.stateObjects[addr]; ok {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// load from trie
	encoded, err := s.trie.TryGet(addr.Bytes())
	if err != nil {
		s.setErr(fmt.Errorf("get stateObject (%s) error: %v", addr, err))
		return nil
	}

	if encoded == nil || len(encoded) == 0 {
		return nil
	}

	// decode account with RLP
	acc, err := DecodeAccount(encoded)
	if err != nil {
		s.setErr(fmt.Errorf("decode stateObject (%s) error: %v", addr, err))
		return nil
	}

	obj := newStateObject(s, addr, *acc)
	s.setStateObject(obj)
	return obj
}

func (s *stateDB) setErr(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}
