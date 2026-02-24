package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s Server) AddGame(c *gin.Context) {
	var jsonObject AddGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	errorDetails := ValidateStringLength("title", jsonObject.Title, 1, 100, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}

	dbGame, err := s.queries.CreateGame(c.Request.Context(), db.CreateGameParams{
		Title:          jsonObject.Title,
		SanitizedTitle: SanitizeTitle(jsonObject.Title),
	})
	if err != nil {
		log.Printf("Error creating game: %v", err)
		internalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, FromGame(dbGame))
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

func (s Server) UpdateGame(c *gin.Context, gameId string) {
	var jsonObject UpdateGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}
	errorDetails := ValidateStringLength("title", jsonObject.Title, 1, 100, []ErrorDetail{})
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId, errorDetails)
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}

	err = s.queries.EditGame(c.Request.Context(), db.EditGameParams{
		ID:             gameUUID,
		Title:          jsonObject.Title,
		SanitizedTitle: SanitizeTitle(jsonObject.Title),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

	gameList := make([]GameStatus, len(dbGameStatusList))
	for i, dbGameStatus := range dbGameStatusList {
		gameList[i] = FromVwGameStatus(dbGameStatus)
	}

	c.JSON(http.StatusOK, GameList{Games: gameList})
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

	gameList := make([]GameStatus, len(dbGameStatusList))
	for i, dbGameStatus := range dbGameStatusList {
		gameList[i] = FromVwGameStatus(dbGameStatus)
	}

	c.JSON(http.StatusOK, GameList{Games: gameList})
}
