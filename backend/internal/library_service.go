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

func WithinTx[T any](s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (*T, error)) (*T, error) {
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
		err = tx.Commit(ctx)
		if err != nil {
			return nil, err
		}
		tx = nil
	}
	return result, nil
}
