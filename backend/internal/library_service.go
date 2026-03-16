package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
)

type LibraryService struct {
	Database db.DB
	queries  *db.Queries
}

// withinTxImpl is the non-generic implementation function variable. Tests
// can replace this to control transaction behavior in unit tests.
var withinTxImpl func(s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (any, error)) (any, error) = func(s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (any, error)) (any, error) {
	var (
		tx  pgx.Tx
		err error
	)
	if optTx != nil {
		tx = optTx
	} else {
		tx, err = s.Database.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return nil, err
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback(ctx)
			}
		}()
	}

	result, err := fn(tx)
	if err != nil {
		return nil, err
	}
	if optTx == nil {
		if err = tx.Commit(ctx); err != nil {
			return nil, err
		}
		// prevent the deferred rollback from running
		tx = nil
	}
	return result, nil
}

// WithinTx is a generic, type-safe wrapper around the non-generic
// `withinTxImpl`. Tests should override `withinTxImpl` when they need to
// mock transaction behavior; they must not attempt to assign a generic
// function literal to `WithinTx` (Go does not allow function literals with
// type parameters).
func WithinTx[T any](s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (*T, error)) (*T, error) {
	wrapper := func(tx pgx.Tx) (any, error) {
		res, err := fn(tx)
		return any(res), err
	}
	out, err := withinTxImpl(s, ctx, optTx, wrapper)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return out.(*T), nil
}
