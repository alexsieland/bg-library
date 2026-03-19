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
	GameApi        *GameApi
	PlayToWinApi   *PlayToWinApi
}

func NewServer() Server {
	database := db.NewLibraryDatabase()
	var libService = internal.NewLibraryService(database)
	var gameSrv = internal.NewGameService(libService)
	var patronSrv = internal.NewPatronService(libService)
	var transSrv = internal.NewTransactionService(libService)
	var ptwSrv = internal.NewPlayToWinService(libService)
	transSrv.SetGameService(gameSrv)
	ptwSrv.SetGameService(gameSrv)
	gameSrv.SetPlayToWinService(ptwSrv)

	return Server{
		LibService:     libService,
		PatronApi:      NewPatronApi(libService, patronSrv),
		TransactionApi: NewTransactionApi(libService, transSrv),
		GameApi:        NewGamesApi(libService, gameSrv),
		PlayToWinApi:   NewPlayToWinApi(libService, ptwSrv),
	}
}

func (s *Server) GetHealth(c *gin.Context) {
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

func (s *Server) AddPatron(c *gin.Context) {
	var request AddPatronJSONRequestBody
	extractRequestBody[AddPatronJSONRequestBody](c, &request)
	if !c.IsAborted() {
		patron, err := s.PatronApi.AddPatron(c.Request.Context(), request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.JSON(http.StatusOK, patron)
	}
}

func (s *Server) GetPatron(c *gin.Context, patronId types.UUID) {
	patron, err := s.PatronApi.GetPatron(c.Request.Context(), patronId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, patron)
}

func (s *Server) GetPatronByBarcode(c *gin.Context, patronBarcode string) {
	patron, err := s.PatronApi.GetPatronByBarcode(c.Request.Context(), patronBarcode)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, patron)
}

func (s *Server) DeletePatron(c *gin.Context, patronId types.UUID) {
	err := s.PatronApi.DeletePatron(c.Request.Context(), patronId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) UpdatePatron(c *gin.Context, patronId types.UUID) {
	var request UpdatePatronJSONRequestBody
	extractRequestBody[UpdatePatronJSONRequestBody](c, &request)
	if !c.IsAborted() {
		err := s.PatronApi.UpdatePatron(c.Request.Context(), patronId, request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (s *Server) ListPatrons(c *gin.Context, params ListPatronsParams) {
	patronList, err := s.PatronApi.ListPatrons(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, patronList)
}

func (s *Server) BulkAddPatrons(c *gin.Context) {
	bulkAddResponse, err := s.PatronApi.BulkAddPatrons(c.Request.Context(), c.Request.Body)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, bulkAddResponse)
}

// Transaction API

func (s *Server) CheckInGame(c *gin.Context, params CheckInGameParams) {
	err := s.TransactionApi.CheckInGame(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) CheckOutGame(c *gin.Context) {
	var request CheckOutGameJSONRequestBody
	extractRequestBody[CheckOutGameJSONRequestBody](c, &request)
	if !c.IsAborted() {
		transaction, err := s.TransactionApi.CheckOutGame(c.Request.Context(), request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.JSON(http.StatusCreated, transaction)
	}
}

func (s *Server) ListTransactionEvents(c *gin.Context, params ListTransactionEventsParams) {
	transactionEvents, err := s.TransactionApi.ListTransactionEvents(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, transactionEvents)
}

// Game API

func (s *Server) AddGame(c *gin.Context) {
	var request CreateGameRequest
	extractRequestBody[CreateGameRequest](c, &request)
	if !c.IsAborted() {
		game, err := s.GameApi.AddGame(c.Request.Context(), request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.JSON(http.StatusCreated, game)
	}
}

func (s *Server) GetGameByBarcode(c *gin.Context, gameBarcode string) {
	game, err := s.GameApi.GetGameByBarcode(c.Request.Context(), gameBarcode)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, game)
}

func (s *Server) DeleteGame(c *gin.Context, gameId types.UUID) {
	err := s.GameApi.DeleteGame(c.Request.Context(), gameId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) GetGame(c *gin.Context, gameId types.UUID) {
	game, err := s.GameApi.GetGame(c.Request.Context(), gameId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, game)
}

func (s *Server) UpdateGame(c *gin.Context, gameId types.UUID) {
	var request CreateGameRequest
	extractRequestBody[CreateGameRequest](c, &request)
	if !c.IsAborted() {
		err := s.GameApi.UpdateGame(c.Request.Context(), gameId, request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (s *Server) ListGames(c *gin.Context, params ListGamesParams) {
	gameStatusList, err := s.GameApi.ListGames(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, gameStatusList)
}

func (s *Server) BulkAddGames(c *gin.Context) {
	bulkAddResponse, err := s.GameApi.BulkAddGames(c.Request.Context(), c.Request.Body)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, bulkAddResponse)
}

// Play To Win API

func (s *Server) GetPlayToWinGameEntries(c *gin.Context, playToWinId types.UUID) {
	ptwEntries, err := s.PlayToWinApi.GetPlayToWinGameEntries(c.Request.Context(), playToWinId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, ptwEntries)
}

func (s *Server) RemovePlayToWinGameByGameId(c *gin.Context, gameId types.UUID) {
	var request RemovePlayToWinGameRequest
	extractRequestBody[RemovePlayToWinGameRequest](c, &request)
	if !c.IsAborted() {
		err := s.PlayToWinApi.RemovePlayToWinGameByGameId(c.Request.Context(), gameId, request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (s *Server) AddPlayToWinGameByGameId(c *gin.Context, gameId types.UUID) {
	_, err := s.PlayToWinApi.AddPlayToWinGameByGameId(c.Request.Context(), gameId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Server) DeletePlayToWinGame(c *gin.Context, ptwId types.UUID) {
	var request RemovePlayToWinGameRequest
	extractRequestBody[RemovePlayToWinGameRequest](c, &request)
	if !c.IsAborted() {
		err := s.PlayToWinApi.DeletePlayToWinGame(c.Request.Context(), ptwId, request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
	}
}

func (s *Server) GetPlayToWinGame(c *gin.Context, ptwId types.UUID) {
	ptwGame, err := s.PlayToWinApi.GetPlayToWinGameOverview(c.Request.Context(), ptwId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, ptwGame)
}

func (s *Server) UpdatePlayToWinGame(c *gin.Context, ptwId types.UUID) {
	var request UpdatePlayToWinGameJSONRequestBody
	extractRequestBody[UpdatePlayToWinGameJSONRequestBody](c, &request)
	if !c.IsAborted() {
		err := s.PlayToWinApi.UpdatePlayToWinGame(c.Request.Context(), ptwId, request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func (s *Server) ListPlayToWinGames(c *gin.Context, params ListPlayToWinGamesParams) {
	ptwGames, err := s.PlayToWinApi.ListPlayToWinGames(c.Request.Context(), params)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, ptwGames)
}

func (s *Server) AddPlayToWinSession(c *gin.Context) {
	var request CreatePlayToWinSessionRequest
	extractRequestBody[CreatePlayToWinSessionRequest](c, &request)
	if !c.IsAborted() {
		ptwSession, err := s.PlayToWinApi.RecordPlayToWinSession(c.Request.Context(), request)
		handleError(c, err)
		if c.IsAborted() {
			return
		}
		c.JSON(http.StatusCreated, ptwSession)
	}
}

// Play To Win Raffle API

func (s *Server) DrawPlayToWinRaffle(c *gin.Context, ptwId types.UUID) {
	ptwEntry, err := s.PlayToWinApi.DrawPlayToWinRaffle(c.Request.Context(), ptwId)
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.JSON(http.StatusOK, ptwEntry)
}

func (s *Server) ResetPlayToWinRaffle(c *gin.Context) {
	err := s.PlayToWinApi.ResetPlayToWinRaffle(c.Request.Context())
	handleError(c, err)
	if c.IsAborted() {
		return
	}
	c.Status(http.StatusNoContent)
}
