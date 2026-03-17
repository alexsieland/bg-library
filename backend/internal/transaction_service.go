package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TransactionService struct {
	libraryService *LibraryService
}

func NewTransactionService(libService *LibraryService) *TransactionService {
	return &TransactionService{libraryService: libService}
}

func (s TransactionService) CheckOutGame(ctx context.Context, gameId pgtype.UUID, patronId pgtype.UUID, optTx pgx.Tx) (db.Transaction, error) {
	transaction, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.Transaction, error) {
		gameStatus, err := s.libraryService.queries.WithTx(tx).GetGameStatus(ctx, gameId)
		if err != nil {
			return nil, err
		}

		if !gameStatus.CheckinTimestamp.Valid && gameStatus.PatronID.Valid {
			if patronId == gameStatus.PatronID {
				//Game is already checked out by this patron, so we return the current status of the game
				return &db.Transaction{
					ID:                gameStatus.TransactionID,
					GameID:            gameId,
					PatronID:          patronId,
					CheckoutTimestamp: gameStatus.CheckoutTimestamp,
					CheckinTimestamp:  gameStatus.CheckinTimestamp,
				}, nil
			}
			return nil, ErrCheckOutConflict
		}

		params := db.CheckOutGameParams{
			GameID:   gameId,
			PatronID: patronId,
		}

		transaction, err := s.libraryService.queries.CheckOutGame(ctx, params)
		if err != nil {
			return nil, err
		}
		return &transaction, nil
	})

	if err != nil {
		return db.Transaction{}, wrapDatabaseError(err)
	}
	return *transaction, nil
}

func (s TransactionService) CheckInGame(ctx context.Context, transactionId pgtype.UUID, optTx pgx.Tx) error {
	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.Patron, error) {
		err := s.libraryService.queries.WithTx(tx).CheckInGame(ctx, transactionId)
		return nil, err
	})

	if err != nil {
		return wrapDatabaseError(err)
	}
	return nil
}

func (s TransactionService) GetGameStatus(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryGameStatus, error) {
	//TODO implement me
	panic("implement me")
}

func (s TransactionService) SearchTransactionEvents(ctx context.Context, params db.SearchTransactionEventsParams, optTx pgx.Tx) ([]db.VwLibraryTransactionEvent, error) {
	//TODO implement me
	panic("implement me")
}
