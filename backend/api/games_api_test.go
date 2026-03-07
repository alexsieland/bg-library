package api

import (
	"bytes"
	"encoding/base64"
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

func TestAddGame(t *testing.T) {
	t.Run("Should return 201 Created when valid game is added", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}  // created_at
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response Game
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, title, response.Title)
		assert.Equal(t, gameID, response.GameId)
		assert.False(t, response.IsPlayToWin)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when JSON is malformed", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBufferString("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 400 Bad Request when title is too long", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: string(make([]byte, 101))})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when database returns an error", func(t *testing.T) {
		server, mockDB := setupTestServer()
		title := "Catan"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})

	t.Run("Should return 201 Created with barcode when barcode is provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"
		barcode := "9780000000001"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(5).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, Barcode: &barcode})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response Game
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, title, response.Title)
		assert.Equal(t, gameID, response.GameId)
		assert.NotNil(t, response.Barcode)
		assert.Equal(t, barcode, *response.Barcode)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when barcode exceeds 48 characters", func(t *testing.T) {
		server, _ := setupTestServer()
		title := "Catan"
		barcode := string(make([]byte, 49)) // 49 chars

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, Barcode: &barcode})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 201 Created and call addPlayToWin when isPlayToWin flag is set", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		ptwID := uuid.New()
		title := "Catan"
		isPlayToWin := true

		// CreateGame call
		mockCreateRow := new(MockRow)
		mockCreateRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockCreateRow).Once()

		// CreatePlayToWinGame call
		mockPtwRow := new(MockRow)
		mockPtwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(4).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockPtwRow).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, IsPlayToWin: &isPlayToWin})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response Game
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, title, response.Title)
		// Both CreateGame and CreatePlayToWinGame DB calls are made
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when addPlayToWin fails after game creation", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"
		isPlayToWin := true

		// CreateGame succeeds
		mockCreateRow := new(MockRow)
		mockCreateRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockCreateRow).Once()

		// CreatePlayToWinGame fails
		mockPtwRow := new(MockRow)
		mockPtwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ptw db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockPtwRow).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, IsPlayToWin: &isPlayToWin})
		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
	})
}

func TestDeleteGame(t *testing.T) {
	t.Run("Should return 204 No Content when game is deleted", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/games/"+gameID.String(), nil)
		server.DeleteGame(c, gameID.String())

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when gameId is invalid UUID", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		server.DeleteGame(c, "invalid-uuid")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs on delete", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/games/"+gameID.String(), nil)
		server.DeleteGame(c, gameID.String())

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestGetGame(t *testing.T) {
	t.Run("Should return 200 OK when game is found", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Text) = pgtype.Text{Valid: false}          // barcode
			*args.Get(4).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/"+gameID.String(), nil)
		server.GetGame(c, gameID.String())

		assert.Equal(t, http.StatusOK, w.Code)
		var response Game
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, title, response.Title)
		assert.Equal(t, gameID, response.GameId)
	})

	t.Run("Should return 404 Not Found when game does not exist", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/"+gameID.String(), nil)
		server.GetGame(c, gameID.String())

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetGameByBarcode(t *testing.T) {
	t.Run("Should return 200 OK with a list of games when games are found by barcode", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"
		barcode := "1234567890"

		mockRows := new(MockRows)
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
			*args.Get(4).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
		}).Return(nil)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
		server.GetGameByBarcode(c, barcode)

		assert.Equal(t, http.StatusOK, w.Code)
		var response GameList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Games, 1)
		assert.Equal(t, title, response.Games[0].Title)
		assert.Equal(t, gameID, response.Games[0].GameId)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 404 Not Found when no games match the barcode", func(t *testing.T) {
		server, mockDB := setupTestServer()
		barcode := "1234567890"

		mockRows := new(MockRows)
		mockRows.On("Next").Return(false)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
		server.GetGameByBarcode(c, barcode)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when barcode is empty", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/barcode/", nil)
		server.GetGameByBarcode(c, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 400 Bad Request when barcode exceeds 48 characters", func(t *testing.T) {
		server, _ := setupTestServer()
		barcode := "1234567890123456789012345678901234567890123456789" // 49 chars

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
		server.GetGameByBarcode(c, barcode)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 500 Internal Server Error when database error occurs", func(t *testing.T) {
		server, mockDB := setupTestServer()
		barcode := "1234567890"

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(nil, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
		server.GetGameByBarcode(c, barcode)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
		mockDB.AssertExpectations(t)
	})
}

func TestUpdateGame(t *testing.T) {
	t.Run("Should return 204 No Content when game is updated", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Updated Catan"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdateGameJSONRequestBody{Title: title})
		c.Request = httptest.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdateGame(c, gameID.String())

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Should return 404 Not Found when updating non-existent game", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Updated Catan"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: "23503"})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdateGameJSONRequestBody{Title: title})
		c.Request = httptest.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdateGame(c, gameID.String())

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Should return 204 No Content when game is updated with barcode", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Updated Catan"
		barcode := "9780000000001"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, title, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdateGameJSONRequestBody{Title: title, Barcode: &barcode})
		c.Request = httptest.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdateGame(c, gameID.String())

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})
}

func TestListGames(t *testing.T) {
	t.Run("Should return 200 OK with list of games when called without title", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"

		mockRows := new(MockRows)
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(7).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
			*args.Get(8).(*pgtype.UUID) = pgtype.UUID{Valid: false} // play_to_win_game_id
		}).Return(nil)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(999), int32(0)}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games", nil)
		server.ListGames(c, ListGamesParams{})

		assert.Equal(t, http.StatusOK, w.Code)
		var response GameStatusList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Games, 1)
		assert.Equal(t, title, response.Games[0].Game.Title)
	})

	t.Run("Should return 200 OK with searched games when title is provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		title := "Catan"

		mockRows := new(MockRows)
		mockRows.On("Next").Return(false)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{"%" + SanitizeTitle(title) + "%", int32(999), int32(0)}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games?title=Catan", nil)
		server.ListGames(c, ListGamesParams{Title: &title})

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should call listCheckedOutGames when CheckedOut is true", func(t *testing.T) {
		server, mockDB := setupTestServer()
		checkedOut := true

		mockRows := new(MockRows)
		mockRows.On("Next").Return(false)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(999), int32(0)}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/games?checkedOut=true", nil)
		server.ListGames(c, ListGamesParams{CheckedOut: &checkedOut})

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})
}

func TestBulkAddGames(t *testing.T) {
	t.Run("Should return 201 Created with imported count when valid CSV is provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID1 := uuid.New()
		gameID2 := uuid.New()
		title1 := "Catan"
		title2 := "Ticket to Ride"

		// CSV content: "Catan"\n"Ticket to Ride"
		csvContent := title1 + "\n" + title2
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// First game
		mockRow1 := new(MockRow)
		mockRow1.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID1, Valid: true}
			*args.Get(1).(*string) = title1
			*args.Get(2).(*string) = SanitizeTitle(title1)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title1, SanitizeTitle(title1), pgtype.Text{Valid: false}}).Return(mockRow1)

		// Second game
		mockRow2 := new(MockRow)
		mockRow2.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID2, Valid: true}
			*args.Get(1).(*string) = title2
			*args.Get(2).(*string) = SanitizeTitle(title2)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title2, SanitizeTitle(title2), pgtype.Text{Valid: false}}).Return(mockRow2)

		mockTx.On("Commit", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddGames(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response BulkAddResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Imported)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should skip invalid records and continue processing when validation fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		validTitle := "Catan"
		invalidTitle := "" // Empty title should fail validation

		csvContent := validTitle + "\n" + invalidTitle
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// Only the valid game
		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = validTitle
			*args.Get(2).(*string) = SanitizeTitle(validTitle)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{validTitle, SanitizeTitle(validTitle), pgtype.Text{Valid: false}}).Return(mockRow)

		mockTx.On("Commit", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddGames(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response BulkAddResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.Imported)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when transaction fails to begin", func(t *testing.T) {
		server, mockDB := setupTestServer()
		csvContent := "Catan"
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(nil, errors.New("transaction error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddGames(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "transaction error")
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs during insert", func(t *testing.T) {
		server, mockDB := setupTestServer()
		title := "Catan"
		csvContent := title
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)
		mockTx.On("Rollback", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddGames(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when transaction commit fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		gameID := uuid.New()
		title := "Catan"
		csvContent := title
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
			*args.Get(1).(*string) = title
			*args.Get(2).(*string) = SanitizeTitle(title)
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(5).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)

		mockTx.On("Commit", mock.Anything).Return(errors.New("commit error"))
		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddGames(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "commit error")
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should handle empty CSV and return 201 with zero imported", func(t *testing.T) {
		server, mockDB := setupTestServer()
		csvContent := ""
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddGames(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response BulkAddResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Imported)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}
