package rawdb

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
)

// ReadCanonicalHash read canonical chain block hash by block number
func ReadCanonicalHash(db IDatabaseReader, number uint64) common.Hash {
	data, err := db.Get(headerHashKey(number))
	if err != nil {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteCanonicalHash
func WriteCanonicalHash(db IDatabaseWriter, number uint64, hash common.Hash) error {
	err := db.Put(headerHashKey(number), hash.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func DeleteCanonicalHash(db IDatabaseWriter, number uint64) error {
	return db.Delete(headerHashKey(number))
}

// ReadHeader 讀取區塊頭
func ReadHeader(db IDatabaseReader, hash common.Hash, number uint64) *types.Header {
	data, err := db.Get(headerKey(number, hash))
	if err != nil || len(data) == 0 {
		return nil
	}

	header := new(types.Header)
	if err := common.RlpDecodeBytes(data, header); err != nil {
		return nil
	}
	return header
}
