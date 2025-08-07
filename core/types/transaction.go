package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"frizo-blockchain/common"
	"frizo-blockchain/crypto"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"sync/atomic"
)

// Transaction represent 1 txn in 1 block
// Each Transaction is 1 value transfer or 1 contract call
type Transaction struct {
	enableCache bool

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
	chainId *big.Int
	// sender's txn nonce
	AccountNonce uint64 `json:"nonce"`

	// Price per gas unit（wei）
	GasPrice *big.Int `json:"gasPrice"`

	// GasLimit this txn max gas usage (prevent infinite loop and control cost)
	GasLimit *big.Int `json:"gasLimit"`
	GasCost  *big.Int `json:"gasCost"`

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
	Signature crypto.Signature `json:"signature"`
}

// ========= Transaction Constructor =========

// NewTransaction create a new TXN
func NewTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit *big.Int, gasPrice *big.Int, data []byte) *Transaction {
	if amount == nil {
		amount = new(big.Int)
	}
	if gasPrice == nil {
		gasPrice = new(big.Int)
	}

	// only create txdata
	d := txdata{
		chainId:      big.NewInt(common.ChainID),
		AccountNonce: nonce,
		Recipient:    to,
		Amount:       new(big.Int).Set(amount),
		GasLimit:     gasLimit,
		GasCost:      new(big.Int).SetUint64(common.TxGas), // default txn gas fee
		GasPrice:     new(big.Int).Set(gasPrice),
		Payload:      data,
	}

	return &Transaction{data: d, enableCache: common.TxnCacheSwitch}
}

// NewContractCreation create a new contract creation txn
// to address is nil
func NewContractCreation(nonce uint64, amount *big.Int, gasLimit *big.Int, gasPrice *big.Int, data []byte) *Transaction {
	return NewTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

// ========= Transaction Getters =========
func (tx *Transaction) Nonce() uint64       { return tx.data.AccountNonce }
func (tx *Transaction) GasPrice() *big.Int  { return new(big.Int).Set(tx.data.GasPrice) }
func (tx *Transaction) GasCost() *big.Int   { return new(big.Int).Set(tx.data.GasCost) }
func (tx *Transaction) GasLimit() *big.Int  { return tx.data.GasLimit }
func (tx *Transaction) To() *common.Address { return tx.data.Recipient }
func (tx *Transaction) Value() *big.Int     { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Data() []byte        { return tx.data.Payload }
func (tx *Transaction) From() (common.Address, error) {
	return tx.Sender()
}
func (tx *Transaction) ChainID() *big.Int {
	return new(big.Int).Set(tx.data.chainId)
}

// ========= Transaction Func =========

// Hash return tx hash
// - for txn's unique ID, build merkle tree and query index
func (tx *Transaction) Hash() common.Hash {
	// load from cache
	if hash := tx.hash.Load(); tx.enableCache && hash != nil {
		return hash.(common.Hash)
	}
	rawData := []interface{}{
		tx.data.chainId,
		tx.data.AccountNonce,
		tx.data.GasPrice,
		tx.data.GasLimit,
		tx.data.GasCost,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
	}
	bytes, _ := common.RlpEncodeToBytes(rawData)
	h := crypto.Keccak256Hash(bytes)
	tx.hash.Store(h)
	return h
}

// SignTx use private to sign txn
// 1. calculate tx_hash
// 2. sign with private key
// 3. store V, R, S
func (tx *Transaction) SignTx(privateKey *ecdsa.PrivateKey) error {
	// 1. calculate tx_hash
	h := tx.Hash()

	// 2. sign with private key
	sig, err := crypto.SignMessage(privateKey, h.Bytes()[:])
	if err != nil {
		return err
	}

	if !sig.Validate() {
		return common.ErrInvalidSignature
	}
	// 3 store V, R, S
	tx.data.Signature = sig

	return nil
}

// VerifySignature validate signature, return bool
// unable to check：
// - 1. Account Nonce（need account state）
// - 2. Account Balance（need account state）
// - 3. Gas（need executed）
func (tx *Transaction) VerifySignature() bool {
	_, err := tx.Sender()
	return err == nil
}

// Sender return txn sender address by sign verify
func (tx *Transaction) Sender() (common.Address, error) {
	if from := tx.from.Load(); tx.enableCache && from != nil {
		return from.(common.Address), nil
	}

	if !tx.data.Signature.Validate() {
		log.Warn("[types][Sender] failed to perform signature Validate")
		return common.Address{}, common.ErrInvalidSignature
	}

	h := tx.Hash()

	// restore address from SRV
	pubKey, err := crypto.Ecrecover(h[:], tx.data.Signature)

	if err != nil {
		log.Warn("[types][Sender] failed to perform Ecrecover", "err", err)
		return common.Address{}, common.ErrInvalidSignature
	}

	if !crypto.VerifySignature(pubKey, h[:], tx.data.Signature) {
		log.Warn("[types][Sender] failed to perform VerifySignature", "err", err)
		return common.Address{}, common.ErrInvalidSignature
	}

	address := crypto.PubkeyToAddress(pubKey)
	tx.from.Store(address)

	return address, nil
}

// TotalCost return txn cost
// TotalCost = transfer amount + (gas limit × gas price)
func (tx *Transaction) TotalCost() *big.Int {
	total := new(big.Int).Mul(tx.data.GasPrice, tx.data.GasCost)
	total.Add(total, tx.data.Amount)
	return total
}

// IsContractCreation is contract creation
func (tx *Transaction) IsContractCreation() bool {
	return tx.data.Recipient == nil
}

// Size return txn size（for restrict block size）
func (tx *Transaction) Size() uint64 {
	if size := tx.size.Load(); tx.enableCache && size != nil {
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

func (tx *Transaction) SetValue(value *big.Int) {
	tx.data.Amount = value
}

// Encode enocde to RLP
func (tx *Transaction) Encode() []byte {
	rawData := []interface{}{
		tx.data.chainId,
		tx.data.AccountNonce,
		tx.data.GasPrice,
		tx.data.GasLimit,
		tx.data.GasCost,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
		tx.data.Signature.Bytes(),
	}
	bytes, err := common.RlpEncodeToBytes(rawData)
	if err != nil {
		panic(err)
	}
	return bytes
}

// DecodeToTxn decode rlp bytes to TXN
func DecodeToTxn(encoded []byte) (*Transaction, error) {
	if len(encoded) == 0 {
		return nil, errors.New("empty encoded data")
	}

	var decoded []interface{}
	err := common.RlpDecodeBytes(encoded, &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RLP: %w", err)
	}

	// verify len
	if len(decoded) != 9 {
		return nil, fmt.Errorf("invalid transaction format: expected 10 fields, got %d", len(decoded))
	}
	tx := &Transaction{
		data: txdata{},
	}

	// 1.  chainId
	chainId, err := decodeBigInt(decoded[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode chainId: %w", err)
	}
	tx.data.chainId = chainId

	// 2.  AccountNonce
	nonce, err := decodeUint64(decoded[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode nonce: %w", err)
	}
	tx.data.AccountNonce = nonce

	// 3.  GasPrice
	gasPrice, err := decodeBigInt(decoded[2])
	if err != nil {
		return nil, fmt.Errorf("failed to decode gasPrice: %w", err)
	}
	tx.data.GasPrice = gasPrice

	// 4.  GasLimit
	gasLimit, err := decodeBigInt(decoded[3])
	if err != nil {
		return nil, fmt.Errorf("failed to decode gasLimit: %w", err)
	}
	tx.data.GasLimit = gasLimit

	// 5.  GasCost
	gasCost, err := decodeBigInt(decoded[4])
	if err != nil {
		return nil, fmt.Errorf("failed to decode gasCost: %w", err)
	}
	tx.data.GasCost = gasCost

	// 6.  Recipient (可能為 nil)
	recipient, err := decodeAddress(decoded[5])
	if err != nil {
		return nil, fmt.Errorf("failed to decode recipient: %w", err)
	}
	tx.data.Recipient = recipient

	// 7.  Amount
	amount, err := decodeBigInt(decoded[6])
	if err != nil {
		return nil, fmt.Errorf("failed to decode amount: %w", err)
	}
	tx.data.Amount = amount

	// 8.  Payload
	payload, err := decodeBytes(decoded[7])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	tx.data.Payload = payload

	// 9.  Signature
	sigBytes, err := decodeBytes(decoded[8])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}
	if len(sigBytes) > 0 {
		signature, err := crypto.NewSignature(sigBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signature: %w", err)
		}
		tx.data.Signature = signature
	}

	// calculate hash
	tx.Hash()

	return tx, nil
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
