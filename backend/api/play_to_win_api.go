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

func (s Server) addPlayToWin(c *gin.Context, gameId types.UUID, errorDetails []ErrorDetail, optTx *pgx.Tx) error {
	var err error
	var tx pgx.Tx
	if tx == nil {
		tx, err = s.Database.BeginTx(c, pgx.TxOptions{})
		if err != nil {
			defer tx.Rollback(c)
			return err
		}
	} else {
		tx = *optTx
	}

	_, err = s.queries.WithTx(tx).CreatePlayToWinGame(c.Request.Context(), uuidToPgTypeUUID(gameId))

	if err != nil {
		// if unique constraint violation, then this means the game id is a duplicate, so attempt to restore it instead
		if isUniqueConstraintViolation(err) {
			err = s.queries.RestorePlayToWinGame(c.Request.Context(), uuidToPgTypeUUID(gameId))
			if err != nil {
				return err
			}
			tx = nil
			return nil
		}
		return err
	}
	tx = nil
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

func (s Server) removePlayToWin(c *gin.Context, gameId types.UUID, deletionReason *string, deletionComment *string, tx *pgx.Tx) error {
	var errorDetails []ErrorDetail
	if deletionComment != nil {
		errorDetails = ValidateStringLength("deletionComment", *deletionComment, 0, 500, errorDetails)
	}
	reason, errorDetails := playToWinGameDeletionReason(deletionReason, errorDetails)
	if len(errorDetails) > 0 {
		return errValidation
	}

	deleteParams := db.DeletePlayToWinGameParams{
		GameID:                uuidToPgTypeUUID(gameId),
		DeletionReason:        reason,
		DeletionReasonComment: stringToPgText(deletionComment),
	}

	var err error
	if tx != nil {
		err = s.queries.WithTx(*tx).DeletePlayToWinGame(c.Request.Context(), deleteParams)
	} else {
		err = s.queries.DeletePlayToWinGame(c.Request.Context(), deleteParams)
	}
	if err != nil {
		log.Printf("Error deleting play to win game: %v", err)
		return err
	}

	return nil
}

func (s Server) RemovePlayToWinGame(c *gin.Context, gameId types.UUID) {
	dbGame, err := s.queries.GetGame(c.Request.Context(), uuidToPgTypeUUID(gameId))
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
		errorDetails := ValidateStringLength("removalComment", *request.RemovalComment, 0, 500, []ErrorDetail{})
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
		GameID:                dbGame.PlayToWinGameID,
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
	dbPtwEntries, err := s.queries.GetPlayToWinEntries(c, uuidToPgTypeUUID(playToWinId))
	if err != nil {
		log.Printf("Error getting play to win entries: %v", err)
		internalError(c, err)
		return
	}

	ptwEntryList := PlayToWinEntryList{
		Entries: make([]PlayToWinEntry, len(dbPtwEntries)),
	}
	for i, dbPtwEntry := range dbPtwEntries {
		ptwEntryList.Entries[i] = PlayToWinEntry{
			EntryId:         pgUUIDToUUID(dbPtwEntry.EntryID),
			EntrantName:     dbPtwEntry.EntrantName,
			EntrantUniqueId: dbPtwEntry.EntrantUniqueID,
		}
	}

	c.JSON(http.StatusOK, ptwEntryList)
}

type ptwEntry struct {
	EntrantName     string `json:"entrantName"`
	EntrantUniqueId string `json:"entrantUniqueId"`
}

func (s Server) addPlayToWinEntry(c *gin.Context, ptwSessionId pgtype.UUID, entry ptwEntry, tx pgx.Tx) (db.PlayToWinEntry, error) {
	playToWinEntryParams := db.CreatePlayToWinEntryParams{
		SessionID:       ptwSessionId,
		EntrantName:     entry.EntrantName,
		EntrantUniqueID: entry.EntrantUniqueId,
	}
	return s.queries.WithTx(tx).CreatePlayToWinEntry(c.Request.Context(), playToWinEntryParams)
}

func (s Server) AddPlayToWinSession(c *gin.Context) {
	//Check that request body is valid json
	var jsonObject CreatePlayToWinSessionRequest
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	//Validate request body fields
	var errorDetails []ErrorDetail
	var playtimeMinutes pgtype.Int4
	if jsonObject.PlaytimeMinutes != nil {
		playtimeMinutes = pgtype.Int4{
			Int32: int32(*jsonObject.PlaytimeMinutes),
			Valid: true,
		}
		errorDetails = ValidateIntMin("playtimeMinutes", *jsonObject.PlaytimeMinutes, 0, errorDetails)
	}

	//Create all ptw entries and validate before creating the session
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

	// Create the play to win session params
	ptwSessionParams := db.CreatePlayToWinSessionParams{
		PlayToWinID:     uuidToPgTypeUUID(jsonObject.PlayToWinId),
		PlaytimeMinutes: playtimeMinutes,
	}

	// Create the play to win session
	tx, err := s.Database.BeginTx(c, pgx.TxOptions{})
	if err != nil {
		log.Printf("Error creating play to win session: %v", err)
		internalError(c, err)
		return
	}
	defer tx.Rollback(c)

	dbPtwSession, err := s.queries.WithTx(tx).CreatePlayToWinSession(c, ptwSessionParams)
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

	// Create the play to win session response
	var dbPtwSessionPlayTime *int32
	if dbPtwSession.PlaytimeMinutes.Valid {
		dbPtwSessionPlayTime = &dbPtwSession.PlaytimeMinutes.Int32
	}
	ptwSession := PlayToWinSession{
		PlayToWinEntries: make([]PlayToWinEntry, len(ptwEntries)),
		PlaytimeMinutes:  dbPtwSessionPlayTime,
		SessionId:        pgUUIDToUUID(dbPtwSession.ID),
	}

	// Create all play to win entries for session
	for i, entry := range ptwEntries {
		entry, err := s.addPlayToWinEntry(c, dbPtwSession.ID, entry, tx)
		if err != nil {
			log.Printf("Error creating play to win entry: %v", err)
			internalError(c, err)
			return
		}

		// Add entries to the session response
		ptwSession.PlayToWinEntries[i] = PlayToWinEntry{
			EntrantName:     entry.EntrantName,
			EntrantUniqueId: entry.EntrantUniqueID,
			EntryId:         pgUUIDToUUID(entry.ID),
		}
	}

	// Set tx to nil so that it is not closed by the defer
	tx = nil
	c.JSON(http.StatusCreated, ptwSession)
}
