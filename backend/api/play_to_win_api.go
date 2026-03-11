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

func (s Server) addPlayToWin(c *gin.Context, gameId types.UUID, optTx *pgx.Tx) error {
	var (
		err error
		tx  pgx.Tx
	)

	if optTx != nil {
		tx = *optTx
	} else {
		tx, err = s.Database.BeginTx(c, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback(c.Request.Context())
			}
		}()
	}

	_, err = s.queries.WithTx(tx).CreatePlayToWinGame(c.Request.Context(), uuidToPgTypeUUID(gameId))
	if err != nil {
		// If unique constraint violation, this is idempotent: restore soft-deleted row if needed.
		if isUniqueConstraintViolation(err) {
			err = s.queries.WithTx(tx).RestorePlayToWinGame(c.Request.Context(), uuidToPgTypeUUID(gameId))
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if optTx == nil {
		err := tx.Commit(c.Request.Context())
		if err != nil {
			log.Printf("Error committing play to win game transaction: %v", err)
			return err
		}
		tx = nil
	}
	return nil
}

func (s Server) AddPlayToWinGame(c *gin.Context, gameId types.UUID) {
	err := s.addPlayToWin(c, gameId, nil)
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
	var errorDetails ErrorDetails
	if deletionComment != nil {
		errorDetails.ValidateStringLength("deletionComment", *deletionComment, 0, 500)
	}
	reason := playToWinGameDeletionReason(deletionReason, &errorDetails)
	if !errorDetails.Empty() {
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
		if isNotFound(err) {
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
		var errorDetails ErrorDetails
		errorDetails.ValidateStringLength("removalComment", *request.RemovalComment, 0, 500)
		if !errorDetails.Empty() {
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
	var errorDetails ErrorDetails
	if jsonObject.PlaytimeMinutes != nil {
		errorDetails.ValidateIntMin("playtimeMinutes", *jsonObject.PlaytimeMinutes, 0)
	}

	//Create all ptw entries and validate before creating the session
	ptwEntries := make([]ptwEntry, len(jsonObject.Entries))
	for i, entry := range jsonObject.Entries {
		errorDetails.ValidateStringLength("entrantName", entry.EntrantName, 1, 100)
		errorDetails.ValidateStringLength("entrantUniqueId", entry.EntrantUniqueId, 1, 100)
		ptwEntries[i] = ptwEntry{
			EntrantName:     entry.EntrantName,
			EntrantUniqueId: entry.EntrantUniqueId,
		}
	}
	if !errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}

	// Create the play to win session params
	ptwSessionParams := db.CreatePlayToWinSessionParams{
		PlayToWinID:     uuidToPgTypeUUID(jsonObject.PlayToWinId),
		PlaytimeMinutes: int32ToPgInt4(jsonObject.PlaytimeMinutes),
	}

	// Create the play to win session
	tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		log.Printf("Error creating play to win session: %v", err)
		internalError(c, err)
		return
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(c.Request.Context())
		}
	}()

	dbPtwSession, err := s.queries.WithTx(tx).CreatePlayToWinSession(c.Request.Context(), ptwSessionParams)
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

	if err = tx.Commit(c.Request.Context()); err != nil {
		log.Printf("Error committing play to win session transaction: %v", err)
		internalError(c, err)
		return
	}
	tx = nil
	c.JSON(http.StatusCreated, ptwSession)
}

func (s Server) ListPlayToWinGames(c *gin.Context, params ListPlayToWinGamesParams) {
	var (
		sanitizedTitle string
		limit          int32 = 100
		offset         int32 = 0
		errorDetails   ErrorDetails
	)
	if params.Limit != nil {
		limit = *params.Limit
		errorDetails.ValidateIntMin("limit", limit, 1)
		errorDetails.ValidateIntMax("limit", limit, 100)
	}
	if params.Offset != nil {
		offset = *params.Offset
		errorDetails.ValidateIntMin("offset", offset, 0)
	}
	if params.Title != nil && *params.Title != "" {
		sanitizedTitle = SanitizeTitle(*params.Title)
		errorDetails.ValidateStringLength("title", sanitizedTitle, 1, 100)
	} else {
		sanitizedTitle = ""
	}
	if !errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}
	requestParams := db.ListPlayToWinGamesParams{
		SanitizedTitle: "%" + sanitizedTitle + "%",
		Limit:          limit,
		Offset:         offset,
	}

	dbPTWGames, err := s.queries.ListPlayToWinGames(c, requestParams)
	if err != nil {
		log.Printf("Error listing play to win games: %v", err)
		internalError(c, err)
		return
	}

	ptwGameList := PlayToWinGameList{
		Games: make([]PlayToWinGame, len(dbPTWGames)),
	}
	for i, dbPTWGame := range dbPTWGames {
		ptwGameList.Games[i] = PlayToWinGame{
			GameId:      pgUUIDToUUID(dbPTWGame.GameID),
			PlayToWinId: pgUUIDToUUID(dbPTWGame.PlayToWinID),
			Title:       dbPTWGame.GameTitle,
		}
	}

	c.JSON(http.StatusOK, ptwGameList)
}
