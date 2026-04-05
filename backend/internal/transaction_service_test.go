package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionServiceCheckOutGame(t *testing.T) {
	t.Run("Should panic when gameService is not set", func(t *testing.T) {
		lib, _ := setupTestLibraryService()
		svc := NewTransactionService(lib)

		assert.Panics(t, func() {
			_, _ = svc.CheckOutGame(context.Background(), pgtype.UUID{Valid: false}, pgtype.UUID{Valid: false}, nil)
		})
	})

	t.Run("Should return error when GetGameStatus fails", func(t *testing.T) {
		// setup lib service with mocked DB returning an error for GetGameStatus
		svc, mockDB := setupTestTransactionService()

		gameId := uuid.New()
		mockRow := new(MockRow)
		MockRowScanError(mockRow, 9, errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameId, Valid: true}}).Return(mockRow).Once()

		_, err := svc.CheckOutGame(context.Background(), pgtype.UUID{Bytes: gameId, Valid: true}, pgtype.UUID{Valid: false}, nil)

		assert.Error(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return existing transaction when already checked out by same patron", func(t *testing.T) {
		svc, ctx, mockTx, mockDB := setupTransactionServiceWithMockTx(t)

		gameId := uuid.New()
		patronId := uuid.New()
		txId := uuid.New()
		now := time.Now().UTC()

		status := makeVwGameStatus(gameId, &patronId, &txId, true, now)

		mockRow := new(MockRow)
		MockVwGameStatusScan(mockRow, status, nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{status.GameID}).Return(mockRow).Once()

		// WithinTx should commit when fn returns without error
		mockTx.On("Commit", ctx).Return(nil).Once()

		tx, err := svc.CheckOutGame(ctx, status.GameID, status.PatronID, nil)

		assert.NoError(t, err)
		assert.Equal(t, status.TransactionID, tx.ID)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return conflict when already checked out by different patron", func(t *testing.T) {
		svc, ctx, mockTx, mockDB := setupTransactionServiceWithMockTx(t)

		gameId := uuid.New()
		currentPatron := uuid.New()
		requestingPatron := uuid.New()

		status := makeVwGameStatus(gameId, &currentPatron, nil, true, time.Now().UTC())

		mockRow := new(MockRow)
		MockVwGameStatusScan(mockRow, status, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{status.GameID}).Return(mockRow).Once()

		// Expect rollback when conflict occurs
		mockTx.On("Rollback", ctx).Return(nil).Once()

		_, err := svc.CheckOutGame(ctx, status.GameID, pgtype.UUID{Bytes: requestingPatron, Valid: true}, nil)

		assert.ErrorIs(t, err, ErrCheckOutConflict)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should perform normal checkout when game is available", func(t *testing.T) {
		svc, ctx, mockTx, mockDB := setupTransactionServiceWithMockTx(t)

		gameId := uuid.New()
		patronId := uuid.New()
		txId := uuid.New()
		now := time.Now().UTC()

		// Game status: not checked out
		status := makeVwGameStatus(gameId, nil, nil, false, time.Time{})

		mockStatusRow := new(MockRow)
		MockVwGameStatusScan(mockStatusRow, status, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{status.GameID}).Return(mockStatusRow).Once()

		// For CheckOutGame query, return a transaction row
		mockTxRow := new(MockRow)
		dbTx := makeTransaction(txId, gameId, patronId, now)
		MockTransactionScan(mockTxRow, dbTx, nil)
		// queries.CheckOutGame uses QueryRow with (gameId, patronId)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{dbTx.GameID, pgtype.UUID{Bytes: patronId, Valid: true}}).Return(mockTxRow).Once()

		mockTx.On("Commit", ctx).Return(nil).Once()

		tx, err := svc.CheckOutGame(ctx, dbTx.GameID, pgtype.UUID{Bytes: patronId, Valid: true}, nil)

		assert.NoError(t, err)
		assert.Equal(t, dbTx.ID, tx.ID)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestTransactionServiceCheckInGame(t *testing.T) {
	t.Run("Should check in and commit when no transaction provided", func(t *testing.T) {
		lib, _ := setupTestLibraryService()
		ctx := context.Background()
		mockTx := MockWithinTx(t)

		svc := NewTransactionService(lib)

		txId := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		// Expect Exec on tx to update the transaction
		mockTx.On("Exec", mock.Anything, mock.Anything, []any{txId}).Return(pgconn.CommandTag{}, nil).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		err := svc.CheckInGame(ctx, txId, nil)

		assert.NoError(t, err)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return wrapped database error when Exec fails", func(t *testing.T) {
		lib, mockDB := setupTestLibraryService()
		svc := NewTransactionService(lib)

		tx := new(MockTx)
		// Simulate Exec returning error through WithTx path
		tx.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, errors.New("exec failed")).Once()

		// Override withinTxImpl to call fn with provided tx
		old := withinTxImpl
		withinTxImpl = func(s *LibraryService, ctx context.Context, optTx pgx.Tx, fn func(tx pgx.Tx) (any, error)) (any, error) {
			return fn(tx)
		}
		t.Cleanup(func() { withinTxImpl = old })

		err := svc.CheckInGame(context.Background(), pgtype.UUID{Bytes: uuid.New(), Valid: true}, tx)

		assert.Error(t, err)
		mockDB.AssertExpectations(t)
	})
}

func TestTransactionServiceListTransactionEvents(t *testing.T) {
	t.Run("Should return rows from DB when no transaction provided", func(t *testing.T) {
		lib, mockDB := setupTestLibraryService()
		svc := NewTransactionService(lib)

		now := time.Now().UTC()
		rowA := db.SearchTransactionEventsRow{
			TransactionID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
			GameID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			GameTitle:       "Catan",
			PatronID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			PatronFullName:  "Alice",
			EventType:       db.TransactionEventTypeCheckOut,
			EventTimestamp:  pgtype.Timestamp{Time: now, Valid: true},
			PlayToWinGameID: pgtype.UUID{Valid: false},
		}
		rows := new(MockRows)
		MockSearchTransactionEventsRows(rows, []db.SearchTransactionEventsRow{rowA}, nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: "", Valid: true}, "", int32(10), int32(0)}).Return(rows, nil).Once()

		result, err := svc.ListTransactionEvents(context.Background(), nil, nil, 10, 0, nil)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		mockDB.AssertExpectations(t)
		rows.AssertExpectations(t)
	})

	t.Run("Should return wrapped DB error when query fails", func(t *testing.T) {
		lib, mockDB := setupTestLibraryService()
		svc := NewTransactionService(lib)

		mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("query failed")).Once()

		res, err := svc.ListTransactionEvents(context.Background(), nil, nil, 5, 0, nil)

		assert.Nil(t, res)
		assert.Error(t, err)
		mockDB.AssertExpectations(t)
	})
}

// ---- Helper mocks for scanning and rows ----

func MockVwGameStatusScan(row *MockRow, status db.VwGameStatus, err error) {
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = status.GameID
		*args.Get(1).(*string) = status.GameTitle
		*args.Get(2).(*string) = status.SanitizedTitle
		*args.Get(3).(*pgtype.UUID) = status.PatronID
		*args.Get(4).(*pgtype.Text) = status.PatronFullName
		*args.Get(5).(*pgtype.UUID) = status.TransactionID
		*args.Get(6).(*pgtype.Timestamp) = status.CheckoutTimestamp
		*args.Get(7).(*pgtype.Timestamp) = status.CheckinTimestamp
		*args.Get(8).(*pgtype.UUID) = status.PtwGameID
	}).Return(err)
}

func MockTransactionScan(row *MockRow, trans db.Transaction, err error) {
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = trans.ID
		*args.Get(1).(*pgtype.UUID) = trans.GameID
		*args.Get(2).(*pgtype.UUID) = trans.PatronID
		*args.Get(3).(*pgtype.Timestamp) = trans.CheckoutTimestamp
		*args.Get(4).(*pgtype.Timestamp) = trans.CheckinTimestamp
	}).Return(err)
}

func MockSearchTransactionEventsRows(rows *MockRows, items []db.SearchTransactionEventsRow, err error) {
	for range items {
		rows.On("Next").Return(true).Once()
	}
	rows.On("Next").Return(false).Once()

	if len(items) > 0 {
		idx := 0
		rows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			item := items[idx]
			idx++
			*args.Get(0).(*pgtype.UUID) = item.TransactionID
			*args.Get(1).(*pgtype.UUID) = item.GameID
			*args.Get(2).(*string) = item.GameTitle
			*args.Get(3).(*pgtype.UUID) = item.PatronID
			*args.Get(4).(*string) = item.PatronFullName
			*args.Get(5).(*db.TransactionEventType) = item.EventType
			*args.Get(6).(*pgtype.Timestamp) = item.EventTimestamp
			*args.Get(7).(*pgtype.UUID) = item.PlayToWinGameID
		}).Return(nil).Times(len(items))
	}

	rows.On("Close").Return().Once()
	rows.On("Err").Return(err).Once()
}

// Setup helpers to mirror patrons_service_test.go style
func setupTestTransactionService() (TransactionService, *MockDatabase) {
	libService, mockDB := setupTestLibraryService()
	svc := TransactionService{libraryService: libService}
	svc.SetGameService(NewGameService(libService))
	return svc, mockDB
}

func setupTransactionServiceWithMockTx(t *testing.T) (TransactionService, context.Context, *MockTx, *MockDatabase) {
	t.Helper()
	ctx := t.Context()
	svc, mockDB := setupTestTransactionService()
	mockTx := MockWithinTx(t)
	return svc, ctx, mockTx, mockDB
}

// makeVwGameStatus builds a db.VwGameStatus for test convenience.
// If patronID or txID is nil we mark those fields invalid.
func makeVwGameStatus(gameID uuid.UUID, patronID *uuid.UUID, txID *uuid.UUID, checkedOut bool, checkoutTime time.Time) db.VwGameStatus {
	status := db.VwGameStatus{
		GameID:            pgtype.UUID{Bytes: gameID, Valid: true},
		GameTitle:         "Test Game",
		SanitizedTitle:    "test game",
		PatronFullName:    pgtype.Text{Valid: false},
		CheckoutTimestamp: pgtype.Timestamp{Valid: false},
		CheckinTimestamp:  pgtype.Timestamp{Valid: true},
		PtwGameID:         pgtype.UUID{Valid: false},
	}
	if patronID != nil {
		status.PatronID = pgtype.UUID{Bytes: *patronID, Valid: true}
		status.PatronFullName = pgtype.Text{String: "Patron", Valid: true}
	}
	if txID != nil {
		status.TransactionID = pgtype.UUID{Bytes: *txID, Valid: true}
	}
	if checkedOut {
		status.CheckoutTimestamp = pgtype.Timestamp{Time: checkoutTime, Valid: true}
		status.CheckinTimestamp = pgtype.Timestamp{Valid: false}
	}
	return status
}

func makeTransaction(id uuid.UUID, gameID uuid.UUID, patronID uuid.UUID, checkoutTime time.Time) db.Transaction {
	return db.Transaction{
		ID:                pgtype.UUID{Bytes: id, Valid: true},
		GameID:            pgtype.UUID{Bytes: gameID, Valid: true},
		PatronID:          pgtype.UUID{Bytes: patronID, Valid: true},
		CheckoutTimestamp: pgtype.Timestamp{Time: checkoutTime, Valid: true},
	}
}
