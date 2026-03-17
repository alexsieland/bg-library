package api

import (
	"context"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type transactionService interface {
	CheckOutGame(ctx context.Context, gameId pgtype.UUID, patronId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryTransaction, error)
	CheckInGame(ctx context.Context, transactionId pgtype.UUID, optTx pgx.Tx) error
	GetGameStatus(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryGameStatus, error)
	SearchTransactionEvents(ctx context.Context, params db.SearchTransactionEventsParams, optTx pgx.Tx) ([]db.VwLibraryTransactionEvent, error)
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

func (s Server) ListTransactionEvents(c *gin.Context, params ListTransactionEventsParams) {
	dbArgs, errorDetails := getSearchTransactionEventsParams(params)
	if !errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}
	transactions, err := s.queries.SearchTransactionEvents(c.Request.Context(), dbArgs)
	if err != nil {
		log.Printf("Error listing transactions: %v", err)
		internalError(c, err)
		return
	}

	var events []TransactionEvent
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

	c.JSON(http.StatusOK, TransactionEventList{Transactions: events})
}

func getSearchTransactionEventsParams(params ListTransactionEventsParams) (db.SearchTransactionEventsParams, ErrorDetails) {
	var errorDetails ErrorDetails
	sanitizedTitle := pgtype.Text{String: "", Valid: true}
	if params.GameTitle != nil {
		errorDetails.ValidateStringLength("gameTitle", *params.GameTitle, 1, 100)
		sanitizedTitle = pgtype.Text{
			String: SanitizeTitle(*params.GameTitle),
			Valid:  true,
		}
	}
	patronFullName := ""
	if params.PatronName != nil {
		errorDetails.ValidateStringLength("patronName", *params.PatronName, 1, 100)
		patronFullName = *params.PatronName
	}
	var limit int32 = 100
	if params.Limit != nil {
		errorDetails.ValidateIntMin("limit", *params.Limit, 1)
		errorDetails.ValidateIntMax("limit", *params.Limit, 100)
		limit = int32(*params.Limit)
	}
	var offset int32 = 0
	if params.Offset != nil {
		errorDetails.ValidateIntMin("offset", *params.Offset, 0)
		offset = int32(*params.Offset)
	}
	if !errorDetails.Empty() {
		return db.SearchTransactionEventsParams{}, errorDetails
	}
	return db.SearchTransactionEventsParams{
		SanitizedTitle: sanitizedTitle,
		PatronFullName: patronFullName,
		Limit:          limit,
		Offset:         offset,
	}, errorDetails
}
