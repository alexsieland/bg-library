package api

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	//TODO implement me
	panic("implement me")
}

func (s Server) DeleteGame(c *gin.Context, gameId string) {
	//TODO implement me
	panic("implement me")
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
	//TODO implement me
	panic("implement me")
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
		gameList[i] = ConvertToOpenAPIGameStatus(dbGameStatus)
	}

	c.JSON(http.StatusOK, GameList{Games: gameList})
}

func (s Server) AddPatron(c *gin.Context) {
	dbPatron, err := s.queries.CreatePatron(c.Request.Context(), c.Params.ByName("name"))
	if err != nil {
		log.Printf("Error creating patron: %v", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, ConvertToOpenAPIPatron(dbPatron))
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

	c.JSON(http.StatusOK, ConvertToOpenAPIPatron(dbPatron))
}

func (s Server) UpdatePatron(c *gin.Context, patronId string) {
	//TODO implement me
	panic("implement me")
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
		patronList[i] = ConvertToOpenAPIPatron(dbPatron)
	}
	c.JSON(http.StatusOK, patronList)
}
