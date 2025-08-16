package blockchain

import (
	"frizo-blockchain/common"
	"frizo-blockchain/core/types"
	"math/big"
)

type Genesis struct {
	Timestamp  uint64
	GasLimit   *big.Int
	Difficulty *big.Int
	Alloc      map[common.Address]GenesisAccount

	ExtraData []byte
}

func (g *Genesis) ToBlock() *types.Block {
	header := &types.Header{
		ParentHash: common.Hash{},
		Number:     big.NewInt(0),
		Timestamp:  g.Timestamp,
		GasLimit:   g.GasLimit,
		GasUsed:    big.NewInt(0),
		Extra:      g.ExtraData,
	}

	return types.NewBlock(header, nil, nil)
}

type GenesisAccount struct {
	Balance *big.Int
	Nonce   uint64
	Code    []byte                      // TODO: Phase-1 not support
	Storage map[common.Hash]common.Hash // TODO: Phase-1 not support
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

// DefaultGenesis default genesis setup
func DefaultGenesis() *Genesis {
	return &Genesis{
		Timestamp:  1640995200, // 2022-01-01 00:00:00
		GasLimit:   big.NewInt(5000000),
		Difficulty: big.NewInt(1),
		Alloc: map[common.Address]GenesisAccount{
			// pre alloc some test account
			common.HexToAddress("0x521147b68d948f24A341E5C0Da50Fa5AA39A06D7"): {
				Balance: new(big.Int).Mul(big.NewInt(1000), common.Ether), // 1000 ETH
				Nonce:   1,
			},
			common.HexToAddress("0xA2D969E82524001Cb6a2357dBF5922B04aD2FCD8"): {
				Balance: new(big.Int).Mul(big.NewInt(100), common.Ether), // 100 ETH
				Nonce:   0,
			},
			// Kai's address
			common.HexToAddress("0x08bC5CB835c658f71A551c241709535Bcd519059"): {
				Balance: new(big.Int).Mul(big.NewInt(100), common.Ether), // 100 ETH
				Nonce:   0,
			},
			// Harry's address
			common.HexToAddress("0xaEc403Ac73a363e007dF97513F547E18189EaDaB"): {
				Balance: new(big.Int).Mul(big.NewInt(100), common.Ether), // 100 ETH
				Nonce:   0,
			},
		},
	}
}
