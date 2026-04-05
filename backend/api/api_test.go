package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	t.Run("Should return 200 OK", func(t *testing.T) {
		server, _, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		server.GetHealth(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRegisterSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a temp directory to simulate swagger folder
	tmpDir, err := os.MkdirTemp("", "swagger-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// In api.go: r.StaticFS("/swagger", http.Dir(filepath.Join("..", "swagger")))
	// This is tricky to test because it depends on the directory structure.
	// We can at least test the redirect.

	RegisterSwagger(r)

	t.Run("Should redirect to index.html when /swagger is accessed", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/swagger", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMovedPermanently, w.Code)
		assert.Equal(t, "/swagger/", w.Header().Get("Location"))
	})
}
