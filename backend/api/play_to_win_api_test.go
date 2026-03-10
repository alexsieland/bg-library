package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- AddPlayToWinGame ---

func TestAddPlayToWinGame(t *testing.T) {
	t.Run("Should return 204 No Content when game is successfully marked as Play to Win", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwID := uuid.New()

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
				*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(4).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
				*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		server.AddPlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 404 Not Found when game does not exist (foreign key violation)", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		pgErr := &pgconn.PgError{Code: "23503"}
		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgErr)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		server.AddPlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 204 No Content when game is already marked as Play to Win (idempotent)", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		pgErr := &pgconn.PgError{Code: "23505"}
		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgErr)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		server.AddPlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when DB returns an unexpected error", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("unexpected db error"))
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		server.AddPlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestAddPlayToWin(t *testing.T) {
	t.Run("Should return nil when game is successfully marked as Play to Win", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwID := uuid.New()

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
				*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(4).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
				*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		err := server.addPlayToWin(c, gameID, []ErrorDetail{}, nil)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return nil when game is already marked as Play to Win (idempotent)", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(&pgconn.PgError{Code: "23505"})
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		err := server.addPlayToWin(c, gameID, []ErrorDetail{}, nil)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return error when DB returns an unexpected error", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		expectedErr := errors.New("unexpected db error")

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(expectedErr)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/game/gameId/"+gameID.String(), nil)

		err := server.addPlayToWin(c, gameID, []ErrorDetail{}, nil)

		assert.ErrorIs(t, err, expectedErr)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

// --- RemovePlayToWinGame ---

func mockGetGameRow(mockDB *MockDatabase, gameID uuid.UUID, ptwGameID uuid.UUID) {
	mockRow := new(MockRow)
	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = "Catan"
			*args.Get(2).(*string) = "catan"
			*args.Get(3).(*pgtype.Text) = pgtype.Text{Valid: false}
			*args.Get(4).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwGameID, Valid: true}
			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		}).Return(nil)
	mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
}

func validRemoveBody(t *testing.T, reason RemovePlayToWinGameRequestRemovalReason, comment *string) *bytes.Buffer {
	t.Helper()
	body, _ := json.Marshal(RemovePlayToWinGameJSONRequestBody{
		RemovalReason:  reason,
		RemovalComment: comment,
	})
	return bytes.NewBuffer(body)
}

func TestRemovePlayToWinGame(t *testing.T) {
	t.Run("Should return 204 No Content when game is successfully removed from Play to Win", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwGameID := uuid.New()

		mockGetGameRow(mockDB, gameID, ptwGameID)
		mockDB.On("Exec", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: ptwGameID, Valid: true},
			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeClaimed, Valid: true},
			pgtype.Text{Valid: false},
		}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "claimed", nil))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should pass the removal comment to the DB when provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwGameID := uuid.New()
		comment := "Librarian confirmed the game was claimed"

		mockGetGameRow(mockDB, gameID, ptwGameID)
		mockDB.On("Exec", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: ptwGameID, Valid: true},
			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeClaimed, Valid: true},
			pgtype.Text{String: comment, Valid: true},
		}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "claimed", &comment))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when request body is malformed JSON", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwGameID := uuid.New()

		mockGetGameRow(mockDB, gameID, ptwGameID)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), bytes.NewBufferString("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when removal comment exceeds 500 characters", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwGameID := uuid.New()
		longComment := string(make([]byte, 501))

		mockGetGameRow(mockDB, gameID, ptwGameID)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "other", &longComment))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 404 Not Found when game does not exist", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgx.ErrNoRows)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "mistake", nil))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when GetGame returns an unexpected error", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "other", nil))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when DeletePlayToWinGame fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwGameID := uuid.New()

		mockGetGameRow(mockDB, gameID, ptwGameID)
		mockDB.On("Exec", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: ptwGameID, Valid: true},
			db.NullPlayToWinGameDeletionType{PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake, Valid: true},
			pgtype.Text{Valid: false},
		}).Return(pgconn.CommandTag{}, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/ptw/game/gameId/"+gameID.String(), validRemoveBody(t, "mistake", nil))
		c.Request.Header.Set("Content-Type", "application/json")

		server.RemovePlayToWinGame(c, gameID)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
	})
}
