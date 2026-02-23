package api

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Server struct {
	Database *db.LibraryDatabase
	queries  *db.Queries
}

func NewServer() Server {
	database := db.NewLibraryDatabase()
	return Server{
		Database: database,
		queries:  db.New(database),
	}
}

func internalError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalError(err))
}

func notFound(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(NOTFOUND, "Resource not found"))
}

func malformedJson(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(MALFORMEDREQUEST, "JSON body is malformed"))
}

func validationError(c *gin.Context, errorDetails []ErrorDetail) {
	c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponseWithDetails(VALIDATIONERROR, "Validation error", errorDetails))
}

func (s Server) GetApiV1Health(c *gin.Context) {
	_, err := s.Database.Exec(c.Request.Context(), "SELECT 1;")
	if err != nil {
		log.Printf("Error checking database health: %v", err)
		c.JSON(http.StatusServiceUnavailable, NewErrorResponse(SERVICEUNAVAILABLE, "Database is unavailable"))
		return
	}
	c.Status(http.StatusOK)
}

func (s Server) CheckInGame(c *gin.Context, params CheckInGameParams) {
	transactionUUID, errorDetails := ConvertToPgTypeUUID("TransactionId", params.TransactionId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
	}
	err := s.queries.CheckInGame(c.Request.Context(), transactionUUID)
	if err != nil {
		log.Printf("Error checking in game: %v", err)
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s Server) CheckOutGame(c *gin.Context) {
	var jsonObject CheckOutGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		//TODO setup validation error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", jsonObject.GameId.String(), []ErrorDetail{})
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", jsonObject.PatronId.String(), errorDetails)
	gameStatus, err := s.queries.GetGameStatus(c.Request.Context(), gameUUID)
	if err != nil {
		log.Printf("Error getting game status: %v", err)
		internalError(c, err)
		return
	}
	if !gameStatus.CheckinTimestamp.Valid && gameStatus.PatronID.Valid {
		if patronUUID == gameStatus.PatronID {
			//Game is already checked out, so we return the current status of the game
			c.JSON(http.StatusCreated, LibraryTransaction{
				GameId:    uuid.MustParse(gameStatus.GameID.String()),
				Id:        uuid.MustParse(gameStatus.TransactionID.String()),
				PatronId:  uuid.MustParse(gameStatus.PatronID.String()),
				Timestamp: gameStatus.CheckoutTimestamp.Time,
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Game is already checked out"})
		return
	}

	transaction, err := s.queries.CheckOutGame(c.Request.Context(), db.CheckOutGameParams{
		GameID:   gameUUID,
		PatronID: patronUUID,
	})
	if err != nil {
		log.Printf("Error checking out game: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, FromTransaction(transaction))
}

func (s Server) AddGame(c *gin.Context) {
	var jsonObject AddGameJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		//TODO setup validation error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if jsonObject.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title is required"})
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

	c.Status(http.StatusNoContent)
}

func (s Server) GetGame(c *gin.Context, gameId string) {
	gameUUID, errorDetails := ConvertToPgTypeUUID("GameId", gameId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	dbGame, err := s.queries.GetGame(c.Request.Context(), gameUUID)
	if err != nil {
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
		log.Printf("Error updating game: %v", err)
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
		dbGameStatusList, err = s.queries.SearchCheckedOutGames(c.Request.Context(), db.SearchCheckedOutGamesParams{
			SanitizedTitle: SanitizeTitle(*params.Title),
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
		dbGameStatusList, err = s.queries.SearchGameStatus(c.Request.Context(), db.SearchGameStatusParams{
			SanitizedTitle: SanitizeTitle(*params.Title),
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

func (s Server) AddPatron(c *gin.Context) {
	var jsonObject AddPatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}
	errorDetails := ValidateStringLength("name", jsonObject.Name, 1, 100, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	dbPatron, err := s.queries.CreatePatron(c.Request.Context(), jsonObject.Name)
	if err != nil {
		log.Printf("Error creating patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, FromPatron(dbPatron))
}

func (s Server) DeletePatron(c *gin.Context, patronId string) {
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", patronId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	err := s.queries.DeletePatron(c.Request.Context(), patronUUID)
	if err != nil {
		log.Printf("Error deleting patron: %v", err)
		internalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (s Server) GetPatron(c *gin.Context, patronId string) {
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", patronId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	dbPatron, err := s.queries.GetPatron(c.Request.Context(), patronUUID)
	if err != nil {
		log.Printf("Error getting patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, FromVwLibraryPatron(dbPatron))
}

func (s Server) UpdatePatron(c *gin.Context, patronId string) {
	var jsonObject UpdatePatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}
	errorDetails := ValidateStringLength("name", jsonObject.Name, 1, 100, []ErrorDetail{})
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", patronId, errorDetails)
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	err = s.queries.EditPatron(c.Request.Context(), db.EditPatronParams{
		ID:       patronUUID,
		FullName: jsonObject.Name,
	})
	if err != nil {
		log.Printf("Error updating patron: %v", err)
		internalError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s Server) ListPatrons(c *gin.Context, params ListPatronsParams) {
	var dbPatronList []db.VwLibraryPatron
	if params.Name == nil {
		var err error
		dbPatronList, err = s.queries.ListPatrons(c.Request.Context(), db.ListPatronsParams{
			Limit:  999,
			Offset: 0,
		})
		if err != nil {
			log.Printf("Error listing patrons: %v", err)
			internalError(c, err)
			return
		}
	} else {
		name := *params.Name
		var err error
		dbPatronList, err = s.queries.SearchPatrons(c.Request.Context(), db.SearchPatronsParams{
			FullName: name,
			Limit:    999,
			Offset:   0,
		})
		if err != nil {
			log.Printf("Error saerching patrons: %v", err)
			internalError(c, err)
			return
		}
	}

	patronList := make([]Patron, len(dbPatronList))
	for i, dbPatron := range dbPatronList {
		patronList[i] = FromVwLibraryPatron(dbPatron)
	}
	c.JSON(http.StatusOK, patronList)
}
