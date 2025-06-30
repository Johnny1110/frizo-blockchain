package core

import (
	"encoding/binary"
	"frizo-blockchain/internal/types"
	"io"
)

type Header struct {
	// Version block format version
	Version uint32
	// PrevBlock previous block's hash
	PrevBlock types.Hash
	// Timestamp bock create time
	Timestamp int64
	// Height block's height on this chain (increase from 0), also the number seq of block
	Height uint32
	// POW diff
	Nonce uint64
}

// EncodeBinary encode all header info to binary
func (h *Header) EncodeBinary(w io.Writer) error {
	// binary.LittleEndian: Byte Order (from little byte to big byte)
	if err := binary.Write(w, binary.LittleEndian, h.Version); err != nil {
		return err
	}
	if _, err := w.Write(h.PrevBlock[:]); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, h.Timestamp); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, h.Height); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, h.Nonce); err != nil {
		return err
	}
	return nil
}

// DecodeBinary decode all header info from binary
func (h *Header) DecodeBinary(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.Version); err != nil {
		return err
	}
	if _, err := io.ReadFull(r, h.PrevBlock[:]); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Timestamp); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Height); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.Nonce); err != nil {
		return err
	}
	return nil
}

type Block struct {
	Header       Header
	Transactions []Transaction
}

func (b *Block) EncodeBinary(w io.Writer) error {
	// 1. write Header
	if err := b.Header.EncodeBinary(w); err != nil {
		return err
	}

	// 2. include txn count
	txCount := int32(len(b.Transactions))
	if err := binary.Write(w, binary.LittleEndian, txCount); err != nil {
		return err
	}

	// 3. write Transactions
	for _, tx := range b.Transactions {
		if err := tx.EncodeBinary(w); err != nil {
			return err
		}
	}

	return nil
}

func (b *Block) DecodeBinary(r io.Reader) error {
	// 1. read Header
	if err := b.Header.DecodeBinary(r); err != nil {
		return err
	}

	// 2. read Transaction count
	var txCount uint32
	if err := binary.Read(r, binary.LittleEndian, &txCount); err != nil {
		return err
	}

	// 3. read TransactionS
	b.Transactions = make([]Transaction, txCount)
	for i := uint32(0); i < txCount; i++ {
		if err := b.Transactions[i].DecodeBinary(r); err != nil {
			return err
		}
	}

	return nil
}
