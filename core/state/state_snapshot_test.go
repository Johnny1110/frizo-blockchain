package state

import (
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/trie"
	"testing"
)

// this test is created for learning state snapshot, revert, commit

type kvState struct {
	baseTrie *trie.ModifiedMerklePatriciaTree

	cache     map[string][]byte
	snapshots []snapshot
}

type snapshot struct {
	id  int
	log []changeLog
}

type changeLog struct {
	key  string
	prev []byte
}

func (s *kvState) Snapshot() int {
	id := len(s.snapshots)
	s.snapshots = append(s.snapshots, snapshot{
		id:  id,
		log: []changeLog{},
	})
	return id
}

func (s *kvState) Put(key []byte, val []byte) {
	currentSnapshotVersion := len(s.snapshots)
	hexKey := common.Bytes2Hex(key)

	if currentSnapshotVersion > 0 {
		if _, exists := s.cache[hexKey]; !exists {
			// 從 MPT 中找出 key 並存入上一個 snapshot 中 ( 失敗時可以用上一個 snapshot 還原)
			prev, _ := s.Get(key)
			top := &s.snapshots[currentSnapshotVersion-1]
			top.log = append(top.log, changeLog{
				key:  hexKey,
				prev: prev,
			})
		}
	}
	// 把 value 寫入
	s.cache[hexKey] = val
}

func (s *kvState) Get(key []byte) ([]byte, bool) {
	hexKey := common.Bytes2Hex(key)
	if val, ok := s.cache[hexKey]; ok {
		return val, true
	}
	val, err := s.baseTrie.Get(key)
	if err != nil {
		return nil, false
	}
	return val, true
}

func (s *kvState) Revert(id int) error {
	if id >= len(s.snapshots) {
		return errors.New("snapshot id out of range")
	}

	snap := s.snapshots[id]

	for _, log := range snap.log {
		if log.prev == nil {
			// if no val found
			delete(s.cache, log.key)
		} else {
			// let previous key * value cover current cache
			s.cache[log.key] = log.prev
		}
	}
	// remove other snapshot
	s.snapshots = s.snapshots[:id]
	return nil
}

func (s *kvState) Commit() {
	for key, val := range s.cache {
		err := s.baseTrie.Put(common.Hex2Bytes(key), val)
		if err != nil {
			panic(err)
		}
	}
	s.cache = map[string][]byte{}
	s.snapshots = []snapshot{}
}

func Test_snapshot_revert_commit(t *testing.T) {

	state := &kvState{
		baseTrie:  trie.NewMPT(),
		cache:     map[string][]byte{},
		snapshots: []snapshot{},
	}

	spVersion_1 := state.Snapshot()
	fmt.Println("snapshots:", spVersion_1)

	state.Put([]byte("AAA"), []byte("111"))
	state.Put([]byte("BBB"), []byte("222"))

	spVersion_2 := state.Snapshot()
	fmt.Println("snapshots:", spVersion_2)

	state.Put([]byte("CCC"), []byte("333"))
	state.Revert(spVersion_2)

	spVersion_3 := state.Snapshot()
	fmt.Println("snapshots:", spVersion_3)
	state.Put([]byte("AAA"), []byte("000"))
	state.Revert(spVersion_3)

	state.Commit()

	state.baseTrie.PrintTree()
}
