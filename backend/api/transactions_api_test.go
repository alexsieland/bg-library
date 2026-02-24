package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckInGame(t *testing.T) {
	t.Run("Should return 204 No Content when game is checked in successfully", func(t *testing.T) {
		server, mockDB := setupTestServer()
		transactionID := uuid.New()

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: transactionID, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions/checkin", nil)
		server.CheckInGame(c, CheckInGameParams{TransactionId: transactionID.String()})

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when transactionId is invalid", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions/checkin", nil)
		server.CheckInGame(c, CheckInGameParams{TransactionId: "invalid-uuid"})

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs", func(t *testing.T) {
		server, mockDB := setupTestServer()
		transactionID := uuid.New()

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: transactionID, Valid: true}}).Return(pgconn.CommandTag{}, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions/checkin", nil)
		server.CheckInGame(c, CheckInGameParams{TransactionId: transactionID.String()})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestCheckOutGame(t *testing.T) {
	t.Run("Should return 201 Created when game is checked out successfully", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		patronID := uuid.New()
		transactionID := uuid.New()
		now := time.Now().UTC()

		// 1. GetGameStatus - Game is available
		mockRowStatus := new(MockRow)
		mockRowStatus.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = "Catan"
			*args.Get(2).(*string) = "catan"
			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(7).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRowStatus).Once()

		// 2. CheckOutGame - Perform checkout
		mockRowCheckOut := new(MockRow)
		mockRowCheckOut.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: transactionID, Valid: true}
			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Time: now, Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.UUID{Bytes: patronID, Valid: true}}).Return(mockRowCheckOut).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(CheckOutRequest{GameId: gameID, PatronId: patronID})
		c.Request = httptest.NewRequest("POST", "/transactions/checkout", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.CheckOutGame(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response LibraryTransaction
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, gameID, response.GameId)
		assert.Equal(t, patronID, response.PatronId)
		assert.Equal(t, transactionID, response.Id)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 201 Created when game is already checked out by the same patron", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		patronID := uuid.New()
		transactionID := uuid.New()
		now := time.Now().UTC()

		mockRowStatus := new(MockRow)
		mockRowStatus.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = "Catan"
			*args.Get(2).(*string) = "catan"
			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(4).(*pgtype.Text) = pgtype.Text{String: "John Doe", Valid: true}
			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Bytes: transactionID, Valid: true}
			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Time: now, Valid: true}
			*args.Get(7).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRowStatus)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(CheckOutRequest{GameId: gameID, PatronId: patronID})
		c.Request = httptest.NewRequest("POST", "/transactions/checkout", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.CheckOutGame(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response LibraryTransaction
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, gameID, response.GameId)
		assert.Equal(t, patronID, response.PatronId)
		assert.Equal(t, transactionID, response.Id)
	})

	t.Run("Should return 409 Conflict when game is already checked out by another patron", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		patronID := uuid.New()
		otherPatronID := uuid.New()
		transactionID := uuid.New()
		now := time.Now().UTC()

		mockRowStatus := new(MockRow)
		mockRowStatus.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = "Catan"
			*args.Get(2).(*string) = "catan"
			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Bytes: otherPatronID, Valid: true}
			*args.Get(4).(*pgtype.Text) = pgtype.Text{String: "Jane Doe", Valid: true}
			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Bytes: transactionID, Valid: true}
			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Time: now, Valid: true}
			*args.Get(7).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRowStatus)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(CheckOutRequest{GameId: gameID, PatronId: patronID})
		c.Request = httptest.NewRequest("POST", "/transactions/checkout", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.CheckOutGame(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Game is already checked out by another patron")
	})

	t.Run("Should return 400 Bad Request when JSON is malformed", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/transactions/checkout", bytes.NewBufferString("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		server.CheckOutGame(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs during status check", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		patronID := uuid.New()

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(CheckOutRequest{GameId: gameID, PatronId: patronID})
		c.Request = httptest.NewRequest("POST", "/transactions/checkout", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.CheckOutGame(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}
