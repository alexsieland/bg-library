package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type LibraryService struct {
	context  *gin.Context
	Database db.DB
	queries  *db.Queries
}

func (s *LibraryService) GetRequestContext() context.Context {
	return s.context.Request.Context()
}

func WithinTx[T any](s *LibraryService, optTx *pgx.Tx, fn func(tx pgx.Tx) (*T, error)) (*T, error) {
	var tx pgx.Tx
	if optTx == nil {
		tx, err := s.Database.BeginTx(s.GetRequestContext(), pgx.TxOptions{})
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = tx.Rollback(s.GetRequestContext())
		}()
	}

	result, err := fn(tx)
	if err != nil {
		return nil, err
	}
	if optTx == nil {
		_ = tx.Commit(s.GetRequestContext())
	}
	return result, nil
}
