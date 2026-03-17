package api

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactionApiCheckInGame(t *testing.T) {
	t.Run("Should forward converted transaction id when checking in a game", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		txId := uuid.New()

		fixture.service.On("CheckInGame", fixture.ctx, testUUID(txId), nil).Return(nil).Once()

		err := fixture.api.CheckInGame(fixture.ctx, CheckInGameParams{TransactionId: types.UUID(txId)})

		assert.NoError(t, err)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return the service error when check in fails", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		txId := uuid.New()
		expectedErr := errors.New("checkin failed")

		fixture.service.On("CheckInGame", fixture.ctx, testUUID(txId), nil).Return(expectedErr).Once()

		err := fixture.api.CheckInGame(fixture.ctx, CheckInGameParams{TransactionId: types.UUID(txId)})

		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

func TestTransactionApiCheckOutGame(t *testing.T) {
	t.Run("Should return converted transaction when checkout succeeds", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		gameId := uuid.New()
		patronId := uuid.New()
		txId := uuid.New()
		now := time.Now().UTC()

		dbTx := db.Transaction{
			ID:                pgtype.UUID{Bytes: txId, Valid: true},
			GameID:            pgtype.UUID{Bytes: gameId, Valid: true},
			PatronID:          pgtype.UUID{Bytes: patronId, Valid: true},
			CheckoutTimestamp: pgtype.Timestamp{Time: now, Valid: true},
		}

		fixture.service.On("CheckOutGame", fixture.ctx, testUUID(gameId), testUUID(patronId), nil).Return(dbTx, nil).Once()

		resp, err := fixture.api.CheckOutGame(fixture.ctx, CheckOutGameJSONRequestBody{
			GameId:   types.UUID(gameId),
			PatronId: types.UUID(patronId),
		})

		assert.NoError(t, err)
		expected := LibraryTransaction{
			GameId:    pgUUIDToUUID(dbTx.GameID),
			Id:        pgUUIDToUUID(dbTx.ID),
			PatronId:  pgUUIDToUUID(dbTx.PatronID),
			Timestamp: dbTx.CheckoutTimestamp.Time,
		}
		assert.Equal(t, expected, resp)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return the service error when checkout fails", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		gameId := uuid.New()
		patronId := uuid.New()
		expectedErr := errors.New("checkout failed")

		fixture.service.On("CheckOutGame", fixture.ctx, testUUID(gameId), testUUID(patronId), nil).Return(db.Transaction{}, expectedErr).Once()

		_, err := fixture.api.CheckOutGame(fixture.ctx, CheckOutGameJSONRequestBody{
			GameId:   types.UUID(gameId),
			PatronId: types.UUID(patronId),
		})

		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

func TestTransactionApiListTransactionEvents(t *testing.T) {
	t.Run("Should return a converted transaction event list when request is valid", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		now := time.Now().UTC()

		txA := db.SearchTransactionEventsRow{
			TransactionID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
			GameID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			GameTitle:       "Catan",
			PatronID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			PatronFullName:  "Alice",
			EventType:       db.TransactionEventTypeCheckOut,
			EventTimestamp:  pgtype.Timestamp{Time: now, Valid: true},
			PlayToWinGameID: pgtype.UUID{Valid: false},
		}
		txB := db.SearchTransactionEventsRow{
			TransactionID:   pgtype.UUID{Bytes: uuid.New(), Valid: true},
			GameID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
			GameTitle:       "Pandemic",
			PatronID:        pgtype.UUID{Bytes: uuid.New(), Valid: true},
			PatronFullName:  "Bob",
			EventType:       db.TransactionEventTypeCheckIn,
			EventTimestamp:  pgtype.Timestamp{Time: now.Add(time.Second), Valid: true},
			PlayToWinGameID: pgtype.UUID{Valid: false},
		}

		expectedRows := []db.SearchTransactionEventsRow{txA, txB}

		fixture.service.On("ListTransactionEvents", fixture.ctx, (*string)(nil), (*string)(nil), int32(100), int32(0), nil).Return(expectedRows, nil).Once()

		list, err := fixture.api.ListTransactionEvents(fixture.ctx, ListTransactionEventsParams{})

		assert.NoError(t, err)

		var expectedEvents []TransactionEvent
		for _, r := range expectedRows {
			expectedEvents = append(expectedEvents, TransactionEvent{
				TransactionId:  pgUUIDToUUID(r.TransactionID),
				Game:           FromGame(db.Game{ID: r.GameID, Title: r.GameTitle}, r.PlayToWinGameID.Valid),
				Patron:         FromPatron(db.Patron{ID: r.PatronID, FullName: r.PatronFullName}),
				EventTimestamp: r.EventTimestamp.Time,
				EventType:      TransactionEventEventType(r.EventType),
			})
		}

		assert.Equal(t, TransactionEventList{Transactions: expectedEvents}, list)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return validation details when gameTitle is invalid", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		tooLong := strings.Repeat("t", 101)

		events, err := fixture.api.ListTransactionEvents(fixture.ctx, ListTransactionEventsParams{GameTitle: &tooLong})

		assert.Equal(t, TransactionEventList{}, events)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "gameTitle", Message: "Length must be between 1 and 100"}}})
	})

	t.Run("Should return the service error when ListTransactionEvents fails", func(t *testing.T) {
		fixture := newTransactionApiTestFixture(t).build()
		expectedErr := errors.New("list failed")

		fixture.service.On("ListTransactionEvents", fixture.ctx, (*string)(nil), (*string)(nil), int32(100), int32(0), nil).Return(nil, expectedErr).Once()

		events, err := fixture.api.ListTransactionEvents(fixture.ctx, ListTransactionEventsParams{})

		assert.Equal(t, TransactionEventList{}, events)
		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

// Fixture and mock for TransactionApi tests

type transactionApiTestFixture struct {
	ctx     context.Context
	service *mockTransactionService
	api     *TransactionApi
	tx      pgx.Tx
	txErr   error
}

func newTransactionApiTestFixture(t *testing.T) *transactionApiTestFixture {
	return &transactionApiTestFixture{
		ctx:     t.Context(),
		service: new(mockTransactionService),
	}
}

func (f *transactionApiTestFixture) withTx(tx pgx.Tx) *transactionApiTestFixture {
	f.tx = tx
	return f
}

func (f *transactionApiTestFixture) withTxError(err error) *transactionApiTestFixture {
	f.txErr = err
	return f
}

func (f *transactionApiTestFixture) build() *transactionApiTestFixture {
	f.api = newTestTransactionApi(f.service, f.tx, f.txErr)
	return f
}

func newTestTransactionApi(service transactionService, tx pgx.Tx, beginErr error) *TransactionApi {
	return &TransactionApi{
		service: service,
		beginTx: func(context.Context) (pgx.Tx, error) {
			return tx, beginErr
		},
	}
}

type mockTransactionService struct {
	mock.Mock
}

func (m *mockTransactionService) CheckOutGame(ctx context.Context, gameId pgtype.UUID, patronId pgtype.UUID, optTx pgx.Tx) (db.Transaction, error) {
	args := m.Called(ctx, gameId, patronId, optTx)
	if args.Get(0) == nil {
		return db.Transaction{}, args.Error(1)
	}
	return args.Get(0).(db.Transaction), args.Error(1)
}

func (m *mockTransactionService) CheckInGame(ctx context.Context, transactionId pgtype.UUID, optTx pgx.Tx) error {
	args := m.Called(ctx, transactionId, optTx)
	return args.Error(0)
}

func (m *mockTransactionService) ListTransactionEvents(ctx context.Context, sanitizedTitle *string, patronFullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.SearchTransactionEventsRow, error) {
	args := m.Called(ctx, sanitizedTitle, patronFullName, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.SearchTransactionEventsRow), args.Error(1)
}
