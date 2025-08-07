package types

import (
	"bytes"
	"fmt"
	"frizo-blockchain/common"
	"math/big"
)

// decodeBigInt 解碼為 *big.Int
func decodeBigInt(data interface{}) (*big.Int, error) {
	if data == nil {
		return big.NewInt(0), nil
	}

	switch v := data.(type) {
	case *big.Int:
		return v, nil
	case []byte:
		if len(v) == 0 {
			return big.NewInt(0), nil
		}
		return new(big.Int).SetBytes(v), nil
	case uint64:
		return big.NewInt(int64(v)), nil
	case int64:
		return big.NewInt(v), nil
	case int:
		return big.NewInt(int64(v)), nil
	case uint:
		return big.NewInt(int64(v)), nil
	default:
		return nil, fmt.Errorf("cannot decode type %T to *big.Int", data)
	}
}

// decodeUint64 解碼為 uint64
func decodeUint64(data interface{}) (uint64, error) {
	if data == nil {
		return 0, nil
	}

	switch v := data.(type) {
	case uint64:
		return v, nil
	case uint:
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative value cannot be converted to uint64")
		}
		return uint64(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("negative value cannot be converted to uint64")
		}
		return uint64(v), nil
	case []byte:
		if len(v) == 0 {
			return 0, nil
		}
		if len(v) > 8 {
			return 0, fmt.Errorf("byte slice too long for uint64")
		}
		var result uint64
		for _, b := range v {
			result = (result << 8) | uint64(b)
		}
		return result, nil
	case *big.Int:
		if v.Sign() < 0 {
			return 0, fmt.Errorf("negative value cannot be converted to uint64")
		}
		if v.BitLen() > 64 {
			return 0, fmt.Errorf("value too large for uint64")
		}
		return v.Uint64(), nil
	default:
		return 0, fmt.Errorf("cannot decode type %T to uint64", data)
	}
}

// decodeAddress 解碼為 *common.Address
func decodeAddress(data interface{}) (*common.Address, error) {
	if data == nil {
		return nil, nil
	}

	switch v := data.(type) {
	case []byte:
		if len(v) == 0 {
			return nil, nil
		}
		if len(v) != common.AddressLength {
			return nil, fmt.Errorf("invalid address length: expected %d, got %d",
				common.AddressLength, len(v))
		}
		addr := common.BytesToAddress(v)
		return &addr, nil
	case string:
		if len(v) == 0 {
			return nil, nil
		}
		addr := common.HexToAddress(v)
		return &addr, nil
	case common.Address:
		return &v, nil
	case *common.Address:
		return v, nil
	default:
		return nil, fmt.Errorf("cannot decode type %T to address", data)
	}
}

// decodeBytes 解碼為 []byte
func decodeBytes(data interface{}) ([]byte, error) {
	if data == nil {
		return []byte{}, nil
	}

	switch v := data.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("cannot decode type %T to []byte", data)
	}
}

// === 批量解碼函數（用於解碼多個交易）===

// DecodeTransactions 解碼多個交易
func DecodeTransactions(data []byte) ([]*Transaction, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var encodedTxs [][]byte
	err := common.RlpDecodeBytes(data, &encodedTxs)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transactions list: %w", err)
	}

	txs := make([]*Transaction, len(encodedTxs))
	for i, encodedTx := range encodedTxs {
		tx, err := DecodeToTxn(encodedTx)
		if err != nil {
			return nil, fmt.Errorf("failed to decode transaction %d: %w", i, err)
		}
		txs[i] = tx
	}

	return txs, nil
}

// === 用於測試的輔助函數 ===

// ValidateDecoding
func ValidateDecoding(tx *Transaction) error {
	encoded := tx.Encode()

	decoded, err := DecodeToTxn(encoded)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	if tx.data.AccountNonce != decoded.data.AccountNonce {
		return fmt.Errorf("nonce mismatch: expected %d, got %d",
			tx.data.AccountNonce, decoded.data.AccountNonce)
	}

	if tx.data.GasPrice.Cmp(decoded.data.GasPrice) != 0 {
		return fmt.Errorf("gasPrice mismatch")
	}

	if tx.data.Amount.Cmp(decoded.data.Amount) != 0 {
		return fmt.Errorf("amount mismatch")
	}

	if (tx.data.Recipient == nil) != (decoded.data.Recipient == nil) {
		return fmt.Errorf("recipient nil status mismatch")
	}

	if tx.data.Recipient != nil && *tx.data.Recipient != *decoded.data.Recipient {
		return fmt.Errorf("recipient address mismatch")
	}

	if !bytes.Equal(tx.data.Payload, decoded.data.Payload) {
		return fmt.Errorf("payload mismatch")
	}

	return nil
}

// decodeHash 解碼為 common.Hash
func decodeHash(data interface{}) (common.Hash, error) {
	if data == nil {
		return common.Hash{}, nil
	}

	switch v := data.(type) {
	case []byte:
		if len(v) == 0 {
			return common.Hash{}, nil
		}
		if len(v) != common.HashLength {
			return common.Hash{}, fmt.Errorf("invalid hash length: expected %d, got %d",
				common.HashLength, len(v))
		}
		return common.BytesToHash(v), nil
	case string:
		return common.HexToHash(v), nil
	case common.Hash:
		return v, nil
	default:
		return common.Hash{}, fmt.Errorf("cannot decode type %T to Hash", data)
	}
}

// decodeUint 解碼為 uint
func decodeUint(data interface{}) (uint, error) {
	if data == nil {
		return 0, nil
	}

	switch v := data.(type) {
	case uint:
		return v, nil
	case uint64:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("negative value cannot be converted to uint")
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative value cannot be converted to uint")
		}
		return uint(v), nil
	case []byte:
		if len(v) == 0 {
			return 0, nil
		}
		if len(v) > 8 {
			return 0, fmt.Errorf("byte slice too long for uint")
		}
		var result uint64
		for _, b := range v {
			result = (result << 8) | uint64(b)
		}
		return uint(result), nil
	case *big.Int:
		if v.Sign() < 0 {
			return 0, fmt.Errorf("negative value cannot be converted to uint")
		}
		if v.BitLen() > 64 {
			return 0, fmt.Errorf("value too large for uint")
		}
		return uint(v.Uint64()), nil
	default:
		return 0, fmt.Errorf("cannot decode type %T to uint", data)
	}
}

// decodeBloom 解碼為 Bloom
func decodeBloom(data interface{}) (Bloom, error) {
	if data == nil {
		return Bloom{}, nil
	}

	switch v := data.(type) {
	case []byte:
		if len(v) == 0 {
			return Bloom{}, nil
		}
		if len(v) != common.BloomBytes {
			return Bloom{}, fmt.Errorf("invalid bloom length: expected %d, got %d",
				common.BloomBytes, len(v))
		}
		var bloom Bloom
		copy(bloom[:], v)
		return bloom, nil
	case Bloom:
		return v, nil
	default:
		return Bloom{}, fmt.Errorf("cannot decode type %T to Bloom", data)
	}
}
