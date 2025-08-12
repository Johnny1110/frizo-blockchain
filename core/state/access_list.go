package state

import "frizo-blockchain/common"

type accessList struct {
}

func newAccessList() *accessList {
	return &accessList{}
}

func (l *accessList) AddAddress(addr common.Address) bool {
	// TODO
	return true
}

func (l *accessList) AddSlot(addr common.Address, slot common.Hash) (bool, bool) {
	// TODO
	return true, true
}

func (l *accessList) ContainsAddress(addr common.Address) bool {
	// TODO
	return true
}

func (l *accessList) Contains(addr common.Address, slot common.Hash) (bool, bool) {
	// TODO
	return true, true
}

func (l *accessList) Copy() *accessList {
	return nil
}

func (l *accessList) DeleteAddress(address common.Address) {
	// TODO
}

func (l *accessList) DeleteSlot(address common.Address, hash common.Hash) {
	// TODO
}
