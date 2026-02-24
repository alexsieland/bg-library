package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type MockDB struct {
	ConnectFunc  func() error
	CloseFunc    func()
	ExecFunc     func(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error)
	QueryFunc    func(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error)
	QueryRowFunc func(ctx context.Context, s string, i ...interface{}) pgx.Row
}

func (m *MockDB) Connect() error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc()
	}
	return nil
}

func (m *MockDB) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func (m *MockDB) Exec(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
	return m.ExecFunc(ctx, s, i...)
}

func (m *MockDB) Query(ctx context.Context, s string, i ...interface{}) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, s, i...)
	}
	return nil, nil
}

func (m *MockDB) QueryRow(ctx context.Context, s string, i ...interface{}) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, s, i...)
	}
	return nil
}

func TestGetHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should return 200 OK when the database is healthy", func(t *testing.T) {
		mockDB := &MockDB{
			ExecFunc: func(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		}
		server := Server{Database: mockDB}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		server.GetHealth(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Should return 503 Service Unavailable when the database returns an error", func(t *testing.T) {
		mockDB := &MockDB{
			ExecFunc: func(ctx context.Context, s string, i ...interface{}) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("db error")
			},
		}
		server := Server{Database: mockDB}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		server.GetHealth(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "Database is unavailable")
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
		validationError(c, details)
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
		assert.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
	})
}
