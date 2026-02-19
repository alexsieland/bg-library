package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func (s Server) GetApiV1Health(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (s Server) CheckInGame(c *gin.Context) {
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

func (Server) GetGame(ctx *gin.Context) {
	resp := Game{
		GameId:    "776e4960-ce84-4b73-a71e-62839db0ecab",
		Title:     "Catan",
		CreatedAt: time.Now(),
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s Server) UpdateGame(c *gin.Context, gameId string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) ListGames(c *gin.Context) {
	//TODO implement me
	panic("implement me")
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
