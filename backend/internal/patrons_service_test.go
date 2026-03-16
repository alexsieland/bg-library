package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPatronServiceInsertPatron(t *testing.T) {
	t.Run("Should insert patron when no transaction is provided", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)

		expectedPatron := testPatron(uuid.New(), "Jane Doe", nil)
		mockRow := new(MockRow)
		MockPatronScan(mockRow, expectedPatron, nil)

		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{"Jane Doe", pgtype.Text{Valid: false}}).Return(mockRow).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		patron, err := service.InsertPatron(ctx, "Jane Doe", nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatron, patron)
		mockRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return already exists error when insert violates a unique constraint", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		barcode := "P-1001"

		mockRow := new(MockRow)
		MockRowScanError(mockRow, 5, &pgconn.PgError{Code: "23505"})

		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{"Jane Doe", pgtype.Text{String: barcode, Valid: true}}).Return(mockRow).Once()
		mockTx.On("Rollback", ctx).Return(nil).Once()

		patron, err := service.InsertPatron(ctx, "Jane Doe", &barcode, nil)

		assert.Equal(t, db.Patron{}, patron)
		assert.ErrorIs(t, err, ErrAlreadyExists)
		mockRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should use the provided transaction when one is supplied", func(t *testing.T) {
		service, _, _ := setupPatronServiceWithMockTx(t)
		providedTx := new(MockTx)
		barcode := "P-1002"

		expectedPatron := testPatron(uuid.New(), "Alex Doe", &barcode)
		mockRow := new(MockRow)
		MockPatronScan(mockRow, expectedPatron, nil)

		providedTx.On("QueryRow", mock.Anything, mock.Anything, []any{"Alex Doe", pgtype.Text{String: barcode, Valid: true}}).Return(mockRow).Once()

		patron, err := service.InsertPatron(context.Background(), "Alex Doe", &barcode, providedTx)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatron, patron)
		mockRow.AssertExpectations(t)
		providedTx.AssertExpectations(t)
	})
}

func TestPatronServiceDeletePatron(t *testing.T) {
	t.Run("Should delete patron when no transaction is provided", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		patronID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockTx.On("Exec", mock.Anything, mock.Anything, []any{patronID}).Return(pgconn.CommandTag{}, nil).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		err := service.DeletePatron(ctx, patronID, nil)

		assert.NoError(t, err)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return not found error when delete affects no patron", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		patronID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockTx.On("Exec", mock.Anything, mock.Anything, []any{patronID}).Return(pgconn.CommandTag{}, pgx.ErrNoRows).Once()
		mockTx.On("Rollback", ctx).Return(nil).Once()

		err := service.DeletePatron(ctx, patronID, nil)

		assert.ErrorIs(t, err, ErrNotFound)
		mockTx.AssertExpectations(t)
	})
}

func TestPatronServiceGetPatron(t *testing.T) {
	t.Run("Should return patron when no transaction is provided", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		barcode := "P-2001"
		expectedPatron := testLibraryPatron(uuid.New(), "Jordan Doe", &barcode)
		mockRow := new(MockRow)

		MockVwLibraryPatronScan(mockRow, expectedPatron, nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{expectedPatron.ID}).Return(mockRow).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		patron, err := service.GetPatron(ctx, expectedPatron.ID, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatron, patron)
		mockRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return not found error when patron does not exist", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		patronID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		mockRow := new(MockRow)

		MockRowScanError(mockRow, 4, pgx.ErrNoRows)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{patronID}).Return(mockRow).Once()
		mockTx.On("Rollback", ctx).Return(nil).Once()

		patron, err := service.GetPatron(ctx, patronID, nil)

		assert.Equal(t, db.VwLibraryPatron{}, patron)
		assert.ErrorIs(t, err, ErrNotFound)
		mockRow.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestPatronServiceGetPatronByBarcode(t *testing.T) {
	t.Run("Should return patron from the database when no transaction is provided", func(t *testing.T) {
		service, _, mockDB := setupPatronServiceWithDB(t)
		barcode := "P-3001"
		expectedPatron := testLibraryPatron(uuid.New(), "Morgan Doe", &barcode)
		mockRow := new(MockRow)

		MockVwLibraryPatronScan(mockRow, expectedPatron, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRow).Once()

		patron, err := service.GetPatronByBarcode(context.Background(), barcode, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatron, patron)
		mockRow.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should use the provided transaction when barcode lookup runs inside a transaction", func(t *testing.T) {
		service, _, _ := setupPatronServiceWithDB(t)
		providedTx := new(MockTx)
		barcode := "P-3002"
		expectedPatron := testLibraryPatron(uuid.New(), "Taylor Doe", &barcode)
		mockRow := new(MockRow)

		MockVwLibraryPatronScan(mockRow, expectedPatron, nil)
		providedTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRow).Once()

		patron, err := service.GetPatronByBarcode(context.Background(), barcode, providedTx)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatron, patron)
		mockRow.AssertExpectations(t)
		providedTx.AssertExpectations(t)
	})

	t.Run("Should return not found error when barcode does not match a patron", func(t *testing.T) {
		service, _, mockDB := setupPatronServiceWithDB(t)
		barcode := "P-3003"
		mockRow := new(MockRow)

		MockRowScanError(mockRow, 4, pgx.ErrNoRows)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRow).Once()

		patron, err := service.GetPatronByBarcode(context.Background(), barcode, nil)

		assert.Equal(t, db.VwLibraryPatron{}, patron)
		assert.ErrorIs(t, err, ErrNotFound)
		mockRow.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
}

func TestPatronServiceUpdatePatron(t *testing.T) {
	t.Run("Should update patron when no transaction is provided", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		patronID := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockTx.On("Exec", mock.Anything, mock.Anything, []any{patronID, "Updated Name", pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, nil).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		err := service.UpdatePatron(ctx, patronID, "Updated Name", nil, nil)

		assert.NoError(t, err)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should use the provided transaction when updating a barcode", func(t *testing.T) {
		service, _, _ := setupPatronServiceWithMockTx(t)
		providedTx := new(MockTx)
		patronID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		barcode := "P-4001"

		providedTx.On("Exec", mock.Anything, mock.Anything, []any{patronID, "Updated Name", pgtype.Text{String: barcode, Valid: true}}).Return(pgconn.CommandTag{}, nil).Once()

		err := service.UpdatePatron(context.Background(), patronID, "Updated Name", &barcode, providedTx)

		assert.NoError(t, err)
		providedTx.AssertExpectations(t)
	})

	t.Run("Should return already exists error when update violates a unique constraint", func(t *testing.T) {
		service, ctx, mockTx := setupPatronServiceWithMockTx(t)
		patronID := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		barcode := "P-4002"

		mockTx.On("Exec", mock.Anything, mock.Anything, []any{patronID, "Updated Name", pgtype.Text{String: barcode, Valid: true}}).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: "23505"}).Once()
		mockTx.On("Rollback", ctx).Return(nil).Once()

		err := service.UpdatePatron(ctx, patronID, "Updated Name", &barcode, nil)

		assert.ErrorIs(t, err, ErrAlreadyExists)
		mockTx.AssertExpectations(t)
	})
}

func TestPatronServiceListPatrons(t *testing.T) {
	t.Run("Should list patrons when search name is not provided", func(t *testing.T) {
		service, ctx, mockDB := setupPatronServiceWithDB(t)
		barcode := "P-5001"
		expectedPatrons := []db.VwLibraryPatron{
			testLibraryPatron(uuid.New(), "Casey Doe", &barcode),
			testLibraryPatron(uuid.New(), "Logan Doe", nil),
		}
		mockRows := new(MockRows)

		MockVwLibraryPatronRows(mockRows, expectedPatrons, nil)
		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(25), int32(50)}).Return(mockRows, nil).Once()

		patrons, err := service.ListPatrons(ctx, nil, 25, 50, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatrons, patrons)
		mockRows.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should search patrons when a name filter is provided", func(t *testing.T) {
		service, _, _ := setupPatronServiceWithDB(t)
		providedTx := new(MockTx)
		searchName := "Doe"
		expectedPatrons := []db.VwLibraryPatron{testLibraryPatron(uuid.New(), "Jamie Doe", nil)}
		mockRows := new(MockRows)

		MockVwLibraryPatronRows(mockRows, expectedPatrons, nil)
		providedTx.On("Query", mock.Anything, mock.Anything, []any{"%Doe%", int32(10), int32(5)}).Return(mockRows, nil).Once()

		patrons, err := service.ListPatrons(context.Background(), &searchName, 10, 5, providedTx)

		assert.NoError(t, err)
		assert.Equal(t, expectedPatrons, patrons)
		mockRows.AssertExpectations(t)
		providedTx.AssertExpectations(t)
	})

	t.Run("Should return the query error when listing patrons fails", func(t *testing.T) {
		service, ctx, mockDB := setupPatronServiceWithDB(t)
		expectedErr := errors.New("query failed")

		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(15), int32(30)}).Return(nil, expectedErr).Once()

		patrons, err := service.ListPatrons(ctx, nil, 15, 30, nil)

		assert.Nil(t, patrons)
		assert.ErrorIs(t, err, expectedErr)
		mockDB.AssertExpectations(t)
	})
}

func setupTestPatronService() (PatronService, *MockDatabase) {
	libService, mockDB := setupTestLibraryService()
	return PatronService{libService: libService}, mockDB
}

func setupPatronServiceWithMockTx(t *testing.T) (PatronService, context.Context, *MockTx) {
	ctx := t.Context()
	service, _ := setupTestPatronService()
	mockTx := MockWithinTx(t)
	return service, ctx, mockTx
}

func setupPatronServiceWithDB(t *testing.T) (PatronService, context.Context, *MockDatabase) {
	ctx := t.Context()
	service, mockDB := setupTestPatronService()
	return service, ctx, mockDB
}

func testPatron(id uuid.UUID, fullName string, barcode *string) db.Patron {
	return db.Patron{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		FullName:  fullName,
		CreatedAt: pgtype.Timestamp{Valid: true},
		DeletedAt: pgtype.Timestamp{Valid: false},
		Barcode:   testBarcode(barcode),
	}
}

func testLibraryPatron(id uuid.UUID, fullName string, barcode *string) db.VwLibraryPatron {
	return db.VwLibraryPatron{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		FullName:  fullName,
		Barcode:   testBarcode(barcode),
		CreatedAt: pgtype.Timestamp{Valid: true},
	}
}

func testBarcode(barcode *string) pgtype.Text {
	if barcode == nil {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: *barcode, Valid: true}
}
