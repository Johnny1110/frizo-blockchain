package common

import (
	"errors"
	"frizo-blockchain/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

// CalcGasLimit calculate new gas limit for new block
// based on parent block gas limit (dynamic adjustment)
func CalcGasLimit(parentGasLimit, desiredLimit uint64) uint64 {
	limit := parentGasLimit

	// calculate delta changing
	delta := parentGasLimit / 1024

	if desiredLimit < parentGasLimit {
		limit = parentGasLimit - delta
	} else {
		limit = parentGasLimit + delta
	}

	// setup min limit
	if limit < BlockMinGasLimit {
		limit = BlockMinGasLimit
	} else if limit > BlockMaxGasLimit {
		limit = BlockMaxGasLimit
	}

	return limit
}

// IntrinsicGas calculate txn gas consume (basic fee + data len fee)
func IntrinsicGas(data []byte, contractCreation bool) (uint64, error) {
	var gas uint64

	if contractCreation {
		gas = TxGasContractCreation
	} else {
		gas = TxGas
	}

	// calculate data len gas
	if len(data) > 0 {
		var nonZeroCount uint64
		for _, byt := range data {
			if byt != 0 {
				nonZeroCount++
			}
		}

		gas += nonZeroCount * TxDataNonZeroGas
		gas += (uint64(len(data)) - nonZeroCount) * TxDataZeroGas
	}

	return gas, nil
}

// DeriveAddress calculate
// contract address = Keccak256(sender address + nonce)[12:]
func DeriveAddress(addr common.Address, nonce uint64) (common.Address, error) {
	rawData := []interface{}{
		addr.Bytes(),
		big.NewInt(int64(nonce)).Bytes(),
	}
	bytes, err := crypto.RlpEncodeToBytes(rawData)
	if err != nil {
		log.Error("Failed to derive address", "err", err)
		return common.Address{}, errors.New("failed to derive address")
	}

	return common.BytesToAddress(crypto.Keccak256(bytes)[12:]), nil
}
