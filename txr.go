package opera_txr

import (
	"context"
)

type TxrInterface interface {
	// Tx runs the provided function fn within a transaction context TxCtx.
	//
	// Panics if:
	//   - ctx is nil (programming error: caller must provide a valid context)
	//   - nested calls (makes no sense and likely indicates a design flaw)
	//   - fn is nil (programming error: transaction body must be provided)
	//
	// Returns the error returned by fn, or a runtime error if processing fails.
	Tx(ctx context.Context, fn func(txCtx *TxCtx) error) error
}
