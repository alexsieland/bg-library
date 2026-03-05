package api

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) CheckInGame(c *gin.Context, params CheckInGameParams) {
	transactionUUID, errorDetails := ConvertToPgTypeUUID("TransactionId", params.TransactionId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	err := s.queries.CheckInGame(c.Request.Context(), transactionUUID)
	if err != nil {
		log.Printf("Error checking in game: %v", err)
		internalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (s Server) CheckOutGame(c *gin.Context) {
	var jsonObject CheckOutGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", jsonObject.GameId.String(), []ErrorDetail{})
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", jsonObject.PatronId.String(), errorDetails)
	gameStatus, err := s.queries.GetGameStatus(c.Request.Context(), gameUUID)
	if err != nil {
		log.Printf("Error getting game status: %v", err)
		internalError(c, err)
		return
	}
	if !gameStatus.CheckinTimestamp.Valid && gameStatus.PatronID.Valid {
		if patronUUID == gameStatus.PatronID {
			//Game is already checked out by this patron, so we return the current status of the game
			c.JSON(http.StatusCreated, LibraryTransaction{
				GameId:    uuid.MustParse(gameStatus.GameID.String()),
				Id:        uuid.MustParse(gameStatus.TransactionID.String()),
				PatronId:  uuid.MustParse(gameStatus.PatronID.String()),
				Timestamp: gameStatus.CheckoutTimestamp.Time,
			})
			return
		}
		conflict(c, "Game is already checked out by another patron")
		return
	}

	transaction, err := s.queries.CheckOutGame(c.Request.Context(), db.CheckOutGameParams{
		GameID:   gameUUID,
		PatronID: patronUUID,
	})
	if err != nil {
		log.Printf("Error checking out game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, FromTransaction(transaction))
}

func (s Server) ListTransactionEvents(c *gin.Context, params ListTransactionEventsParams) {
	dbArgs, verr := getSearchTransactionEventsParams(params)
	if len(verr) > 0 {
		validationError(c, verr)
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
		game := FromGame(db.Game{ID: transaction.GameID, Title: transaction.GameTitle})
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

func getSearchTransactionEventsParams(params ListTransactionEventsParams) (db.SearchTransactionEventsParams, []ErrorDetail) {
	var errorDetails []ErrorDetail
	sanitizedTitle := pgtype.Text{String: "", Valid: true}
	if params.GameTitle != nil {
		ValidateStringLength("gameTitle", *params.GameTitle, 1, 100, errorDetails)
		sanitizedTitle = pgtype.Text{
			String: SanitizeTitle(*params.GameTitle),
			Valid:  true,
		}
	}
	patronFullName := ""
	if params.PatronName != nil {
		ValidateStringLength("patronName", *params.GameTitle, 1, 100, errorDetails)
		patronFullName = *params.PatronName
	}
	var limit int32 = 100
	if params.Limit != nil {
		ValidateIntMin("limit", *params.Limit, 1, errorDetails)
		ValidateIntMax("limit", *params.Limit, 100, errorDetails)
		limit = int32(*params.Limit)
	}
	var offset int32 = 0
	if params.Offset != nil {
		ValidateIntMin("offset", *params.Limit, 1, errorDetails)
		offset = int32(*params.Offset)
	}
	if len(errorDetails) > 0 {
		return db.SearchTransactionEventsParams{}, errorDetails
	}
	return db.SearchTransactionEventsParams{
		SanitizedTitle: sanitizedTitle,
		PatronFullName: patronFullName,
		Limit:          limit,
		Offset:         offset,
	}, errorDetails
}
