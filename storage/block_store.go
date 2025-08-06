package storage

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
)

type BlockStore struct {
	db *ChainDatabase
}

type TxLookupEntry struct {
	BlockHash  common.Hash
	BlockIndex uint64
	Index      uint64
}

func (bs *BlockStore) Write(block *types.Block, receipts types.Receipts) error {
	batch := bs.db.blockDB.NewBatch()

	// 1. write into block
	headerData, err := common.RlpEncodeToBytes(block.Header())
	if err != nil {
		return err
	}
	batch.Put(headerKey(block.NumberU64(), block.Hash()), headerData)

	// 2. write into body
	bodyData, err := common.RlpEncodeToBytes(block.Body())
	if err != nil {
		return err
	}
	batch.Put(bodyKey(block.NumberU64(), block.Hash()), bodyData)

	// 3. receipt
	receiptData, err := common.RlpEncodeToBytes(receipts)
	if err != nil {
		return err
	}
	batch.Put(receiptsKey(block.NumberU64(), block.Hash()), receiptData)

	// 4. update canonical
	batch.Put(canonicalKey(block.NumberU64()), block.Hash().Bytes())

	batch.Put(headBlockKey, block.Hash().Bytes())
	batch.Put(headHeaderKey, block.Hash().Bytes())

	// 5. index
	indexBatch := bs.db.indexDB.NewBatch()
	for i, tx := range block.Transactions() {
		lookupData, _ := common.RlpEncodeToBytes(&TxLookupEntry{
			BlockHash:  block.Hash(),
			BlockIndex: block.NumberU64(),
			Index:      uint64(i),
		})
		indexBatch.Put(txLookupKey(tx.Hash()), lookupData)
	}

	// exec batch write
	if err := batch.Write(); err != nil {
		return err
	}
	if err := indexBatch.Write(); err != nil {
		return err
	}

	return nil
}

// ReadBlock read block from BlockStore
func (bs *BlockStore) ReadBlock(hash common.Hash, number uint64) (*types.Block, error) {
	// 1. read block header
	headerData, err := bs.db.blockDB.Get(headerKey(number, hash))
	if err != nil {
		return nil, err
	}
	var header types.Header
	if err := common.RlpDecodeBytes(headerData, &header); err != nil {
		return nil, err
	}

	// 2. read block body
	bodyData, err := bs.db.blockDB.Get(bodyKey(number, hash))
	if err != nil {
		return nil, err
	}
	var body types.Transactions
	if err := common.RlpDecodeBytes(bodyData, &body); err != nil {
		return nil, err
	}

	// 3. read receipts
	receiptData, err := bs.db.blockDB.Get(receiptsKey(number, hash))
	if err != nil {
		return nil, err
	}
	var receipts types.Receipts
	if err := common.RlpDecodeBytes(receiptData, &receipts); err != nil {
		return nil, err
	}

	// 4. make block
	block := types.NewBlock(&header, body, receipts)

	return block, nil
}

// ReadReceipts read receipts from block store
func (bs *BlockStore) ReadReceipts(hash common.Hash, number uint64) (types.Receipts, error) {
	data, err := bs.db.blockDB.Get(receiptsKey(number, hash))
	if err != nil {
		return nil, err
	}

	var receipts types.Receipts
	if err := common.RlpDecodeBytes(data, &receipts); err != nil {
		return nil, err
	}

	return receipts, nil
}

// DeleteBlock delete block
func (bs *BlockStore) DeleteBlock(hash common.Hash, number uint64) error {
	batch := bs.db.blockDB.NewBatch()

	batch.Delete(headerKey(number, hash))
	batch.Delete(bodyKey(number, hash))
	batch.Delete(receiptsKey(number, hash))

	return batch.Write()
}
