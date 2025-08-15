package common

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

const (
	// HashLength is the expected length of the hash
	HashLength = 32
	// AddressLength is the expected length of the address
	AddressLength = 20
)

// Hash represents the 32 byte Keccak256 hash of arbitrary data.
type Hash [HashLength]byte

// Address represents the 20 byte address of an Ethereum account.
type Address [AddressLength]byte

// BytesToHash sets b to hash.
// If b is larger than HashLength, b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

// BigToHash sets byte representation of b to hash.
// If b is larger than HashLength, b will be cropped from the left.
func BigToHash(b *big.Int) Hash {
	return BytesToHash(b.Bytes())
}

// HexToHash sets byte representation of s to hash.
// If s is larger than HashLength, s will be cropped from the left.
func HexToHash(s string) Hash {
	return BytesToHash(FromHex(s))
}

// SetBytes sets the hash to the value of b.
// If b is larger than HashLength, b will be cropped from the left.
func (h *Hash) SetBytes(b []byte) {
	if len(b) > HashLength {
		b = b[len(b)-HashLength:]
	}
	copy(h[HashLength-len(b):], b)
}

// Bytes gets the byte representation of the hash.
func (h Hash) Bytes() []byte {
	return h[:]
}

// Big converts a hash to a big integer.
func (h Hash) Big() *big.Int {
	return new(big.Int).SetBytes(h[:])
}

// Hex converts a hash to a hex string.
func (h Hash) Hex() string {
	return "0x" + hex.EncodeToString(h[:])
}

// String implements the stringer interface and is used also by the logger.
func (h Hash) String() string {
	return h.Hex()
}

// IsZero returns true if the hash is all zeros.
func (h Hash) IsZero() bool {
	return h == Hash{}
}

// BytesToAddress returns Address with value b.
// If b is larger than AddressLength, b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

// BigToAddress returns Address with byte values of b.
// If b is larger than AddressLength, the address will be cropped from the left.
func BigToAddress(b *big.Int) Address {
	return BytesToAddress(b.Bytes())
}

// HexToAddress returns Address with byte values of s.
// If s is larger than AddressLength, it will be cropped from the left.
func HexToAddress(s string) Address {
	return BytesToAddress(FromHex(s))
}

// SetBytes sets the address to the value of b.
// If b is larger than AddressLength, b will be cropped from the left.
func (a *Address) SetBytes(b []byte) {
	if len(b) > AddressLength {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// Bytes gets the byte representation of the address.
func (a Address) Bytes() []byte {
	return a[:]
}

// Big converts an address to a big integer.
func (a Address) Big() *big.Int {
	return new(big.Int).SetBytes(a[:])
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	return "0x" + hex.EncodeToString(a[:])
}

// String implements the stringer interface and is used also by the logger.
func (a Address) String() string {
	return a.Hex()
}

// IsZero returns true if the address is all zeros.
func (a Address) IsZero() bool {
	return a == Address{}
}

// Format implements fmt.Formatter.
func (a Address) Format(s fmt.State, c rune) {
	switch c {
	case 'v', 's':
		fmt.Fprintf(s, "%s", a.Hex())
	case 'q':
		fmt.Fprintf(s, "%q", a.Hex())
	case 'x', 'X':
		fmt.Fprintf(s, "%"+string(c), a[:])
	default:
		fmt.Fprintf(s, "%"+string(c), a[:])
	}
}

// Format implements fmt.Formatter.
func (h Hash) Format(s fmt.State, c rune) {
	switch c {
	case 'v', 's':
		fmt.Fprintf(s, "%s", h.Hex())
	case 'q':
		fmt.Fprintf(s, "%q", h.Hex())
	case 'x', 'X':
		fmt.Fprintf(s, "%"+string(c), h[:])
	default:
		fmt.Fprintf(s, "%"+string(c), h[:])
	}
}

// GenesisAccount Genesis Account
type GenesisAccount struct {
	Address common.Address
	Balance *big.Int
	Code    []byte
	Storage map[common.Hash]common.Hash
	Nonce   uint64
}

// GenesisBlock GenesisBlock config
type GenesisBlock struct {
	Timestamp  uint64
	ParentHash common.Hash
	ExtraData  []byte
	GasLimit   uint64

	// Alloc pre set account
	Alloc map[common.Address]GenesisAccount

	// chain config
	Config *ChainConfig
}

// ChainConfig Chain Config
type ChainConfig struct {
	ChainID *big.Int // Chain ID, for sign txn

	// Genesis
	GenesisHash   common.Hash
	ConsensusType string // "pos" or "poa"

	// Block
	BlockTime    time.Duration
	MaxBlockSize uint64
	MaxBlockGas  uint64

	// Transaction
	MinGasPrice *big.Int
	MaxTxSize   uint64

	// PoS setup
	MinStake          *big.Int // min stake value (32 or 128)
	ValidatorSetSize  uint64   // Validator Set Size
	BlockPeriod       uint64   // block creation interval（secs）
	EpochLength       uint64   // epoch length（block count）
	FinalizationDelay uint64   // finalized confirmation delay

	// Gas setup
	InitialGasLimit uint64 // init gas limit
}
