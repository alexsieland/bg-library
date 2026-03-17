package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PlayToWinService struct {
	LibraryService *LibraryService
}

func NewPlayToWinService(libService *LibraryService) *PlayToWinService {
	return &PlayToWinService{LibraryService: libService}
}

// InsertPlayToWinGroup inserts a new play to win group into the database.
// This call is idempotent. If the play to win group already exists, the existing play to win group will be returned.
func (s PlayToWinService) InsertPlayToWinGroup(ctx context.Context, groupName string, optTx pgx.Tx) (db.VwPlayToWinGroup, error) {
	ptwGroup, err := WithinTx(s.LibraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwPlayToWinGroup, error) {
		cpTx, err := tx.Begin(ctx)
		if err != nil {
			return nil, err
		}

		newPtwGroup, err := s.LibraryService.queries.WithTx(cpTx).CreatePlayToWinGroup(ctx, groupName)
		if err != nil {
			_ = cpTx.Rollback(ctx)

			if isUniqueConstraintViolation(err) {
				ptwGroup, err := s.LibraryService.queries.WithTx(tx).GetPlayToWinGroupByName(ctx, groupName)
				if err != nil {
					return nil, err
				}
				return &ptwGroup, err
			}

			return nil, err
		}

		_ = cpTx.Commit(ctx)
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
func (s PlayToWinService) InsertPlayToWin(ctx context.Context, gameId pgtype.UUID, ptwGroupId pgtype.UUID, optTx pgx.Tx) error {

	params := db.CreatePlayToWinGameParams{
		GameID:     pgtype.UUID{},
		PtwGroupID: pgtype.UUID{},
	}

	return nil
}

// DeletePlayToWin deletes a play to win from the database.
// This call is idempotent. If the play to win does not exist, it will be ignored.
func (s PlayToWinService) DeletePlayToWin(ctx context.Context, ptwGameId pgtype.UUID, deletionReason *string, deletionReasonComment *string, optTx pgx.Tx) error {
	dbDeletionReason, err := playToWinGameDeletionReason(deletionReason)
	if err != nil {
		return err
	}
	params := db.DeletePlayToWinGameByPlayToWinIdParams{
		ID:                    ptwGameId,
		DeletionReason:        dbDeletionReason,
		DeletionReasonComment: stringToPgText(deletionReasonComment),
	}

	_, err = WithinTx(s.LibraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.LibraryService.queries.WithTx(tx).DeletePlayToWinGameByPlayToWinId(ctx, params)
		return nil, err
	})

	return wrapDatabaseError(err)
}
