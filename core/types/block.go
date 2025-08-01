package types

import (
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/crypto"
	"frizo-blockchain/trie"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"sync/atomic"
)

type Receipts []*Receipt

type Transactions []*Transaction

// Len return txn size
func (s Transactions) Len() int { return len(s) }

// GetRlp return index = i txn's RLP encoding
func (s Transactions) GetRlp(i int) ([]byte, error) {
	// 實際應該返回 RLP 編碼
	if i < 0 || i >= len(s) {
		return nil, errors.New("invalid index")
	}
	txn := s[i]
	rawData := []interface{}{
		txn.Hash(),
		txn.Nonce(),
		txn.To(),
		txn.Value(),
		txn.Data(),
		txn.Size(),
		txn.GasLimit(),
		txn.GasPrice(),
	}

	bytes, err := common.RlpEncodeToBytes(rawData)
	if err != nil {
		log.Error("encode txn failed", "err", err)
		panic("encode txn failed")
	}

	return bytes, nil
}

// Block Represent 1 block in blockchain
// Block = BlockHeader + TxnList
// - light-weight client only need block header (merkle root)
// - BlockHeader contains all potential data for verify txn
type Block struct {
	header       *Header
	transactions Transactions
	receipts     Receipts

	// cache
	hash atomic.Value
	size atomic.Value
}

// Header BlockHeader
// include 1 block's all metadata
type Header struct {
	// ParentHash ref to last block (linked list)
	ParentHash common.Hash `json:"parentHash"`
	// Number block height (remark block's position in whole chain)
	Number *big.Int `json:"number"`
	// Timestamp block createTime, also for smart contract usage
	Timestamp uint64 `json:"timestamp"`
	// txn merkle root (transaction trie root)
	TxHashRoot common.Hash `json:"transactionsRoot"`
	// receipt merkle root (receipt trie root)
	ReceiptHashRoot common.Hash `json:"receiptsRoot"`
	// state MPT root (state trie root)
	StateHashRoot common.Hash `json:"stateRoot"`
	// GasLimit gas limit for this block (prevent block oversize, dynamic adjustment)
	GasLimit *big.Int `json:"gasLimit"`
	// GasUsed used gas
	GasUsed *big.Int `json:"gasUsed"`
	// Extra extra data 32 byte (validator can put custom data into this field)
	Extra []byte `json:"extraData"`
	// MixDigest for random number
	MixDigest common.Hash `json:"mixHash"`
	// Nonce for protocol remark
	Nonce uint64 `json:"nonce"`
}

func NewHeader(parentHash common.Hash, number *big.Int, timestamp uint64,
	gasLimit, gasUsed *big.Int,
	extra []byte, mixDigest common.Hash, nonce uint64) *Header {
	return &Header{
		ParentHash: parentHash,
		Number:     number,
		Timestamp:  timestamp,
		GasLimit:   gasLimit,
		GasUsed:    gasUsed,
		Extra:      extra,
		MixDigest:  mixDigest,
		Nonce:      nonce,
	}
}

// NewBlock constructor for block
func NewBlock(header *Header, txs []*Transaction, receipts []*Receipt) *Block {
	b := &Block{header: CopyHeader(header)}

	// Txn tree
	if len(txs) > 0 {
		b.transactions = make(Transactions, len(txs))
		copy(b.transactions, txs)
		b.header.TxHashRoot = b.calculateTxHash()
	}

	// Receipt tree
	if len(receipts) > 0 {
		b.receipts = make(Receipts, len(receipts))
		copy(b.receipts, receipts)
		b.header.ReceiptHashRoot = b.calculateReceiptHash()
	}

	return b
}

// NewBlockWithHeader create block only with blocker header
func NewBlockWithHeader(header *Header) *Block {
	return &Block{header: CopyHeader(header)}
}

// CopyHeader blocker header deep copy
func CopyHeader(h *Header) *Header {
	cpy := *h
	if h.Number != nil {
		cpy.Number = new(big.Int).Set(h.Number)
	}
	if len(h.Extra) > 0 {
		cpy.Extra = make([]byte, len(h.Extra))
		copy(cpy.Extra, h.Extra)
	}
	return &cpy
}

// calculateTxHash calculate txn merkle root
func (b *Block) calculateTxHash() common.Hash {
	if len(b.transactions) == 0 {
		return common.Hash{}
	}

	txHashes := make([][]byte, len(b.transactions))
	for i, tx := range b.transactions {
		txHashes[i] = tx.Hash().Bytes()
	}
	tree := b.GetTxnTree()
	return tree.GetRootHash()
}

// calculateReceiptHash calculate receipt merkle root
func (b *Block) calculateReceiptHash() common.Hash {
	if len(b.receipts) == 0 {
		return common.Hash{}
	}

	tree := b.GetReceiptTree()
	return tree.GetRootHash()
}

// Hash return block's hash (header's hash)
// block hash = block header hash
func (b *Block) Hash() common.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	h := b.header.Hash()
	b.hash.Store(h)
	return h
}

// Hash return header's hash
func (h *Header) Hash() common.Hash {
	rawData := []interface{}{
		h.ParentHash.Hex(),
		h.Number.String(),
		h.Timestamp,
		h.TxHashRoot.Hex(),
		h.ReceiptHashRoot.Hex(),
		h.StateHashRoot.Hex(),
		h.GasLimit,
		h.GasUsed,
	}

	bytes, err := common.RlpEncodeToBytes(rawData)
	if err != nil {
		log.Error("encode header failed", "err", err)
		panic("encode header failed")
	}

	return crypto.Keccak256Hash(bytes)
}

// ========= Block Getters =========

func (b *Block) Header() *Header            { return CopyHeader(b.header) }
func (b *Block) Transactions() Transactions { return b.transactions }
func (b *Block) Number() *big.Int {
	return new(big.Int).Set(b.header.Number)
}
func (b *Block) GasLimit() *big.Int       { return b.header.GasLimit }
func (b *Block) GasUsed() *big.Int        { return b.header.GasUsed }
func (b *Block) Timestamp() uint64        { return b.header.Timestamp }
func (b *Block) ParentHash() common.Hash  { return b.header.ParentHash }
func (b *Block) TxHash() common.Hash      { return b.header.TxHashRoot }
func (b *Block) ReceiptHash() common.Hash { return b.header.ReceiptHashRoot }
func (b *Block) StateRoot() common.Hash   { return b.header.StateHashRoot }
func (b *Block) Extra() []byte            { return common.CopyBytes(b.header.Extra) }

// Size return block size
func (b *Block) Size() uint64 {
	if size := b.size.Load(); size != nil {
		return size.(uint64)
	}

	size := uint64(0)
	size += common.HashLength * 6 // all hash
	size += 8 * 3                 // uint64
	size += uint64(len(b.header.Extra))

	for _, tx := range b.transactions {
		size += tx.Size()
	}

	b.size.Store(size)
	return size
}

// Validate validate block
func (b *Block) Validate() error {
	if b.header == nil {
		return fmt.Errorf("block header is nil")
	}

	if b.header.Number == nil {
		return fmt.Errorf("block number is nil")
	}

	if b.header.Timestamp == 0 {
		return fmt.Errorf("block timestamp is zero")
	}

	// validate gas usage lower than gas limit
	if b.header.GasUsed.Cmp(b.header.GasLimit) > 0 {
		return fmt.Errorf("gas used (%d) exceeds gas limit (%d)",
			b.header.GasUsed, b.header.GasLimit)
	}

	// validate txn hash root
	if len(b.transactions) > 0 {
		calculatedTxHash := b.calculateTxHash()
		if calculatedTxHash != b.header.TxHashRoot {
			return fmt.Errorf("transaction root hash mismatch")
		}
	}

	return nil
}

func (b *Block) String() string {
	return fmt.Sprintf("Block(#%v): Size: %v, Hash: %s, TxCount: %d, GasUsed: %d",
		b.Number(),
		b.Size(),
		b.Hash().Hex()[:8],
		len(b.transactions),
		b.header.GasUsed,
	)
}

func (b *Block) DebugTxnTree() {
	fmt.Println("Txn Tree:")
	b.GetTxnTree().PrintTree()
}

func (b *Block) DebugReceiptTree() {
	fmt.Println("Receipt Tree:")
	b.GetReceiptTree().PrintTree()
}

func (b Block) GetTxnTree() *trie.MerkleTree {
	txHashes := make([][]byte, len(b.transactions))
	for i, tx := range b.transactions {
		txHashes[i] = tx.Hash().Bytes()
	}
	return trie.NewMerkleTree(txHashes, nil) // using default hashFunc (nil)
}

func (b Block) GetReceiptTree() *trie.MerkleTree {
	receiptHashes := make([][]byte, len(b.receipts))
	for i, receipt := range b.receipts {
		receiptHashes[i] = receipt.Hash().Bytes()
	}
	return trie.NewMerkleTree(receiptHashes, nil) // using default hashFunc (nil)
}
