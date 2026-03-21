package internal

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

// MockWithinTx provides a centralized test helper that overrides the
// package-level WithinTx implementation to use a test MockTx. Tests can
// call MockWithinTx(t) to obtain the *MockTx and set expectations on it.
// The original WithinTx implementation is restored automatically via t.Cleanup.
func MockWithinTx(t *testing.T) *MockTx {
	t.Helper()
	old := withinTxImpl
	mtx := new(MockTx)
	withinTxImpl = func(s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (any, error)) (any, error) {
		if optTx != nil {
			return fn(optTx)
		}
		res, err := fn(mtx)
		if err != nil {
			_ = mtx.Rollback(ctx)
			return nil, err
		}
		if err := mtx.Commit(ctx); err != nil {
			return nil, err
		}
		return res, nil
	}
	t.Cleanup(func() { withinTxImpl = old })
	return mtx
}
