package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameService struct {
	libraryService *LibraryService
	ptwService     PlayToWinServiceInterface
}

func NewGameService(libService *LibraryService) *GameService {
	return &GameService{libraryService: libService}
}

func (s *GameService) SetPlayToWinService(ptwService PlayToWinServiceInterface) {
	s.ptwService = ptwService
}

func (s *GameService) listGames(ctx context.Context, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	params := db.ListGamesParams{Limit: limit, Offset: offset}

	if optTx == nil {
		return s.libraryService.queries.ListGames(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).ListGames(ctx, params)
}

func (s *GameService) searchGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	if gameTitle == nil || *gameTitle == "" {
		return s.listGames(ctx, limit, offset, optTx)
	}

	sanitizedTitle := SanitizeTitle(*gameTitle)

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

func (s *GameService) ListGames(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
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

func (s *GameService) listGameStatuses(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error) {
	params := db.ListGamesStatusParams{Limit: limit, Offset: offset}

	if optTx == nil {
		return s.libraryService.queries.ListGamesStatus(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).ListGamesStatus(ctx, params)
}

func (s *GameService) searchGameStatuses(ctx context.Context, checkedOut *bool, gameTitle *string, gameBarcode *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error) {
	var sanitizedTitle *string
	if gameTitle != nil {
		st := GenerateDBRegexString(SanitizeTitle(*gameTitle))
		sanitizedTitle = &st
	}
	params := db.SearchGameStatusParams{
		CheckedOut:     boolToPgBool(checkedOut),
		SanitizedTitle: stringToPgText(sanitizedTitle),
		GameBarcode:    stringToPgText(gameBarcode),
		Limit:          limit,
		Offset:         offset,
	}
	if optTx == nil {
		return s.libraryService.queries.SearchGameStatus(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).SearchGameStatus(ctx, params)
}

func (s *GameService) ListGameStatuses(ctx context.Context, checkedOut *bool, gameTitle *string, gameBarcode *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwGameStatus, error) {
	gameStatuses, err := s.searchGameStatuses(ctx, checkedOut, gameTitle, gameBarcode, limit, offset, optTx)
	return wrapErrorOrReturn(&gameStatuses, []db.VwGameStatus{}, err)
}

func (s *GameService) GetGamesByBarcode(ctx context.Context, barcode string, optTx pgx.Tx) ([]db.VwLibraryGame, error) {
	var (
		games []db.VwLibraryGame
		err   error
	)

	if optTx == nil {
		games, err = s.libraryService.queries.GetGameByBarcode(ctx, stringToPgText(&barcode))
	} else {
		games, err = s.libraryService.queries.WithTx(optTx).GetGameByBarcode(ctx, stringToPgText(&barcode))
	}

	return wrapErrorOrReturn(&games, []db.VwLibraryGame{}, err)
}

func (s *GameService) GetGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryGame, error) {
	var (
		game db.VwLibraryGame
		err  error
	)

	if optTx == nil {
		game, err = s.libraryService.queries.GetGame(ctx, gameId)
	} else {
		game, err = s.libraryService.queries.WithTx(optTx).GetGame(ctx, gameId)
	}

	return wrapErrorOrReturn(&game, db.VwLibraryGame{}, err)
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
	return wrapErrorOrReturn(&gameStatus, db.VwGameStatus{}, nil)
}

func (s *GameService) InsertGame(ctx context.Context, title string, barcode *string, isPlayToWin bool, optTx pgx.Tx) (db.VwLibraryGame, error) {
	// ptwService is only required if creating a PlayToWin game

	createGameParams := db.CreateGameParams{
		Title:          title,
		SanitizedTitle: SanitizeTitle(title),
		Barcode:        stringToPgText(barcode),
	}

	game, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwLibraryGame, error) {
		newGame, err := s.libraryService.queries.WithTx(tx).CreateGame(ctx, createGameParams)
		if err != nil {
			return nil, err
		}

		var ptwGame db.VwPlayToWinGame
		if isPlayToWin {
			if s.ptwService == nil {
				return nil, ErrInvalidState
			}
			ptwGame, err = s.ptwService.InsertPlayToWinGame(ctx, newGame.ID, tx)
			if err != nil {
				return nil, err
			}
		}

		libraryGame := db.VwLibraryGame{
			ID:              newGame.ID,
			DisplayTitle:    newGame.Title,
			Title:           newGame.Title,
			SanitizedTitle:  newGame.SanitizedTitle,
			Barcode:         newGame.Barcode,
			PlayToWinGameID: ptwGame.ID,
			CreatedAt:       newGame.CreatedAt,
		}

		return &libraryGame, nil
	})

	return wrapErrorOrReturn(game, db.VwLibraryGame{}, err)
}

func (s *GameService) UpdateGame(ctx context.Context, gameId pgtype.UUID, title string, barcode *string, optTx pgx.Tx) error {
	params := db.EditGameParams{
		ID:             gameId,
		DisplayTitle:   stringToPgText(&title),
		SanitizedTitle: SanitizeTitle(title),
		Barcode:        stringToPgText(barcode),
	}

	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).EditGame(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s *GameService) SetIsPlayToWin(ctx context.Context, gameId pgtype.UUID, isPlayToWin bool, optTx pgx.Tx) error {
	if s.ptwService == nil {
		panic("ptwService must be set before calling SetIsPlayToWin")
	}

	gameStatus, err := s.GetGameStatus(ctx, gameId, optTx)
	if err != nil {
		return err
	}

	_, err = WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		if isPlayToWin && !gameStatus.PtwGameID.Valid {
			_, err = s.ptwService.InsertPlayToWinGame(ctx, gameId, tx)
		} else if !isPlayToWin && gameStatus.PtwGameID.Valid {
			reason := db.NullPlayToWinGameDeletionType{
				PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeMistake,
				Valid:                     true,
			}
			err = s.ptwService.DeletePlayToWinGameByLibraryGameId(ctx, gameId, reason, nil, tx)
		}
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s *GameService) DeleteGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) error {
	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeleteGame(ctx, gameId)
		return nil, err
	})

	return wrapDatabaseError(err)
}
