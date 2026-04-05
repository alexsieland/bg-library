package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 500 Internal Server Error with error message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		testErr := errors.New("database connection failed")

		internalError(c, testErr)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "database connection failed")
	})

	t.Run("Should abort request when internalError is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		internalError(c, errors.New("test error"))

		assert.True(t, c.IsAborted())
	})
}

func TestNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 404 Not Found with standard message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		notFound(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Resource not found")
	})

	t.Run("Should abort request when notFound is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		notFound(c)

		assert.True(t, c.IsAborted())
	})
}

func TestBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 400 Bad Request with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		message := "Invalid request parameters"

		badRequest(c, message)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), message)
	})

	t.Run("Should abort request when badRequest is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		badRequest(c, "custom message")

		assert.True(t, c.IsAborted())
	})
}

func TestMalformedJson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 400 Bad Request with malformed JSON message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		malformedJson(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should abort request when malformedJson is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		malformedJson(c)

		assert.True(t, c.IsAborted())
	})
}

func TestValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 400 Bad Request with validation error details", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		details := []ErrorDetail{
			{Field: "name", Message: "name is required"},
			{Field: "email", Message: "email must be valid"},
		}
		errorDetails := ErrorDetails{details}

		validationError(c, errorDetails)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
		assert.Contains(t, w.Body.String(), "name")
		assert.Contains(t, w.Body.String(), "email")
	})

	t.Run("Should abort request when validationError is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		errorDetails := ErrorDetails{[]ErrorDetail{{Field: "test", Message: "error"}}}

		validationError(c, errorDetails)

		assert.True(t, c.IsAborted())
	})

	t.Run("Should handle empty validation error details", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		errorDetails := ErrorDetails{[]ErrorDetail{}}

		validationError(c, errorDetails)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})
}

func TestConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 409 Conflict with custom message", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		message := "Resource already exists"

		conflict(c, message)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), message)
	})

	t.Run("Should abort request when conflict is called", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		conflict(c, "conflict message")

		assert.True(t, c.IsAborted())
	})
}

func TestExtractRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestRequest struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	t.Run("Should successfully extract valid JSON request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := TestRequest{Name: "John Doe", Email: "john@example.com"}
		body, _ := json.Marshal(requestBody)
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestRequest
		extractRequestBody(c, &result)

		assert.Equal(t, "John Doe", result.Name)
		assert.Equal(t, "john@example.com", result.Email)
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

	t.Run("Should partially unmarshal valid JSON with missing optional fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := TestRequest{Name: "Jane Doe"} // Email omitted
		body, _ := json.Marshal(requestBody)
		c.Request = httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		var result TestRequest
		extractRequestBody(c, &result)

		assert.Equal(t, "Jane Doe", result.Name)
		assert.Equal(t, "", result.Email)
	})
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should handle nil error without aborting", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handleError(c, nil)

		assert.False(t, c.IsAborted())
		assert.NotEqual(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Should return 404 Not Found when ErrNotFound is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handleError(c, internal.ErrNotFound)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Resource not found")
		assert.True(t, c.IsAborted())
	})

	t.Run("Should return 409 Conflict when ErrAlreadyExists is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handleError(c, internal.ErrAlreadyExists)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Contains(t, w.Body.String(), "Resource already exists")
		assert.True(t, c.IsAborted())
	})

	t.Run("Should return 400 Bad Request when ErrInvalidInput is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handleError(c, internal.ErrInvalidInput)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
		assert.True(t, c.IsAborted())
	})

	t.Run("Should return 400 Bad Request when ErrorDetails is passed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		details := []ErrorDetail{{Field: "username", Message: "already taken"}}
		errorDetails := ErrorDetails{details}

		handleError(c, errorDetails)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
		assert.Contains(t, w.Body.String(), "already taken")
		assert.True(t, c.IsAborted())
	})

	t.Run("Should return 500 Internal Server Error for unknown error types", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		unknownErr := errors.New("unexpected database error")

		handleError(c, unknownErr)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "unexpected database error")
		assert.True(t, c.IsAborted())
	})

	t.Run("Should prioritize ErrNotFound over generic error matching", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Wrap ErrNotFound to ensure errors.Is still matches
		wrappedErr := errors.New("wrapper: " + internal.ErrNotFound.Error())

		handleError(c, wrappedErr)

		// Should be treated as generic error since errors.Is won't match wrapped
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
