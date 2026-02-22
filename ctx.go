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
//
// Panics if:
//   - ctx is nil
//   - tx is nil
func WithTxCtx(ctx context.Context, tx any) context.Context {
	if ctx == nil {
		panic("`ctx` must not be nil")
	} else if tx == nil {
		panic("`tx` must not be nil")
	}

	return context.WithValue(ctx, ctxValueKeyTx, tx)
}

// IsTxCtx
//
// Checks if the given context is derived by WithTxCtx.
//
// Panics if ctx is nil.
func IsTxCtx(ctx context.Context) bool {
	if ctx == nil {
		panic("`ctx` must not be nil")
	}

	return ctx.Value(ctxValueKeyTx) != nil
}

// TxFromCtx
//
// See WithTxCtx for explanation of "any" return type.
//
// Panics if ctx is nil.
func TxFromCtx(ctx context.Context) any {
	if ctx == nil {
		panic("`ctx` must not be nil")
	}

	return ctx.Value(ctxValueKeyTx)
}
