package rawdb

import (
	"bytes"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"frizo-blockchain/storage"
	"github.com/ethereum/go-ethereum/rlp"
)

// ReadCanonicalHash read canonical chain block hash by block number
func ReadCanonicalHash(db storage.IKVStore, number uint64) common.Hash {
	data, err := db.Get(canonicalKey(number))
	if err != nil {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteCanonicalHash
func WriteCanonicalHash(db storage.IKVStore, number uint64, hash common.Hash) error {
	err := db.Put(canonicalKey(number), hash.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func DeleteCanonicalHash(db storage.IKVStore, number uint64) error {
	return db.Delete(canonicalKey(number))
}

// ReadHeader
func ReadHeader(db storage.IKVStore, hash common.Hash, number uint64) *types.Header {
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

func WriteHeader(db storage.IKVStore, header *types.Header) error {
	if header == nil {
		return fmt.Errorf("header is nil")
	}
	encoded, err := rlp.EncodeToBytes(header)
	if err != nil {
		return err
	}

	number := header.Number.Uint64()
	hash := header.Hash()

	return db.Put(headerKey(number, hash), encoded)
}

func DeleteHeader(db storage.IKVStore, hash common.Hash, number uint64) error {
	return db.Delete(headerKey(number, hash))
}

// ReadBody
func ReadBody(db storage.IKVStore, hash common.Hash, number uint64) *types.Body {
	data, err := db.Get(bodyKey(number, hash))
	if err != nil || len(data) == 0 {
		return nil
	}

	body := new(types.Body)
	// rlp decode
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		return nil
	}
	return body
}

func WriteBody(db storage.IKVStore, hash common.Hash, number uint64, body *types.Body) error {
	encoded, err := rlp.EncodeToBytes(body)
	if err != nil {
		return err
	}
	return db.Put(bodyKey(number, hash), encoded)
}

func DeleteBody(db storage.IKVStore, hash common.Hash, number uint64) error {
	return db.Delete(bodyKey(number, hash))
}

func ReadBlock(db storage.IKVStore, hash common.Hash, number uint64) *types.Block {
	// recover header
	header := ReadHeader(db, hash, number)
	if header == nil {
		return nil
	}

	body := ReadBody(db, hash, number)
	if body == nil {
		return nil
	}

	return types.NewBlockWithHeader(header).WithBody(body)
}

func WriteBlock(db storage.IKVStore, block *types.Block) error {
	// write header
	header := block.Header()
	if header == nil {
		return fmt.Errorf("header is nil")
	}
	err := WriteHeader(db, header)
	if err != nil {
		return err
	}

	// write body
	body := block.Body()
	if body == nil {
		return fmt.Errorf("block is nil")
	}
	return WriteBody(db, block.Hash(), block.NumberU64(), body)
}

func DeleteBlock(db storage.IKVStore, hash common.Hash, number uint64) error {
	if err := DeleteHeader(db, hash, number); err != nil {
		return err
	}
	if err := DeleteBody(db, hash, number); err != nil {
		return err
	}
	if err := DeleteReceipts(db, hash, number); err != nil {
		return err
	}
	return nil
}

func ReadReceipts(db storage.IKVStore, hash common.Hash, number uint64) types.Receipts {
	data, err := db.Get(receiptsKey(number, hash))
	if err != nil || len(data) == 0 {
		return nil
	}

	receipts := make(types.Receipts, 0)
	if err := rlp.Decode(bytes.NewReader(data), &receipts); err != nil {
		return nil
	}
	return receipts
}

func WriteReceipts(db storage.IKVStore, hash common.Hash, number uint64, receipts types.Receipts) error {
	encoded, err := rlp.EncodeToBytes(receipts)
	if err != nil {
		return err
	}
	return db.Put(receiptsKey(number, hash), encoded)
}

func DeleteReceipts(db storage.IKVStore, hash common.Hash, number uint64) error {
	return db.Delete(receiptsKey(number, hash))
}

// =================================================================================

// ReadTxLookupEntry
func ReadTxLookupEntry(db storage.IKVStore, hash common.Hash) (common.Hash, uint64, uint64) {
	data, err := db.Get(txLookupKey(hash))
	if err != nil || len(data) == 0 {
		return common.Hash{}, 0, 0
	}

	var entry TxLookupEntry
	if err := rlp.Decode(bytes.NewReader(data), &entry); err != nil {
		return common.Hash{}, 0, 0
	}

	return entry.BlockHash, entry.BlockIndex, entry.Index
}

// WriteTxLookupEntry 寫入交易查詢條目
func WriteTxLookupEntry(db storage.IKVStore, txHash common.Hash, blockHash common.Hash, blockIndex uint64, txIndex uint64) error {
	entry := TxLookupEntry{
		BlockHash:  blockHash,
		BlockIndex: blockIndex,
		Index:      txIndex,
	}

	encoded, err := rlp.EncodeToBytes(entry)
	if err != nil {
		return err
	}

	return db.Put(txLookupKey(txHash), encoded)
}

// DeleteTxLookupEntry 刪除交易查詢條目
func DeleteTxLookupEntry(db storage.IKVStore, txHash common.Hash) error {
	return db.Delete(txLookupKey(txHash))
}

// TxLookupEntry 交易查詢條目
type TxLookupEntry struct {
	BlockHash  common.Hash
	BlockIndex uint64
	Index      uint64
}

// ReadHeadBlockHash 讀取頭區塊哈希
func ReadHeadBlockHash(db storage.IKVStore) common.Hash {
	data, err := db.Get(headBlockKey)
	if err != nil || len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadBlockHash 寫入頭區塊哈希
func WriteHeadBlockHash(db storage.IKVStore, hash common.Hash) error {
	return db.Put(headBlockKey, hash.Bytes())
}

// ReadHeadHeaderHash 讀取頭區塊頭哈希
func ReadHeadHeaderHash(db storage.IKVStore) common.Hash {
	data, err := db.Get(headHeaderKey)
	if err != nil || len(data) == 0 {
		return common.Hash{}
	}
	return common.BytesToHash(data)
}

// WriteHeadHeaderHash 寫入頭區塊頭哈希
func WriteHeadHeaderHash(db storage.IKVStore, hash common.Hash) error {
	return db.Put(headHeaderKey, hash.Bytes())
}
