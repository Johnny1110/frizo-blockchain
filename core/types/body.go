package types

// Body represents the data content of a block (transactions and uncles)
type Body struct {
	Transactions Transactions `json:"transactions"`
	Uncles       []*Header    `json:"uncles"`
}

// create new block body
func NewBody(txns []*Transaction, uncles []*Header) *Body {
	body := &Body{
		Transactions: make([]*Transaction, len(txns)),
		Uncles:       make([]*Header, len(uncles)),
	}

	if len(txns) > 0 {
		copy(body.Transactions, txns)
	}
	if len(uncles) > 0 {
		copy(body.Uncles, uncles)
	}

	return body
}

func (b *Body) IsEmpty() bool {
	return b == nil || (len(b.Transactions) == 0 && len(b.Uncles) == 0)
}

func (b *Body) Copy() *Body {
	newBody := &Body{
		Transactions: make([]*Transaction, len(b.Transactions)),
		Uncles:       make([]*Header, len(b.Uncles)),
	}

	if len(b.Transactions) > 0 {
		copy(newBody.Transactions, b.Transactions)
	}

	if len(b.Uncles) > 0 {
		copy(newBody.Uncles, b.Uncles)
	}
	return newBody
}

func (b *Body) Size() uint64 {
	return uint64(len(b.Transactions))
}
