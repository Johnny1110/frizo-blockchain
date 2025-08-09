package state

import (
	"frizo-blockchain/common"
	"math/big"
)

// Constants
var (
	// EmptyRoot is the root hash of an empty trie
	EmptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// emptyCodeHash is the hash of empty code
	emptyCodeHash = common.Keccak256Hash(nil).Bytes()
)

// State Account
type Account struct {
	Nonce   uint64
	Balance *big.Int

	// contract address usage
	Root     common.Hash // merkle root of the storage trie
	CodeHash []byte
}

func NewAccount() Account {
	return Account{
		Balance:  big.NewInt(0),
		Nonce:    0,
		Root:     EmptyRoot,
		CodeHash: emptyCodeHash,
	}
}

func (a *Account) Copy() *Account {
	return &Account{
		Nonce:    a.Nonce,
		Balance:  new(big.Int).Set(a.Balance),
		Root:     a.Root,
		CodeHash: common.CopyBytes(a.CodeHash),
	}
}

// EncodeAccount with RLP
func (a *Account) EncodeAccount() ([]byte, error) {
	return common.RlpEncodeToBytes(a)
}

func DecodeAccount(data []byte) (*Account, error) {
	var account Account
	if err := common.RlpDecodeBytes(data, &account); err != nil {
		return nil, err
	}
	return &account, nil
}
