
# Operational Layer Transactor

### TL;DR

A simple transaction manager for the operational layer.

### Here's the thing

Transaction management is the responsibility of the _operational layer_.
Usage of _infrastructure layer_ libraries like `database/sql` on the operational layer breaks dependency direction.
Moreover, higher‑level control and error handling (e.g., deadlocks) are useful, so some wrapper is rather required.
Probably, the most laconic form of such wrapper is a method that accepts the transaction body as a closure.

```
var result
err = txr.Tx(func() error {
    result = ...
    return err
})
```

Such closures contain, for example, repository methods calls.
An active transaction must be accessible in repository methods that are called within that transaction.
The most natural way to do this in Go -- send an active transaction through the `context.Context` to each call.
`TxFromCtx()` is used for this -- it gets an actual transaction inside repository methods.

See examples for details.

### Example

```go
package some_package

import (
	"context"
	"github.com/selyukovn/go-opera-txr"
    ".../domain/account"
	// ...
)

type OperationLayerService struct {
	// ...
	txr     opera_txr.TxrInterface
	accRepo account.RepositoryInterface
	// ...
}

func (s *OperationLayerService) SomeUseCase(ctx context.Context, email Email) {
	// ...

	var accId account.Id
	if err := s.txr.Tx(ctx, func(ctx context.Context) error {
		acc, err := s.accRepo.GetByEmail(ctx, email)

		// ...

		accId = acc.Id()
		return err
	}); err != nil {
		// ...
	}

	// ...
}

// ----

func (r *AccountRepositoryImplSql) GetByEmail(ctx context.Context, email Email) (*Account, error) {
	// Here, inside SQL-implementation of the repository,
	// we expect, that SQL-implementation of the `TxrInterface` is used,
	// and `*sql.Tx` is a type of the transaction.
	// So this is how to get actual transaction in 1 line of code:
	tx := opera_txr.TxFromCtx(ctx).(*sql.Tx)

	// Usual usage of *sql.Tx ...
	result, err := tx.QueryRowContext(ctx, "SELECT ...")

	// ...
}
```
