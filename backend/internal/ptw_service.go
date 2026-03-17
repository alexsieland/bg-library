package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PlayToWinService struct {
	libraryService *LibraryService
	gameService    *GameService
}

func NewPlayToWinService(libService *LibraryService) *PlayToWinService {
	return &PlayToWinService{libraryService: libService}
}

func (s *PlayToWinService) SetPlayToWinService(gameService *GameService) {
	s.gameService = gameService
}

func (s *PlayToWinService) GetPlayToWinGameByLibraryGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	var (
		ptwGame db.VwPlayToWinGame
		err     error
	)

	if optTx == nil {
		ptwGame, err = s.libraryService.queries.GetPlayToWinGameByLibraryGameId(ctx, gameId)
	} else {
		ptwGame, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinGameByLibraryGameId(ctx, gameId)
	}
	return wrapErrorOrReturn(&ptwGame, db.VwPlayToWinGame{}, err)
}

func (s *PlayToWinService) GetPlayToWinGroup(ctx context.Context, groupId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	var (
		ptwGroup db.VwPlayToWinGroup
		err      error
	)

	if optTx == nil {
		ptwGroup, err = s.libraryService.queries.GetPlayToWinGroup(ctx, groupId)
	} else {
		ptwGroup, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinGroup(ctx, groupId)
	}
	return wrapErrorOrReturn(&ptwGroup, db.VwPlayToWinGroup{}, err)
}

func (s *PlayToWinService) GetPlayToWinGameOverview(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGameOverview, error) {
	var (
		ptwGameOverview db.VwPlayToWinGameOverview
		err             error
	)

	if optTx == nil {
		ptwGameOverview, err = s.libraryService.queries.GetPlayToWinGameOverview(ctx, ptwGameId)
	} else {
		ptwGameOverview, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinGameOverview(ctx, ptwGameId)
	}
	return wrapErrorOrReturn(&ptwGameOverview, db.VwPlayToWinGameOverview{}, err)
}

func (s *PlayToWinService) listPlayToWinGameOverviews(ctx context.Context, limit int32, offset int32, optTx pgx.Tx) ([]db.VwPlayToWinGameOverview, error) {
	params := db.ListPlayToWinGameOverviewsParams{Limit: limit, Offset: offset}

	if optTx == nil {
		return s.libraryService.queries.ListPlayToWinGameOverviews(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).ListPlayToWinGameOverviews(ctx, params)
}

func (s *PlayToWinService) searchPlayToWinGameOverview(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwPlayToWinGameOverview, error) {
	if gameTitle == nil || *gameTitle == "" {
		return s.listPlayToWinGameOverviews(ctx, limit, offset, optTx)
	}

	sanitizedTitle := ""
	if gameTitle != nil && *gameTitle != "" {
		sanitizedTitle = SanitizeTitle(*gameTitle)
	}

	params := db.SearchPlayToWinGameOverviewsParams{
		SanitizedTitle: GenerateDBRegexString(sanitizedTitle),
		Limit:          limit,
		Offset:         offset,
	}

	if optTx == nil {
		return s.libraryService.queries.SearchPlayToWinGameOverviews(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).SearchPlayToWinGameOverviews(ctx, params)
}

func (s *PlayToWinService) ListPlayToWinGameOverviews(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwPlayToWinGameOverview, error) {
	ptwGameOverviews, err := s.searchPlayToWinGameOverview(ctx, gameTitle, limit, offset, optTx)
	return wrapErrorOrReturn(&ptwGameOverviews, []db.VwPlayToWinGameOverview{}, err)
}

// InsertPlayToWinGroup inserts a new play to win group into the database.
// This call is idempotent. If the play to win group already exists, the existing play to win group will be returned.
func (s *PlayToWinService) InsertPlayToWinGroup(ctx context.Context, groupName string, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	ptwGroup, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwPlayToWinGroup, error) {
		// Create checkpoint transaction
		cpTx, err := tx.Begin(ctx)
		if err != nil {
			return nil, err
		}

		// Attempt to create the play to win group
		newPtwGroup, err := s.libraryService.queries.WithTx(cpTx).CreatePlayToWinGroup(ctx, groupName)
		if err != nil {
			// Checkpoint transaction failed, so rollback
			_ = cpTx.Rollback(ctx)

			// Check if the error is due to a unique constraint violation, and if so, return the existing play to win group
			if isUniqueConstraintViolation(err) {
				ptwGroup, err := s.libraryService.queries.WithTx(tx).GetPlayToWinGroupByName(ctx, groupName)
				if err != nil {
					return nil, err
				}
				return &ptwGroup, err
			}

			return nil, err
		}

		// Commit checkpoint transaction
		_ = cpTx.Commit(ctx)

		// Convert the play to win group to the view model and return it
		ptwGroup := db.VwPlayToWinGroup{
			ID:        newPtwGroup.ID,
			Name:      newPtwGroup.Name,
			CreatedAt: newPtwGroup.CreatedAt,
		}
		return &ptwGroup, err
	})

	return wrapErrorOrReturn(ptwGroup, db.VwPlayToWinGroup{}, err)
}

// InsertPlayToWin inserts a new play to win into the database.
// This call is idempotent. If the play to win already exists, it will be ignored.
// If the play to win was deleted, it will be restored.
func (s *PlayToWinService) InsertPlayToWin(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	game, err := s.gameService.GetGame(ctx, gameId, optTx)
	if err != nil {
		return db.VwPlayToWinGame{}, wrapDatabaseError(err)
	}

	ptwGame, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwPlayToWinGame, error) {
		ptwGroup, err := s.InsertPlayToWinGroup(ctx, game.Title, tx)
		if err != nil {
			return nil, err
		}
		params := db.CreatePlayToWinGameParams{
			GameID:     game.ID,
			PtwGroupID: ptwGroup.ID,
		}

		// Create checkpoint transaction
		cpTx, err := tx.Begin(ctx)
		if err != nil {
			return nil, err
		}

		// Attempt to create the play to win game
		var ptwGame db.VwPlayToWinGame
		newPtwGame, err := s.libraryService.queries.WithTx(cpTx).CreatePlayToWinGame(ctx, params)
		if err != nil {
			// Rollback checkpoint transaction if there's an error
			_ = cpTx.Rollback(ctx)

			// Check if the error is due to a unique constraint violation
			if isUniqueConstraintViolation(err) {
				err := s.libraryService.queries.WithTx(tx).RestorePlayToWinGameByLibraryGameId(ctx, gameId)
				if err != nil {
					return nil, err
				}
				ptwGame, err = s.GetPlayToWinGameByLibraryGame(ctx, gameId, tx)
				return &ptwGame, err
			}
			return nil, err
		}

		_ = cpTx.Commit(ctx)
		ptwGame = db.VwPlayToWinGame{
			ID:         newPtwGame.ID,
			GameID:     newPtwGame.GameID,
			PtwGroupID: newPtwGame.PtwGroupID,
			CreatedAt:  newPtwGame.CreatedAt,
		}
		return &ptwGame, nil
	})

	return wrapErrorOrReturn(ptwGame, db.VwPlayToWinGame{}, err)
}

// DeletePlayToWin deletes a play to win from the database.
// This call is idempotent. If the play to win does not exist, it will be ignored.
func (s *PlayToWinService) DeletePlayToWin(ctx context.Context, ptwGameId pgtype.UUID, deletionReason *string, deletionReasonComment *string, optTx pgx.Tx) error {
	dbDeletionReason, err := playToWinGameDeletionReason(deletionReason)
	if err != nil {
		return err
	}
	params := db.DeletePlayToWinGameByPlayToWinIdParams{
		ID:                    ptwGameId,
		DeletionReason:        dbDeletionReason,
		DeletionReasonComment: stringToPgText(deletionReasonComment),
	}

	_, err = WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeletePlayToWinGameByPlayToWinId(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}
