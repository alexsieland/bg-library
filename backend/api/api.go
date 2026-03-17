package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
)

type Server struct {
	LibService     *internal.LibraryService
	PatronApi      *PatronApi
	TransactionApi *TransactionApi
}

func NewServer() Server {
	database := db.NewLibraryDatabase()
	var libService = internal.NewLibraryService(database)

	return Server{
		LibService:     libService,
		PatronApi:      NewPatronApi(libService),
		TransactionApi: NewTransactionApi(libService),
	}
}

func (s Server) GetHealth(c *gin.Context) {
	c.Status(http.StatusOK)
}

func RegisterSwagger(r *gin.Engine) {
	swaggerDir := filepath.Join("..", "swagger")
	if os.Getenv("IS_DOCKER") == "true" {
		swaggerDir = "swagger"
	}

	// Serve api.yaml with dynamic server URL
	r.GET("/swagger/api.yaml", func(c *gin.Context) {
		swaggerFile := filepath.Join(swaggerDir, "api.yaml")
		content, err := os.ReadFile(swaggerFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewInternalError(err))
			return
		}

		// Get the server URL from the environment variable, default to http://localhost:8080
		serverURL := os.Getenv("API_URL")
		if serverURL == "" {
			serverURL = "http://localhost:8080"
		}

		// Replace the placeholder in the YAML
		yamlContent := string(content)
		yamlContent = strings.ReplaceAll(yamlContent, "${API_URL}", serverURL)

		c.Header("Content-Type", "application/yaml")
		c.String(http.StatusOK, yamlContent)
	})

	// Serve the index.html file itself from disk so relative URLs inside it resolve correctly
	r.GET("/swagger/", func(c *gin.Context) {
		indexPath := filepath.Join(swaggerDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		} else {
			c.JSON(http.StatusInternalServerError, NewInternalError(err))
		}
	})
}

// Patron API

func (s Server) AddPatron(c *gin.Context) {
	var request AddPatronJSONRequestBody
	extractRequestBody[AddPatronJSONRequestBody](c, request)
	if !c.IsAborted() {
		patron, err := s.PatronApi.AddPatron(c.Request.Context(), request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.JSON(http.StatusOK, patron)
	}
}

func (s Server) GetPatron(c *gin.Context, patronId types.UUID) {
	patron, err := s.PatronApi.GetPatron(c.Request.Context(), patronId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, patron)
}

func (s Server) GetPatronByBarcode(c *gin.Context, patronBarcode string) {
	patron, err := s.PatronApi.GetPatronByBarcode(c.Request.Context(), patronBarcode)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, patron)
}

func (s Server) DeletePatron(c *gin.Context, patronId types.UUID) {
	err := s.PatronApi.DeletePatron(c.Request.Context(), patronId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}

func (s Server) UpdatePatron(c *gin.Context, patronId types.UUID) {
	var request UpdatePatronJSONRequestBody
	extractRequestBody[UpdatePatronJSONRequestBody](c, request)
	if !c.IsAborted() {
		err := s.PatronApi.UpdatePatron(c.Request.Context(), patronId, request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (s Server) ListPatrons(c *gin.Context, params ListPatronsParams) {
	patronList, err := s.PatronApi.ListPatrons(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, patronList)
}

func (s Server) BulkAddPatrons(c *gin.Context) {
	bulkAddResponse, err := s.PatronApi.BulkAddPatrons(c.Request.Context(), c.Request.Body)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, bulkAddResponse)
}

// Transaction API

func (s Server) CheckInGame(c *gin.Context, params CheckInGameParams) {
	err := s.TransactionApi.CheckInGame(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}

func (s Server) CheckOutGame(c *gin.Context) {
	var request CheckOutGameJSONRequestBody
	extractRequestBody[CheckOutGameJSONRequestBody](c, request)
	if !c.IsAborted() {
		transaction, err := s.TransactionApi.CheckOutGame(c.Request.Context(), request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.JSON(http.StatusCreated, transaction)
	}
}

func (s Server) ListTransactionEvents(c *gin.Context, params ListTransactionEventsParams) {
	transactionEvents, err := s.TransactionApi.ListTransactionEvents(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, transactionEvents)
}

// Game API

func (s Server) AddGame(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetGameByBarcode(c *gin.Context, gameBarcode string) {
	//TODO implement me
	panic("implement me")
}

func (s Server) DeleteGame(c *gin.Context, gameId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetGame(c *gin.Context, gameId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) UpdateGame(c *gin.Context, gameId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) ListGames(c *gin.Context, params ListGamesParams) {
	//TODO implement me
	panic("implement me")
}

func (s Server) BulkAddGames(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

// Play To Win API

func (s Server) GetPlayToWinGameEntries(c *gin.Context, playToWinId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) RemovePlayToWinGameByGameId(c *gin.Context, gameId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) AddPlayToWinGameByGameId(c *gin.Context, gameId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) DeletePlayToWinGame(c *gin.Context, ptwId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) GetPlayToWinGame(c *gin.Context, ptwId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) UpdatePlayToWinGame(c *gin.Context, ptwId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) ListPlayToWinGames(c *gin.Context, params ListPlayToWinGamesParams) {
	//TODO implement me
	panic("implement me")
}

func (s Server) AddPlayToWinSession(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

// Play To Win Raffle API

func (s Server) DrawPlayToWinRaffle(c *gin.Context, ptwId types.UUID) {
	//TODO implement me
	panic("implement me")
}

func (s Server) ResetPlayToWinRaffle(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
