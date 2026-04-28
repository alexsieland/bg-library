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
func (m *mockGameService) ListGameStatuses(ctx context.Context, checkedOut *bool, gameTitle *string, gameBarcode *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error) {
	args := m.Called(ctx, checkedOut, gameTitle, gameBarcode, limit, offset, optTx)
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

		svc.On("ListGameStatuses", ctx, (*bool)(nil), (*string)(nil), (*string)(nil), int32(100), int32(0), (pgx.Tx)(nil)).Return([]db.VwGameStatus{}, nil).Once()

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

		svc.On("ListGameStatuses", ctx, (*bool)(nil), &title, (*string)(nil), int32(100), int32(0), (pgx.Tx)(nil)).Return(statuses, nil).Once()
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
