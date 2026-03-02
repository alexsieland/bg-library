package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Connect() error
	Close()
	Exec(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, s string, i ...interface{}) pgx.Row
}

type Server struct {
	Database DB
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

func conflict(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusConflict, NewErrorResponse(CONFLICT, message))
}

func (s Server) GetHealth(c *gin.Context) {
	_, err := s.Database.Exec(c.Request.Context(), "SELECT 1;")
	if err != nil {
		log.Printf("Error checking database health: %v", err)
		c.JSON(http.StatusServiceUnavailable, NewErrorResponse(SERVICEUNAVAILABLE, "Database is unavailable"))
		return
	}
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

		// Get the server URL from environment variable, default to http://localhost:8080
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
