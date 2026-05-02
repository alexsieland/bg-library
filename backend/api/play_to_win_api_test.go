package api

import (
	"context"
	"errors"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTx is a mock of the pgx.Tx interface
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := m.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := m.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	args := m.Called()
	return args.Get(0).(pgx.LargeObjects)
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := m.Called(ctx, name, sql)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pgx.Rows), callArgs.Error(1)
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil
	}
	return callArgs.Get(0).(pgx.Row)
}

func (m *MockTx) Conn() *pgx.Conn {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*pgx.Conn)
}

// This test file provides lightweight, fixture-based unit tests for the
// PlayToWin API handlers. It mocks the underlying playToWinService interface and the
// libraryService transaction handling to focus tests on input validation and
// correct wiring.

type mockPlayToWinService struct{ mock.Mock }

func (m *mockPlayToWinService) GetPlayToWinGameByLibraryGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGame), args.Error(1)
}

func (m *mockPlayToWinService) GetPlayToWinGroup(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	args := m.Called(ctx, ptwGroupId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGroup{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGroup), args.Error(1)
}

func (m *mockPlayToWinService) GetPlayToWinGroupByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	args := m.Called(ctx, ptwGameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGroup{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGroup), args.Error(1)
}

func (m *mockPlayToWinService) GetPlayToWinGameOverview(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGameOverview, error) {
	args := m.Called(ctx, ptwGameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGameOverview{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGameOverview), args.Error(1)
}

func (m *mockPlayToWinService) ListPlayToWinGameOverviews(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwPlayToWinGameOverview, error) {
	args := m.Called(ctx, gameTitle, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwPlayToWinGameOverview), args.Error(1)
}

func (m *mockPlayToWinService) ListDeletedPlayToWinGameOverviews(ctx context.Context, deletionReason db.NullPlayToWinGameDeletionType, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwDeletedPlayToWinGameOverview, error) {
	args := m.Called(ctx, deletionReason, gameTitle, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwDeletedPlayToWinGameOverview), args.Error(1)
}

func (m *mockPlayToWinService) GetPlayToWinGameEntriesByGroupId(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	args := m.Called(ctx, ptwGroupId, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwPlayToWinEntry), args.Error(1)
}

func (m *mockPlayToWinService) GetPlayToWinGameEntriesByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	args := m.Called(ctx, ptwGameId, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwPlayToWinEntry), args.Error(1)
}

func (m *mockPlayToWinService) ListPlayToWinEntriesByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	args := m.Called(ctx, ptwGameId, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwPlayToWinEntry), args.Error(1)
}

func (m *mockPlayToWinService) ListPlayToWinEntriesByGroupId(ctx context.Context, groupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	args := m.Called(ctx, groupId, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwPlayToWinEntry), args.Error(1)
}

func (m *mockPlayToWinService) InsertPlayToWinSession(ctx context.Context, ptwGroupId pgtype.UUID, playtimeMinutes *int32, optTx pgx.Tx) (db.PlayToWinSession, error) {
	args := m.Called(ctx, ptwGroupId, playtimeMinutes, optTx)
	if args.Get(0) == nil {
		return db.PlayToWinSession{}, args.Error(1)
	}
	return args.Get(0).(db.PlayToWinSession), args.Error(1)
}

func (m *mockPlayToWinService) InsertPlayToWinEntry(ctx context.Context, ptwSessionId pgtype.UUID, ptwGroupId pgtype.UUID, entrantName string, entrantUniqueID string, optTx pgx.Tx) (db.PlayToWinEntry, error) {
	args := m.Called(ctx, ptwSessionId, ptwGroupId, entrantName, entrantUniqueID, optTx)
	if args.Get(0) == nil {
		return db.PlayToWinEntry{}, args.Error(1)
	}
	return args.Get(0).(db.PlayToWinEntry), args.Error(1)
}

func (m *mockPlayToWinService) UpdatePlayToWinGameWinner(ctx context.Context, ptwGameId pgtype.UUID, entryId pgtype.UUID, optTx pgx.Tx) error {
	args := m.Called(ctx, ptwGameId, entryId, optTx)
	return args.Error(0)
}

func (m *mockPlayToWinService) InsertPlayToWinGroup(ctx context.Context, groupName string, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	args := m.Called(ctx, groupName, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGroup{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGroup), args.Error(1)
}

func (m *mockPlayToWinService) InsertPlayToWinGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGame), args.Error(1)
}

func (m *mockPlayToWinService) DeletePlayToWinGameByPlayToWinId(ctx context.Context, ptwGameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	args := m.Called(ctx, ptwGameId, deletionReason, deletionReasonComment, optTx)
	return args.Error(0)
}

func (m *mockPlayToWinService) DeletePlayToWinGameByLibraryGameId(ctx context.Context, gameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	args := m.Called(ctx, gameId, deletionReason, deletionReasonComment, optTx)
	return args.Error(0)
}

func (m *mockPlayToWinService) DeletePlayToWinEntry(ctx context.Context, entryId pgtype.UUID, deletionReason db.NullPlayToWinEntryDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	args := m.Called(ctx, entryId, deletionReason, deletionReasonComment, optTx)
	return args.Error(0)
}

func (m *mockPlayToWinService) ClaimPlayToWinGame(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) error {
	args := m.Called(ctx, ptwGameId, optTx)
	return args.Error(0)
}

func (m *mockPlayToWinService) ResetPlayToWinGameWinners(ctx context.Context, optTx pgx.Tx) error {
	args := m.Called(ctx, optTx)
	return args.Error(0)
}

// testLibService mirrors the small test lib used by patrons tests: it returns
// a preconfigured tx or error when BeginTx is called.

func newTestPlayToWinApi(service *mockPlayToWinService, tx pgx.Tx, beginErr error) *PlayToWinApi {
	lib := &testLibService{tx: tx, err: beginErr}
	return &PlayToWinApi{libraryService: lib, service: service}
}

// --- Tests ---------------------------------------------------------------

func TestAddPlayToWinGame(t *testing.T) {
	t.Run("Should return converted play to win game when service succeeds", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		mockTx := new(MockTx)
		fixture := newTestPlayToWinApi(svc, mockTx, nil)
		ctx := context.Background()
		gameID := uuid.New()
		ptwID := uuid.New()

		dbPtw := db.VwPlayToWinGame{
			ID:         pgtype.UUID{Bytes: ptwID, Valid: true},
			GameID:     pgtype.UUID{Bytes: gameID, Valid: true},
			PtwGroupID: pgtype.UUID{Valid: false},
			GroupName:  pgtype.Text{String: "Test Group", Valid: true},
			CreatedAt:  pgtype.Timestamp{Valid: true},
			WinnerID:   pgtype.UUID{Valid: false},
		}

		dbOverview := db.VwPlayToWinGameOverview{
			PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
			GameID:         pgtype.UUID{Bytes: gameID, Valid: true},
			PtwGroupID:     pgtype.UUID{Valid: false},
			GameTitle:      "Test Game",
			SanitizedTitle: "test game",
			CreatedAt:      pgtype.Timestamp{Valid: true},
			WinnerID:       pgtype.UUID{Valid: false},
		}

		svc.On("InsertPlayToWinGame", ctx, pgtype.UUID{Bytes: gameID, Valid: true}, mockTx).Return(dbPtw, nil).Once()
		svc.On("GetPlayToWinGameOverview", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, mockTx).Return(dbOverview, nil).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		got, err := fixture.AddPlayToWinGameByGameId(ctx, types.UUID(gameID))
		assert.NoError(t, err)
		assert.Equal(t, ptwID.String(), got.PlayToWinId.String())
		svc.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should propagate service error when insert fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		mockTx := new(MockTx)
		fixture := newTestPlayToWinApi(svc, mockTx, nil)
		ctx := context.Background()
		gameID := uuid.New()
		expected := errors.New("insert failed")

		svc.On("InsertPlayToWinGame", ctx, pgtype.UUID{Bytes: gameID, Valid: true}, mockTx).Return(db.VwPlayToWinGame{}, expected).Once()
		mockTx.On("Rollback", ctx).Return(nil).Once()

		got, err := fixture.AddPlayToWinGameByGameId(ctx, types.UUID(gameID))
		assert.Equal(t, PlayToWinGame{}, got)
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestRemovePlayToWinGame(t *testing.T) {
	t.Run("Should succeed when service deletes by ptw id successfully", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		gameID := uuid.New()

		dbRemovalReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
			Valid:                     true,
		}
		svc.On("DeletePlayToWinGameByLibraryGameId", ctx, pgtype.UUID{Bytes: gameID, Valid: true}, dbRemovalReason, (*string)(nil), (pgx.Tx)(nil)).Return(nil).Once()

		err := fixture.RemovePlayToWinGameByGameId(ctx, types.UUID(gameID), RemovePlayToWinGameRequest{RemovalReason: RemovePlayToWinGameRequestRemovalReasonMistake})
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when delete fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		gameID := uuid.New()
		expected := errors.New("delete failed")

		dbRemovalReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
			Valid:                     true,
		}
		svc.On("DeletePlayToWinGameByLibraryGameId", ctx, pgtype.UUID{Bytes: gameID, Valid: true}, dbRemovalReason, (*string)(nil), (pgx.Tx)(nil)).Return(expected).Once()

		err := fixture.RemovePlayToWinGameByGameId(ctx, types.UUID(gameID), RemovePlayToWinGameRequest{RemovalReason: RemovePlayToWinGameRequestRemovalReasonMistake})
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})
}

func TestGetPlayToWinGameEntries(t *testing.T) {
	t.Run("Should return converted entries when service succeeds", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		entryID := uuid.New()

		dbEntries := []db.VwPlayToWinEntry{
			{
				ID:              pgtype.UUID{Bytes: entryID, Valid: true},
				EntrantName:     "Test Entrant",
				EntrantUniqueID: "unique-id-1",
				CreatedAt:       pgtype.Timestamp{Valid: true},
			},
		}

		svc.On("GetPlayToWinGameEntriesByPlayToWinGameId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(dbEntries, nil).Once()

		got, err := fixture.GetPlayToWinGameEntries(ctx, types.UUID(ptwID))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(got.Entries))
		assert.Equal(t, entryID.String(), got.Entries[0].EntryId.String())
		assert.Equal(t, "Test Entrant", got.Entries[0].EntrantName)
		assert.Equal(t, "unique-id-1", got.Entries[0].EntrantUniqueId)
		svc.AssertExpectations(t)
	})

	t.Run("Should return empty list when no entries found", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()

		svc.On("GetPlayToWinGameEntriesByPlayToWinGameId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return([]db.VwPlayToWinEntry{}, nil).Once()

		got, err := fixture.GetPlayToWinGameEntries(ctx, types.UUID(ptwID))
		assert.NoError(t, err)
		assert.Equal(t, 0, len(got.Entries))
		svc.AssertExpectations(t)
	})
}

func TestGetPlayToWinGameOverview(t *testing.T) {
	t.Run("Should return converted overview when service succeeds", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()

		dbOverview := db.VwPlayToWinGameOverview{
			PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
			GameID:         pgtype.UUID{Valid: true},
			PtwGroupID:     pgtype.UUID{Valid: false},
			GameTitle:      "Test Game",
			SanitizedTitle: "test game",
			CreatedAt:      pgtype.Timestamp{Valid: true},
			WinnerID:       pgtype.UUID{Valid: false},
		}

		svc.On("GetPlayToWinGameOverview", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(dbOverview, nil).Once()

		got, err := fixture.GetPlayToWinGameOverview(ctx, types.UUID(ptwID))
		assert.NoError(t, err)
		assert.Equal(t, ptwID.String(), got.PlayToWinId.String())
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when lookup fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		expected := errors.New("lookup failed")

		svc.On("GetPlayToWinGameOverview", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(db.VwPlayToWinGameOverview{}, expected).Once()

		got, err := fixture.GetPlayToWinGameOverview(ctx, types.UUID(ptwID))
		assert.Equal(t, PlayToWinGame{}, got)
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})
}

func TestListPlayToWinGames(t *testing.T) {
	t.Run("Should return converted games when service succeeds", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()

		dbGames := []db.VwPlayToWinGameOverview{
			{
				PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
				GameID:         pgtype.UUID{Valid: true},
				PtwGroupID:     pgtype.UUID{Valid: false},
				GameTitle:      "Test Game",
				SanitizedTitle: "test game",
				CreatedAt:      pgtype.Timestamp{Valid: true},
				WinnerID:       pgtype.UUID{Valid: false},
			},
		}

		svc.On("ListPlayToWinGameOverviews", ctx, (*string)(nil), int32(10), int32(0), (pgx.Tx)(nil)).Return(dbGames, nil).Once()

		got, err := fixture.ListPlayToWinGames(ctx, ListPlayToWinGamesParams{Limit: ptrInt32(10), Offset: ptrInt32(0)})
		assert.NoError(t, err)
		assert.Equal(t, 1, len(got.Games))
		svc.AssertExpectations(t)
	})

	t.Run("Should return empty list when no games found", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()

		svc.On("ListPlayToWinGameOverviews", ctx, (*string)(nil), int32(10), int32(0), (pgx.Tx)(nil)).Return([]db.VwPlayToWinGameOverview{}, nil).Once()

		got, err := fixture.ListPlayToWinGames(ctx, ListPlayToWinGamesParams{Limit: ptrInt32(10), Offset: ptrInt32(0)})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(got.Games))
		svc.AssertExpectations(t)
	})

	t.Run("Should return deleted games when deletionReason is provided", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		gameID := uuid.New()

		deletionReason := ListPlayToWinGamesParamsDeletionReasonMistake
		dbDeletionReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
			Valid:                     true,
		}
		dbGames := []db.VwDeletedPlayToWinGameOverview{
			{
				PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
				GameID:         pgtype.UUID{Bytes: gameID, Valid: true},
				GameTitle:      "Deleted Game",
				SanitizedTitle: "deleted game",
				DeletionReason: dbDeletionReason,
				WinnerID:       pgtype.UUID{Valid: false},
			},
		}

		svc.On("ListDeletedPlayToWinGameOverviews", ctx, dbDeletionReason, (*string)(nil), int32(100), int32(0), (pgx.Tx)(nil)).Return(dbGames, nil).Once()

		got, err := fixture.ListPlayToWinGames(ctx, ListPlayToWinGamesParams{DeletionReason: &deletionReason})
		assert.NoError(t, err)
		assert.Len(t, got.Games, 1)
		assert.Equal(t, ptwID, got.Games[0].PlayToWinId)
		assert.Equal(t, gameID, got.Games[0].GameId)
		assert.Equal(t, "Deleted Game", got.Games[0].Title)
		assert.Nil(t, got.Games[0].Winner)
		svc.AssertExpectations(t)
	})

	t.Run("Should return deleted games with winner when deletionReason is claimed", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		gameID := uuid.New()
		winnerID := uuid.New()

		deletionReason := ListPlayToWinGamesParamsDeletionReasonClaimed
		dbDeletionReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeClaimed,
			Valid:                     true,
		}
		dbGames := []db.VwDeletedPlayToWinGameOverview{
			{
				PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
				GameID:         pgtype.UUID{Bytes: gameID, Valid: true},
				GameTitle:      "Claimed Game",
				SanitizedTitle: "claimed game",
				DeletionReason: dbDeletionReason,
				WinnerID:       pgtype.UUID{Bytes: winnerID, Valid: true},
				WinnerName:     pgtype.Text{String: "Alice", Valid: true},
				WinnerUniqueID: pgtype.Text{String: "alice-1", Valid: true},
			},
		}

		svc.On("ListDeletedPlayToWinGameOverviews", ctx, dbDeletionReason, (*string)(nil), int32(100), int32(0), (pgx.Tx)(nil)).Return(dbGames, nil).Once()

		got, err := fixture.ListPlayToWinGames(ctx, ListPlayToWinGamesParams{DeletionReason: &deletionReason})
		assert.NoError(t, err)
		assert.Len(t, got.Games, 1)
		assert.Equal(t, "Claimed Game", got.Games[0].Title)
		require.NotNil(t, got.Games[0].Winner)
		assert.Equal(t, winnerID, got.Games[0].Winner.EntryId)
		assert.Equal(t, "Alice", got.Games[0].Winner.EntrantName)
		assert.Equal(t, "alice-1", got.Games[0].Winner.EntrantUniqueId)
		svc.AssertExpectations(t)
	})

	t.Run("Should return empty list when no deleted games match deletion reason", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()

		deletionReason := ListPlayToWinGamesParamsDeletionReasonOther
		dbDeletionReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeOther,
			Valid:                     true,
		}

		svc.On("ListDeletedPlayToWinGameOverviews", ctx, dbDeletionReason, (*string)(nil), int32(100), int32(0), (pgx.Tx)(nil)).Return([]db.VwDeletedPlayToWinGameOverview{}, nil).Once()

		got, err := fixture.ListPlayToWinGames(ctx, ListPlayToWinGamesParams{DeletionReason: &deletionReason})
		assert.NoError(t, err)
		assert.Empty(t, got.Games)
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when ListDeletedPlayToWinGameOverviews fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()

		deletionReason := ListPlayToWinGamesParamsDeletionReasonMistake
		dbDeletionReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
			Valid:                     true,
		}

		svc.On("ListDeletedPlayToWinGameOverviews", ctx, dbDeletionReason, (*string)(nil), int32(100), int32(0), (pgx.Tx)(nil)).Return(nil, errors.New("db error")).Once()

		_, err := fixture.ListPlayToWinGames(ctx, ListPlayToWinGamesParams{DeletionReason: &deletionReason})
		assert.Error(t, err)
		svc.AssertExpectations(t)
	})
}

func TestRecordPlayToWinSession(t *testing.T) {
	t.Run("Should return converted session when service succeeds", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		mockTx := new(MockTx)
		fixture := newTestPlayToWinApi(svc, mockTx, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		groupID := uuid.New()
		sessionID := uuid.New()
		entryID := uuid.New()
		playtime := int32(30)

		dbGroup := db.VwPlayToWinGroup{
			ID:   pgtype.UUID{Bytes: groupID, Valid: true},
			Name: "Test Group",
		}
		dbSession := db.PlayToWinSession{
			ID:              pgtype.UUID{Bytes: sessionID, Valid: true},
			PtwGroupID:      pgtype.UUID{Bytes: groupID, Valid: true},
			PlaytimeMinutes: pgtype.Int4{Int32: playtime, Valid: true},
		}
		dbEntry := db.PlayToWinEntry{
			ID:              pgtype.UUID{Bytes: entryID, Valid: true},
			EntrantName:     "Test Entrant",
			EntrantUniqueID: "unique-id-1",
		}

		req := CreatePlayToWinSessionRequest{
			PlayToWinId:     types.UUID(ptwID),
			PlaytimeMinutes: &playtime,
			Entries: []struct {
				EntrantName     string `json:"entrantName"`
				EntrantUniqueId string `json:"entrantUniqueId"`
			}{
				{EntrantName: "Test Entrant", EntrantUniqueId: "unique-id-1"},
			},
		}

		svc.On("GetPlayToWinGroupByPlayToWinGameId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(dbGroup, nil).Once()
		svc.On("InsertPlayToWinSession", ctx, pgtype.UUID{Bytes: groupID, Valid: true}, req.PlaytimeMinutes, mockTx).Return(dbSession, nil).Once()
		svc.On("InsertPlayToWinEntry", ctx, pgtype.UUID{Bytes: sessionID, Valid: true}, pgtype.UUID{Bytes: groupID, Valid: true}, "Test Entrant", "unique-id-1", mockTx).Return(dbEntry, nil).Once()
		mockTx.On("Commit", ctx).Return(nil).Once()

		got, err := fixture.RecordPlayToWinSession(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, sessionID.String(), got.SessionId.String())
		assert.Equal(t, 1, len(got.PlayToWinEntries))
		assert.Equal(t, entryID.String(), got.PlayToWinEntries[0].EntryId.String())
		assert.Equal(t, "Test Entrant", got.PlayToWinEntries[0].EntrantName)
		svc.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should propagate service error when insert fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		mockTx := new(MockTx)
		fixture := newTestPlayToWinApi(svc, mockTx, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		groupID := uuid.New()
		expected := errors.New("session insert failed")

		dbGroup := db.VwPlayToWinGroup{
			ID:   pgtype.UUID{Bytes: groupID, Valid: true},
			Name: "Test Group",
		}

		req := CreatePlayToWinSessionRequest{
			PlayToWinId: types.UUID(ptwID),
			Entries: []struct {
				EntrantName     string `json:"entrantName"`
				EntrantUniqueId string `json:"entrantUniqueId"`
			}{
				{EntrantName: "Test Entrant", EntrantUniqueId: "unique-id-1"},
			},
		}

		svc.On("GetPlayToWinGroupByPlayToWinGameId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(dbGroup, nil).Once()
		svc.On("InsertPlayToWinSession", ctx, pgtype.UUID{Bytes: groupID, Valid: true}, (*int32)(nil), mockTx).Return(db.PlayToWinSession{}, expected).Once()
		mockTx.On("Rollback", ctx).Return(nil).Once()

		got, err := fixture.RecordPlayToWinSession(ctx, req)
		assert.Equal(t, PlayToWinSession{}, got)
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}

func TestUpdatePlayToWinGame(t *testing.T) {
	t.Run("Should succeed when service updates successfully", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		entryID := uuid.New()

		svc.On("UpdatePlayToWinGameWinner", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, pgtype.UUID{Bytes: entryID, Valid: true}, (pgx.Tx)(nil)).Return(nil).Once()

		winnerUUID := types.UUID(entryID)
		err := fixture.UpdatePlayToWinGame(ctx, types.UUID(ptwID), UpdatePlayToWinGame{WinnerId: &winnerUUID})
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when update fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		entryID := uuid.New()
		expected := errors.New("update failed")

		svc.On("UpdatePlayToWinGameWinner", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, pgtype.UUID{Bytes: entryID, Valid: true}, (pgx.Tx)(nil)).Return(expected).Once()

		winnerUUID := types.UUID(entryID)
		err := fixture.UpdatePlayToWinGame(ctx, types.UUID(ptwID), UpdatePlayToWinGame{WinnerId: &winnerUUID})
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})
}

func TestDeletePlayToWinGame(t *testing.T) {
	t.Run("Should succeed when service deletes successfully", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()

		dbRemovalReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
			Valid:                     true,
		}
		svc.On("DeletePlayToWinGameByPlayToWinId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, dbRemovalReason, (*string)(nil), (pgx.Tx)(nil)).Return(nil).Once()

		err := fixture.DeletePlayToWinGame(ctx, types.UUID(ptwID), RemovePlayToWinGameRequest{RemovalReason: RemovePlayToWinGameRequestRemovalReasonMistake})
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when delete fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		expected := errors.New("delete failed")

		dbRemovalReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
			Valid:                     true,
		}
		svc.On("DeletePlayToWinGameByPlayToWinId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, dbRemovalReason, (*string)(nil), (pgx.Tx)(nil)).Return(expected).Once()

		err := fixture.DeletePlayToWinGame(ctx, types.UUID(ptwID), RemovePlayToWinGameRequest{RemovalReason: RemovePlayToWinGameRequestRemovalReasonMistake})
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})
}

func TestClaimPlayToWinGame(t *testing.T) {
	t.Run("Should succeed when service claims successfully", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()

		svc.On("ClaimPlayToWinGame", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(nil).Once()

		err := fixture.ClaimPlayToWinGame(ctx, types.UUID(ptwID))
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when claim fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		expected := errors.New("claim failed")

		svc.On("ClaimPlayToWinGame", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(expected).Once()

		err := fixture.ClaimPlayToWinGame(ctx, types.UUID(ptwID))
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})
}

func TestDrawPlayToWinRaffle(t *testing.T) {
	t.Run("Should return converted entry when service succeeds", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		ptwID := uuid.New()
		entryID := uuid.New()

		dbEntries := []db.VwPlayToWinEntry{
			{
				ID:              pgtype.UUID{Bytes: entryID, Valid: true},
				EntrantName:     "Winner",
				EntrantUniqueID: "winner-id",
				CreatedAt:       pgtype.Timestamp{Valid: true},
			},
		}

		svc.On("GetPlayToWinGameEntriesByPlayToWinGameId", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, (pgx.Tx)(nil)).Return(dbEntries, nil).Once()
		svc.On("UpdatePlayToWinGameWinner", ctx, pgtype.UUID{Bytes: ptwID, Valid: true}, pgtype.UUID{Bytes: entryID, Valid: true}, (pgx.Tx)(nil)).Return(nil).Once()

		got, err := fixture.DrawPlayToWinRaffle(ctx, types.UUID(ptwID))
		assert.NoError(t, err)
		assert.Equal(t, entryID.String(), got.EntryId.String())
		assert.Equal(t, "Winner", got.EntrantName)
		assert.Equal(t, "winner-id", got.EntrantUniqueId)
		svc.AssertExpectations(t)
	})
}

func TestResetPlayToWinRaffle(t *testing.T) {
	t.Run("Should succeed when service resets successfully", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()

		svc.On("ResetPlayToWinGameWinners", ctx, (pgx.Tx)(nil)).Return(nil).Once()

		err := fixture.ResetPlayToWinRaffle(ctx)
		assert.NoError(t, err)
		svc.AssertExpectations(t)
	})

	t.Run("Should propagate service error when reset fails", func(t *testing.T) {
		svc := new(mockPlayToWinService)
		fixture := newTestPlayToWinApi(svc, nil, nil)
		ctx := context.Background()
		expected := errors.New("reset failed")

		svc.On("ResetPlayToWinGameWinners", ctx, (pgx.Tx)(nil)).Return(expected).Once()

		err := fixture.ResetPlayToWinRaffle(ctx)
		assert.ErrorIs(t, err, expected)
		svc.AssertExpectations(t)
	})
}

// --- Helpers ---------------------------------------------------------------

func ptrInt32(v int32) *int32 {
	return &v
}
