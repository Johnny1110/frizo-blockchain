package state

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"frizo-blockchain/storage"
	"math/big"
)

// core db interface -> abstract MPT access to account access
type StateDB interface {
	// Account
	CreateAccount(address common.Address)
	SubBalance(address common.Address, value *big.Int) error
	AddBalance(address common.Address, amount *big.Int) error
	GetBalance(address common.Address) *big.Int
	GetNonce(address common.Address) uint64
	SetNonce(address common.Address, nonce uint64)

	// Contract
	GetCodeHash(address common.Address) common.Hash
	GetCode(address common.Address) []byte
	SetCode(address common.Address, code []byte)
	GetCodeSize(address common.Address) int

	// Contract Storage Access
	GetState(addr common.Address, hash common.Hash) common.Hash
	SetState(address common.Address, hash1 common.Hash, hash2 common.Hash)
	GetCommittedState(address common.Address, hash common.Hash) common.Hash

	// Account management
	HasSuicided(address common.Address) bool
	Suicide(address common.Address) bool
	Exist(address common.Address) bool
	Empty(address common.Address) bool

	// Snapshot Revert Commit
	Snapshot() int
	Revert(int)
	Commit(deleteEmptyObjects bool) (common.Hash, error)

	// persistence
	IntermediateRoot(deleteEmptyObjects bool) common.Hash
	Finalise(deleteEmptyObjects bool)

	// Logs
	AddLog(*types.Log)
	GetLogs(hash common.Hash) []*types.Log
	Logs() []*types.Log

	// Refunds
	AddRefund(value uint64)
	SubRefund(value uint64)
	GetRefund() uint64

	// Access list (EIP-2929) link: https://medium.com/taipei-ethereum-meetup/eip2929-eip2930-%E7%B0%A1%E4%BB%8B-14a8b580a141
	AddAddressToAccessList(addr common.Address)
	AddSlotToAccessList(addr common.Address, slot common.Hash)
	IsAddressInAccessList(addr common.Address) bool
	IsSlotInAccessList(addr common.Address, slot common.Hash) (addressOk, slotOk bool)

	// Debugging and tools
	ForEachContractStorage(common.Address, func(common.Hash, common.Hash) bool) error
	Copy() StateDB
	Database() Database

	// Additional helper methods
	Reset(root common.Hash) error
	Error() error
}

// Database wraps access to tries and contract code
type Database interface {
	// Trie operations
	OpenTrie(root common.Hash) (Trie, error)
	OpenStorageTrie(addrHash, root common.Hash) (Trie, error)
	CopyTrie(Trie) Trie

	// Contract code operations
	ContractCode(addrHash, codeHash common.Hash) ([]byte, error)
	ContractCodeSize(addrHash, codeHash common.Hash) (int, error)
	ContractCodeWithPrefix(addrHash, codeHash common.Hash) ([]byte, error)

	// Database access
	TrieDB() storage.Database
}

// Trie is the interface for Merkle Patricia Trie operations
type Trie interface {
	// Basic operations
	TryGet(key []byte) ([]byte, error)
	TryUpdate(key, value []byte) error
	TryDelete(key []byte) error

	// Commit and hash
	Commit() (common.Hash, error)
	Hash() common.Hash

	// Iterator
	NodeIterator(startKey []byte) NodeIterator

	// Proof generation
	Prove(key []byte, fromLevel uint, proofDb storage.Database) error
}

// Revision represents a state revision point
type revision struct {
	id           int
	journalIndex int
}
