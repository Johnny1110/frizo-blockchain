package state

import "frizo-blockchain/common"

// ContractStorage represents the storage of a contract
type ContractStorage map[common.Hash]common.Hash

// Copy creates a deep copy of the storage
func (s ContractStorage) Copy() ContractStorage {
	cpy := make(ContractStorage, len(s))
	for key, value := range s {
		cpy[key] = value
	}
	return cpy
}
