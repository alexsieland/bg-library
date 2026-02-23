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

func (s Server) GetApiV1Health(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (s Server) CheckInGame(c *gin.Context, params CheckInGameParams) {
	//TODO implement me
	panic("implement me")
}

func (s Server) CheckOutGame(c *gin.Context) {
	//TODO implement me
	panic("implement me")
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
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, FromGame(dbGame))
}

func (s Server) DeleteGame(c *gin.Context, gameId string) {
	err := s.queries.DeleteGame(c.Request.Context(), ConvertToPgTypeUUID(gameId))
	if err != nil {
		log.Printf("Error deleting game: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (Server) GetGame(c *gin.Context, gameId string) {
	gameUuid, _ := uuid.Parse(gameId)
	resp := Game{
		GameId: gameUuid,
		Title:  "Catan",
	}

	c.JSON(http.StatusOK, resp)
}

func (s Server) UpdateGame(c *gin.Context, gameId string) {
	var jsonObject UpdateGameJSONRequestBody
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

	err = s.queries.EditGame(c.Request.Context(), db.EditGameParams{
		ID:             ConvertToPgTypeUUID(gameId),
		Title:          jsonObject.Title,
		SanitizedTitle: SanitizeTitle(jsonObject.Title),
	})
	if err != nil {
		log.Printf("Error updating game: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s Server) ListGames(c *gin.Context) {
	dbGameStatusList, err := s.queries.ListGamesStatus(c.Request.Context(), db.ListGamesStatusParams{
		Limit:  999,
		Offset: 0,
	})

	if err != nil {
		log.Printf("Error listing games: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
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
		//TODO setup validation error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if jsonObject.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	dbPatron, err := s.queries.CreatePatron(c.Request.Context(), jsonObject.Name)
	if err != nil {
		log.Printf("Error creating patron: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, FromPatron(dbPatron))
}

func (s Server) DeletePatron(c *gin.Context, patronId string) {
	err := s.queries.DeletePatron(c.Request.Context(), ConvertToPgTypeUUID(patronId))
	if err != nil {
		log.Printf("Error deleting patron: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (s Server) GetPatron(c *gin.Context, patronId string) {
	dbPatron, err := s.queries.GetPatron(c.Request.Context(), ConvertToPgTypeUUID(patronId))
	if err != nil {
		log.Printf("Error getting patron: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, FromVwLibraryPatron(dbPatron))
}

func (s Server) UpdatePatron(c *gin.Context, patronId string) {
	var jsonObject UpdatePatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		//TODO setup validation error
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if jsonObject.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	err = s.queries.EditPatron(c.Request.Context(), db.EditPatronParams{
		ID:       ConvertToPgTypeUUID(patronId),
		FullName: jsonObject.Name,
	})
	if err != nil {
		log.Printf("Error updating patron: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (s Server) ListPatrons(c *gin.Context) {
	dbPatronList, err := s.queries.ListPatrons(c.Request.Context(), db.ListPatronsParams{
		Limit:  999,
		Offset: 0,
	})
	if err != nil {
		log.Printf("Error listing patrons: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
	}
	patronList := make([]Patron, len(dbPatronList))
	for i, dbPatron := range dbPatronList {
		patronList[i] = FromVwLibraryPatron(dbPatron)
	}
	c.JSON(http.StatusOK, patronList)
}
