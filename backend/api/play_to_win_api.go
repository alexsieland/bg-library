package api

import (
	"context"
	"log"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

type playToWinService interface {
	GetPlayToWinGameByLibraryGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error)
	GetPlayToWinGroup(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error)
	GetPlayToWinGroupByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGroup, error)
	GetPlayToWinGameOverview(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGameOverview, error)
	ListPlayToWinGameOverviews(ctx context.Context, gameTitle *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwPlayToWinGameOverview, error)
	GetPlayToWinGameEntriesByGroupId(ctx context.Context, ptwGroupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error)
	GetPlayToWinGameEntriesByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error)
	ListPlayToWinEntriesByPlayToWinGameId(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error)
	ListPlayToWinEntriesByGroupId(ctx context.Context, groupId pgtype.UUID, optTx pgx.Tx) ([]db.VwPlayToWinEntry, error)
	InsertPlayToWinSession(ctx context.Context, ptwGroupId pgtype.UUID, playtimeMinutes *int32, optTx pgx.Tx) (db.PlayToWinSession, error)
	InsertPlayToWinEntry(ctx context.Context, ptwSessionId pgtype.UUID, ptwGroupId pgtype.UUID, entrantName string, entrantUniqueID string, optTx pgx.Tx) (db.PlayToWinEntry, error)
	UpdatePlayToWinGameWinner(ctx context.Context, ptwGameId pgtype.UUID, entryId pgtype.UUID, optTx pgx.Tx) error
	InsertPlayToWinGroup(ctx context.Context, groupName string, optTx pgx.Tx) (db.VwPlayToWinGroup, error)
	InsertPlayToWinGame(ctx context.Context, gameId pgtype.UUID, optTx pgx.Tx) (db.VwPlayToWinGame, error)
	DeletePlayToWinGameByPlayToWinId(ctx context.Context, ptwGameId pgtype.UUID, deletionReason *string, deletionReasonComment *string, optTx pgx.Tx) error
	DeletePlayToWinGameByLibraryGameId(ctx context.Context, gameId pgtype.UUID, deletionReason *string, deletionReasonComment *string, optTx pgx.Tx) error
	ClaimPlayToWinGame(ctx context.Context, ptwGameId pgtype.UUID, optTx pgx.Tx) error
	ResetPlayToWinGameWinners(ctx context.Context, optTx pgx.Tx) error
}

type PlayToWinApi struct {
	service playToWinService
	beginTx func(ctx context.Context) (pgx.Tx, error)
}

func NewPlayToWinApi(libService *internal.LibraryService) *PlayToWinApi {
	service := internal.NewPlayToWinService(libService)
	service.SetGameService(internal.NewGameService(libService))
	return &PlayToWinApi{
		service: service,
		beginTx: func(ctx context.Context) (pgx.Tx, error) {
			return libService.Database.BeginTx(ctx, pgx.TxOptions{})
		},
	}

}

type ptwEntry struct {
	EntrantName     string `json:"entrantName"`
	EntrantUniqueId string `json:"entrantUniqueId"`
}

func (api *PlayToWinApi) RecordPlayToWinSession(ctx context.Context, request CreatePlayToWinSessionRequest, optTx pgx.Tx) (PlayToWinSession, error) {
	//Validate request body fields
	var errorDetails ErrorDetails
	if request.PlaytimeMinutes != nil {
		errorDetails.ValidateIntMin("playtimeMinutes", *request.PlaytimeMinutes, 0)
	}

	//Create all ptw entries and validate before creating the session
	ptwEntries := make([]ptwEntry, len(request.Entries))
	for i, entry := range request.Entries {
		errorDetails.ValidateStringLength("entrantName", entry.EntrantName, 1, 100)
		errorDetails.ValidateStringLength("entrantUniqueId", entry.EntrantUniqueId, 1, 100)
		ptwEntries[i] = ptwEntry{
			EntrantName:     entry.EntrantName,
			EntrantUniqueId: entry.EntrantUniqueId,
		}
	}
	if !errorDetails.Empty() {
		return PlayToWinSession{}, errorDetails
	}

	ptwGroup, err := api.service.GetPlayToWinGroupByPlayToWinGameId(ctx, uuidToPgTypeUUID(request.PlayToWinId), nil)
	if err != nil {
		return PlayToWinSession{}, err
	}

	tx, err := api.beginTx(ctx)
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return PlayToWinSession{}, err
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	session, err := api.service.InsertPlayToWinSession(ctx, ptwGroup.ID, request.PlaytimeMinutes, tx)
	if err != nil {
		return PlayToWinSession{}, err
	}

	var entries []PlayToWinEntry
	for _, entryParams := range request.Entries {
		entry, err := api.service.InsertPlayToWinEntry(ctx, session.ID, ptwGroup.ID, entryParams.EntrantName, entryParams.EntrantUniqueId, tx)
		if err != nil {
			return PlayToWinSession{}, err
		}

		entries = append(entries, PlayToWinEntry{
			EntryId:         pgUUIDToUUID(entry.ID),
			EntrantName:     entry.EntrantName,
			EntrantUniqueId: entry.EntrantUniqueID,
		})
	}

	ptwSession := PlayToWinSession{
		PlayToWinEntries: entries,
		PlaytimeMinutes:  pgInt4ToInteger(session.PlaytimeMinutes),
		SessionId:        pgUUIDToUUID(session.ID),
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return PlayToWinSession{}, err
	}

	tx = nil
	return ptwSession, nil
}

func (api *PlayToWinApi) AddPlayToWinGameByGameId(ctx context.Context, gameId types.UUID) (PlayToWinGame, error) {
	tx, err := api.beginTx(ctx)
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return PlayToWinGame{}, err
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	dbPtwGame, err := api.service.InsertPlayToWinGame(ctx, uuidToPgTypeUUID(gameId), tx)
	if err != nil {
		return PlayToWinGame{}, err
	}
	dbPtwGameOverview, err := api.service.GetPlayToWinGameOverview(ctx, dbPtwGame.ID, tx)
	if err != nil {
		return PlayToWinGame{}, err
	}

	var winnerEntry *PlayToWinEntry
	if dbPtwGameOverview.WinnerID.Valid {
		winnerEntry = &PlayToWinEntry{
			EntrantName:     dbPtwGameOverview.WinnerName.String,
			EntrantUniqueId: dbPtwGameOverview.WinnerUniqueID.String,
			EntryId:         pgUUIDToUUID(dbPtwGameOverview.WinnerID),
		}
	}

	ptwGame := PlayToWinGame{
		GameId:      pgUUIDToUUID(dbPtwGame.GameID),
		PlayToWinId: pgUUIDToUUID(dbPtwGame.ID),
		Title:       dbPtwGameOverview.GameTitle,
		Winner:      winnerEntry,
	}

	err = tx.Commit(ctx)
	if err != nil {
		return PlayToWinGame{}, err
	}
	tx = nil
	return ptwGame, nil
}

func (api *PlayToWinApi) RemovePlayToWinGameByGameId(c context.Context, gameId types.UUID, request RemovePlayToWinGameRequest) error {
	var errorDetails ErrorDetails

	validEnums := []string{
		string(db.PlayToWinGameDeletionTypeClaimed),
		string(db.PlayToWinGameDeletionTypeClaimed),
		string(db.PlayToWinGameDeletionTypeClaimed),
	}
	removalReason := string(request.RemovalReason)
	errorDetails.ValidateEnum("removalReason", removalReason, validEnums)

	if request.RemovalComment != nil {
		errorDetails.ValidateStringLength("removalComment", *request.RemovalComment, 0, 500)
	}

	if !errorDetails.Empty() {
		return errorDetails
	}

	return api.service.DeletePlayToWinGameByLibraryGameId(c, uuidToPgTypeUUID(gameId), &removalReason, request.RemovalComment, nil)
}

func (api *PlayToWinApi) GetPlayToWinGameEntries(ctx context.Context, ptwGameId types.UUID) (PlayToWinEntryList, error) {
	dbPtwEntries, err := api.service.GetPlayToWinGameEntriesByPlayToWinGameId(ctx, uuidToPgTypeUUID(ptwGameId), nil)
	if err != nil {
		return PlayToWinEntryList{}, err
	}
	ptwEntries := make([]PlayToWinEntry, len(dbPtwEntries))
	for i, dbPtwEntry := range dbPtwEntries {
		ptwEntries[i] = PlayToWinEntry{
			EntryId:         pgUUIDToUUID(dbPtwEntry.ID),
			EntrantName:     dbPtwEntry.EntrantName,
			EntrantUniqueId: dbPtwEntry.EntrantUniqueID,
		}
	}
	return PlayToWinEntryList{Entries: ptwEntries}, nil
}

func (api *PlayToWinApi) GetPlayToWinGameOverview(ctx context.Context, ptwId types.UUID) (PlayToWinGame, error) {
	dbPtwGame, err := api.service.GetPlayToWinGameOverview(ctx, uuidToPgTypeUUID(ptwId), nil)
	if err != nil {
		return PlayToWinGame{}, err
	}

	ptwGame := FromPlayToWinGameOverview(dbPtwGame)
	return ptwGame, nil
}

func (api *PlayToWinApi) ListPlayToWinGames(c *gin.Context, params ListPlayToWinGamesParams) (PlayToWinGameList, error) {
	var (
		limit        int32 = 100
		offset       int32 = 0
		errorDetails ErrorDetails
	)

	// Validate query params
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
		errorDetails.ValidateStringLength("title", *params.Title, 1, 100)
	}
	if !errorDetails.Empty() {
		return PlayToWinGameList{}, errorDetails
	}

	// Get play to win games based on query params
	dbPTWGames, err := api.service.ListPlayToWinGameOverviews(c.Request.Context(), params.Title, limit, offset, nil)
	if err != nil {
		return PlayToWinGameList{}, err
	}

	ptwGameList := FromPlayToWinGameList(dbPTWGames)
	return ptwGameList, nil
}

func (api *PlayToWinApi) UpdatePlayToWinGame(ctx context.Context, ptwId types.UUID, request UpdatePlayToWinGame) error {
	winnerId := pgtype.UUID{
		Valid: false,
	}
	if request.WinnerId != nil {
		winnerId = pgtype.UUID{
			Bytes: *request.WinnerId,
			Valid: true,
		}
	}

	return api.service.UpdatePlayToWinGameWinner(ctx, uuidToPgTypeUUID(ptwId), winnerId, nil)
}

func (api *PlayToWinApi) DeletePlayToWinGame(ctx context.Context, ptwId types.UUID, request RemovePlayToWinGameRequest) error {
	//// Get request body
	//var request RemovePlayToWinGameRequest
	//err := c.ShouldBindBodyWithJSON(&request)
	//if err != nil {
	//	malformedJson(c)
	//	return
	//}
	//
	//// Check that the game exists
	//ptwGame, err := s.queries.GetPlayToWinGame(c.Request.Context(), uuidToPgTypeUUID(ptwId))
	//if err != nil {
	//	if isNotFound(err) {
	//		// Since delete deletes, if the game is not found we can pretend it was deleted and return 204
	//		c.JSON(http.StatusNoContent, nil)
	//		return
	//	}
	//	log.Printf("Error getting play to win game: %v", err)
	//	internalError(c, err)
	//}
	//
	//// Validate request body fields and convert to db types
	//var errorDetails ErrorDetails
	//if request.RemovalComment != nil {
	//	errorDetails.ValidateStringLength("deletionComment", *request.RemovalComment, 0, 500)
	//}
	//var deletionReason *string
	//if request.RemovalReason != "" {
	//	reason := string(request.RemovalReason)
	//	deletionReason = &reason
	//}
	//
	//reason := playToWinGameDeletionReason(deletionReason, &errorDetails)
	//if !errorDetails.Empty() {
	//	validationError(c, errorDetails)
	//	return
	//}
	//
	//deleteParams := db.DeletePlayToWinGameByPlayToWinIdParams{
	//	ID:                    uuidToPgTypeUUID(ptwId),
	//	DeletionReason:        reason,
	//	DeletionReasonComment: stringToPgText(request.RemovalComment),
	//}
	//
	//// Start transaction
	//tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
	//if err != nil {
	//	log.Printf("Error creating play to win session: %v", err)
	//	internalError(c, err)
	//	return
	//}
	//defer func() {
	//	if tx != nil {
	//		_ = tx.Rollback(c.Request.Context())
	//	}
	//}()
	//
	//// Soft delete the play to win game
	//err = s.queries.WithTx(tx).DeletePlayToWinGameByPlayToWinId(c.Request.Context(), deleteParams)
	//if err != nil {
	//	log.Printf("Error deleting play to win game: %v", err)
	//	internalError(c, err)
	//	return
	//}
	//
	//// If the deletion was a claimed prize, also delete the play to win entry to prevent it from showing up in the duplicate game raffle results
	//if deleteParams.DeletionReason.Valid &&
	//	deleteParams.DeletionReason.PlayToWinGameDeletionType == db.PlayToWinGameDeletionTypeClaimed {
	//
	//	// If claimed prize, soft delete the play to win entry so that entry is not available for potential duplicate game raffles
	//	if ptwGame.WinnerID.Valid {
	//		deleteEntryReason := db.NullPlayToWinEntryDeletionType{
	//			PlayToWinEntryDeletionType: db.PlayToWinEntryDeletionTypeWon,
	//			Valid:                      true,
	//		}
	//		deleteEntryParams := db.DeletePlayToWinEntryParams{
	//			ID:                    ptwGame.WinnerID,
	//			DeletionReason:        deleteEntryReason,
	//			DeletionReasonComment: pgtype.Text{Valid: false},
	//		}
	//		err = s.queries.WithTx(tx).DeletePlayToWinEntry(c.Request.Context(), deleteEntryParams)
	//		if err != nil {
	//			log.Printf("Error deleting play to win entry: %v", err)
	//			internalError(c, err)
	//			return
	//		}
	//	}
	//
	//	// If claimed prize, soft delete the library game because it is no longer available for check out
	//	err = s.queries.WithTx(tx).DeleteGame(c.Request.Context(), ptwGame.GameID)
	//	if err != nil {
	//		log.Printf("Error deleting play to win entry: %v", err)
	//		internalError(c, err)
	//		return
	//	}
	//}
	//
	//// Commit transaction
	//err = tx.Commit(c.Request.Context())
	//if err != nil {
	//	log.Printf("Error committing transaction: %v", err)
	//	internalError(c, err)
	//	return
	//}
	//
	//// Set tx to nil to prevent deferred rollback and return 204
	//tx = nil
	//c.JSON(http.StatusNoContent, nil)
	panic("not implemented")
}

func (api *PlayToWinApi) DrawPlayToWinRaffle(c *gin.Context, ptwId types.UUID) {
	//entries, err := s.queries.GetPlayToWinEntriesByPlayToWinGameId(c.Request.Context(), uuidToPgTypeUUID(ptwId))
	//if err != nil {
	//	log.Printf("Error getting play to win entries: %v", err)
	//	internalError(c, err)
	//	return
	//}
	//
	//if len(entries) == 0 {
	//	notFound(c)
	//	return
	//}
	//
	//selectedPos := 0
	//if len(entries) > 1 {
	//	selectedPos = rand.IntN(len(entries))
	//}
	//selectedEntry := entries[selectedPos]
	//
	//updateParams := db.UpdatePlayToWinWinnerParams{
	//	ID:       uuidToPgTypeUUID(ptwId),
	//	WinnerID: selectedEntry.PtwEntryID,
	//}
	//
	//err = s.queries.UpdatePlayToWinWinner(c.Request.Context(), updateParams)
	//if err != nil {
	//	log.Printf("Error updating play to win entry: %v", err)
	//	internalError(c, err)
	//	return
	//}
	//
	//winner := PlayToWinEntry{
	//	EntrantName:     selectedEntry.EntrantName,
	//	EntrantUniqueId: selectedEntry.EntrantUniqueID,
	//	EntryId:         pgUUIDToUUID(selectedEntry.PtwEntryID),
	//}
	//
	//c.JSON(http.StatusOK, winner)
	panic("not implemented")
}

func (api *PlayToWinApi) ResetPlayToWinRaffle(ctx context.Context) error {
	return api.service.ResetPlayToWinGameWinners(ctx, nil)
}
