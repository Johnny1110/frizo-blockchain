package common

import "errors"

// Common errors
var (
	// ErrInvalidHash is returned when a hash is invalid
	ErrInvalidHash = errors.New("invalid hash")

	// ErrInvalidAddress is returned when an address is invalid
	ErrInvalidAddress = errors.New("invalid address")

	// ErrInvalidSignature is returned when a signature is invalid
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInsufficientFunds is returned when an account has insufficient funds
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrNonceTooLow is returned when the nonce of a transaction is lower than expected
	ErrNonceTooLow = errors.New("nonce too low")

	// ErrNonceTooHigh is returned when the nonce of a transaction is higher than expected
	ErrNonceTooHigh = errors.New("nonce too high")

	// ErrGasLimitReached is returned when the gas limit is exceeded
	ErrGasLimitReached = errors.New("gas limit reached")

	// ErrBlockNotFound is returned when a block is not found
	ErrBlockNotFound = errors.New("block not found")

	// ErrTransactionNotFound is returned when a transaction is not found
	ErrTransactionNotFound = errors.New("transaction not found")

	// ErrInvalidBlock is returned when a block is invalid
	ErrInvalidBlock = errors.New("invalid block")

	// ErrInvalidTransaction is returned when a transaction is invalid
	ErrInvalidTransaction = errors.New("invalid transaction")

	// ErrKnownBlock is returned when a block is already known
	ErrKnownBlock = errors.New("known block")

	// ErrKnownTransaction is returned when a transaction is already known
	ErrKnownTransaction = errors.New("known transaction")

	// ErrInvalidChainID is returned when the chain ID is invalid
	ErrInvalidChainID = errors.New("invalid chain ID")
)
