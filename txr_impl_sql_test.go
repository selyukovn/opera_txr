package opera_txr

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
// NewTxrImplSql
// ---------------------------------------------------------------------------------------------------------------------

func Test_NewTxrImplSql(t *testing.T) {
	db, _, _ := sqlmock.New()
	t.Cleanup(func() {
		_ = db.Close()
	})

	t.Run("PanicForBadClient", func(t *testing.T) {
		// db nil
		assert.Panics(t, func() {
			_ = NewTxrImplSql(nil, 3, time.Millisecond, func(err error) bool { return true })
		})

		// deadlockDetectionFn nil
		assert.Panics(t, func() {
			_ = NewTxrImplSql(db, 3, time.Millisecond, nil)
		})
	})
}

// ---------------------------------------------------------------------------------------------------------------------
// TxrImplSql.Tx
// ---------------------------------------------------------------------------------------------------------------------

func testHelper_TxrImplSql_Tx_setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db, mock
}

func Test_TxrImplSql_Tx(t *testing.T) {
	// Bad client
	// --------------------------------

	t.Run("PanicForBadClient", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		txr := NewTxrImplSql(db, 3, time.Minute, func(err error) bool { return false })

		// no ctx
		assert.Panics(t, func() {
			_ = txr.Tx(nil, func(ctx *TxCtx) error { return nil })
		})

		// no fn
		assert.Panics(t, func() {
			_ = txr.Tx(context.Background(), nil)
		})

		// nested tx
		assert.Panics(t, func() {
			mock.ExpectBegin() // outer transaction started

			_ = txr.Tx(context.Background(), func(ctx *TxCtx) error {
				return txr.Tx(ctx, func(ctx2 *TxCtx) error {
					return nil
				})
			})
		})

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Normal flow
	// --------------------------------

	t.Run("Commit", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		txr := NewTxrImplSql(db, 3, time.Minute, func(err error) bool { return false })

		mock.ExpectBegin()
		mock.ExpectCommit()

		err := txr.Tx(context.Background(), func(ctx *TxCtx) error {
			assert.IsType(t, &sql.Tx{}, ctx.Tx())
			return nil
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RollbackOnFnPanic", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		txr := NewTxrImplSql(db, 3, time.Minute, func(err error) bool { return false })

		mock.ExpectBegin()
		mock.ExpectRollback()

		panicValue := errors.New("panic error")

		assert.PanicsWithValue(t, panicValue, func() {
			_ = txr.Tx(context.Background(), func(ctx *TxCtx) error {
				panic(panicValue)
			})
		})

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RollbackOnFnError", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		txr := NewTxrImplSql(db, 3, 100*time.Millisecond, func(err error) bool { return false })

		expectedErr := errors.New("fn error")

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := txr.Tx(context.Background(), func(ctx *TxCtx) error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Deadlock
	// --------------------------------

	t.Run("RetryOnDeadlock", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		deadlockErr := errors.New("deadlock")
		txr := NewTxrImplSql(db, 2, 10*time.Millisecond, func(err error) bool { return errors.Is(err, deadlockErr) })

		// first attempt: deadlock
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(deadlockErr)

		// second attempt: success
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(nil)

		err := txr.Tx(context.Background(), func(ctx *TxCtx) error {
			return nil
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RetryOnDeadlockLimitExceeded", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		deadlockErr := errors.New("deadlock")
		txr := NewTxrImplSql(db, 2, 10*time.Millisecond, func(err error) bool { return errors.Is(err, deadlockErr) })

		// first attempt: deadlock
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(deadlockErr)

		// second attempt: retry
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(deadlockErr)

		// third attempt: limit exceed
		mock.ExpectBegin()
		mock.ExpectCommit().WillReturnError(deadlockErr)

		err := txr.Tx(context.Background(), func(ctx *TxCtx) error {
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "deadlock retry limit (2) exceeded")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Context cancellation
	// --------------------------------

	t.Run("CancelledByContextBeforeTx", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		txr := NewTxrImplSql(db, 3, 100*time.Millisecond, func(err error) bool { return false })

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // before tx

		err := txr.Tx(ctx, func(ctx *TxCtx) error {
			return errors.New("should not reach here")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("CancelledByContextDuringTx", func(t *testing.T) {
		db, mock := testHelper_TxrImplSql_Tx_setupMockDB(t)
		txr := NewTxrImplSql(db, 3, 100*time.Millisecond, func(err error) bool { return false })

		timeout := 10 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		mock.ExpectBegin()
		mock.ExpectRollback()

		err := txr.Tx(ctx, func(ctx *TxCtx) error {
			time.Sleep(timeout * 2)
			return errors.New("should not reach here")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// ---------------------------------------------------------------------------------------------------------------------
