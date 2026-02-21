package api

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Server struct {
	database *db.LibraryDatabase
	queries  *db.Queries
}

func NewServer() Server {
	d := db.NewLibraryDatabase()
	return Server{
		database: &d,
		queries:  db.New(d),
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
	dbGameStatusList, error := s.queries.ListGamesStatus(c.Request.Context(), db.ListGamesStatusParams{
		Limit:  999,
		Offset: 0,
	})

	if error != nil {
		log.Printf("Error listing games: %v", error)
		c.AbortWithError(http.StatusInternalServerError, error)
	}

	gameList := make([]GameStatus, len(dbGameStatusList))
	for i, dbGameStatus := range dbGameStatusList {
		gameList[i] = ConvertToOpenAPIGameStatus(dbGameStatus)
	}

	c.JSON(http.StatusOK, GameList{Games: gameList})
}

func (s Server) AddPatron(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (s Server) DeletePatron(c *gin.Context, patronId string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetPatron(c *gin.Context, patronId string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) UpdatePatron(c *gin.Context, patronId string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) ListPatrons(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
