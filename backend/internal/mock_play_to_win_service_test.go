package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

// MockPlayToWinService is a mock for the PlayToWinService interface used in
// unit tests where GameService or other services depend on PTW operations.
// Tests can set expectations on methods like InsertPlayToWinGame and
// DeletePlayToWinGameByLibraryGameId.
type MockPlayToWinService struct {
	mock.Mock
}

func (m *MockPlayToWinService) InsertPlayToWinGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGame), args.Error(1)
}

func (m *MockPlayToWinService) DeletePlayToWinGameByLibraryGameId(ctx context.Context, gameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	args := m.Called(ctx, gameId, deletionReason, deletionReasonComment, optTx)
	return args.Error(0)
}

func (m *MockPlayToWinService) GetPlayToWinGameByLibraryGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	args := m.Called(ctx, gameId, optTx)
	if args.Get(0) == nil {
		return db.VwPlayToWinGame{}, args.Error(1)
	}
	return args.Get(0).(db.VwPlayToWinGame), args.Error(1)
}
