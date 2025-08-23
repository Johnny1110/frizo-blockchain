package state

import (
	"bytes"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/storage"
	"math/big"
)

type Code []byte

func (c Code) String() string {
	return common.Bytes2Hex(c)
}

// stateObject represent an account which is bing modified
type stateObject struct {
	address  common.Address
	addrHash common.Hash
	data     Account
	db       *stateDB

	dbErr error

	trie Trie // contract storage trie
	code Code // contract bytecode

	originStorage  ContractStorage // Storage cache of original entries
	pendingStorage ContractStorage // Storage entries to be updated
	dirtyStorage   ContractStorage // Storage entries that need to be flushed to disk

	// Cache flags
	dirtyCode bool // if the code was updated
	suicided  bool
	deleted   bool
}

func (s *stateObject) getKVStore() (storage.IKVStore, error) {
	kvs := s.db.db
	if s.trie == nil {
		return nil, fmt.Errorf("kv store not initialized in state object")
	}
	return kvs, nil
}

// newStateObject create a state object
func newStateObject(sdb *stateDB, addr common.Address, account *Account) *stateObject {
	if account == nil {
		account = NewAccount()
	}

	return &stateObject{
		db:             sdb,
		address:        addr,
		data:           *account,
		originStorage:  make(ContractStorage),
		pendingStorage: make(ContractStorage),
		dirtyStorage:   make(ContractStorage),
	}
}

func (s *stateObject) Address() common.Address {
	return s.address
}

func (s *stateObject) Balance() *big.Int {
	return s.data.Balance
}

func (o *stateObject) SetBalance(balance *big.Int) {
	o.db.journal.append(balanceChange{
		account: &o.address,
		prev:    new(big.Int).Set(o.data.Balance),
	})
	o.data.Balance = balance
}

func (o *stateObject) SubBalance(amount *big.Int) error {
	if amount.Sign() == 0 {
		return nil
	}

	if o.Balance().Cmp(amount) < 0 {
		return common.ErrInsufficientFunds
	}
	o.SetBalance(new(big.Int).Sub(o.Balance(), amount))
	return nil
}

func (o *stateObject) AddBalance(amount *big.Int) error {
	if amount.Sign() == 0 {
		return nil
	}
	o.SetBalance(new(big.Int).Add(o.data.Balance, amount))
	return nil
}

func (o *stateObject) Nonce() uint64 {
	return o.data.Nonce
}

func (o *stateObject) SetNonce(nonce uint64) {
	o.db.journal.append(nonceChange{
		account: &o.address,
		prev:    o.data.Nonce,
	})

	o.data.Nonce = nonce
}

// Code returns the contract code
func (o *stateObject) Code() Code {
	if o.code != nil {
		return o.code
	}

	if bytes.Equal(o.CodeHash().Bytes(), emptyCodeHash.Bytes()) {
		return nil
	}

	// Get contractCode from db (addr_hash, codeHash)
	code, err := o.db.db.Get(codeKey(o.CodeHash()))
	if err != nil {
		o.setErr(fmt.Errorf("failed to get code for address %v: %v", o.addrHash, err))
	}
	// cache code:
	o.code = code
	return code
}

func (o *stateObject) CodeSize() int {
	if o.code != nil {
		return len(o.code)
	}
	if bytes.Equal(o.CodeHash().Bytes(), emptyCodeHash.Bytes()) {
		return 0
	}
	code := o.Code()
	return len(code)
}

func (o *stateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := o.Code()
	o.db.journal.append(codeChange{
		account:  &o.address,
		prevhash: o.CodeHash(),
		prevcode: prevcode,
	})

	o.code = code
	o.data.CodeHash = codeHash
	o.dirtyCode = true
}

func (o *stateObject) CodeHash() common.Hash {
	return o.data.CodeHash
}

func (o *stateObject) GetState(key common.Hash) common.Hash {
	// check from pending contract storage
	if val, pending := o.pendingStorage[key]; pending {
		return val
	}

	// check from o.cache
	if val, cached := o.originStorage[key]; cached {
		return val
	}

	// load from trie
	val := o.getState(key)
	// store into cache
	o.originStorage[key] = val
	return val
}

// GetCommittedState returns the committed value
func (o *stateObject) GetCommittedState(hash common.Hash) common.Hash {
	// check origin state
	if val, cached := o.originStorage[hash]; cached {
		return val
	}
	// load from trie
	val := o.getState(hash)
	o.originStorage[hash] = val
	return val
}

// SetState sets a value in the storage trie
func (o *stateObject) SetState(key common.Hash, value common.Hash) {
	preVal := o.GetState(key)
	if preVal == value {
		return
	}
	// mark into journal
	o.db.journal.append(storageChange{
		account:  &o.address,
		key:      key,
		prevalue: preVal,
	})

	o.pendingStorage[key] = value
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// <private> -------------------------------------------------------------------------
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

// updateRoot updates the storage root
func (o *stateObject) updateRoot() {
	o.updateTrie()
	o.data.Root = o.trie.Hash()
}

// updateTrie updates the storage trie
func (o *stateObject) updateTrie() Trie {
	o.finalise() // move pending to dirty.
	source, err := o.getKVStore()
	if err != nil {
		o.dbErr = err
		return nil
	}
	trie := o.getTrie(source)
	for key, val := range o.dirtyStorage {
		delete(o.dirtyStorage, key)
		// delete if val is zero
		if val == (common.Hash{}) {
			o.setErr(trie.TryDelete(key.Bytes()))
			continue
		}

		// encode and update
		v, _ := common.RlpEncodeToBytes(val.Bytes())
		o.setErr(trie.TryUpdate(key.Bytes(), v))
	}
	return trie
}

// getTrie returns the contract storage trie
func (o *stateObject) getTrie(source storage.IKVStore) Trie {
	if o.trie == nil {
		trie, err := OpenStorageTrie(source, o.addrHash, o.data.Root)
		if err != nil {
			// open trie with empty hash root
			trie, err = OpenStorageTrie(source, o.addrHash, common.Hash{})
			if err != nil {
				o.setErr(fmt.Errorf("failed to open storage trie: %v", err))
			}
		}
		o.trie = trie
	}
	return o.trie
}

// getState retrieves a value from the storage trie
func (o *stateObject) getState(key common.Hash) common.Hash {
	db, err := o.getKVStore()
	if err != nil {
		o.setErr(err)
		return common.Hash{}
	}

	trie := o.getTrie(db)
	encoded, err := trie.TryGet(key.Bytes())
	if err != nil {
		o.setErr(err)
		return common.Hash{}
	}

	var val common.Hash
	if len(encoded) > 0 {
		_, content, _, err := common.RlpSplit(encoded)
		if err != nil {
			o.setErr(err)
		}
		val.SetBytes(content)
	}
	return val
}

func (o *stateObject) markSuicide() {
	o.suicided = true
}

// empty returns whether the account is empty
func (o *stateObject) empty() bool {
	return o.data.Nonce == 0 &&
		o.data.Balance.Sign() == 0 &&
		bytes.Equal(o.data.CodeHash.Bytes(), emptyCodeHash.Bytes())
}

// finalise: storage from pending to dirty
func (o *stateObject) finalise() {
	for key, val := range o.pendingStorage {
		o.dirtyStorage[key] = val
	}
	if len(o.dirtyStorage) > 0 {
		// clean pending
		o.pendingStorage = make(ContractStorage)
	}
}

func (o *stateObject) deepCopy(db *stateDB) *stateObject {
	obj := newStateObject(db, o.address, o.data.Copy())
	// copy trie
	if o.trie != nil {
		obj.trie = o.trie.Copy()
	}
	// storage code and storage stuff
	obj.code = o.code
	obj.dirtyStorage = o.dirtyStorage.Copy()
	obj.originStorage = o.originStorage.Copy()
	obj.pendingStorage = o.pendingStorage.Copy()
	obj.dirtyCode = o.dirtyCode
	obj.suicided = o.suicided
	obj.deleted = o.deleted

	return obj
}

func (o *stateObject) setErr(err error) {
	o.dbErr = err
}

// Helper function to generate code storage key
func codeKey(hash common.Hash) []byte {
	return append([]byte("code-"), hash.Bytes()...)
}
