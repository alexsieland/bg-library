package api

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s Server) CheckInGame(c *gin.Context, params CheckInGameParams) {
	transactionUUID, errorDetails := ConvertToPgTypeUUID("TransactionId", params.TransactionId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
	}
	err := s.queries.CheckInGame(c.Request.Context(), transactionUUID)
	if err != nil {
		log.Printf("Error checking in game: %v", err)
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
