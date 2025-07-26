package types

import (
	"crypto/ecdsa"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"sync/atomic"
	"time"
)

// Transaction represent 1 txn in 1 block
// Each Transaction is 1 value transfer or 1 contract call
type Transaction struct {
	// txn basic data
	data txdata

	// cache value（optimized performance）
	hash atomic.Value // tx_hash
	size atomic.Value // tx_size
	from atomic.Value // from address
}

// txdata txn core data structure
// support serialize and un-serialize
type txdata struct {
	// sender's txn nonce
	AccountNonce uint64 `json:"nonce"`

	// Price per gas unit（wei）
	GasPrice *big.Int `json:"gasPrice"`

	// GasLimit this txn max gas usage (prevent infinite loop and control cost)
	GasLimit uint64 `json:"gas"`

	// Recipient receiver address（nil is create contract）
	// - nil create contract
	// - non-nil value transfer or contract call
	Recipient *common.Address `json:"to"`

	// Amount txn amount/value（wei）
	Amount *big.Int `json:"value"`

	// Payload txn data
	// - value transfer: usually empty
	// - contract creation: contract bytecode
	// - contract call: func selector and params
	Payload []byte `json:"input"`

	// Signature values
	// - standard ECDSA signature format
	// - V for Ecrecover
	// - R, main part of signature
	signature crypto.Signature `json:"signature"`

	// txn create time (only for debug)
	Time time.Time `json:"time"`
}

// ========= Transaction Constructor =========

// NewTransaction create a new TXN
func NewTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if amount == nil {
		amount = new(big.Int)
	}
	if gasPrice == nil {
		gasPrice = new(big.Int)
	}

	// only create txdata
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Amount:       new(big.Int).Set(amount),
		GasLimit:     gasLimit,
		GasPrice:     new(big.Int).Set(gasPrice),
		Payload:      data,
		Time:         time.Now(),
	}

	return &Transaction{data: d}
}

// NewContractCreation create a new contract creation txn
// to address is nil
func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return NewTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

// ========= Transaction Getters =========
func (tx *Transaction) Nonce() uint64       { return tx.data.AccountNonce }
func (tx *Transaction) GasPrice() *big.Int  { return new(big.Int).Set(tx.data.GasPrice) }
func (tx *Transaction) GasLimit() uint64    { return tx.data.GasLimit }
func (tx *Transaction) To() *common.Address { return tx.data.Recipient }
func (tx *Transaction) Value() *big.Int     { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Data() []byte        { return tx.data.Payload }
func (tx *Transaction) Time() time.Time     { return tx.data.Time }

// ========= Transaction Func =========

// Hash return tx hash
// - for txn's unique ID, build merkle tree and query index
func (tx *Transaction) Hash() common.Hash {
	// load from cache
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}

	bytes, _ := crypto.RlpEncodeToBytes(tx.data.AccountNonce)
	h := crypto.Keccak256Hash(bytes)

	// cache
	tx.hash.Store(h)
	return h
}

// SignTx use private to sign txn
// 1. calculate tx_hash
// 2. sign with private key
// 3. store V, R, S
func (tx *Transaction) SignTx(privateKey *ecdsa.PrivateKey) error {
	// 1. calculate tx_hash
	h := tx.signingHash()

	// 2. sign with private key
	sig, err := crypto.SignMessage(privateKey, h[:])
	if err != nil {
		return err
	}

	if !sig.Validate() {
		return common.ErrInvalidSignature
	}
	// 3 store V, R, S
	tx.data.signature = sig

	return nil
}

// signingHash return tx_hash (AccountNonce + Recipient + Amount + GasLimit + Payload)
func (tx *Transaction) signingHash() common.Hash {
	bytes, _ := crypto.RlpEncodeToBytes(fmt.Sprintf("%d,%s,%d,%d,%s,%x",
		tx.data.AccountNonce,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.GasLimit,
		tx.data.GasPrice,
		tx.data.Payload,
	))
	return crypto.Keccak256Hash(bytes)
}

// VerifySignature validate signature, return bool
func (tx *Transaction) VerifySignature() bool {
	_, err := tx.Sender()
	return err == nil
}

// Sender return txn sender address by sign verify
func (tx *Transaction) Sender() (common.Address, error) {
	if from := tx.from.Load(); from != nil {
		return from.(common.Address), nil
	}

	if !tx.data.signature.Validate() {
		log.Warn("[types][Sender] failed to perform signature Validate")
		return common.Address{}, common.ErrInvalidSignature
	}

	h := tx.signingHash()

	// restore address from SRV
	pubKey, err := crypto.Ecrecover(h[:], tx.data.signature)
	if err != nil {
		log.Warn("[types][Sender] failed to perform Ecrecover", "err", err)
		return common.Address{}, common.ErrInvalidSignature
	}

	if !crypto.VerifySignature(pubKey, h[:], tx.data.signature) {
		log.Warn("[types][Sender] failed to perform VerifySignature", "err", err)
		return common.Address{}, common.ErrInvalidSignature
	}

	address := crypto.PubkeyToAddress(pubKey)

	tx.from.Store(address)

	return address, nil
}

// Cost return txn cost
// Cost = transfer amount + (gas limit × gas price)
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.GasPrice, new(big.Int).SetUint64(tx.data.GasLimit))
	total.Add(total, tx.data.Amount)
	return total
}

// IsContractCreation is contract creation
func (tx *Transaction) IsContractCreation() bool {
	return tx.data.Recipient == nil
}

// Size return txn size（for restrict block size）
func (tx *Transaction) Size() uint64 {
	if size := tx.size.Load(); size != nil {
		return size.(uint64)
	}

	// TODO: revamp this
	size := uint64(32 + 32 + 8 + 8 + 32 + len(tx.data.Payload) + 96) // simplify
	tx.size.Store(size)
	return size
}

func (tx *Transaction) String() string {
	to := "contract creation"
	if tx.data.Recipient != nil {
		to = tx.data.Recipient.Hex()
	}

	return fmt.Sprintf("TX(%s): nonce=%d, to=%s, value=%s, gas=%d, gasPrice=%s, dataLen=%d",
		fmt.Sprintf("...%s", tx.Hash().Hex()[:8]),
		tx.data.AccountNonce,
		to,
		tx.data.Amount,
		tx.data.GasLimit,
		tx.data.GasPrice,
		len(tx.data.Payload),
	)
}

// TxByNonce all txn order by nonce
type TxByNonce []*Transaction

func (s TxByNonce) Len() int      { return len(s) }
func (s TxByNonce) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TxByNonce) Less(i, j int) bool {
	return s[i].Nonce() < s[j].Nonce()
}

// TxByPrice order by gas price
type TxByPrice []*Transaction

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].data.GasPrice.Cmp(s[j].data.GasPrice) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
