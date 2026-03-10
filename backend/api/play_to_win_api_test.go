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

		err := server.addPlayToWin(c, gameID, nil)

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

		err := server.addPlayToWin(c, gameID, nil)

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

		err := server.addPlayToWin(c, gameID, nil)

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

// --- GetPlayToWinSessionEntries ---

func TestGetPlayToWinSessionEntries(t *testing.T) {
	t.Run("Should return 200 OK with list of entries when entries exist", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()
		sessionID := uuid.New()
		entryID1 := uuid.New()
		entryID2 := uuid.New()

		mockRows := new(MockRows)
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID1, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: playToWinID, Valid: true}
				*args.Get(3).(*string) = "Alice Smith"
				*args.Get(4).(*string) = "alice123"
				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			}).Return(nil).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID2, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: playToWinID, Valid: true}
				*args.Get(3).(*string) = "Bob Jones"
				*args.Get(4).(*string) = "bob456"
				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			}).Return(nil).Once()
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: playToWinID, Valid: true}}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/ptw/"+playToWinID.String()+"/entries", nil)

		server.GetPlayToWinSessionEntries(c, playToWinID)

		assert.Equal(t, http.StatusOK, w.Code)
		var response PlayToWinEntryList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Entries, 2)
		assert.Equal(t, entryID1, response.Entries[0].EntryId)
		assert.Equal(t, "Alice Smith", response.Entries[0].EntrantName)
		assert.Equal(t, "alice123", response.Entries[0].EntrantUniqueId)
		assert.Equal(t, entryID2, response.Entries[1].EntryId)
		assert.Equal(t, "Bob Jones", response.Entries[1].EntrantName)
		assert.Equal(t, "bob456", response.Entries[1].EntrantUniqueId)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 200 OK with empty list when no entries exist", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()

		mockRows := new(MockRows)
		mockRows.On("Next").Return(false)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: playToWinID, Valid: true}}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/ptw/"+playToWinID.String()+"/entries", nil)

		server.GetPlayToWinSessionEntries(c, playToWinID)

		assert.Equal(t, http.StatusOK, w.Code)
		var response PlayToWinEntryList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Empty(t, response.Entries)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when DB query fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: playToWinID, Valid: true}}).Return(nil, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/ptw/"+playToWinID.String()+"/entries", nil)

		server.GetPlayToWinSessionEntries(c, playToWinID)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
		mockDB.AssertExpectations(t)
	})
}

// --- AddPlayToWinSession ---

func validSessionBody(t *testing.T, playToWinID uuid.UUID, playtimeMinutes *int32, entries []struct{ name, uniqueId string }) *bytes.Buffer {
	t.Helper()

	// Build the request using the generated CreatePlayToWinSessionRequest type
	request := CreatePlayToWinSessionRequest{
		PlayToWinId:     playToWinID,
		PlaytimeMinutes: playtimeMinutes,
		Entries: make([]struct {
			EntrantName     string `json:"entrantName"`
			EntrantUniqueId string `json:"entrantUniqueId"`
		}, len(entries)),
	}

	for i, entry := range entries {
		request.Entries[i].EntrantName = entry.name
		request.Entries[i].EntrantUniqueId = entry.uniqueId
	}

	body, _ := json.Marshal(request)
	return bytes.NewBuffer(body)
}

func TestAddPlayToWinSession(t *testing.T) {
	t.Run("Should return 201 Created when session is successfully created with entries", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()
		sessionID := uuid.New()
		entryID1 := uuid.New()
		entryID2 := uuid.New()
		playtimeMinutes := int32(45)

		entries := []struct{ name, uniqueId string }{
			{name: "Alice Smith", uniqueId: "alice123"},
			{name: "Bob Jones", uniqueId: "bob456"},
		}

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// Mock CreatePlayToWinSession - returns 7 columns
		mockSessionRow := new(MockRow)
		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: playToWinID, Valid: true}
				*args.Get(2).(*pgtype.Int4) = pgtype.Int4{Int32: playtimeMinutes, Valid: true}
				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{Valid: false}
				*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: playToWinID, Valid: true},
			pgtype.Int4{Int32: playtimeMinutes, Valid: true},
		}).Return(mockSessionRow).Once()

		// Mock CreatePlayToWinEntry for first entry - returns 8 columns
		mockEntry1Row := new(MockRow)
		mockEntry1Row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID1, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(2).(*string) = entries[0].name
				*args.Get(3).(*string) = entries[0].uniqueId
				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(6).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{Valid: false}
				*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: sessionID, Valid: true},
			entries[0].name,
			entries[0].uniqueId,
		}).Return(mockEntry1Row).Once()

		// Mock CreatePlayToWinEntry for second entry - returns 8 columns
		mockEntry2Row := new(MockRow)
		mockEntry2Row.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID2, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(2).(*string) = entries[1].name
				*args.Get(3).(*string) = entries[1].uniqueId
				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(6).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{Valid: false}
				*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: sessionID, Valid: true},
			entries[1].name,
			entries[1].uniqueId,
		}).Return(mockEntry2Row).Once()

		mockTx.On("Commit", mock.Anything).Return(nil)
		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, &playtimeMinutes, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response PlayToWinSession
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, sessionID, response.SessionId)
		assert.NotNil(t, response.PlaytimeMinutes)
		assert.Equal(t, playtimeMinutes, *response.PlaytimeMinutes)
		assert.Len(t, response.PlayToWinEntries, 2)
		assert.Equal(t, entryID1, response.PlayToWinEntries[0].EntryId)
		assert.Equal(t, entries[0].name, response.PlayToWinEntries[0].EntrantName)
		assert.Equal(t, entries[0].uniqueId, response.PlayToWinEntries[0].EntrantUniqueId)
		assert.Equal(t, entryID2, response.PlayToWinEntries[1].EntryId)
		assert.Equal(t, entries[1].name, response.PlayToWinEntries[1].EntrantName)
		assert.Equal(t, entries[1].uniqueId, response.PlayToWinEntries[1].EntrantUniqueId)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 201 Created when session is created without playtime", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()
		sessionID := uuid.New()
		entryID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice Smith", uniqueId: "alice123"},
		}

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// Mock CreatePlayToWinSession without playtime - returns 7 columns
		mockSessionRow := new(MockRow)
		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: playToWinID, Valid: true}
				*args.Get(2).(*pgtype.Int4) = pgtype.Int4{Valid: false}
				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{Valid: false}
				*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: playToWinID, Valid: true},
			pgtype.Int4{Valid: false},
		}).Return(mockSessionRow).Once()

		// Mock CreatePlayToWinEntry - returns 8 columns
		mockEntryRow := new(MockRow)
		mockEntryRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: entryID, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(2).(*string) = entries[0].name
				*args.Get(3).(*string) = entries[0].uniqueId
				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(6).(*db.NullPlayToWinEntryDeletionType) = db.NullPlayToWinEntryDeletionType{Valid: false}
				*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: sessionID, Valid: true},
			entries[0].name,
			entries[0].uniqueId,
		}).Return(mockEntryRow).Once()

		mockTx.On("Commit", mock.Anything).Return(nil)
		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response PlayToWinSession
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Nil(t, response.PlaytimeMinutes)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when JSON is malformed", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", bytes.NewBufferString("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 400 Bad Request when playtimeMinutes is negative", func(t *testing.T) {
		server, _ := setupTestServer()
		playToWinID := uuid.New()
		playtimeMinutes := int32(-5)

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: "alice123"},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, &playtimeMinutes, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 400 Bad Request when entrantName is empty", func(t *testing.T) {
		server, _ := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "", uniqueId: "alice123"},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 400 Bad Request when entrantName exceeds 100 characters", func(t *testing.T) {
		server, _ := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: string(make([]byte, 101)), uniqueId: "alice123"},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 400 Bad Request when entrantUniqueId is empty", func(t *testing.T) {
		server, _ := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: ""},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 400 Bad Request when entrantUniqueId exceeds 100 characters", func(t *testing.T) {
		server, _ := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: string(make([]byte, 101))},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 404 Not Found when play to win game does not exist (FK violation)", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: "alice123"},
		}

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		pgErr := &pgconn.PgError{Code: "23503"}
		mockSessionRow := new(MockRow)
		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(pgErr)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: playToWinID, Valid: true},
			pgtype.Int4{Valid: false},
		}).Return(mockSessionRow).Once()

		mockTx.On("Rollback", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when transaction begin fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: "alice123"},
		}

		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(nil, errors.New("tx error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "tx error")
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when CreatePlayToWinSession fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: "alice123"},
		}

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockSessionRow := new(MockRow)
		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("db error"))
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: playToWinID, Valid: true},
			pgtype.Int4{Valid: false},
		}).Return(mockSessionRow).Once()

		mockTx.On("Rollback", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when CreatePlayToWinEntry fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		playToWinID := uuid.New()
		sessionID := uuid.New()

		entries := []struct{ name, uniqueId string }{
			{name: "Alice", uniqueId: "alice123"},
		}

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// Mock successful session creation - 7 columns
		mockSessionRow := new(MockRow)
		mockSessionRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: sessionID, Valid: true}
				*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: playToWinID, Valid: true}
				*args.Get(2).(*pgtype.Int4) = pgtype.Int4{Valid: false}
				*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
				*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
				*args.Get(5).(*db.NullPlayToWinSessionDeletionType) = db.NullPlayToWinSessionDeletionType{Valid: false}
				*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
			}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: playToWinID, Valid: true},
			pgtype.Int4{Valid: false},
		}).Return(mockSessionRow).Once()

		// Mock failed entry creation - 8 columns
		mockEntryRow := new(MockRow)
		mockEntryRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("entry creation error"))
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{
			pgtype.UUID{Bytes: sessionID, Valid: true},
			entries[0].name,
			entries[0].uniqueId,
		}).Return(mockEntryRow).Once()

		mockTx.On("Commit", mock.Anything).Return(nil).Maybe()
		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/ptw/session", validSessionBody(t, playToWinID, nil, entries))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPlayToWinSession(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "entry creation error")
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}
