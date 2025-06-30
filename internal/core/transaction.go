package core

import (
	"io"
)

type Transaction struct {
}

// EncodeBinary encode all txn info to binary
func (h *Transaction) EncodeBinary(w io.Writer) error {
	// TODO
	return nil
}

// DecodeBinary decode all txn info from binary
func (h *Transaction) DecodeBinary(r io.Reader) error {
	// TODO
	return nil
}
