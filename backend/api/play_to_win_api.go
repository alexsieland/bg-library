package api

import (
	"errors"
	"log"
	"math/rand/v2"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

// Get Parent Play To Win ID from Ref ID.
// Intended for use when CRUD operations are performed on a session or entry
// This ensures the id always points the parent play-to-win game rather than the duplicate.
func (s Server) getParentPlayToWinId(c *gin.Context, ptwId pgtype.UUID) pgtype.UUID {
	if !ptwId.Valid {
		return ptwId
	}
	parentId, err := s.queries.GetParentPlayToWinId(c.Request.Context(), ptwId)
	if err == nil {
		if parentId.Valid {
			return parentId
		}
	}
	return ptwId
}

func (s Server) addPlayToWinByGameId(c *gin.Context, gameId types.UUID, optTx *pgx.Tx) error {
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

func (s Server) AddPlayToWinGameByGameId(c *gin.Context, gameId types.UUID) {
	err := s.addPlayToWinByGameId(c, gameId, nil)
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

func (s Server) removePlayToWinByGameId(c *gin.Context, gameId types.UUID, deletionReason *string, deletionComment *string, tx *pgx.Tx) error {
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

func (s Server) RemovePlayToWinGameByGameId(c *gin.Context, gameId types.UUID) {
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

	var request RemovePlayToWinGameRequest
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
	pgPtwId := s.getParentPlayToWinId(c, uuidToPgTypeUUID(playToWinId))
	dbPtwEntries, err := s.queries.GetPlayToWinEntries(c, pgPtwId)
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

func (s Server) addPlayToWinEntry(c *gin.Context, ptwSessionId pgtype.UUID, playToWinID pgtype.UUID, entry ptwEntry, tx pgx.Tx) (db.PlayToWinEntry, error) {
	pgPtwId := s.getParentPlayToWinId(c, playToWinID)
	playToWinEntryParams := db.CreatePlayToWinEntryParams{
		SessionID:       ptwSessionId,
		PlayToWinID:     pgPtwId,
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
	ptwId := s.getParentPlayToWinId(c, uuidToPgTypeUUID(jsonObject.PlayToWinId))
	ptwSessionParams := db.CreatePlayToWinSessionParams{
		PlayToWinID:     ptwId,
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
		entry, err := s.addPlayToWinEntry(c, dbPtwSession.ID, dbPtwSession.PlayToWinID, entry, tx)
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

func (s Server) GetPlayToWinGame(c *gin.Context, ptwId types.UUID) {
	dbPtwGame, err := s.queries.GetPlayToWinGame(c.Request.Context(), uuidToPgTypeUUID(ptwId))
	if err != nil {
		if isNotFound(err) {
			notFound(c)
			return
		}
		log.Printf("Error getting play to win game: %v", err)
		internalError(c, err)
		return
	}

	ptwGame := FromPlayToWinGameOverview(dbPtwGame)
	c.JSON(http.StatusOK, ptwGame)
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

	dbPTWGames, err := s.queries.ListPlayToWinGames(c.Request.Context(), requestParams)
	if err != nil {
		log.Printf("Error listing play to win games: %v", err)
		internalError(c, err)
		return
	}

	ptwGameList := FromPlayToWinGameList(dbPTWGames)

	c.JSON(http.StatusOK, ptwGameList)
}

func (s Server) UpdatePlayToWinGame(c *gin.Context, ptwId types.UUID) {
	var jsonObject UpdatePlayToWinGame
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	winnerId := pgtype.UUID{
		Valid: false,
	}
	if jsonObject.WinnerId != nil {
		winnerId = pgtype.UUID{
			Bytes: *jsonObject.WinnerId,
			Valid: true,
		}
	}

	params := db.UpdatePlayToWinEntryParams{
		ID:       uuidToPgTypeUUID(ptwId),
		WinnerID: winnerId,
	}

	err = s.queries.UpdatePlayToWinEntry(c.Request.Context(), params)
	if err != nil {
		if isForeignKeyConstraintViolation(err) {
			var errorDetails ErrorDetails
			errorDetails.AddErrorDetail("winnerId", "Must reference an entry belonging to this play to win game")
			validationError(c, errorDetails)
			return
		}
		log.Printf("Error updating play to win entry: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s Server) DeletePlayToWinGame(c *gin.Context, ptwId types.UUID) {
	// Get request body
	var request RemovePlayToWinGameRequest
	err := c.ShouldBindBodyWithJSON(&request)
	if err != nil {
		malformedJson(c)
		return
	}

	// Check that the game exists
	dbPtwId := s.getParentPlayToWinId(c, uuidToPgTypeUUID(ptwId))
	ptwGame, err := s.queries.GetPlayToWinGame(c.Request.Context(), dbPtwId)
	if err != nil {
		if isNotFound(err) {
			// Since delete deletes, if the game is not found we can pretend it was deleted and return 204
			c.JSON(http.StatusNoContent, nil)
			return
		}
		log.Printf("Error getting play to win game: %v", err)
		internalError(c, err)
	}

	// Validate request body fields and convert to db types
	var errorDetails ErrorDetails
	if request.RemovalComment != nil {
		errorDetails.ValidateStringLength("deletionComment", *request.RemovalComment, 0, 500)
	}
	var deletionReason *string
	if request.RemovalReason != "" {
		reason := string(request.RemovalReason)
		deletionReason = &reason
	}

	reason := playToWinGameDeletionReason(deletionReason, &errorDetails)
	if !errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}

	deleteParams := db.DeletePlayToWinGameByPlayToWinIdParams{
		ID:                    uuidToPgTypeUUID(ptwId),
		DeletionReason:        reason,
		DeletionReasonComment: stringToPgText(request.RemovalComment),
	}

	// Start transaction
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

	// Soft delete the play to win game
	err = s.queries.WithTx(tx).DeletePlayToWinGameByPlayToWinId(c.Request.Context(), deleteParams)
	if err != nil {
		log.Printf("Error deleting play to win game: %v", err)
		internalError(c, err)
		return
	}

	// If the deletion was a claimed prize, also delete the play to win entry to prevent it from showing up in the duplicate game raffle results
	if deleteParams.DeletionReason.Valid &&
		deleteParams.DeletionReason.PlayToWinGameDeletionType == db.PlayToWinGameDeletionTypeClaimed {

		// If claimed prize, soft delete the play to win entry so that entry is not available for potential duplicate game raffles
		if ptwGame.WinnerID.Valid {
			deleteEntryReason := db.NullPlayToWinEntryDeletionType{
				PlayToWinEntryDeletionType: db.PlayToWinEntryDeletionTypeWon,
				Valid:                      true,
			}
			deleteEntryParams := db.DeletePlayToWinEntryParams{
				ID:                    ptwGame.WinnerID,
				DeletionReason:        deleteEntryReason,
				DeletionReasonComment: pgtype.Text{Valid: false},
			}
			err = s.queries.WithTx(tx).DeletePlayToWinEntry(c.Request.Context(), deleteEntryParams)
			if err != nil {
				log.Printf("Error deleting play to win entry: %v", err)
				internalError(c, err)
				return
			}
		}

		// If claimed prize, soft delete the library game because it is no longer available for check out
		err = s.queries.WithTx(tx).DeleteGame(c.Request.Context(), ptwGame.GameID)
		if err != nil {
			log.Printf("Error deleting play to win entry: %v", err)
			internalError(c, err)
			return
		}
	}

	// Commit transaction
	err = tx.Commit(c.Request.Context())
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		internalError(c, err)
		return
	}

	// Set tx to nil to prevent deferred rollback and return 204
	tx = nil
	c.JSON(http.StatusNoContent, nil)
}

func (s Server) DrawPlayToWinRaffle(c *gin.Context, ptwId types.UUID) {
	pgPtwId := s.getParentPlayToWinId(c, uuidToPgTypeUUID(ptwId))
	if !pgPtwId.Valid {
		notFound(c)
		return
	}

	entries, err := s.queries.GetPlayToWinEntries(c.Request.Context(), pgPtwId)
	if err != nil {
		log.Printf("Error getting play to win entries: %v", err)
		internalError(c, err)
		return
	}

	if len(entries) == 0 {
		notFound(c)
		return
	}

	selectedPos := 0
	if len(entries) > 1 {
		selectedPos = rand.IntN(len(entries))
	}
	selectedEntry := entries[selectedPos]

	updateParams := db.UpdatePlayToWinEntryParams{
		ID:       pgPtwId,
		WinnerID: selectedEntry.EntryID,
	}

	err = s.queries.UpdatePlayToWinEntry(c.Request.Context(), updateParams)
	if err != nil {
		log.Printf("Error updating play to win entry: %v", err)
		internalError(c, err)
		return
	}

	winner := PlayToWinEntry{
		EntrantName:     selectedEntry.EntrantName,
		EntrantUniqueId: selectedEntry.EntrantUniqueID,
		EntryId:         pgUUIDToUUID(selectedEntry.EntryID),
	}

	c.JSON(http.StatusOK, winner)
}

func (s Server) ResetPlayToWinRaffle(c *gin.Context) {
	err := s.queries.ResetPlayToWinGameWinners(c.Request.Context())
	if err != nil {
		log.Printf("Error resetting play to win raffle: %v", err)
		internalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
