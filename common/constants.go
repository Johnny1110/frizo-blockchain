package common

import "math/big"

// Blockchain Constants
const (
	// Block
	BlockMaxGasLimit    = 8000000 // Block Gas Limit
	BlockMinGasLimit    = 5000    // min gas fee for 1 block
	MaxBlockSize        = 1048576 // Block max size (1MB)
	BlockGenerationTime = 15      // Block Generate per（Sec）

	// TXN
	MaxTransactionSize = 32768      // Max txn size (32KB)
	MinGasPrice        = 1000000000 // Min Gas price (1 Gwei)
	SignatureLen       = 65         // sgin len (RSV)

	// Network
	DefaultP2PPort = 30303 // Default P2P port
	DefaultRPCPort = 8545  // Default RPC port
	MaxPeers       = 50    // Max connected peer count

	// DB
	DatabaseCache   = 768 // database cache (MB)
	DatabaseHandles = 1024

	// Version info
	ProtocolVersion = 1 // protocol version
	DatabaseVersion = 1 // DB version

	// BloomBits Bloom filter decimals
	BloomBits = 2048

	// BloomBytes Bloom filter bytes size
	BloomBytes = BloomBits / 8

	// MaxExtraDataSize max extra data size
	MaxExtraDataSize = 32

	// TxGas normal value transfer gas fee
	TxGas = 21000

	// TxGasContractCreation create contract gas fee
	TxGasContractCreation = 53000

	// TxDataZeroGas txn data every zero byte gas fee
	TxDataZeroGas = 4

	// TxDataNonZeroGas txn data every non-zero byte gas fee
	TxDataNonZeroGas = 16
)

// Genesis
var (
	// GenesisDifficulty
	GenesisDifficulty = big.NewInt(131072)

	// GenesisGasLimit
	GenesisGasLimit = uint64(4700000)

	// GenesisNonce
	GenesisNonce = uint64(42)

	// GenesisTimestamp（2024-01-01 00:00:00 UTC）
	GenesisTimestamp = uint64(1704067200)
)

// Reward (create block & staking)
var (
	// BlockReward (5 FRZ)
	BlockReward = big.NewInt(5e18)

	// MinimumStake (32 FRZ)
	MinimumStake = new(big.Int).Mul(
		big.NewInt(32),   // 32
		big.NewInt(1e18), // 10^18
	)
)

// unit
var (
	// Wei is the smallest unit
	Wei = big.NewInt(1)

	// GWei is 10^9 Wei
	GWei = big.NewInt(1e9)

	// Ether is 10^18 Wei (1 FRZ = 10^18 Wei)
	Ether = big.NewInt(1e18)
)
