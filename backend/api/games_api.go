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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

func (s Server) AddGame(c *gin.Context) {
	var jsonObject AddGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	var errorDetails ErrorDetails
	isPlayToWin := false
	if jsonObject.IsPlayToWin != nil {
		isPlayToWin = *jsonObject.IsPlayToWin
	}
	dbGame, err := s.insertGame(c, jsonObject.Title, jsonObject.Barcode, isPlayToWin, &errorDetails, nil)
	if errors.Is(err, errValidation) {
		validationError(c, errorDetails)
		return
	}
	if err != nil {
		log.Printf("Error creating game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, FromGame(dbGame, isPlayToWin))
}

func (s Server) insertGame(c *gin.Context, title string, barcode *string, isPlayToWin bool, errorDetails *ErrorDetails, tx *pgx.Tx) (db.Game, error) {
	errorDetails.ValidateStringLength("title", title, 1, 100)
	if barcode != nil {
		errorDetails.ValidateStringLength("barcode", *barcode, 1, 48)
	}
	if !errorDetails.Empty() {
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
		err = s.addPlayToWin(c, pgUUIDToUUID(game.ID), tx)
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
	var errorDetails ErrorDetails
	recordCount := int32(0)
	firstRow := true
	for {
		record, err := csvReader.Read()
		if firstRow {
			firstRow = false
			continue
		}
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
		errorDetails.ValidateStringLength("title", title, 1, 100)

		var barcode *string
		if len(record) > 1 && record[1] != "" {
			barcode = &record[1]
			errorDetails.ValidateStringLength("barcode", *barcode, 1, 48)
		}

		isPlayToWin := false
		if len(record) > 2 && record[2] == "true" {
			isPlayToWin = true
		}

		_, err = s.insertGame(c, title, barcode, isPlayToWin, &errorDetails, &tx)
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
	if !errorDetails.Empty() {
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

func (s Server) DeleteGame(c *gin.Context, gameId types.UUID) {
	err := s.queries.DeleteGame(c.Request.Context(), uuidToPgTypeUUID(gameId))
	if err != nil {
		log.Printf("Error deleting game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s Server) getGame(c *gin.Context, gameId types.UUID) (db.VwLibraryGame, error) {
	return s.queries.GetGame(c.Request.Context(), uuidToPgTypeUUID(gameId))
}

func (s Server) GetGame(c *gin.Context, gameId types.UUID) {
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
	c.JSON(http.StatusOK, FromVwLibraryGame(dbGame))
}

func (s Server) GetGameByBarcode(c *gin.Context, gameBarcode string) {
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("gameBarcode", gameBarcode, 1, 48)
	if !errorDetails.Empty() {
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

func (s Server) UpdateGame(c *gin.Context, gameId types.UUID) {
	// validate json body
	var jsonObject UpdateGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	// validate field values
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("title", jsonObject.Title, 1, 100)
	if jsonObject.Barcode != nil {
		errorDetails.ValidateStringLength("barcode", *jsonObject.Barcode, 1, 48)
	}
	if !errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}

	// get the current game entry to determine if it is currently a play to win game
	dbGame, err := s.getGame(c, gameId)
	if err != nil {
		// since edit will fail anyways if game is not found, we can safely return a client error here
		if isNotFound(err) {
			notFound(c)
			return
		}
		internalError(c, err)
		return
	}

	// start a db transaction
	tx, err := s.Database.BeginTx(c, pgx.TxOptions{})
	if err != nil {
		internalError(c, err)
		return
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(c.Request.Context())
		}
	}()

	// update the game entry
	err = s.queries.EditGame(c.Request.Context(), db.EditGameParams{
		ID:             uuidToPgTypeUUID(gameId),
		Title:          jsonObject.Title,
		SanitizedTitle: SanitizeTitle(jsonObject.Title),
		Barcode:        stringToPgText(jsonObject.Barcode),
	})
	if err != nil {
		// if FK violation, then this means the game id is invalid, so return 404
		if isForeignKeyConstraintViolation(err) {
			notFound(c)
			return
		}
		log.Printf("Error updating game: %v", err)
		internalError(c, err)
		return
	}

	// if edit included play to win status, then see if we need to add or remove the play to win game entry
	if jsonObject.IsPlayToWin != nil {
		if *jsonObject.IsPlayToWin == true && !dbGame.PlayToWinGameID.Valid {
			err = s.addPlayToWin(c, gameId, &tx)
			if err != nil {
				log.Printf("Error adding play to win game: %v", err)
				internalError(c, err)
				return
			}
		} else if *jsonObject.IsPlayToWin == false && dbGame.PlayToWinGameID.Valid {
			// If game was play to win, but is now not, then remove it with the reason 'mistake'
			reason := string(db.PlayToWinGameDeletionTypeMistake)
			err = s.removePlayToWin(c, gameId, &reason, nil, &tx)
			if err != nil {
				log.Printf("Error removing play to win game: %v", err)
				internalError(c, err)
				return
			}
		}
	}

	// commit the transaction
	err = tx.Commit(c.Request.Context())
	if err != nil {
		internalError(c, err)
		return
	}

	// prevent deferred rollback after a successful commit
	tx = nil

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
