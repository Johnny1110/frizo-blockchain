## state -> access_list.go

```go
package state

import (
"frizo-blockchain/common"
)

// accessList is an EIP-2929 access list implementation
type accessList struct {
addresses map[common.Address]int
slots     []map[common.Hash]struct{}
}

// newAccessList creates a new access list
func newAccessList() *accessList {
return &accessList{
addresses: make(map[common.Address]int),
slots:     make([]map[common.Hash]struct{}, 0),
}
}

// Copy creates a deep copy of the access list
func (al *accessList) Copy() *accessList {
cp := newAccessList()
for k, v := range al.addresses {
cp.addresses[k] = v
}
cp.slots = make([]map[common.Hash]struct{}, len(al.slots))
for i, slotMap := range al.slots {
newSlotMap := make(map[common.Hash]struct{}, len(slotMap))
for k := range slotMap {
newSlotMap[k] = struct{}{}
}
cp.slots[i] = newSlotMap
}
return cp
}

// AddAddress adds an address to the access list
func (al *accessList) AddAddress(address common.Address) bool {
if _, present := al.addresses[address]; present {
return false
}
al.addresses[address] = len(al.slots)
al.slots = append(al.slots, make(map[common.Hash]struct{}))
return true
}

// AddSlot adds a storage slot to the access list
func (al *accessList) AddSlot(address common.Address, slot common.Hash) (addrChange bool, slotChange bool) {
idx, addrPresent := al.addresses[address]
if !addrPresent {
al.addresses[address] = len(al.slots)
slotMap := make(map[common.Hash]struct{})
slotMap[slot] = struct{}{}
al.slots = append(al.slots, slotMap)
return true, true
}

slotMap := al.slots[idx]
if _, slotPresent := slotMap[slot]; !slotPresent {
slotMap[slot] = struct{}{}
return false, true
}
return false, false
}

// DeleteAddress removes an address from the access list
func (al *accessList) DeleteAddress(address common.Address) {
if idx, present := al.addresses[address]; present {
delete(al.addresses, address)
al.slots[idx] = nil
}
}

// DeleteSlot removes a storage slot from the access list
func (al *accessList) DeleteSlot(address common.Address, slot common.Hash) {
if idx, present := al.addresses[address]; present {
slotMap := al.slots[idx]
delete(slotMap, slot)
}
}

// ContainsAddress checks if an address is in the access list
func (al *accessList) ContainsAddress(address common.Address) bool {
_, present := al.addresses[address]
return present
}

// Contains checks if an address and slot are in the access list
func (al *accessList) Contains(address common.Address, slot common.Hash) (addressPresent bool, slotPresent bool) {
idx, addressPresent := al.addresses[address]
if !addressPresent {
return false, false
}
if slotMap := al.slots[idx]; slotMap != nil {
_, slotPresent = slotMap[slot]
}
return true, slotPresent
}
```