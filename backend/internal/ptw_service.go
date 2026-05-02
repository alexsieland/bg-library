package internal

import (
	"context"
	"errors"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PlayToWinService struct {
	libraryService *LibraryService
	gameService    *GameService
}

// PlayToWinServiceInterface defines the subset of methods from PlayToWinService
// that other services depend on. Using an interface allows tests to inject a
// mock implementation.
type PlayToWinServiceInterface interface {
	InsertPlayToWinGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error)
	DeletePlayToWinGameByLibraryGameId(ctx context.Context, gameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error
	GetPlayToWinGameByLibraryGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error)
}

func NewPlayToWinService(libService *LibraryService) *PlayToWinService {
	return &PlayToWinService{libraryService: libService}
}

func (s *PlayToWinService) SetGameService(gameService *GameService) {
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

func (s *PlayToWinService) GetPlayToWinGroup(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	var (
		ptwGroup db.VwPlayToWinGroup
		err      error
	)

	if optTx == nil {
		ptwGroup, err = s.libraryService.queries.GetPlayToWinGroup(ctx, ptwGroupId)
	} else {
		ptwGroup, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinGroup(ctx, ptwGroupId)
	}
	return wrapErrorOrReturn(&ptwGroup, db.VwPlayToWinGroup{}, err)
}

func (s *PlayToWinService) GetPlayToWinGroupByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	var (
		ptwGroup db.VwPlayToWinGroup
		err      error
	)

	if optTx == nil {
		ptwGroup, err = s.libraryService.queries.GetPlayToWinGroupByPlayToWinGameId(ctx, ptwGameId)
	} else {
		ptwGroup, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinGroupByPlayToWinGameId(ctx, ptwGameId)
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

	sanitizedTitle := SanitizeTitle(*gameTitle)

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

func (s *PlayToWinService) ListDeletedPlayToWinGameOverviews(ctx context.Context, deletionReason db.NullPlayToWinGameDeletionType, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwDeletedPlayToWinGameOverview, error) {
	var sanitizedTitle *string = nil
	if gameTitle != nil && *gameTitle != "" {
		st := GenerateDBRegexString(SanitizeTitle(*gameTitle))
		sanitizedTitle = &st
	}

	params := db.ListDeletedPlayToWinGameOverviewsParams{
		DeletionReason: deletionReason,
		SanitizedTitle: stringToPgText(sanitizedTitle),
		Limit:          limit,
		Offset:         offset,
	}

	if optTx == nil {
		return s.libraryService.queries.ListDeletedPlayToWinGameOverviews(ctx, params)
	}
	return s.libraryService.queries.WithTx(optTx).ListDeletedPlayToWinGameOverviews(ctx, params)
}

func (s *PlayToWinService) GetPlayToWinGameEntriesByGroupId(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	var (
		entries []db.VwPlayToWinEntry
		err     error
	)

	if optTx == nil {
		entries, err = s.libraryService.queries.GetPlayToWinEntriesByGroupId(ctx, ptwGroupId)
	} else {
		entries, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinEntriesByGroupId(ctx, ptwGroupId)
	}

	return wrapErrorOrReturn(&entries, []db.VwPlayToWinEntry{}, err)
}

func (s *PlayToWinService) GetPlayToWinGameEntriesByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	var (
		entries []db.VwPlayToWinEntry
		err     error
	)

	if optTx == nil {
		entries, err = s.libraryService.queries.GetPlayToWinEntriesByPlayToWinGameId(ctx, ptwGameId)
	} else {
		entries, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinEntriesByPlayToWinGameId(ctx, ptwGameId)
	}

	return wrapErrorOrReturn(&entries, []db.VwPlayToWinEntry{}, err)
}

func (s *PlayToWinService) ListPlayToWinSessionsByGroupId(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinSession, error) {
	var (
		ptwSession []db.VwPlayToWinSession
		err        error
	)

	if optTx == nil {
		ptwSession, err = s.libraryService.queries.GetPlayToWinSessionsByGroupId(ctx, ptwGroupId)
	} else {
		ptwSession, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinSessionsByGroupId(ctx, ptwGroupId)
	}

	return wrapErrorOrReturn(&ptwSession, []db.VwPlayToWinSession{}, err)
}

func (s *PlayToWinService) ListPlayToWinEntriesByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	var (
		ptwEntries []db.VwPlayToWinEntry
		err        error
	)

	if optTx == nil {
		ptwEntries, err = s.libraryService.queries.GetPlayToWinEntriesByPlayToWinGameId(ctx, ptwGameId)
	} else {
		ptwEntries, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinEntriesByPlayToWinGameId(ctx, ptwGameId)
	}

	return wrapErrorOrReturn(&ptwEntries, []db.VwPlayToWinEntry{}, err)
}

func (s *PlayToWinService) ListPlayToWinEntriesByGroupId(ctx context.Context, groupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error) {
	var (
		ptwEntries []db.VwPlayToWinEntry
		err        error
	)

	if optTx == nil {
		ptwEntries, err = s.libraryService.queries.GetPlayToWinEntriesByGroupId(ctx, groupId)
	} else {
		ptwEntries, err = s.libraryService.queries.WithTx(optTx).GetPlayToWinEntriesByGroupId(ctx, groupId)
	}

	return wrapErrorOrReturn(&ptwEntries, []db.VwPlayToWinEntry{}, err)
}

func (s *PlayToWinService) InsertPlayToWinSession(ctx context.Context, ptwGroupId pgtype.UUID, playtimeMinutes *int32, optTx pgx.Tx) (db.PlayToWinSession, error) {
	params := db.CreatePlayToWinSessionParams{
		PtwGroupID:      ptwGroupId,
		PlaytimeMinutes: int32ToPgInt4(playtimeMinutes),
	}

	ptwSession, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.PlayToWinSession, error) {
		ptwSession, err := s.libraryService.queries.WithTx(tx).CreatePlayToWinSession(ctx, params)
		return &ptwSession, err
	})

	return wrapErrorOrReturn(ptwSession, db.PlayToWinSession{}, err)
}

func (s *PlayToWinService) InsertPlayToWinEntry(ctx context.Context, ptwSessionId pgtype.UUID, ptwGroupId pgtype.UUID, entrantName string, entrantUniqueID string, optTx pgx.Tx) (db.PlayToWinEntry, error) {
	params := db.CreatePlayToWinEntryParams{
		PtwSessionID:    ptwSessionId,
		PtwGroupID:      ptwGroupId,
		EntrantName:     entrantName,
		EntrantUniqueID: entrantUniqueID,
	}
	ptwEntry, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.PlayToWinEntry, error) {
		ptwEntry, err := s.libraryService.queries.WithTx(tx).CreatePlayToWinEntry(ctx, params)
		return &ptwEntry, err
	})

	return wrapErrorOrReturn(ptwEntry, db.PlayToWinEntry{}, err)
}

func (s *PlayToWinService) UpdatePlayToWinGameWinner(ctx context.Context, ptwGameId pgtype.UUID, entryId pgtype.UUID, optTx pgx.Tx) error {
	params := db.UpdatePlayToWinWinnerParams{
		ID:       ptwGameId,
		WinnerID: entryId,
	}

	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).UpdatePlayToWinWinner(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
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

// InsertPlayToWinGame inserts a new play to win into the database.
// This call is idempotent. If the play to win already exists, it will be ignored.
// If the play to win was deleted, it will be restored.
func (s *PlayToWinService) InsertPlayToWinGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error) {
	if s.gameService == nil {
		panic("gameService must be set before calling InsertPlayToWinGame")
	}

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

// DeletePlayToWinGameByPlayToWinId deletes a play to win from the database.
// This call is idempotent. If the play to win does not exist, it will be ignored.
func (s *PlayToWinService) DeletePlayToWinGameByPlayToWinId(ctx context.Context, ptwGameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	params := db.DeletePlayToWinGameByPlayToWinIdParams{
		ID:                    ptwGameId,
		DeletionReason:        deletionReason,
		DeletionReasonComment: stringToPgText(deletionReasonComment),
	}

	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeletePlayToWinGameByPlayToWinId(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}

// DeletePlayToWinGameByLibraryGameId deletes a play to win from the database.
// This call is idempotent. If the play to win does not exist, it will be ignored.
func (s *PlayToWinService) DeletePlayToWinGameByLibraryGameId(ctx context.Context, gameId pgtype.UUID, deletionReason db.NullPlayToWinGameDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	params := db.DeletePlayToWinGameParams{
		GameID:                gameId,
		DeletionReason:        deletionReason,
		DeletionReasonComment: stringToPgText(deletionReasonComment),
	}

	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeletePlayToWinGame(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s *PlayToWinService) DeletePlayToWinSession(ctx context.Context, sessionId pgtype.UUID, deletionReason db.NullPlayToWinSessionDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	params := db.DeletePlayToWinSessionParams{
		ID:                    sessionId,
		DeletionReason:        deletionReason,
		DeletionReasonComment: stringToPgText(deletionReasonComment),
	}

	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeletePlayToWinSession(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s *PlayToWinService) DeletePlayToWinEntry(ctx context.Context, entryId pgtype.UUID, deletionReason db.NullPlayToWinEntryDeletionType, deletionReasonComment *string, optTx pgx.Tx) error {
	params := db.DeletePlayToWinEntryParams{
		ID:                    entryId,
		DeletionReason:        deletionReason,
		DeletionReasonComment: stringToPgText(deletionReasonComment),
	}

	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeletePlayToWinEntry(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}

// ResetPlayToWinGameWinners resets the winners of all won, but unclaimed play to win games.
func (s *PlayToWinService) ResetPlayToWinGameWinners(ctx context.Context, optTx pgx.Tx) error {
	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).ResetPlayToWinGameWinners(ctx)
		return nil, err
	})
	return wrapDatabaseError(err)
}

// ClaimPlayToWinGame claims a play to win game.
// First deletes the play to win game with deletion reason "claimed"
// Then deletes the entry with deletion reason "won"
// Finally, deletes the library game
// If any of the above steps fail, the transaction will be rolled back.
func (s *PlayToWinService) ClaimPlayToWinGame(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) error {
	if s.gameService == nil {
		panic("gameService must be set before calling ClaimPlayToWinGame")
	}

	ptwGameOverview, err := s.GetPlayToWinGameOverview(ctx, ptwGameId, optTx)
	if err != nil {
		if errors.Is(ErrNotFound, err) {
			return nil
		}
		return wrapDatabaseError(err)
	}

	// A game can't be claimed if it doesn't have a winner
	if !ptwGameOverview.WinnerID.Valid {
		return ErrClaimUnwonPtwGame
	}

	_, err = WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		gameDeletionReason := db.NullPlayToWinGameDeletionType{
			PlayToWinGameDeletionType: db.PlayToWinGameDeletionTypeClaimed,
			Valid:                     true,
		}
		err := s.DeletePlayToWinGameByPlayToWinId(ctx, ptwGameId, gameDeletionReason, nil, tx)
		if err != nil {
			return nil, err
		}

		entryDeletionReason := db.NullPlayToWinEntryDeletionType{
			PlayToWinEntryDeletionType: db.PlayToWinEntryDeletionTypeWon,
			Valid:                      true,
		}
		err = s.DeletePlayToWinEntry(ctx, ptwGameOverview.WinnerID, entryDeletionReason, nil, tx)
		if err != nil {
			return nil, err
		}

		err = s.gameService.DeleteGame(ctx, ptwGameOverview.GameID, tx)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return wrapDatabaseError(err)
}

// RestoreClaimedPlayToWinGame restores a previously claimed play to win game.
// First restores the play to win game
// Then restores the library game
// Finally, restores the winner entry if present
// If any of the above steps fail, the transaction will be rolled back.
func (s *PlayToWinService) RestoreClaimedPlayToWinGame(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) error {
	if s.gameService == nil {
		panic("gameService must be set before calling RestoreClaimedPlayToWinGame")
	}

	// Look up the deleted game to get gameId and winnerId
	var deletedGame db.VwDeletedPlayToWinGameOverview
	var err error

	if optTx == nil {
		deletedGame, err = s.libraryService.queries.GetDeletedPlayToWinGameOverview(ctx, ptwGameId)
	} else {
		deletedGame, err = s.libraryService.queries.WithTx(optTx).GetDeletedPlayToWinGameOverview(ctx, ptwGameId)
	}

	if err != nil {
		if errors.Is(ErrNotFound, err) {
			return nil
		}
		return wrapDatabaseError(err)
	}

	_, err = WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		// Restore the PTW game
		err := s.libraryService.queries.WithTx(tx).RestorePlayToWinGame(ctx, ptwGameId)
		if err != nil {
			return nil, err
		}

		// Restore the library game
		err = s.libraryService.queries.WithTx(tx).RestoreGame(ctx, deletedGame.GameID)
		if err != nil {
			return nil, err
		}

		// Restore the winner entry if present
		if deletedGame.WinnerID.Valid {
			err = s.libraryService.queries.WithTx(tx).RestorePlayToWinEntry(ctx, deletedGame.WinnerID)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return wrapDatabaseError(err)
}
