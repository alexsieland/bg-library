package internal

import (
	"context"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGameServiceListAndSearch(t *testing.T) {
	t.Run("Should list games when no title is provided", func(t *testing.T) {
		svc, ctx, mockDB := setupGameServiceWithDB(t)

		expected := []db.VwLibraryGame{
			makeLibraryGame(uuid.New(), "Catan", nil),
		}

		rows := new(MockRows)
		MockVwLibraryGameRows(rows, expected, nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(10), int32(0)}).Return(rows, nil).Once()

		got, err := svc.ListGames(ctx, nil, 10, 0, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)

		rows.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should search games when title provided using optTx", func(t *testing.T) {
		svc, _, mockTx, _ := setupGameServiceWithMockTx(t)

		search := "Catan"
		expected := []db.VwLibraryGame{makeLibraryGame(uuid.New(), "Catan", nil)}
		rows := new(MockRows)
		MockVwLibraryGameRows(rows, expected, nil)

		sanitized := GenerateDBRegexString(SanitizeTitle(search))
		// When using an optTx, the tx's Query will be called with (sanitized, limit, offset)
		mockTx.On("Query", mock.Anything, mock.Anything, []any{sanitized, int32(10), int32(0)}).Return(rows, nil).Once()

		got, err := svc.ListGames(context.Background(), &search, 10, 0, mockTx)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)

		rows.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestGameServiceGetGamesByBarcode(t *testing.T) {
	t.Run("Should return games when barcode provided and no tx", func(t *testing.T) {
		svc, mockDB := setupTestGameService()

		barcode := "B-123"
		expected := []db.VwLibraryGame{makeLibraryGame(uuid.New(), "Catan", &barcode)}
		rows := new(MockRows)
		MockVwLibraryGameRows(rows, expected, nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(rows, nil).Once()

		got, err := svc.GetGamesByBarcode(context.Background(), barcode, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)

		rows.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
}

func TestGameServiceGetGame(t *testing.T) {
	t.Run("Should return game when found", func(t *testing.T) {
		svc, mockDB := setupTestGameService()
		id := uuid.New()
		expected := makeLibraryGame(id, "Terraforming Mars", nil)

		row := new(MockRow)
		MockVwLibraryGameScan(row, expected, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{expected.ID}).Return(row).Once()

		got, err := svc.GetGame(context.Background(), expected.ID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)

		mockDB.AssertExpectations(t)
		row.AssertExpectations(t)
	})

	t.Run("Should return not found when no rows", func(t *testing.T) {
		svc, mockDB := setupTestGameService()
		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockRow := new(MockRow)
		// vw_library_games has 7 fields
		MockRowScanError(mockRow, 7, pgx.ErrNoRows)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{id}).Return(mockRow).Once()

		g, err := svc.GetGame(context.Background(), id, nil)
		assert.Equal(t, db.VwLibraryGame{}, g)
		assert.ErrorIs(t, err, ErrNotFound)

		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestGameServiceGetGameStatus(t *testing.T) {
	t.Run("Should return game status when found", func(t *testing.T) {
		svc, mockDB := setupTestGameService()

		gameId := uuid.New()
		patronId := uuid.New()
		txId := uuid.New()
		now := time.Now().UTC()

		status := db.VwGameStatus{
			GameID:            pgtype.UUID{Bytes: gameId, Valid: true},
			GameTitle:         "Some Game",
			SanitizedTitle:    "some game",
			PatronID:          pgtype.UUID{Bytes: patronId, Valid: true},
			PatronFullName:    pgtype.Text{String: "Patron", Valid: true},
			TransactionID:     pgtype.UUID{Bytes: txId, Valid: true},
			CheckoutTimestamp: pgtype.Timestamp{Time: now, Valid: true},
			CheckinTimestamp:  pgtype.Timestamp{Valid: false},
			PtwGameID:         pgtype.UUID{Valid: false},
		}

		mockRow := new(MockRow)
		MockVwGameStatusScan(mockRow, status, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{status.GameID}).Return(mockRow).Once()

		got, err := svc.GetGameStatus(context.Background(), status.GameID, nil)
		assert.NoError(t, err)
		assert.Equal(t, status, got)

		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestGameServiceInsertGame(t *testing.T) {
	t.Run("Should insert game (non-ptw) when no tx provided", func(t *testing.T) {
		svc, ctx, mockTx, _ := setupGameServiceWithMockTx(t)

		title := "New Game"
		barcode := "B-1000"
		created := db.Game{
			ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Title:          title,
			DisplayTitle:   pgtype.Text{String: title, Valid: true},
			SanitizedTitle: SanitizeTitle(title),
			CreatedAt:      pgtype.Timestamp{Valid: true},
			DeletedAt:      pgtype.Timestamp{Valid: false},
			Barcode:        pgtype.Text{String: barcode, Valid: true},
		}

		row := new(MockRow)
		MockGameScan(row, created, nil)

		expectedArgs := []any{title, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}
		mockTx.On("QueryRow", mock.Anything, mock.Anything, expectedArgs).Return(row).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		libGame, err := svc.InsertGame(ctx, title, &barcode, false, nil)
		assert.NoError(t, err)

		assert.Equal(t, created.ID, libGame.ID)
		assert.Equal(t, created.Title, libGame.Title)
		assert.Equal(t, created.SanitizedTitle, libGame.SanitizedTitle)
		assert.Equal(t, created.Barcode, libGame.Barcode)

		row.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

// ---- Setup helpers ----

func setupTestGameService() (GameService, *MockDatabase) {
	libService, mockDB := setupTestLibraryService()
	return GameService{libraryService: libService}, mockDB
}

func setupGameServiceWithDB(t *testing.T) (GameService, context.Context, *MockDatabase) {
	t.Helper()
	ctx := t.Context()
	svc, mockDB := setupTestGameService()
	return svc, ctx, mockDB
}

func setupGameServiceWithMockTx(t *testing.T) (GameService, context.Context, *MockTx, *MockDatabase) {
	t.Helper()
	ctx := t.Context()
	svc, mockDB := setupTestGameService()
	mockTx := MockWithinTx(t)
	return svc, ctx, mockTx, mockDB
}

// ...helpers moved to db_mock_test.go...

func TestGameServiceInsertGame_PTW(t *testing.T) {
	t.Run("Should insert game (ptw) and create play-to-win when no tx provided", func(t *testing.T) {
		svc, ctx, mockTx, _ := setupGameServiceWithMockTx(t)

		title := "New PTW Game"
		barcode := "B-PTW-1"
		created := db.Game{
			ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Title:          title,
			DisplayTitle:   pgtype.Text{String: title, Valid: true},
			SanitizedTitle: SanitizeTitle(title),
			CreatedAt:      pgtype.Timestamp{Valid: true},
			DeletedAt:      pgtype.Timestamp{Valid: false},
			Barcode:        pgtype.Text{String: barcode, Valid: true},
		}

		row := new(MockRow)
		MockGameScan(row, created, nil)

		expectedArgs := []any{title, SanitizeTitle(title), pgtype.Text{String: barcode, Valid: true}}
		mockTx.On("QueryRow", mock.Anything, mock.Anything, expectedArgs).Return(row).Once()

		// Prepare PTW mock service and expectation
		ptwGameID := uuid.New()
		ptwGame := makeVwPlayToWinGame(ptwGameID, uuid.UUID(created.ID.Bytes), title)

		mockPtw := new(MockPlayToWinService)
		svc.ptwService = mockPtw
		// Expect InsertPlayToWinGame to be called with the created game ID
		mockPtw.On("InsertPlayToWinGame", mock.Anything, created.ID, mock.Anything).Return(ptwGame, nil).Once()

		mockTx.On("Commit", ctx).Return(nil).Once()

		libGame, err := svc.InsertGame(ctx, title, &barcode, true, nil)
		assert.NoError(t, err)

		// play-to-win id should be set on the returned library game
		assert.Equal(t, ptwGame.ID, libGame.PlayToWinGameID)

		mockPtw.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestGameServiceSetIsPlayToWin(t *testing.T) {
	t.Run("Should panic when ptwService is not set", func(t *testing.T) {
		lib, _ := setupTestLibraryService()
		svc := GameService{libraryService: lib}

		assert.Panics(t, func() {
			_ = svc.SetIsPlayToWin(context.Background(), pgtype.UUID{Valid: false}, true, nil)
		})
	})

	t.Run("Should add play-to-win when flag true and no existing ptw", func(t *testing.T) {
		svc, ctx, mockTx, mockDB := setupGameServiceWithMockTx(t)

		gameId := uuid.New()
		// Game status: no ptw
		status := db.VwGameStatus{
			GameID:            pgtype.UUID{Bytes: gameId, Valid: true},
			GameTitle:         "Some Game",
			SanitizedTitle:    "some game",
			PatronID:          pgtype.UUID{Valid: false},
			PatronFullName:    pgtype.Text{Valid: false},
			TransactionID:     pgtype.UUID{Valid: false},
			CheckoutTimestamp: pgtype.Timestamp{Valid: false},
			CheckinTimestamp:  pgtype.Timestamp{Valid: true},
			PtwGameID:         pgtype.UUID{Valid: false},
		}

		mockRow := new(MockRow)
		MockVwGameStatusScan(mockRow, status, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{status.GameID}).Return(mockRow).Once()

		mockPtw := new(MockPlayToWinService)
		svc.ptwService = mockPtw

		// Expect InsertPlayToWinGame to be called
		mockPtw.On("InsertPlayToWinGame", mock.Anything, status.GameID, mock.Anything).Return(makeVwPlayToWinGame(uuid.New(), uuid.UUID(status.GameID.Bytes), status.GameTitle), nil).Once()

		mockTx.On("Commit", ctx).Return(nil).Once()

		err := svc.SetIsPlayToWin(ctx, status.GameID, true, nil)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
		mockPtw.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should remove play-to-win when flag false and ptw exists", func(t *testing.T) {
		svc, ctx, mockTx, mockDB := setupGameServiceWithMockTx(t)

		gameId := uuid.New()
		ptwId := uuid.New()
		// Game status: ptw exists
		status := db.VwGameStatus{
			GameID:            pgtype.UUID{Bytes: gameId, Valid: true},
			GameTitle:         "Some Game",
			SanitizedTitle:    "some game",
			PatronID:          pgtype.UUID{Valid: false},
			PatronFullName:    pgtype.Text{Valid: false},
			TransactionID:     pgtype.UUID{Valid: false},
			CheckoutTimestamp: pgtype.Timestamp{Valid: false},
			CheckinTimestamp:  pgtype.Timestamp{Valid: true},
			PtwGameID:         pgtype.UUID{Bytes: ptwId, Valid: true},
		}

		mockRow := new(MockRow)
		MockVwGameStatusScan(mockRow, status, nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{status.GameID}).Return(mockRow).Once()

		mockPtw := new(MockPlayToWinService)
		svc.ptwService = mockPtw

		// Expect DeletePlayToWinGameByLibraryGameId to be called (we accept any args for deletion reason/comment/tx)
		mockPtw.On("DeletePlayToWinGameByLibraryGameId", mock.Anything, status.GameID, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockTx.On("Commit", ctx).Return(nil).Once()

		err := svc.SetIsPlayToWin(ctx, status.GameID, false, nil)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
		mockPtw.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}
