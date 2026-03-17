package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameService struct {
	libraryService *LibraryService
	ptwService     *PlayToWinService
}

func NewGameService(libService *LibraryService) *GameService {
	return &GameService{libraryService: libService}
}

func (s *GameService) SetPlayToWinService(ptwService *PlayToWinService) {
	s.ptwService = ptwService
}

func (s GameService) listGames(ctx context.Context, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	params := db.ListGamesParams{Limit: limit, Offset: offset}

	if optTx == nil {
		return s.libraryService.queries.ListGames(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).ListGames(ctx, params)
}

func (s GameService) searchGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	if gameTitle == nil || *gameTitle == "" {
		return s.listGames(ctx, limit, offset, optTx)
	}

	sanitizedTitle := ""
	if gameTitle != nil && *gameTitle != "" {
		sanitizedTitle = SanitizeTitle(*gameTitle)
	}

	params := db.SearchGamesParams{
		SanitizedTitle: GenerateDBRegexString(sanitizedTitle),
		Limit:          limit,
		Offset:         offset,
	}

	if optTx == nil {
		return s.libraryService.queries.SearchGames(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).SearchGames(ctx, params)
}

func (s GameService) ListGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	var (
		games []db.VwLibraryGame
		err   error
	)

	if gameTitle == nil || *gameTitle == "" {
		games, err = s.listGames(ctx, limit, offset, optTx)
	} else {
		games, err = s.searchGames(ctx, gameTitle, limit, offset, optTx)
	}

	if err != nil {
		return nil, wrapDatabaseError(err)
	}
	return games, nil
}

func (s GameService) GetGamesByBarcode(ctx context.Context, barcode string, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	var (
		games []db.VwLibraryGame
		err   error
	)

	if optTx == nil {
		games, err = s.libraryService.queries.GetGameByBarcode(ctx, stringToPgText(&barcode))
	} else {
		games, err = s.libraryService.queries.WithTx(optTx).GetGameByBarcode(ctx, stringToPgText(&barcode))
	}

	if err != nil {
		return nil, wrapDatabaseError(err)
	}
	return games, nil
}

func (s GameService) GetGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryGame, error) {
	var (
		game db.VwLibraryGame
		err  error
	)

	if optTx == nil {
		game, err = s.libraryService.queries.GetGame(ctx, gameId)
	} else {
		game, err = s.libraryService.queries.WithTx(optTx).GetGame(ctx, gameId)
	}

	if err != nil {
		return db.VwLibraryGame{}, wrapDatabaseError(err)
	}
	return game, nil
}

func (s GameService) GetGameStatus(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwGameStatus, error) {
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
	return wrapErrorOrReturn(gameStatus, db.VwGameStatus{}, nil)
}

func (s GameService) InsertGame(ctx context.Context, title string, barcode *string, isPlayToWin bool, optTx pgx.Tx) (db.Game, error) {
	createGameParams := db.CreateGameParams{
		Title:          title,
		SanitizedTitle: SanitizeTitle(title),
		Barcode:        stringToPgText(barcode),
	}

	game, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.Game, error) {
		game, err := s.libraryService.queries.WithTx(tx).CreateGame(ctx, createGameParams)
		if err != nil {
			return nil, err
		}

		if isPlayToWin {

		}

		return game, nil
	})
	if err != nil {
		return db.Game{}, wrapDatabaseError(err)
	}
	return *game, nil
}
