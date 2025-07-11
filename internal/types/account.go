package types

import (
	"math/big"
)

// Account represents a blockchain account
type Account struct {
	Address     *Address `json:"address"`
	Balance     *big.Int `json:"balance"`
	Nonce       uint64   `json:"nonce"`
	CodeHash    string   `json:"codeHash"`
	StorageRoot string   `json:"storageRoot"`
}

// NewAccount creates a new account
func NewAccount(address *Address) *Account {
	return &Account{
		Address:     address,
		Balance:     big.NewInt(0),
		Nonce:       0,
		CodeHash:    "",
		StorageRoot: "",
	}
}

// NewAccountWithBalance creates a new account with initial balance
func NewAccountWithBalance(address *Address, balance *big.Int) *Account {
	return &Account{
		Address:     address,
		Balance:     balance,
		Nonce:       0,
		CodeHash:    "",
		StorageRoot: "",
	}
}

// AddBalance adds to the account balance
func (a *Account) AddBalance(amount *big.Int) {
	a.Balance = new(big.Int).Add(a.Balance, amount)
}

// SubBalance subtracts from the account balance
func (a *Account) SubBalance(amount *big.Int) bool {
	if a.Balance.Cmp(amount) < 0 {
		return false
	}
	a.Balance = new(big.Int).Sub(a.Balance, amount)
	return true
}

// HasSufficientBalance checks if account has sufficient balance
func (a *Account) HasSufficientBalance(amount *big.Int) bool {
	return a.Balance.Cmp(amount) >= 0
}

// IncrementNonce increments the account nonce
func (a *Account) IncrementNonce() {
	a.Nonce++
}

// SetCode sets the contract code hash
func (a *Account) SetCode(codeHash string) {
	a.CodeHash = codeHash
}

// IsContract checks if the account is a contract account
func (a *Account) IsContract() bool {
	return a.CodeHash != ""
}

// IsEmpty checks if the account is empty (zero balance, zero nonce, no code)
func (a *Account) IsEmpty() bool {
	return a.Balance.Cmp(big.NewInt(0)) == 0 && a.Nonce == 0 && a.CodeHash == ""
}

// Clone creates a copy of the account
func (a *Account) Clone() *Account {
	return &Account{
		Address:     a.Address,
		Balance:     new(big.Int).Set(a.Balance),
		Nonce:       a.Nonce,
		CodeHash:    a.CodeHash,
		StorageRoot: a.StorageRoot,
	}
}