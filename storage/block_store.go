package storage

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
)

type BlockStore struct {
	db *ChainDatabase
}

func NewBlockStore(db *ChainDatabase) *BlockStore {
	return &BlockStore{db: db}
}

type TxLookupEntry struct {
	BlockHash  common.Hash
	BlockIndex uint64
	Index      uint64
}

func (bs *BlockStore) Write(block *types.Block) error {
	batch := bs.db.blockDB.NewBatch()

	// 1. write into block
	headerData, err := common.RlpEncodeToBytes(block.Header())
	if err != nil {
		return err
	}
	err = batch.Put(headerKey(block.NumberU64(), block.Hash()), headerData)
	if err != nil {
		return err
	}

	// 2. write into body (body is block's all txn rlp encoded bytes)
	bodyData, err := common.RlpEncodeToBytes(block.RlpEncodeTxns())
	if err != nil {
		return err
	}
	err = batch.Put(bodyKey(block.NumberU64(), block.Hash()), bodyData)
	if err != nil {
		return err
	}

	// 3. receipt
	receiptData, err := common.RlpEncodeToBytes(block.RlpEncodeReceipts())
	if err != nil {
		return err
	}
	err = batch.Put(receiptsKey(block.NumberU64(), block.Hash()), receiptData)
	if err != nil {
		return err
	}

	// 4. update canonical
	err = batch.Put(canonicalKey(block.NumberU64()), block.Hash().Bytes())
	if err != nil {
		return err
	}

	err = batch.Put(headBlockKey, block.Hash().Bytes())
	if err != nil {
		return err
	}
	err = batch.Put(headHeaderKey, block.Hash().Bytes())
	if err != nil {
		return err
	}

	// 5. index
	indexBatch := bs.db.indexDB.NewBatch()
	for i, tx := range block.Transactions() {
		lookupData, _ := common.RlpEncodeToBytes(&TxLookupEntry{
			BlockHash:  block.Hash(),
			BlockIndex: block.NumberU64(),
			Index:      uint64(i),
		})
		err = indexBatch.Put(txLookupKey(tx.Hash()), lookupData)
		if err != nil {
			return err
		}
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

	var encodedBody []interface{}
	err = common.RlpDecodeBytes(bodyData, &encodedBody)
	if err != nil {
		return nil, err
	}
	var body types.Transactions
	for _, encoded := range encodedBody {
		if b, ok := encoded.([]byte); ok {
			if txn, err := types.DecodeToTxn(b); err == nil {
				body = append(body, txn)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 3. read receipts
	receiptData, err := bs.db.blockDB.Get(receiptsKey(number, hash))
	if err != nil {
		return nil, err
	}
	var encodedReceiptsBody []interface{}
	err = common.RlpDecodeBytes(receiptData, &encodedReceiptsBody)
	if err != nil {
		return nil, err
	}
	var receipts types.Receipts
	for _, encoded := range encodedReceiptsBody {
		if b, ok := encoded.([]byte); ok {
			if txn, err := types.DecodeToReceipt(b); err == nil {
				receipts = append(receipts, txn)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
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

	err := batch.Delete(headerKey(number, hash))
	if err != nil {
		return err
	}
	err = batch.Delete(bodyKey(number, hash))
	if err != nil {
		return err
	}
	err = batch.Delete(receiptsKey(number, hash))
	if err != nil {
		return err
	}

	return batch.Write()
}
