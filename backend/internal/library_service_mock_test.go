package internal

import (
	"context"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
)

// setupTestLibraryService constructs a LibraryService wired with a MockDatabase
// and sqlc-generated Queries for use in unit tests.
func setupTestLibraryService() (*LibraryService, *MockDatabase) {
	mockDB := new(MockDatabase)
	queries := db.New(mockDB)
	lib := &LibraryService{
		Database: mockDB,
		queries:  queries,
	}
	return lib, mockDB
}

// MockWithinTx overrides the package-level WithinTx to use a test MockTx.
// It returns the created *MockTx so tests can set expectations on it. The
// original WithinTx is restored automatically via t.Cleanup.
func MockWithinTx(t *testing.T) *MockTx {
	t.Helper()
	old := withinTxImpl
	mtx := new(MockTx)
	withinTxImpl = func(s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (any, error)) (any, error) {
		if optTx != nil {
			return fn(optTx)
		}
		// Simulate owning the transaction: run the function, then commit or rollback.
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
