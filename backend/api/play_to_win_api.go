package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

func (s Server) addPlayToWin(c *gin.Context, gameId types.UUID, errorDetails []ErrorDetail, tx *pgx.Tx) error {
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId.String(), []ErrorDetail{})
	if len(errorDetails) > 0 {
		return errValidation
	}

	var err error
	if tx != nil {
		_, err = s.queries.WithTx(*tx).CreatePlayToWinGame(c.Request.Context(), gameUUID)
	} else {
		_, err = s.queries.CreatePlayToWinGame(c.Request.Context(), gameUUID)
	}

	if err != nil {
		var pgErr *pgconn.PgError

		// check for specific Postgres error codes:
		if errors.As(err, &pgErr) {

			// 23505 is the error code for a unique constraint violation, which would occur if the game is already marked as play to win
			if pgErr.Code == "23505" {
				return nil
			}
		}
		return err
	}
	return nil
}

func (s Server) AddPlayToWinGame(c *gin.Context, gameId types.UUID) {
	var errorDetails []ErrorDetail
	err := s.addPlayToWin(c, gameId, errorDetails, nil)
	// 23503 is the error code for a foreign key violation, which would occur if the game ID does not exist
	if err != nil {
		var pgErr *pgconn.PgError
		if pgErr.Code == "23503" {
			notFound(c)
			return
		}
		log.Printf("Error creating play to win game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s Server) RemovePlayToWinGame(c *gin.Context, gameId types.UUID) {
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId.String(), []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}

	dbGame, err := s.queries.GetGame(c.Request.Context(), gameUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			notFound(c)
			return
		}
		log.Printf("Error getting game: %v", err)
		internalError(c, err)
		return
	}

	var request RemovePlayToWinGameJSONRequestBody
	if err = c.ShouldBindJSON(&request); err != nil {
		malformedJson(c)
		return
	}

	deletionType := db.NullPlayToWinGameDeletionType{
		PlayToWinGameDeletionType: db.PlayToWinGameDeletionType(request.RemovalReason),
		Valid:                     true,
	}

	var comment pgtype.Text
	if request.RemovalComment != nil {
		errorDetails = ValidateStringLength("removalComment", *request.RemovalComment, 0, 500, []ErrorDetail{})
		if len(errorDetails) > 0 {
			validationError(c, errorDetails)
			return
		}
		comment = pgtype.Text{
			String: *request.RemovalComment,
			Valid:  true,
		}
	}

	params := db.DeletePlayToWinGameParams{
		ID:                    dbGame.PlayToWinGameID,
		DeletionReason:        deletionType,
		DeletionReasonComment: comment,
	}

	err = s.queries.DeletePlayToWinGame(c.Request.Context(), params)
	if err != nil {
		log.Printf("Error deleting play to win game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
