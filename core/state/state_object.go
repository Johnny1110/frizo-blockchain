package state

import (
	"frizo-blockchain/common"
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

	trie Trie // storage trie
	code Code // contract bytecode

	originStorage  ContractStorage // Storage cache of original entries
	pendingStorage ContractStorage // Storage entries to be updated
	dirtyStorage   ContractStorage // Storage entries that need to be flushed to disk

	// Cache flags
	dirtyCode bool // if the code was updated
	suicided  bool
	deleted   bool
}

func newStateObject(sdb *stateDB, addr common.Address, account Account) *stateObject {
	return nil
}

func (o *stateObject) SetBalance(balance *big.Int) {
	o.data.Balance = balance
}

func (o *stateObject) SubBalance(amount *big.Int) error {
	o.data.Balance = new(big.Int).Sub(o.data.Balance, amount)
}

func (o *stateObject) AddBalance(amount *big.Int) error {
	o.data.Balance = new(big.Int).Add(o.data.Balance, amount)
}

func (o *stateObject) Balance() *big.Int {
	return o.data.Balance
}

func (o *stateObject) Nonce() uint64 {
	return o.data.Nonce
}

func (o *stateObject) SetNonce(nonce uint64) {
	o.data.Nonce = nonce
}

func (o *stateObject) CodeHash() []byte {
	// TODO
	return nil
}

func (o *stateObject) Code(s *stateDB) Code {
	return o.code
}

func (o *stateObject) SetCode(hash common.Hash, code []byte) {
	// TODO
}

func (o *stateObject) CodeSize(db Database) int {
	// TODO
}

func (o *stateObject) GetState(db Database, hash common.Hash) common.Hash {
	// TODO
}

func (o *stateObject) SetState(db Database, key common.Hash, value common.Hash) {
	// TODO
}

func (o *stateObject) GetCommittedState(db Database, hash common.Hash) common.Hash {
	// TODO
}

func (o *stateObject) markSuicide() {
	// TODO
}

func (o *stateObject) empty() bool {
	// TODO
	return false
}

func (o *stateObject) finalise() {
	// TODO
}

func (o *stateObject) ForEachContractStorage(cb func(key common.Hash, value common.Hash) bool) error {
	// TODO
}

func (o *stateObject) deepCopy(state *stateDB) *stateObject {
	// TODO
}
