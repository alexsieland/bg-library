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
		// if unique constraint violation, then this means the game id is a duplicate, so return nil as it is already marked as play to win
		if isUniqueConstraintViolation(err) {
			return nil
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
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
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
	if err = c.ShouldBindBodyWithJSON(&request); err != nil {
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

func (s Server) GetPlayToWinSessionEntries(c *gin.Context, playToWinId types.UUID) {
	//TODO implement me
	panic("implement me")
}

type ptwEntry struct {
	EntrantName     string `json:"entrantName"`
	EntrantUniqueId string `json:"entrantUniqueId"`
}

func (s Server) AddPlayToWinSession(c *gin.Context) {
	var jsonObject CreatePlayToWinSessionRequest
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	var errorDetails []ErrorDetail
	ptwId, errorDetails := ConvertToPgTypeUUID("PlayToWinId", jsonObject.PlayToWinId.String(), errorDetails)
	var playtimeMinutes pgtype.Int4
	if jsonObject.PlaytimeMinutes != nil {
		playtimeMinutes = pgtype.Int4{
			Int32: int32(*jsonObject.PlaytimeMinutes),
			Valid: true,
		}
		errorDetails = ValidateIntMin("playtimeMinutes", *jsonObject.PlaytimeMinutes, 0, errorDetails)
	}
	ptwEntries := make([]ptwEntry, len(jsonObject.Entries))
	for i, entry := range jsonObject.Entries {
		errorDetails = ValidateStringLength("entrantName", entry.EntrantName, 1, 100, errorDetails)
		errorDetails = ValidateStringLength("entrantUniqueId", entry.EntrantUniqueId, 1, 100, errorDetails)
		ptwEntries[i] = ptwEntry{
			EntrantName:     entry.EntrantName,
			EntrantUniqueId: entry.EntrantUniqueId,
		}
	}
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	ptwSessionParams := db.CreatePlayToWinSessionParams{
		PlayToWinID:     ptwId,
		PlaytimeMinutes: playtimeMinutes,
	}

	tx, err := s.Database.BeginTx(c, pgx.TxOptions{})
	if err != nil {
		log.Printf("Error creating play to win session: %v", err)
		internalError(c, err)
		return
	}
	defer tx.Rollback(c)

	ptwSession, err := s.queries.WithTx(tx).CreatePlayToWinSession(c, ptwSessionParams)
	if err != nil {
		// if FK violation, then this means the play to win game id is invalid, so return 404
		if isForeignKeyConstraintViolation(err) {
			notFound(c)
			return
		}
		log.Printf("Error creating play to win session: %v", err)
		internalError(c, err)
		return
	}

	// Set tx to nil so that it is not closed by the defer
	tx = nil
	//TODO do the thing
}
