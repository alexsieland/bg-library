package api

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
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
	r.StaticFS("/swagger", http.Dir(filepath.Join("..", "swagger")))

	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}
