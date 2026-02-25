package txr

import (
	"context"
)

const ctxValueKeyTx = "txr.Tx"

// WithTxCtx derives context with provided transaction.
//
// Use it within your own implementation of the TxrInterface.
//
// `tx` is of type `any` because this is an abstraction.
// Specific type of the transaction (or whatever it may be — perhaps just an identifier)
// depends on the specific implementation of the TxrInterface and related repositories implementations.
//
// Panics if:
//   - `ctx` is nil.
//   - `tx` is nil.
func WithTxCtx(ctx context.Context, tx any) context.Context {
	if ctx == nil {
		panic("`ctx` must not be nil")
	} else if tx == nil {
		panic("`tx` must not be nil")
	}

	return context.WithValue(ctx, ctxValueKeyTx, tx)
}

// IsTxCtx checks if the given context is derived by WithTxCtx.
//
// Panics if `ctx` is nil.
func IsTxCtx(ctx context.Context) bool {
	if ctx == nil {
		panic("`ctx` must not be nil")
	}

	return ctx.Value(ctxValueKeyTx) != nil
}

// TxFromCtx gets transaction from the context derived by WithTxCtx.
//
// Panics if `ctx` is nil.
//
// Returns `nil` if IsTxCtx returns false for `ctx`.
func TxFromCtx(ctx context.Context) any {
	if ctx == nil {
		panic("`ctx` must not be nil")
	}

	return ctx.Value(ctxValueKeyTx)
}
