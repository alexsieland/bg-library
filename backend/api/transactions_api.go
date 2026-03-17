package api

import (
	"context"
	"log"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type transactionService interface {
	CheckOutGame(ctx context.Context, gameId pgtype.UUID, patronId pgtype.UUID, optTx pgx.Tx) (db.Transaction, error)
	CheckInGame(ctx context.Context, transactionId pgtype.UUID, optTx pgx.Tx) error
	ListTransactionEvents(ctx context.Context, sanitizedTitle *string, patronFullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.SearchTransactionEventsRow, error)
}

type TransactionApi struct {
	service transactionService
	beginTx func(ctx context.Context) (pgx.Tx, error)
}

func NewTransactionApi(libService *internal.LibraryService) *TransactionApi {
	service := internal.NewTransactionService(libService)
	return &TransactionApi{
		service: service,
		beginTx: func(ctx context.Context) (pgx.Tx, error) {
			return libService.Database.BeginTx(ctx, pgx.TxOptions{})
		},
	}
}

func (api TransactionApi) CheckInGame(ctx context.Context, params CheckInGameParams) error {
	err := api.service.CheckInGame(ctx, uuidToPgTypeUUID(params.TransactionId), nil)
	if err != nil {
		log.Printf("Error checking in game: %v", err)
		return err
	}
	return nil
}

func (api TransactionApi) CheckOutGame(ctx context.Context, request CheckOutGameJSONRequestBody) error {
	_, err := api.service.CheckOutGame(ctx, uuidToPgTypeUUID(request.GameId), uuidToPgTypeUUID(request.PatronId), nil)
	if err != nil {
		log.Printf("Error checking out game: %v", err)
		return err
	}
	return nil
}

func (api TransactionApi) ListTransactionEvents(ctx context.Context, params ListTransactionEventsParams) (TransactionEventList, error) {
	var (
		errorDetails ErrorDetails
		limit        int32 = 100
		offset       int32 = 0
		events       []TransactionEvent
	)
	if params.GameTitle != nil {
		errorDetails.ValidateStringLength("gameTitle", *params.GameTitle, 1, 100)
	}
	if params.PatronName != nil {
		errorDetails.ValidateStringLength("patronName", *params.PatronName, 1, 100)
	}
	if params.Limit != nil {
		errorDetails.ValidateIntMin("limit", *params.Limit, 1)
		errorDetails.ValidateIntMax("limit", *params.Limit, 100)
		limit = *params.Limit
	}
	if params.Offset != nil {
		errorDetails.ValidateIntMin("offset", *params.Offset, 0)
		offset = *params.Offset
	}
	if !errorDetails.Empty() {
		return TransactionEventList{}, errorDetails
	}

	transactions, err := api.service.ListTransactionEvents(ctx, params.GameTitle, params.PatronName, limit, offset, nil)
	if err != nil {
		log.Printf("Error listing transaction events: %v", err)
		return TransactionEventList{}, err
	}

	for _, transaction := range transactions {
		game := FromGame(db.Game{ID: transaction.GameID, Title: transaction.GameTitle}, transaction.PlayToWinGameID.Valid)
		patron := FromPatron(db.Patron{ID: transaction.PatronID, FullName: transaction.PatronFullName})
		eventType := TransactionEventEventType(transaction.EventType)
		events = append(events, TransactionEvent{
			TransactionId:  pgUUIDToUUID(transaction.TransactionID),
			Game:           game,
			Patron:         patron,
			EventTimestamp: transaction.EventTimestamp.Time,
			EventType:      eventType,
		})
	}

	return TransactionEventList{Transactions: events}, nil
}
