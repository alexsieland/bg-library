package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// This test file provides lightweight, fixture-based unit tests for the
// Game API handlers. It mocks the underlying gameService interface and the
// libraryService transaction handling to focus tests on input validation and
// correct wiring.

type mockGameService struct{ mock.Mock }

func (m *mockGameService) ListGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	args := m.Called(ctx, gameTitle, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwLibraryGame), args.Error(1)
}
func (m *mockGameService) ListGameStatuses(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error) {
	args := m.Called(ctx, gameTitle, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwGameStatus), args.Error(1)
}
func (m *mockGameService) ListCheckedOutGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error) {
	args := m.Called(ctx, gameTitle, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwGameStatus), args.Error(1)
}
func (m *mockGameService) InsertGame(ctx context.Context, title string, barcode *string, isPlayToWin bool, optTx pgx.Tx) (db.VwLibraryGame, error) {
	args := m.Called(ctx, title, barcode, isPlayToWin, optTx)
	if args.Get(0) == nil {
		return db.VwLibraryGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwLibraryGame), args.Error(1)
}
func (m *mockGameService) UpdateGame(ctx context.Context, gameId pgtype.UUID, title string, barcode *string, optTx pgx.Tx) error {
	args := m.Called(ctx, gameId, title, barcode, optTx)
	return args.Error(0)
}
func (m *mockGameService) GetGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwLibraryGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwLibraryGame), args.Error(1)
}
func (m *mockGameService) GetGamesByBarcode(ctx context.Context, barcode string, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	args := m.Called(ctx, barcode, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwLibraryGame), args.Error(1)
}
func (m *mockGameService) GetGameStatus(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwGameStatus, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwGameStatus{}, args.Error(1)
	}
	return args.Get(0).(db.VwGameStatus), args.Error(1)
}
func (m *mockGameService) DeleteGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) error {
	args := m.Called(ctx, gameId, optTx)
	return args.Error(0)
}
func (m *mockGameService) SetIsPlayToWin(ctx context.Context, gameId pgtype.UUID, isPlayToWin bool, optTx pgx.Tx) error {
	args := m.Called(ctx, gameId, isPlayToWin, optTx)
	return args.Error(0)
}

// testLibService mirrors the small test lib used by patrons tests: it returns
// a preconfigured tx or error when BeginTx is called.

func newTestGameApi(service *mockGameService, tx pgx.Tx, beginErr error) *GameApi {
	lib := &testLibService{tx: tx, err: beginErr}
	return &GameApi{libraryService: lib, service: service}
}

// --- Tests ---------------------------------------------------------------

func TestAddGame(t *testing.T) {
	t.Run("Should return converted game when service succeeds", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		id := uuid.New()
		title := "Catan"
		dbg := db.VwLibraryGame{
			ID:              pgtype.UUID{Bytes: id, Valid: true},
			DisplayTitle:    title,
			Title:           title,
			SanitizedTitle:  internal.SanitizeTitle(title),
			Barcode:         pgtype.Text{Valid: false},
			PlayToWinGameID: pgtype.UUID{Valid: false},
			CreatedAt:       pgtype.Timestamp{Valid: true},
		}

		svc.On("InsertGame", ctx, title, (*string)(nil), false, (pgx.Tx)(nil)).Return(dbg, nil).Once()

		got, err := fixture.AddGame(ctx, CreateGameRequest{Title: title})
		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryGame(dbg), got)
		svc.AssertExpectations(t)
	})

	t.Run("Should return validation error when title is too long", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		longTitle := strings.Repeat("x", 101)

		got, err := fixture.AddGame(ctx, CreateGameRequest{Title: longTitle})
		assert.Equal(t, Game{}, got)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "title", Message: "Length must be between 1 and 100"}}})
	})

	t.Run("Should propagate service error when insert fails", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		title := "Catan"
		expected := errors.New("insert failed")

		svc.On("InsertGame", ctx, title, (*string)(nil), false, (pgx.Tx)(nil)).Return(db.VwLibraryGame{}, expected).Once()

		got, err := fixture.AddGame(ctx, CreateGameRequest{Title: title})
		assert.Equal(t, Game{}, got)
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})

	t.Run("Should return converted game when barcode provided", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		id := uuid.New()
		title := "Catan"
		barcode := "9780000000001"
		dbg := db.VwLibraryGame{
			ID:              pgtype.UUID{Bytes: id, Valid: true},
			DisplayTitle:    title,
			Title:           title,
			SanitizedTitle:  internal.SanitizeTitle(title),
			Barcode:         pgtype.Text{String: barcode, Valid: true},
			PlayToWinGameID: pgtype.UUID{Valid: false},
			CreatedAt:       pgtype.Timestamp{Valid: true},
		}

		svc.On("InsertGame", ctx, title, &barcode, false, (pgx.Tx)(nil)).Return(dbg, nil).Once()

		got, err := fixture.AddGame(ctx, CreateGameRequest{Title: title, Barcode: &barcode})
		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryGame(dbg), got)
		svc.AssertExpectations(t)
	})

	t.Run("Should return validation error when barcode is too long", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		tooLong := strings.Repeat("b", 49)

		got, err := fixture.AddGame(ctx, CreateGameRequest{Title: "Ok", Barcode: &tooLong})
		assert.Equal(t, Game{}, got)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "barcode", Message: "Length must be between 1 and 48"}}})
	})
}

func TestBulkAddGames(t *testing.T) {
	t.Run("Should import rows and commit when CSV is valid", func(t *testing.T) {
		svc := new(mockGameService)
		tx := &stubTx{}
		fixture := newTestGameApi(svc, tx, nil)
		ctx := context.Background()

		title1 := "Catan"
		title2 := "Ticket"

		svc.On("InsertGame", ctx, title1, (*string)(nil), false, mock.Anything).Return(db.VwLibraryGame{}, nil).Once().Run(func(args mock.Arguments) {
			// ensure service received our tx
			assert.Same(t, tx, args.Get(4))
		})
		svc.On("InsertGame", ctx, title2, (*string)(nil), false, mock.Anything).Return(db.VwLibraryGame{}, nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(4))
		})

		csv := "title\n" + title1 + "\n" + title2 + "\n"
		resp, err := fixture.BulkAddGames(ctx, encodedCSVBody(csv))
		assert.NoError(t, err)
		assert.Equal(t, BulkAddResponse{Imported: 2}, resp)
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 0, tx.rollbackCount)
		svc.AssertExpectations(t)
	})

	t.Run("Should return error when request body is not valid base64", func(t *testing.T) {
		svc := new(mockGameService)
		tx := &stubTx{}
		fixture := newTestGameApi(svc, tx, nil)
		ctx := context.Background()

		resp, err := fixture.BulkAddGames(ctx, io.NopCloser(strings.NewReader("%%%")))
		assert.Error(t, err)
		assert.Equal(t, BulkAddResponse{}, resp)
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
	})

	t.Run("Should accumulate validation errors and skip inserts when CSV rows are invalid", func(t *testing.T) {
		svc := new(mockGameService)
		tx := &stubTx{}
		fixture := newTestGameApi(svc, tx, nil)
		ctx := context.Background()

		tooLong := strings.Repeat("b", 49)
		csv := "title,barcode\n,\nValid," + tooLong + "\n"
		resp, err := fixture.BulkAddGames(ctx, encodedCSVBody(csv))
		assert.Equal(t, BulkAddResponse{}, resp)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "title", Message: "Length must be between 1 and 100"}, {Field: "barcode", Message: "Length must be between 1 and 48"}}})
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
	})
}

func TestDeleteGetListBulkErrors(t *testing.T) {
	t.Run("Should return service error when delete fails", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		id := uuid.New()
		expected := errors.New("delete failed")

		svc.On("DeleteGame", ctx, testUUID(id), (pgx.Tx)(nil)).Return(expected).Once()

		err := fixture.DeleteGame(ctx, types.UUID(id))
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})

	t.Run("Should return error when GetGame not found", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		id := uuid.New()

		svc.On("GetGame", ctx, testUUID(id), (pgx.Tx)(nil)).Return(db.VwLibraryGame{}, errors.New("not found")).Once()

		got, err := fixture.GetGame(ctx, types.UUID(id))
		assert.Equal(t, Game{}, got)
		assert.Error(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should return db error when GetGameByBarcode fails", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		barcode := "123456"
		expected := errors.New("db err")

		svc.On("GetGamesByBarcode", ctx, barcode, (pgx.Tx)(nil)).Return(nil, expected).Once()

		got, err := fixture.GetGameByBarcode(ctx, barcode)
		assert.Equal(t, GameList{}, got)
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})

	t.Run("Should return empty list when ListGames returns no items", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()

		svc.On("ListGameStatuses", ctx, (*string)(nil), mock.Anything, mock.Anything, (pgx.Tx)(nil)).Return([]db.VwGameStatus{}, nil).Once()

		resp, err := fixture.ListGames(ctx, ListGamesParams{})
		assert.NoError(t, err)
		assert.Len(t, resp.Games, 0)
		svc.AssertExpectations(t)
	})

	t.Run("Should return transaction error when begin fails in bulk import", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, errors.New("begin failed"))

		resp, err := fixture.BulkAddGames(context.Background(), encodedCSVBody("title\nCatan\n"))
		assert.Equal(t, BulkAddResponse{}, resp)
		assert.Error(t, err)
	})

	t.Run("Should return commit error when tx commit fails in bulk import", func(t *testing.T) {
		svc := new(mockGameService)
		tx := &stubTx{commitErr: errors.New("commit failed")}
		fixture := newTestGameApi(svc, tx, nil)

		svc.On("InsertGame", mock.Anything, "Catan", (*string)(nil), false, mock.Anything).Return(db.VwLibraryGame{}, nil).Once()

		resp, err := fixture.BulkAddGames(context.Background(), encodedCSVBody("title\nCatan\n"))
		assert.Equal(t, BulkAddResponse{}, resp)
		assert.Error(t, err)
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		svc.AssertExpectations(t)
	})
}

func TestDeleteGetUpdateListGames(t *testing.T) {
	t.Run("Should forward converted id when deleting a game", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		id := uuid.New()

		svc.On("DeleteGame", ctx, testUUID(id), (pgx.Tx)(nil)).Return(nil).Once()
		err := fixture.DeleteGame(ctx, types.UUID(id))
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should return a converted game when the service finds one", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		id := uuid.New()
		title := "Catan"
		dbg := db.VwLibraryGame{ID: pgtype.UUID{Bytes: id, Valid: true}, DisplayTitle: title, Title: title, SanitizedTitle: internal.SanitizeTitle(title), Barcode: pgtype.Text{Valid: false}, CreatedAt: pgtype.Timestamp{Valid: true}}

		svc.On("GetGame", ctx, testUUID(id), (pgx.Tx)(nil)).Return(dbg, nil).Once()
		got, err := fixture.GetGame(ctx, types.UUID(id))
		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryGame(dbg), got)
		svc.AssertExpectations(t)
	})

	t.Run("Should return validation error when the barcode is invalid", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		tooLong := strings.Repeat("b", 49)
		got, err := fixture.GetGameByBarcode(context.Background(), tooLong)
		assert.Equal(t, GameList{}, got)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "gameBarcode", Message: "Length must be between 1 and 48"}}})
	})

	t.Run("Should delegate to the service when listing games with a title", func(t *testing.T) {
		svc := new(mockGameService)
		fixture := newTestGameApi(svc, nil, nil)
		ctx := context.Background()
		title := "Search"
		statuses := []db.VwGameStatus{}

		svc.On("ListGameStatuses", ctx, &title, mock.Anything, mock.Anything, (pgx.Tx)(nil)).Return(statuses, nil).Once()
		_, err := fixture.ListGames(ctx, ListGamesParams{Title: &title})
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should commit transaction when update and set play-to-win succeed", func(t *testing.T) {
		svc := new(mockGameService)
		tx := &stubTx{}
		fixture := newTestGameApi(svc, tx, nil)
		ctx := context.Background()
		id := uuid.New()
		title := "Updated"

		svc.On("UpdateGame", ctx, testUUID(id), title, (*string)(nil), mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
			// optTx should be the last argument; assert same if present
			if len(args) > 0 {
				last := args[len(args)-1]
				assert.Same(t, tx, last)
			}
		})
		svc.On("SetIsPlayToWin", ctx, testUUID(id), false, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
			if len(args) > 0 {
				last := args[len(args)-1]
				assert.Same(t, tx, last)
			}
		})

		err := fixture.UpdateGame(ctx, id, CreateGameRequest{Title: title})
		assert.NoError(t, err)
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 0, tx.rollbackCount)
		svc.AssertExpectations(t)
	})
}

// Handler-level tests that exercise Server error handling behavior.
// These sit above the legacy, DB-mock-based commented tests for reference.
func TestServerHandlerErrors(t *testing.T) {
	t.Run("Should return 400 Bad Request when AddGame JSON is malformed", func(t *testing.T) {
		server, _, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/games", strings.NewReader("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddGame(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 404 Not Found when an internal ErrNotFound is handled", func(t *testing.T) {
		// create a Gin context and directly exercise handleError to assert mapping
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handleError(c, internal.ErrNotFound)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBufferString("{invalid json}"))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "JSON body is malformed")
//	})
//
//	t.Run("Should return 400 Bad Request when title is too long", func(t *testing.T) {
//		server, _ := setupTestServer()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(AddGameJSONRequestBody{Title: string(make([]byte, 101))})
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 500 Internal Server Error when database returns an error", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		title := "Catan"
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title})
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//	})
//
//	t.Run("Should return 201 Created with barcode when barcode is provided", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Catan"
//		barcode := "9780000000001"
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
//		}).Return(nil)
//
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, Barcode: &barcode})
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response Game
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, title, response.Title)
//		assert.Equal(t, gameID, response.GameId)
//		assert.NotNil(t, response.Barcode)
//		assert.Equal(t, barcode, *response.Barcode)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when barcode exceeds 48 characters", func(t *testing.T) {
//		server, _ := setupTestServer()
//		title := "Catan"
//		barcode := string(make([]byte, 49)) // 49 chars
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, Barcode: &barcode})
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//	})
//
//	t.Run("Should return 201 Created and call addPlayToWinByGameId when isPlayToWin flag is set", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwID := uuid.New()
//		groupID := uuid.New()
//		title := "Catan"
//		isPlayToWin := true
//
//		// CreateGame call (not in tx)
//		mockCreateRow := new(MockRow)
//		mockCreateRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockCreateRow).Once()
//
//		// addPlayToWinByGameId: BeginTx
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil).Once()
//
//		// addPlayToWinByGameId: GetGame via tx
//		mockTxGetGameRow := new(MockRow)
//		mockTxGetGameRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockTxGetGameRow).Once()
//
//		// getOrCreatePlayToWinGroup (1st call): group does not exist yet, create it
//		sp1 := new(MockTx)
//		mockTx.On("Begin", mock.Anything).Return(sp1, nil).Once()
//		mockGroupNotFoundRow := new(MockRow)
//		mockGroupNotFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
//		sp1.On("QueryRow", mock.Anything, mock.Anything, []any{title}).Return(mockGroupNotFoundRow)
//		sp1.On("Rollback", mock.Anything).Return(nil)
//		mockCreateGroupRow := new(MockRow)
//		mockCreateGroupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title}).Return(mockCreateGroupRow).Once()
//
//		// getOrCreatePlayToWinGroup (2nd call - debug): group now exists
//		sp2 := new(MockTx)
//		mockTx.On("Begin", mock.Anything).Return(sp2, nil).Once()
//		mockGroupFoundRow := new(MockRow)
//		mockGroupFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//		sp2.On("QueryRow", mock.Anything, mock.Anything, []any{title}).Return(mockGroupFoundRow)
//		sp2.On("Commit", mock.Anything).Return(nil)
//
//		// CreatePlayToWinGame savepoint
//		sp3 := new(MockTx)
//		mockTx.On("Begin", mock.Anything).Return(sp3, nil).Once()
//		mockPtwRow := new(MockRow)
//		mockPtwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
//			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false} // winner_id
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
//			*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//		sp3.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.UUID{Bytes: groupID, Valid: true}}).Return(mockPtwRow)
//		sp3.On("Commit", mock.Anything).Return(nil)
//
//		// Final tx commit
//		mockTx.On("Commit", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, IsPlayToWin: &isPlayToWin})
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response Game
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, title, response.Title)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when addPlayToWinByGameId fails after game creation", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Catan"
//		isPlayToWin := true
//
//		// CreateGame succeeds
//		mockCreateRow := new(MockRow)
//		mockCreateRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockCreateRow).Once()
//
//		// addPlayToWinByGameId: BeginTx succeeds
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil).Once()
//
//		// GetGame via tx fails
//		mockGetGameRow := new(MockRow)
//		mockGetGameRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("ptw db error"))
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockGetGameRow).Once()
//		mockTx.On("Rollback", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(AddGameJSONRequestBody{Title: title, IsPlayToWin: &isPlayToWin})
//		c.Request = httptest.NewRequest("POST", "/games", bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.AddGame(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//}
//
//func TestDeleteGame(t *testing.T) {
//	t.Run("Should return 204 No Content when game is deleted", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/games/"+gameID.String(), nil)
//		server.DeleteGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when DB error occurs on delete", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(pgconn.CommandTag{}, errors.New("db error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("DELETE", "/games/"+gameID.String(), nil)
//		server.DeleteGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//	})
//}
//
//func TestGetGame(t *testing.T) {
//	t.Run("Should return 200 OK when game is found", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Catan"
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}          // barcode
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
//		}).Return(nil)
//
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/"+gameID.String(), nil)
//		server.GetGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response Game
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, title, response.Title)
//		assert.Equal(t, gameID, response.GameId)
//	})
//
//	t.Run("Should return 404 Not Found when game does not exist", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/"+gameID.String(), nil)
//		server.GetGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusNotFound, w.Code)
//	})
//}
//
//func TestGetGameByBarcode(t *testing.T) {
//	t.Run("Should return 200 OK with a list of games when games are found by barcode", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Catan"
//		barcode := "1234567890"
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
//		}).Return(nil)
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
//		server.GetGameByBarcode(c, barcode)
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response GameList
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Len(t, response.Games, 1)
//		assert.Equal(t, title, response.Games[0].Title)
//		assert.Equal(t, gameID, response.Games[0].GameId)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 404 Not Found when no games match the barcode", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		barcode := "1234567890"
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(false)
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
//		server.GetGameByBarcode(c, barcode)
//
//		assert.Equal(t, http.StatusNotFound, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when barcode is empty", func(t *testing.T) {
//		server, _ := setupTestServer()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/barcode/", nil)
//		server.GetGameByBarcode(c, "")
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//	})
//
//	t.Run("Should return 400 Bad Request when barcode exceeds 48 characters", func(t *testing.T) {
//		server, _ := setupTestServer()
//		barcode := "1234567890123456789012345678901234567890123456789" // 49 chars
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
//		server.GetGameByBarcode(c, barcode)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//	})
//
//	t.Run("Should return 500 Internal Server Error when database error occurs", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		barcode := "1234567890"
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(nil, errors.New("db error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games/barcode/"+barcode, nil)
//		server.GetGameByBarcode(c, barcode)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//		mockDB.AssertExpectations(t)
//	})
//}
//
//func TestUpdateGame(t *testing.T) {
//	t.Run("Should return 204 No Content when game is updated", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Updated Catan"
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}          // barcode
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
//		}).Return(nil)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.Text{String: title, Valid: true}, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, nil)
//		mockTx.On("Commit", mock.Anything).Return(nil)
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(UpdateGameJSONRequestBody{Title: title})
//		c.Request = httptest.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.UpdateGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 404 Not Found when updating non-existent game", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Updated Catan"
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}          // barcode
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
//		}).Return(nil)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.Text{String: title, Valid: true}, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, &pgconn.PgError{Code: "23503"})
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(UpdateGameJSONRequestBody{Title: title})
//		c.Request = httptest.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.UpdateGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusNotFound, w.Code)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 204 No Content when game is updated with barcode", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Updated Catan"
//		barcode := "9780000000001"
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}          // barcode
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}          // play_to_win_game_id
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true} // created_at
//		}).Return(nil)
//		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockRow)
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.Text{String: title, Valid: true}, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}).Return(pgconn.CommandTag{}, nil)
//		mockTx.On("Commit", mock.Anything).Return(nil)
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		body, _ := json.Marshal(UpdateGameJSONRequestBody{Title: title, Barcode: &barcode})
//		c.Request = httptest.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(body))
//		c.Request.Header.Set("Content-Type", "application/json")
//
//		server.UpdateGame(c, types.UUID(gameID))
//
//		assert.Equal(t, http.StatusNoContent, w.Code)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//}
//
//func TestListGames(t *testing.T) {
//	t.Run("Should return 200 OK with list of games when called without title", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Catan"
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(true).Once()
//		mockRows.On("Next").Return(false).Once()
//		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*string) = SanitizeTitle(title)
//			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(7).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(8).(*pgtype.UUID) = pgtype.UUID{Valid: false} // play_to_win_game_id
//		}).Return(nil)
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(999), int32(0)}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games", nil)
//		server.ListGames(c, ListGamesParams{})
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		var response GameStatusList
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Len(t, response.Games, 1)
//		assert.Equal(t, title, response.Games[0].Game.Title)
//	})
//
//	t.Run("Should return 200 OK with searched games when title is provided", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		title := "Catan"
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(false)
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{"%" + SanitizeTitle(title) + "%", int32(999), int32(0)}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games?title=Catan", nil)
//		server.ListGames(c, ListGamesParams{Title: &title})
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should call listCheckedOutGames when CheckedOut is true", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		checkedOut := true
//
//		mockRows := new(MockRows)
//		mockRows.On("Next").Return(false)
//		mockRows.On("Close").Return()
//		mockRows.On("Err").Return(nil)
//
//		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(999), int32(0)}).Return(mockRows, nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("GET", "/games?checkedOut=true", nil)
//		server.ListGames(c, ListGamesParams{CheckedOut: &checkedOut})
//
//		assert.Equal(t, http.StatusOK, w.Code)
//		mockDB.AssertExpectations(t)
//	})
//}
//
//func TestBulkAddGames(t *testing.T) {
//	t.Run("Should return 201 Created with imported count when valid CSV is provided", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID1 := uuid.New()
//		gameID2 := uuid.New()
//		title1 := "Catan"
//		title2 := "Ticket to Ride"
//
//		// CSV content: "Catan"\n"Ticket to Ride"
//		csvContent := "title\n" + title1 + "\n" + title2
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		// First game
//		mockRow1 := new(MockRow)
//		mockRow1.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID1, Valid: true}
//			*args.Get(1).(*string) = title1
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title1)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title1, SanitizeTitle(title1), pgtype.Text{Valid: false}}).Return(mockRow1)
//
//		// Second game
//		mockRow2 := new(MockRow)
//		mockRow2.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID2, Valid: true}
//			*args.Get(1).(*string) = title2
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title2)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title2, SanitizeTitle(title2), pgtype.Text{Valid: false}}).Return(mockRow2)
//
//		mockTx.On("Commit", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response BulkAddResponse
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, int32(2), response.Imported)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 400 Bad Request when any CSV record fails validation", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		validTitle := "Catan"
//		invalidTitle := string(make([]byte, 101)) // Too long title should fail validation
//
//		csvContent := "title\n" + validTitle + "\n" + invalidTitle
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		// Valid row is still attempted before final validation response is returned.
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = validTitle
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(validTitle)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{validTitle, SanitizeTitle(validTitle), pgtype.Text{Valid: false}}).Return(mockRow)
//		mockTx.On("Rollback", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusBadRequest, w.Code)
//		assert.Contains(t, w.Body.String(), "Validation error")
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when transaction fails to begin", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		csvContent := "Catan"
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(nil, errors.New("transaction error"))
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "transaction error")
//		mockDB.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when DB error occurs during insert", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		title := "Catan"
//		csvContent := "title\n" + title
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)
//		mockTx.On("Rollback", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "db error")
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should return 500 Internal Server Error when transaction commit fails", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		title := "Catan"
//		csvContent := "title\n" + title
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		mockRow := new(MockRow)
//		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{Valid: false}}).Return(mockRow)
//
//		mockTx.On("Commit", mock.Anything).Return(errors.New("commit error"))
//		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusInternalServerError, w.Code)
//		assert.Contains(t, w.Body.String(), "commit error")
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should create a game with barcode and Play to Win when CSV columns are provided", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		gameID := uuid.New()
//		ptwID := uuid.New()
//		groupID := uuid.New()
//		title := "Heat"
//		barcode := "9780000000001"
//
//		csvContent := "title,barcode,isPlayToWin\n" + title + "," + barcode + ",true"
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//
//		// CreateGame via tx
//		mockGameRow := new(MockRow)
//		mockGameRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // display_title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}).Return(mockGameRow).Once()
//
//		// addPlayToWinByGameId: GetGame via tx
//		mockTxGetGameRow := new(MockRow)
//		mockTxGetGameRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(1).(*string) = title // display_title
//			*args.Get(2).(*string) = title // title
//			*args.Get(3).(*string) = SanitizeTitle(title)
//			*args.Get(4).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
//			*args.Get(5).(*pgtype.UUID) = pgtype.UUID{Valid: false}
//			*args.Get(6).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}}).Return(mockTxGetGameRow).Once()
//
//		// getOrCreatePlayToWinGroup (1st call): group does not exist yet, create it
//		sp1 := new(MockTx)
//		mockTx.On("Begin", mock.Anything).Return(sp1, nil).Once()
//		mockGroupNotFoundRow := new(MockRow)
//		mockGroupNotFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
//		sp1.On("QueryRow", mock.Anything, mock.Anything, []any{title}).Return(mockGroupNotFoundRow)
//		sp1.On("Rollback", mock.Anything).Return(nil)
//		mockCreateGroupRow := new(MockRow)
//		mockCreateGroupRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//		}).Return(nil)
//		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{title}).Return(mockCreateGroupRow).Once()
//
//		// getOrCreatePlayToWinGroup (2nd call - debug): group now exists
//		sp2 := new(MockTx)
//		mockTx.On("Begin", mock.Anything).Return(sp2, nil).Once()
//		mockGroupFoundRow := new(MockRow)
//		mockGroupFoundRow.On("Scan", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(1).(*string) = title
//			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//		}).Return(nil)
//		sp2.On("QueryRow", mock.Anything, mock.Anything, []any{title}).Return(mockGroupFoundRow)
//		sp2.On("Commit", mock.Anything).Return(nil)
//
//		// CreatePlayToWinGame savepoint
//		sp3 := new(MockTx)
//		mockTx.On("Begin", mock.Anything).Return(sp3, nil).Once()
//		mockPtwRow := new(MockRow)
//		mockPtwRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
//			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: ptwID, Valid: true}
//			*args.Get(1).(*pgtype.UUID) = pgtype.UUID{Bytes: gameID, Valid: true}
//			*args.Get(2).(*pgtype.UUID) = pgtype.UUID{Bytes: groupID, Valid: true}
//			*args.Get(3).(*pgtype.UUID) = pgtype.UUID{Valid: false} // winner_id
//			*args.Get(4).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
//			*args.Get(5).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false}
//			*args.Get(6).(*db.NullPlayToWinGameDeletionType) = db.NullPlayToWinGameDeletionType{Valid: false}
//			*args.Get(7).(*pgtype.Text) = pgtype.Text{Valid: false}
//		}).Return(nil)
//		sp3.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: gameID, Valid: true}, pgtype.UUID{Bytes: groupID, Valid: true}}).Return(mockPtwRow)
//		sp3.On("Commit", mock.Anything).Return(nil)
//
//		mockTx.On("Commit", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response BulkAddResponse
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, int32(1), response.Imported)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//
//	t.Run("Should handle empty CSV and return 201 with zero imported", func(t *testing.T) {
//		server, mockDB := setupTestServer()
//		csvContent := ""
//		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))
//
//		mockTx := new(MockTx)
//		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
//		mockTx.On("Commit", mock.Anything).Return(nil)
//
//		w := httptest.NewRecorder()
//		c, _ := gin.CreateTestContext(w)
//		c.Request = httptest.NewRequest("POST", "/games/bulk", bytes.NewBufferString(base64Content))
//		c.Request.Header.Set("Content-Type", "text/plain")
//
//		server.BulkAddGames(c)
//
//		assert.Equal(t, http.StatusCreated, w.Code)
//		var response BulkAddResponse
//		err := json.Unmarshal(w.Body.Bytes(), &response)
//		assert.NoError(t, err)
//		assert.Equal(t, int32(0), response.Imported)
//		mockDB.AssertExpectations(t)
//		mockTx.AssertExpectations(t)
//	})
//}
