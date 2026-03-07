package api

import (
	"errors"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

func (s Server) AddPlayToWinGame(c *gin.Context, gameId types.UUID) {
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId.String(), []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}

	_, err := s.queries.CreatePlayToWinGame(c.Request.Context(), gameUUID)
	var pgErr *pgconn.PgError
	if err != nil {
		// check for specific Postgres error codes:
		if errors.As(err, &pgErr) {
			// 23503 is the error code for a foreign key violation, which would occur if the game ID does not exist
			if pgErr.Code == "23503" {
				notFound(c)
				return
			}

			// 23505 is the error code for a unique constraint violation, which would occur if the game is already marked as play to win
			if pgErr.Code == "23505" {
				c.JSON(http.StatusNoContent, nil)
				return
			}
		}
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
		}
	}

	var request RemovePlayToWinGameJSONRequestBody
	err = c.ShouldBindJSON(request)
	if err != nil {
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
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
