package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alexsieland/bg-library/internal"
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

func TestErrorHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 500 Internal Server Error when internalError helper is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		internalError(c, errors.New("test error"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "test error")
	})

	t.Run("Should return 404 Not Found when notFound helper is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		notFound(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Resource not found")
	})

	t.Run("Should return 400 Bad Request when malformedJson helper is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		malformedJson(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 400 Bad Request when validationError helper is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		details := []ErrorDetail{{Field: "test", Message: "invalid"}}
		validationError(c, ErrorDetails{details})
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
		assert.Contains(t, w.Body.String(), "invalid")
	})

	t.Run("Should return 409 Conflict when conflict helper is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		conflict(c, "conflict message")
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "conflict message")
	})

	t.Run("Should return 400 Bad Request when badRequest helper is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		badRequest(c, "bad request message")
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "bad request message")
	})
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should handle nil error by returning nil", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handleError(c, nil)
		// Should not abort or return any error status
		assert.NotEqual(t, http.StatusBadRequest, w.Code)
		assert.NotEqual(t, http.StatusNotFound, w.Code)
		assert.NotEqual(t, http.StatusConflict, w.Code)
		assert.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Should return 404 Not Found when ErrNotFound is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handleError(c, internal.ErrNotFound)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Resource not found")
	})

	t.Run("Should return 409 Conflict when ErrAlreadyExists is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handleError(c, internal.ErrAlreadyExists)
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Resource already exists")
	})

	t.Run("Should return 400 Bad Request when ErrInvalidInput is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handleError(c, internal.ErrInvalidInput)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})

	t.Run("Should return 400 Bad Request when ErrorDetails is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		details := []ErrorDetail{{Field: "name", Message: "required"}}
		errorDetails := ErrorDetails{details}
		handleError(c, errorDetails)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error for unknown error types", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		handleError(c, errors.New("unknown error"))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "unknown error")
	})
}

func TestExtractRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestRequest struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("Should successfully extract valid JSON request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := TestRequest{Name: "John", Age: 30}
		body, _ := json.Marshal(requestBody)
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestRequest
		extractRequestBody(c, &result)

		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("Should return 400 Bad Request when JSON is malformed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestRequest
		extractRequestBody(c, &result)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 400 Bad Request when request body is empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(""))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestRequest
		extractRequestBody(c, &result)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should partially unmarshal valid JSON with missing fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := TestRequest{Name: "Alice"} // Age is omitted
		body, _ := json.Marshal(requestBody)
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestRequest
		extractRequestBody(c, &result)

		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 0, result.Age) // Default int value
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
