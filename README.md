
# Operational Layer Transactor

### TL;DR

A simple abstraction of the operational layer for transaction management, with ability for use in concurrent goroutines.

### Here's the thing

Transaction management is the responsibility of the operational layer.

Direct use of infrastructure libraries like `database/sql` is not allowed here.
Moreover, higher‑level control and error handling (e.g., deadlocks) are required,
which necessitates some kind of wrapper around such libraries.
Probably, the most concise form of such a wrapper is a method that accepts the transaction body as a closure.

```
var result
err = txr.Tx(func() error {
    result = ...
    return err
})
```


Instances of such a wrapper, like repository instances, are usually created as singletons
and then injected into operational services.
An active transaction must be accessible in repository methods that are called within that transaction.

----

The operational layer can perform parallelization (as it is also an element of coordination)
of independent actions within a single scenario by creating goroutines as the most natural approach.
For example:

```
func (s *OperationLayerService) SomeUseCase() {
    go func() {
        // ... tx1 begin ...
        // using domain repositories, etc.
        // ... tx1 end ...
    })

    // ... tx2 begin ...
    // using domain repositories, etc.
    // ... tx2 end ...
}
```

----

This leads to the problem of obtaining the active transaction inside repository methods.

Injecting the wrapper as a holder of the active transaction into repository implementations to retrieve the active transaction
is unacceptable, because concurrent transactions could overlap via a shared wrapper instance,
which would certainly lead to unpredictable consequences.
That is, an active transaction cannot be shared across multiple goroutines,
and therefore the wrapper cannot serve as the holder of the active transaction.

Synchronization is obviously unacceptable and thus not considered.

Go does not provide tools to access the current goroutine as an entity,
so it is impossible to "bind" a transaction to the current goroutine and then retrieve it inside a repository method.

Passing the active transaction as an argument to a domain repository method is not allowed,
as it violates the dependency direction between layers,
and creating a corresponding abstraction in the domain layer looks awkward.

Creating new repository instances with the current transaction injected in each goroutine via a factory
is a technically viable option, somewhat reminiscent of Unit‑of‑Work, but still rather cumbersome and equally awkward.

A more implicit yet simple approach is to pass the active transaction via `context.Context`.
Repositories and similar components represent external dependencies,
access to which often requires timeouts.
This means that such timeouts must be passed to every repository method call.
Passing a context as the first argument to such methods is already an established practice,
and the passed context can be augmented with arbitrary data — for instance, the active transaction itself.

P.S.: It should be clarified that we are talking about repositories that mimic a storage, not a collection,
since mimicking a collection requires implementing and/or using ORM systems,
which can negatively affect the domain layer and introduce additional overhead.

### Example

```go
package some_operational_layer_package

type OperationLayerService struct {
	// ...
	txr     opera_txr.TxrInterface
	accRepo domain.AccountRepositoryInterface
	// ...
}

func (s *OperationLayerService) SomeUseCase(email Email) {
	// ...

	var accId account.Id
	if err := s.txr.Tx(ctx, func(txCtx *opera_txr.TxCtx) error {
		acc, err := s.accRepo.GetByEmail(txCtx, email)
		// ... if err != nil { return err } ...

		err = acc.DoSomething()
		// ... if err != nil { return err } ...

		err = s.accRepo.Update(txCtx, acc)
		// ... if err != nil { return err } ...

		accId = acc.Id()
		return err
	}); err != nil {
		// ...
	}

	// ...
}
```
