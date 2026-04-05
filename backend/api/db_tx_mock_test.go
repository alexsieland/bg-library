package api

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type stubTx struct {
	commitCount   int
	rollbackCount int
	commitErr     error
	rollbackErr   error
}

func (s *stubTx) Begin(context.Context) (pgx.Tx, error) {
	return nil, nil
}

func (s *stubTx) Commit(context.Context) error {
	s.commitCount++
	return s.commitErr
}

func (s *stubTx) Rollback(context.Context) error {
	s.rollbackCount++
	return s.rollbackErr
}

func (s *stubTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (s *stubTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (s *stubTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (s *stubTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (s *stubTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (s *stubTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (s *stubTx) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (s *stubTx) Conn() *pgx.Conn {
	return nil
}
