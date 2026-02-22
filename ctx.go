package opera_txr

import (
	"context"
)

const ctxValueKeyTx = "opera_txr.Tx"

// WithTxCtx
//
// tx is of type any because this is an abstraction.
// The concrete type of the transaction (or whatever it may be — perhaps just an identifier)
// will depend on the specific implementation of TxrInterface and related repository implementations.
func WithTxCtx(ctx context.Context, tx any) context.Context {
	return context.WithValue(ctx, ctxValueKeyTx, tx)
}

// IsTxCtx
//
// Checks if the given context is derived by WithTxCtx.
func IsTxCtx(ctx context.Context) bool {
	return ctx.Value(ctxValueKeyTx) != nil
}

// TxFromCtx
//
// See WithTxCtx for explanation of "any" return type.
func TxFromCtx(ctx context.Context) any {
	return ctx.Value(ctxValueKeyTx)
}
