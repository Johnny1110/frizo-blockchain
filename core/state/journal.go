package state

import (
	"frizo-blockchain/common"
	"math/big"
)

// journalEntry is a modification entry in the state change journal
type journalEntry interface {
	// revert undoes the changes introduced by this entry
	revert(*stateDB)
	// dirtied returns the address modified by this entry
	dirtied() *common.Address
}

// journal contains the list of state modifications applied since the last state commit
type journal struct {
	entries []journalEntry         // Current changes
	dirties map[common.Address]int // Dirty accounts and the number of changes
}

// newJournal creates a new journal
func newJournal() *journal {
	return &journal{
		dirties: make(map[common.Address]int),
	}
}

// append adds a new entry to the journal
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if addr := entry.dirtied(); addr != nil {
		j.dirties[*addr]++
	}
}

// revert undoes state changes along the journal by snapshot
func (j *journal) revert(sdb *stateDB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		j.entries[i].revert(sdb)

		// Decrease dirty count
		if addr := j.entries[i].dirtied(); addr != nil {
			if j.dirties[*addr]--; j.dirties[*addr] == 0 {
				delete(j.dirties, *addr)
			}
		}
	}
	// remove snapshot
	j.entries = j.entries[:snapshot]
}

// length returns the length of the journal
func (j *journal) length() int {
	return len(j.entries)
}

// define all journalEntry implement types: ---------------------------
type (
	// createObjectChange creates a new account
	createObjectChange struct {
		account *common.Address
	}

	// resetObjectChange resets an account
	resetObjectChange struct {
		prev *stateObject
	}

	// suicideChange marks an account for deletion
	suicideChange struct {
		account     *common.Address
		prev        bool // whether account was already suicided
		prevBalance *big.Int
	}

	// balanceChange changes the balance of an account
	balanceChange struct {
		account *common.Address
		prev    *big.Int
	}

	// nonceChange changes the nonce of an account
	nonceChange struct {
		account *common.Address
		prev    uint64
	}

	// storageChange changes a storage slot
	storageChange struct {
		account  *common.Address
		key      common.Hash
		prevalue common.Hash
	}

	// codeChange changes the code of an account
	codeChange struct {
		account  *common.Address
		prevcode []byte
		prevhash common.Hash
	}

	// refundChange changes the refund amount
	refundChange struct {
		prev uint64
	}

	// addLogChange adds a log
	addLogChange struct {
		txhash common.Hash
	}

	// accessListAddAccountChange adds an account to the access list
	accessListAddAccountChange struct {
		address *common.Address
	}

	// accessListAddSlotChange adds a storage slot to the access list
	accessListAddSlotChange struct {
		address *common.Address
		slot    *common.Hash
	}
)

// Implement revert & dirtied methods for each journal entry type

// createObjectChange >>>
func (c *createObjectChange) revert(db *stateDB) {
	delete(db.stateObjects, *c.account)
	delete(db.stateObjectsDirty, *c.account)
	delete(db.stateObjectsPending, *c.account)
}
func (c *createObjectChange) dirtied() *common.Address {
	return c.account
}

// resetObjectChange >>>
func (c *resetObjectChange) revert(db *stateDB) {
	db.setStateObject(c.prev)
}

func (c *resetObjectChange) dirtied() *common.Address {
	return nil
}

// suicideChange >>>
func (ch suicideChange) revert(s *stateDB) {
	obj := s.getStateObject(*ch.account)
	if obj != nil {
		obj.suicided = ch.prev
		obj.data.Balance = ch.prevBalance
	}
}

func (ch suicideChange) dirtied() *common.Address {
	return ch.account
}

// balanceChange >>>
func (ch balanceChange) revert(s *stateDB) {
	obj := s.getStateObject(*ch.account)
	if obj != nil {
		obj.data.Balance = ch.prev
	}
}

func (ch balanceChange) dirtied() *common.Address {
	return ch.account
}

// nonceChange >>>
func (ch nonceChange) revert(s *stateDB) {
	obj := s.getStateObject(*ch.account)
	if obj != nil {
		obj.data.Nonce = ch.prev
	}
}

func (ch nonceChange) dirtied() *common.Address {
	return ch.account
}

// contract storageChange >>>
func (ch storageChange) revert(s *stateDB) {
	if obj := s.getStateObject(*ch.account); obj != nil {
		obj.pendingStorage[ch.key] = ch.prevalue
	}
}

func (ch storageChange) dirtied() *common.Address {
	return ch.account
}

// codeChange >>>
func (ch codeChange) revert(s *stateDB) {
	if obj := s.getStateObject(*ch.account); obj != nil {
		obj.code = ch.prevcode
		obj.data.CodeHash = ch.prevcode
		obj.dirtyCode = true
	}
}

func (ch codeChange) dirtied() *common.Address {
	return ch.account
}

// refundChange >>>
func (ch refundChange) revert(s *stateDB) {
	s.refund = ch.prev
}

func (ch refundChange) dirtied() *common.Address {
	return nil
}

// addLogChange >>>
func (ch addLogChange) revert(s *stateDB) {
	logs := s.logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.txhash)
	} else {
		s.logs[ch.txhash] = logs[:len(logs)-1]
	}
	s.logSize--
}

func (ch addLogChange) dirtied() *common.Address {
	return nil
}

// accessListAddAccountChange >>>
func (ch accessListAddAccountChange) revert(s *stateDB) {
	s.accessList.DeleteAddress(*ch.address)
}

func (ch accessListAddAccountChange) dirtied() *common.Address {
	return nil
}

// accessListAddSlotChange >>>
func (ch accessListAddSlotChange) revert(s *stateDB) {
	s.accessList.DeleteSlot(*ch.address, *ch.slot)
}

func (ch accessListAddSlotChange) dirtied() *common.Address {
	return nil
}
