package opera_txr

import (
	"context"
)

const ctxValueKeyTx = "opera_txr.TxCtx.Tx"

type TxCtx struct {
	context.Context
}

// WithTxCtx
//
// tx is of type any because this is an abstraction.
// The concrete type of the transaction (or whatever it may be — perhaps just an identifier)
// will depend on the specific implementation of TxrInterface and related repository implementations.
func WithTxCtx(ctx context.Context, tx any) *TxCtx {
	ctx = context.WithValue(ctx, ctxValueKeyTx, tx)
	return &TxCtx{ctx}
}

// IsInTxCtx
//
// Checks if the given context derives from TxCtx.
func IsInTxCtx(ctx context.Context) bool {
	return ctx.Value(ctxValueKeyTx) != nil
}

// Tx
//
// See WithTxCtx for explanation of "any" return type.
func (c *TxCtx) Tx() any {
	return c.Context.Value(ctxValueKeyTx)
}
