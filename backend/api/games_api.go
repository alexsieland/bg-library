package api

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s Server) AddGame(c *gin.Context) {
	var jsonObject AddGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	var errorDetails []ErrorDetail
	isPlayToWin := false
	if jsonObject.IsPlayToWin != nil {
		isPlayToWin = *jsonObject.IsPlayToWin
	}
	dbGame, err := s.insertGame(c, jsonObject.Title, jsonObject.Barcode, isPlayToWin, errorDetails, nil)
	if errors.Is(err, errValidation) {
		validationError(c, errorDetails)
	}
	if err != nil {
		log.Printf("Error creating game: %v", err)
		internalError(c, err)
		return
	}

	//isPlayToWin is always false because games must exist to be marked as play to win
	c.JSON(http.StatusCreated, FromGame(dbGame, isPlayToWin))
}

func (s Server) insertGame(c *gin.Context, title string, barcode *string, isPlayToWin bool, errorDetails []ErrorDetail, tx *pgx.Tx) (db.Game, error) {
	errorDetails = ValidateStringLength("title", title, 1, 100, errorDetails)
	if barcode != nil {
		errorDetails = ValidateStringLength("barcode", *barcode, 1, 48, errorDetails)
	}
	if len(errorDetails) > 0 {
		return db.Game{}, errValidation
	}

	dbBarcode := pgtype.Text{Valid: false}
	if barcode != nil {
		dbBarcode = pgtype.Text{String: *barcode, Valid: true}
	}
	createGameParams := db.CreateGameParams{
		Title:          title,
		SanitizedTitle: SanitizeTitle(title),
		Barcode:        dbBarcode,
	}

	var game db.Game
	var err error
	if tx != nil {
		game, err = s.queries.WithTx(*tx).CreateGame(c.Request.Context(), createGameParams)
	} else {
		game, err = s.queries.CreateGame(c.Request.Context(), createGameParams)
	}

	if err != nil {
		return db.Game{}, err
	}

	if isPlayToWin {
		err = s.addPlayToWin(c, pgUUIDToUUID(game.ID), errorDetails, tx)
	}
	return game, err
}

func (s Server) BulkAddGames(c *gin.Context) {
	decodedReader := base64.NewDecoder(base64.StdEncoding, c.Request.Body)
	csvReader := csv.NewReader(decodedReader)

	// Start a db transaction
	tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		log.Printf("Error creating transaction: %v", err)
		internalError(c, err)
		return
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(c.Request.Context())
		}
	}()

	// Process each row
	var errorDetails []ErrorDetail
	recordCount := 0
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV: %v", err)
			internalError(c, err)
			return
		}
		if len(record) == 0 {
			continue
		}
		title := record[0]
		var barcode *string
		// Disable the ability to set the barcode on bulk add for now
		// TODO Add this back in once barcode implementation is complete
		//if len(record) > 1 && record[1] != "" {
		//	barcode = &record[1]
		//}

		_, err = s.insertGame(c, title, barcode, false, errorDetails, &tx)
		if errors.Is(err, errValidation) {
			continue
		}
		if err != nil {
			log.Printf("Error adding game: %v", err)
			internalError(c, err)
			return
		}
		recordCount++
	}

	//If there are any validation errors, rollback the transaction
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}

	//If there are no validation errors, commit the transaction
	err = tx.Commit(c.Request.Context())
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		internalError(c, err)
	}
	tx = nil // Prevent deferred rollback after a successful commit

	c.JSON(http.StatusCreated, BulkAddResponse{Imported: recordCount})
}

func (s Server) DeleteGame(c *gin.Context, gameId string) {
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	err := s.queries.DeleteGame(c.Request.Context(), gameUUID)
	if err != nil {
		log.Printf("Error deleting game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s Server) GetGame(c *gin.Context, gameId string) {
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId, []ErrorDetail{})
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
	c.JSON(http.StatusOK, FromVwLibraryGame(dbGame))
}

func (s Server) GetGameByBarcode(c *gin.Context, gameBarcode string) {
	errorDetails := ValidateStringLength("gameBarcode", gameBarcode, 1, 48, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	var barcode = pgtype.Text{String: gameBarcode, Valid: true}
	dbGames, err := s.queries.GetGameByBarcode(c.Request.Context(), barcode)
	if err != nil {
		log.Printf("Error getting game: %v", err)
		internalError(c, err)
		return
	}
	if dbGames == nil || len(dbGames) == 0 {
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, FromVwLibraryGames(dbGames))
}

func (s Server) UpdateGame(c *gin.Context, gameId string) {
	var jsonObject UpdateGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}
	errorDetails := ValidateStringLength("title", jsonObject.Title, 1, 100, []ErrorDetail{})
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId, errorDetails)
	if jsonObject.Barcode != nil {
		errorDetails = ValidateStringLength("barcode", *jsonObject.Barcode, 1, 48, errorDetails)
	}
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}

	dbBarcode := pgtype.Text{Valid: false}
	if jsonObject.Barcode != nil {
		dbBarcode = pgtype.Text{String: *jsonObject.Barcode, Valid: true}
	}

	err = s.queries.EditGame(c.Request.Context(), db.EditGameParams{
		ID:             gameUUID,
		Title:          jsonObject.Title,
		SanitizedTitle: SanitizeTitle(jsonObject.Title),
		Barcode:        dbBarcode,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		// Postgres 23503 is the error code for a unique constraint violation
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			notFound(c)
			return
		}
		log.Printf("Error updating game: %v", err)
		internalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (s Server) listCheckedOutGames(c *gin.Context, params ListGamesParams) {
	var dbGameStatusList []db.VwGameStatus
	if params.Title == nil {
		var err error
		dbGameStatusList, err = s.queries.ListCheckedOutGames(c.Request.Context(), db.ListCheckedOutGamesParams{
			Limit:  999,
			Offset: 0,
		})

		if err != nil {
			log.Printf("Error listing checked out games: %v", err)
			internalError(c, err)
		}

	} else {
		var err error
		title := *params.Title
		dbGameStatusList, err = s.queries.SearchCheckedOutGames(c.Request.Context(), db.SearchCheckedOutGamesParams{
			SanitizedTitle: "%" + SanitizeTitle(title) + "%",
			Limit:          999,
			Offset:         0,
		})
		if err != nil {
			log.Printf("Error searching checked out games: %v", err)
			internalError(c, err)
			return
		}
	}

	gameStatusList := make([]GameStatus, len(dbGameStatusList))
	for i, dbGameStatus := range dbGameStatusList {
		gameStatusList[i] = FromVwGameStatus(dbGameStatus)
	}

	c.JSON(http.StatusOK, GameStatusList{Games: gameStatusList})
}

func (s Server) ListGames(c *gin.Context, params ListGamesParams) {
	if params.CheckedOut != nil && *params.CheckedOut {
		s.listCheckedOutGames(c, params)
		return
	}
	var dbGameStatusList []db.VwGameStatus
	if params.Title == nil {
		var err error
		dbGameStatusList, err = s.queries.ListGamesStatus(c.Request.Context(), db.ListGamesStatusParams{
			Limit:  999,
			Offset: 0,
		})

		if err != nil {
			log.Printf("Error listing games: %v", err)
			internalError(c, err)
		}

	} else {
		var err error
		title := *params.Title
		dbGameStatusList, err = s.queries.SearchGameStatus(c.Request.Context(), db.SearchGameStatusParams{
			SanitizedTitle: "%" + SanitizeTitle(title) + "%",
			Limit:          999,
			Offset:         0,
		})
		if err != nil {
			log.Printf("Error searching games: %v", err)
			internalError(c, err)
			return
		}
	}

	gameStatusList := make([]GameStatus, len(dbGameStatusList))
	for i, dbGameStatus := range dbGameStatusList {
		gameStatusList[i] = FromVwGameStatus(dbGameStatus)
	}

	c.JSON(http.StatusOK, GameStatusList{Games: gameStatusList})
}
