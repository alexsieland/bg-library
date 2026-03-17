package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameService struct {
	libraryService *LibraryService
}

func NewGameService(libService *LibraryService) *GameService {
	return &GameService{libraryService: libService}
}

func (s *GameService) GetGameStatus(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwGameStatus, error) {
	var (
		gameStatus db.VwGameStatus
		err        error
	)

	if optTx == nil {
		gameStatus, err = s.libraryService.queries.GetGameStatus(ctx, gameId)
	} else {
		gameStatus, err = s.libraryService.queries.WithTx(optTx).GetGameStatus(ctx, gameId)
	}
	if err != nil {
		return db.VwGameStatus{}, wrapDatabaseError(err)
	}

	return gameStatus, nil
}
