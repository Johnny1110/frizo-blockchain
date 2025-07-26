package types

import (
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

// Receipt represent txn's Receipt
// - log txn execute result（success/failed）
// - log consumed gas
// - store emit event（for DApp query）
// - provide proof of txn
type Receipt struct {
	Type uint8 `json:"type,omitempty"`

	// PostState state root after exec
	PostState []byte `json:"root"`

	// Status exec status（1=success, 0=failed）
	Status uint64 `json:"status"`

	// CumulativeGasUsed (區塊中到這筆交易為止的累計 gas 使用量)
	CumulativeGasUsed uint64 `json:"cumulativeGasUsed"`

	// Bloom bloom filter
	// - for filter log and fast query log
	Bloom Bloom `json:"logsBloom"`

	// Logs txn log emit
	Logs []*Log `json:"logs"`

	// TxHash tx hash -> transaction
	TxHash common.Hash `json:"transactionHash"`

	// ContractAddress created contract address（nil of not a contract creation）
	ContractAddress *common.Address `json:"contractAddress,omitempty"`

	// GasUsed consumed gas
	GasUsed uint64 `json:"gasUsed"`

	// BlockHash block hash
	BlockHash common.Hash `json:"blockHash,omitempty"`

	// BlockNumber block number
	BlockNumber *big.Int `json:"blockNumber,omitempty"`

	// TransactionIndex transaction index in block
	TransactionIndex uint `json:"transactionIndex"`
}

// Log transaction log (for contract emit or transfer value log)
type Log struct {
	// Address who create this log
	Address common.Address `json:"address"`

	// Topics log topic list (max len is 4)
	// - Topics[0] event sign hash
	// - Topics[1-3] index params（searchable）
	// -
	Topics []common.Hash `json:"topics"`

	// Data non-index log data (store event non-index param)
	Data []byte `json:"data"`

	// for query usage:
	BlockNumber uint64      `json:"blockNumber,omitempty"`
	TxHash      common.Hash `json:"transactionHash,omitempty"`
	TxIndex     uint        `json:"transactionIndex,omitempty"`
	BlockHash   common.Hash `json:"blockHash,omitempty"`
	Index       uint        `json:"logIndex,omitempty"`

	Removed bool `json:"removed,omitempty"`
}

// Bloom 2048 decimal bloom filter
type Bloom [common.BloomBytes]byte

// NewReceipt create new receipt
func NewReceipt(root []byte, failed bool, cumulativeGasUsed uint64) *Receipt {
	r := &Receipt{
		PostState:         root,
		CumulativeGasUsed: cumulativeGasUsed,
	}
	if failed {
		r.Status = 0
	} else {
		r.Status = 1
	}
	return r
}

// Hash Receipt hash
func (r *Receipt) Hash() common.Hash {
	rawData := []interface{}{
		r.Status,
		r.CumulativeGasUsed,
		r.TxHash.Hex(),
		r.GasUsed,
	}

	bytes, err := crypto.RlpEncodeToBytes(rawData)
	if err != nil {
		log.Error("Failed to encode receipt", "err", err)
		return common.Hash{}
	}

	return crypto.Keccak256Hash(bytes)
}

// Failed is failed
func (r *Receipt) Failed() bool {
	return r.Status == 0
}

// Success is success
func (r *Receipt) Success() bool {
	return r.Status == 1
}

func (r *Receipt) String() string {
	status := "Success"
	if r.Failed() {
		status = "Failed"
	}
	return fmt.Sprintf("Receipt(%s): Status=%s, GasUsed=%d, Logs=%d",
		r.TxHash.Hex()[:8],
		status,
		r.GasUsed,
		len(r.Logs),
	)
}

// ========= bloom filter func ========= TODO: study bloom filter.

// Add 向布隆過濾器添加數據
func (b *Bloom) Add(data []byte) {
	// 布隆過濾器實現
	// 使用多個哈希函數將數據映射到位圖中
	hash := crypto.Keccak256(data)
	for i := 0; i < 3; i++ {
		bit := (uint(hash[i]) + (uint(hash[i+1]) << 8)) & 2047
		b[bit/8] |= 1 << (bit % 8)
	}
}

// Contains 檢查布隆過濾器是否可能包含數據
func (b *Bloom) Contains(data []byte) bool {
	hash := crypto.Keccak256(data)
	for i := 0; i < 3; i++ {
		bit := (uint(hash[i]) + (uint(hash[i+1]) << 8)) & 2047
		if b[bit/8]&(1<<(bit%8)) == 0 {
			return false
		}
	}
	return true
}

// CreateBloom 從收據列表創建布隆過濾器
func CreateBloom(receipts []*Receipt) Bloom {
	var bloom Bloom
	for _, receipt := range receipts {
		bloom.Or(&receipt.Bloom)
	}
	return bloom
}

// Or 執行布隆過濾器的或操作
func (b *Bloom) Or(other *Bloom) {
	for i := range b {
		b[i] |= other[i]
	}
}

// LogsBloom 從日誌列表創建布隆過濾器
func LogsBloom(logs []*Log) Bloom {
	var bloom Bloom
	for _, log := range logs {
		bloom.Add(log.Address.Bytes())
		for _, topic := range log.Topics {
			bloom.Add(topic.Bytes())
		}
	}
	return bloom
}
