package internal

import (
	"github.com/alexsieland/bg-library/db"
)

// setupTestLibraryService constructs a LibraryService wired with a MockDatabase
// and sqlc-generated Queries for use in unit tests.
func setupTestLibraryService() (*LibraryService, *MockDatabase) {
	mockDB := new(MockDatabase)
	queries := db.New(mockDB)
	lib := &LibraryService{
		database: mockDB,
		queries:  queries,
	}
	return lib, mockDB
}

// MockWithinTx overrides the package-level WithinTx to use a test MockTx.
// It returns the created *MockTx so tests can set expectations on it. The
// original WithinTx is restored automatically via t.Cleanup.
// ...existing code...
