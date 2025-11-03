package opera_txr

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
// Struct
// ---------------------------------------------------------------------------------------------------------------------

type TxrImplSql struct {
	db                       *sql.DB
	deadlockMaxRetries       uint
	deadlockMinRetryInterval time.Duration
}

// ---------------------------------------------------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------------------------------------------------

func NewTxrImplSql(
	db *sql.DB,
	deadlockMaxRetries uint,
	deadlockMinRetryInterval time.Duration,
) *TxrImplSql {
	return &TxrImplSql{
		db:                       db,
		deadlockMaxRetries:       deadlockMaxRetries,
		deadlockMinRetryInterval: deadlockMinRetryInterval,
	}
}

// ---------------------------------------------------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------------------------------------------------

// Tx runs the provided function fn within a transaction context TxCtx.
//
// Panics if:
//   - ctx is nil (programming error: caller must provide a valid context)
//   - nested calls (makes no sense and likely indicates a design flaw)
//   - fn is nil (programming error: transaction body must be provided)
//   - fn panics
//
// Returns the error returned by fn, or a runtime error if processing fails.
func (t *TxrImplSql) Tx(ctx context.Context, fn func(txCtx *TxCtx) error) error {
	return t.processTx(true, ctx, fn)
}

func (t *TxrImplSql) processTx(
	// todo : perhaps, there should be RO/RW-transactions ???
	isWritable bool,
	ctx context.Context,
	fn func(txCtx *TxCtx) error,
) error {
	if ctx == nil {
		panic(fmt.Errorf("%T : ctx must not be nil", t))
	} else if IsInTxCtx(ctx) {
		panic(fmt.Errorf("%T : nested transactions are not allowed", t))
	}

	if fn == nil {
		panic(fmt.Errorf("%T : fn (transaction body) must not be nil", t))
	}

	// --

	var deadlockRetries uint

	for {
		err := t.tx(isWritable, ctx, fn)

		if err == nil {
			break
		}

		if strings.Contains(strings.ToLower(err.Error()), "deadlock") ||
			strings.Contains(strings.ToLower(err.Error()), "lock wait timeout") ||
			strings.Contains(strings.ToLower(err.Error()), "could not obtain lock") {
			if deadlockRetries == t.deadlockMaxRetries {
				return fmt.Errorf(
					"%T : deadlock retry limit (%d) exceeded. Originally caused by : %w",
					t,
					t.deadlockMaxRetries,
					err,
				)
			}

			deadlockRetries++

			select {
			case <-ctx.Done():
				err = fmt.Errorf(
					"%T : transaction retry #%d (originally caused by: %v) cancelled by context: %w",
					t,
					deadlockRetries,
					err,
					ctx.Err(),
				)
			case <-time.After(t.deadlockMinRetryInterval * time.Duration(deadlockRetries)):
				continue
			}
		}

		return err
	}

	// --

	// Unreachable (loop either returns or continues)
	return nil
}

func (t *TxrImplSql) tx(
	isWritable bool,
	ctx context.Context,
	fn func(txCtx *TxCtx) error,
) error {
	sqlTx, err := t.db.BeginTx(ctx, nil)

	defer func() {
		if sqlTx != nil {
			// According to the inner code, rollback will not be executed, if transaction is already done.
			// So if transaction was committed (successfully or not), we should not expect an additional rollback.
			_ = sqlTx.Rollback()
		}
	}()

	if err != nil {
		return err
	}

	err = t.txFn(ctx, sqlTx, fn)

	if err == nil && isWritable {
		err = sqlTx.Commit()
	}

	return err
}

func (t *TxrImplSql) txFn(
	ctx context.Context,
	sqlTx *sql.Tx,
	fn func(txCtx *TxCtx) error,
) error {
	// Here we block current goroutine until ctx is done or fn is completed.
	// Fn runs in a separate goroutine, but appears synchronous to client code
	// due to blocking by the select statement and panic propagation.

	var err error

	txCtx := WithTxCtx(ctx, sqlTx)

	type fnChanResult = struct {
		err   error
		panic any
	}

	fnChan := make(chan fnChanResult)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fnChan <- fnChanResult{panic: r}
			}
			close(fnChan)
		}()

		err := fn(txCtx)

		fnChan <- fnChanResult{err: err}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case r := <-fnChan:
		if r.panic != nil {
			panic(r.panic)
		}

		err = r.err
	}

	return err
}
