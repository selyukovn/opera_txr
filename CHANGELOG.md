## [0.4.1] - 2026-02-25

Project renamed from `go-opera-txr` to `go-txr`.

## [0.4.0] - 2026-02-22

### BREAKING CHANGES

- `TxCtx` is removed, `context.Context` abstraction is used instead. 
  Use `TxFromCtx()` to get transaction values from context.

- `IsInTxCtx()` is renamed to `InTxCtx()`.

### IMPROVEMENTS

- `WithTxCtx()`, `InTxCtx()`, `TxFromCtx()` panic in case of invalid arguments.

- A bit more readable go-doc.

---

## [0.3.0] - 2025-12-14

### BREAKING CHANGES
- `TxrImplSql.Tx` function is no longer responsible for checking `ctx.Done` during transaction execution.
  It required the transaction to be executed in a goroutine, and the panic propagation --
  so, the stack trace from the original panic point to the propagation point was lost,
  since the stack trace was counted from the panic propagation point, not from the original one.

### IMPROVEMENTS
- Simplified README.md

---

## [0.2.0] - 2025-11-04

*Note: v0.1.0 has been retracted.*

### BREAKING CHANGES
- `TxrImplSql` now requires `deadlockDetectionFn` function to handle driver-specific deadlock error detection 
  (e.g. MySQL deadlock error code = 1213, PostgreSQL = 40P01, etc.).

### FEATURES
- Added tests for `TxrImplSql`

### BUGS FIXED
- `NewTxrImplSql` panics on invalid (nil) arguments `db` or `deadlockDetectionFn`.
- `TxrImplSql.Tx` no longer ignores context cancellation during `fn` execution.

### IMPROVEMENTS
- Downgraded minimal Go version to `1.18`  (no functional dependency on higher versions).
- Updated `TxrImplSql` and `TxrInterface.Tx` documentation

---

## [0.1.0] - 2025-11-03

*Note: This release lacked a CHANGELOG section.*

### FEATURES
- Defined basic abstractions.
- Initial release of `TxrImplSql` with retries on deadlock.
