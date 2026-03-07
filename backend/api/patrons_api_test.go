package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddPatron(t *testing.T) {
	t.Run("Should return 201 Created when valid patron is added", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "John Doe"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = name
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{name, pgtype.Text{Valid: false}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddPatronJSONRequestBody{Name: name})
		c.Request = httptest.NewRequest("POST", "/patrons", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPatron(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response Patron
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, name, response.Name)
		assert.Equal(t, patronID, response.PatronId)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when JSON is malformed", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons", bytes.NewBufferString("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPatron(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "JSON body is malformed")
	})

	t.Run("Should return 400 Bad Request when name is too long", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddPatronJSONRequestBody{Name: string(make([]byte, 101))})
		c.Request = httptest.NewRequest("POST", "/patrons", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPatron(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when DB returns error", func(t *testing.T) {
		server, mockDB := setupTestServer()
		name := "John Doe"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{name, pgtype.Text{Valid: false}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddPatronJSONRequestBody{Name: name})
		c.Request = httptest.NewRequest("POST", "/patrons", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPatron(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})

	t.Run("Should return 201 Created with barcode when barcode is provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "John Doe"
		barcode := "9780000000001"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = name
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(4).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
		}).Return(nil)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{name, pgtype.Text{String: barcode, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddPatronJSONRequestBody{Name: name, Barcode: &barcode})
		c.Request = httptest.NewRequest("POST", "/patrons", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPatron(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response Patron
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, name, response.Name)
		assert.Equal(t, patronID, response.PatronId)
		assert.NotNil(t, response.Barcode)
		assert.Equal(t, barcode, *response.Barcode)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when barcode exceeds 48 characters", func(t *testing.T) {
		server, _ := setupTestServer()
		name := "John Doe"
		barcode := string(make([]byte, 49)) // 49 chars

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(AddPatronJSONRequestBody{Name: name, Barcode: &barcode})
		c.Request = httptest.NewRequest("POST", "/patrons", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.AddPatron(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})
}

func TestDeletePatron(t *testing.T) {
	t.Run("Should return 204 No Content when patron is deleted", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/patrons/"+patronID.String(), nil)
		server.DeletePatron(c, patronID.String())

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when patronId is invalid UUID", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		server.DeletePatron(c, "invalid-uuid")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs on delete", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}}).Return(pgconn.CommandTag{}, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/patrons/"+patronID.String(), nil)
		server.DeletePatron(c, patronID.String())

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestGetPatron(t *testing.T) {
	t.Run("Should return 200 OK when patron is found", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "John Doe"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = name
			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // barcode
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/"+patronID.String(), nil)
		server.GetPatron(c, patronID.String())

		assert.Equal(t, http.StatusOK, w.Code)
		var response Patron
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, name, response.Name)
		assert.Equal(t, patronID, response.PatronId)
	})

	t.Run("Should return 404 Not Found when patron does not exist", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/"+patronID.String(), nil)
		server.GetPatron(c, patronID.String())

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Should return 400 Bad Request when patronId is invalid UUID", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		server.GetPatron(c, "invalid-uuid")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs on get", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/"+patronID.String(), nil)
		server.GetPatron(c, patronID.String())

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestGetPatronByBarcode(t *testing.T) {
	t.Run("Should return 200 OK when patron is found by barcode", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "John Doe"
		barcode := "1234567890"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = name
			*args.Get(2).(*pgtype.Text) = pgtype.Text{String: barcode, Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/barcode/"+barcode, nil)
		server.GetPatronByBarcode(c, barcode)

		assert.Equal(t, http.StatusOK, w.Code)
		var response Patron
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, name, response.Name)
		assert.Equal(t, patronID, response.PatronId)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 400 Bad Request when barcode is empty", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/barcode/", nil)
		server.GetPatronByBarcode(c, "")

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 400 Bad Request when barcode exceeds 48 characters", func(t *testing.T) {
		server, _ := setupTestServer()
		barcode := "1234567890123456789012345678901234567890123456789" // 49 chars

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/barcode/"+barcode, nil)
		server.GetPatronByBarcode(c, barcode)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Should return 404 Not Found when patron does not exist for barcode", func(t *testing.T) {
		server, mockDB := setupTestServer()
		barcode := "1234567890"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/barcode/"+barcode, nil)
		server.GetPatronByBarcode(c, barcode)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Should return 500 Internal Server Error when database error occurs", func(t *testing.T) {
		server, mockDB := setupTestServer()
		barcode := "1234567890"

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, []any{pgtype.Text{String: barcode, Valid: true}}).Return(mockRow)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons/barcode/"+barcode, nil)
		server.GetPatronByBarcode(c, barcode)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
		mockDB.AssertExpectations(t)
	})
}

func TestUpdatePatron(t *testing.T) {
	t.Run("Should return 204 No Content when patron is updated", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "Updated Name"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}, name, pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdatePatronJSONRequestBody{Name: name})
		c.Request = httptest.NewRequest("PUT", "/patrons/"+patronID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdatePatron(c, patronID.String())

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Should return 404 Not Found when updating non-existent patron", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "Updated Name"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}, name, pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, pgx.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdatePatronJSONRequestBody{Name: name})
		c.Request = httptest.NewRequest("PUT", "/patrons/"+patronID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdatePatron(c, patronID.String())

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Should return 400 Bad Request when patronId is invalid UUID", func(t *testing.T) {
		server, _ := setupTestServer()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdatePatronJSONRequestBody{Name: "Name"})
		c.Request = httptest.NewRequest("PUT", "/patrons/invalid-uuid", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdatePatron(c, "invalid-uuid")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation error")
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs on update", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "Updated Name"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}, name, pgtype.Text{Valid: false}}).Return(pgconn.CommandTag{}, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdatePatronJSONRequestBody{Name: name})
		c.Request = httptest.NewRequest("PUT", "/patrons/"+patronID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdatePatron(c, patronID.String())

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})

	t.Run("Should return 204 No Content when patron is updated with barcode", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "Updated Name"
		barcode := "9780000000001"

		mockDB.On("Exec", mock.Anything, mock.Anything, []any{pgtype.UUID{Bytes: patronID, Valid: true}, name, pgtype.Text{String: barcode, Valid: true}}).Return(pgconn.CommandTag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(UpdatePatronJSONRequestBody{Name: name, Barcode: &barcode})
		c.Request = httptest.NewRequest("PUT", "/patrons/"+patronID.String(), bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		server.UpdatePatron(c, patronID.String())

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockDB.AssertExpectations(t)
	})
}

func TestListPatrons(t *testing.T) {
	t.Run("Should return 200 OK with list of patrons when called without search", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "John Doe"

		mockRows := new(MockRows)
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = name
			*args.Get(2).(*pgtype.Text) = pgtype.Text{Valid: false} // barcode
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
		}).Return(nil)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(999), int32(0)}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons", nil)
		server.ListPatrons(c, ListPatronsParams{})

		assert.Equal(t, http.StatusOK, w.Code)
		var response PatronList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Patrons, 1)
		assert.Equal(t, name, response.Patrons[0].Name)
	})

	t.Run("Should return 200 OK with searched patrons when name is provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		name := "John"

		mockRows := new(MockRows)
		mockRows.On("Next").Return(false)
		mockRows.On("Close").Return()
		mockRows.On("Err").Return(nil)

		mockDB.On("Query", mock.Anything, mock.Anything, []any{"%" + name + "%", int32(999), int32(0)}).Return(mockRows, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons?name=John", nil)
		server.ListPatrons(c, ListPatronsParams{Name: &name})

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs on list", func(t *testing.T) {
		server, mockDB := setupTestServer()

		mockDB.On("Query", mock.Anything, mock.Anything, []any{int32(999), int32(0)}).Return(nil, errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/patrons", nil)
		server.ListPatrons(c, ListPatronsParams{})

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
	})
}

func TestBulkAddPatrons(t *testing.T) {
	t.Run("Should return 201 Created with imported count when valid CSV is provided", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID1 := uuid.New()
		patronID2 := uuid.New()
		name1 := "John Doe"
		name2 := "Jane Smith"

		// CSV content: "John Doe"\n"Jane Smith"
		csvContent := name1 + "\n" + name2
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// First patron
		mockRow1 := new(MockRow)
		mockRow1.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID1, Valid: true}
			*args.Get(1).(*string) = name1
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{name1, pgtype.Text{Valid: false}}).Return(mockRow1)

		// Second patron
		mockRow2 := new(MockRow)
		mockRow2.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID2, Valid: true}
			*args.Get(1).(*string) = name2
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{name2, pgtype.Text{Valid: false}}).Return(mockRow2)

		mockTx.On("Commit", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddPatrons(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response BulkAddResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 2, response.Imported)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should skip invalid records and continue processing when validation fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		validName := "John Doe"
		invalidName := "" // Empty name should fail validation

		csvContent := validName + "\n" + invalidName
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		// Only the valid patron
		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = validName
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{validName, pgtype.Text{Valid: false}}).Return(mockRow)

		mockTx.On("Commit", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddPatrons(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response BulkAddResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 1, response.Imported)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when transaction fails to begin", func(t *testing.T) {
		server, mockDB := setupTestServer()
		csvContent := "John Doe"
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(nil, errors.New("transaction error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddPatrons(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "transaction error")
		mockDB.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when DB error occurs during insert", func(t *testing.T) {
		server, mockDB := setupTestServer()
		name := "John Doe"
		csvContent := name
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{name, pgtype.Text{Valid: false}}).Return(mockRow)
		mockTx.On("Rollback", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddPatrons(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "db error")
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should return 500 Internal Server Error when transaction commit fails", func(t *testing.T) {
		server, mockDB := setupTestServer()
		patronID := uuid.New()
		name := "John Doe"
		csvContent := name
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)

		mockRow := new(MockRow)
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = pgtype.UUID{Bytes: patronID, Valid: true}
			*args.Get(1).(*string) = name
			*args.Get(2).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: true}
			*args.Get(3).(*pgtype.Timestamp) = pgtype.Timestamp{Valid: false} // deleted_at
			*args.Get(4).(*pgtype.Text) = pgtype.Text{Valid: false}           // barcode
		}).Return(nil)
		mockTx.On("QueryRow", mock.Anything, mock.Anything, []any{name, pgtype.Text{Valid: false}}).Return(mockRow)

		mockTx.On("Commit", mock.Anything).Return(errors.New("commit error"))
		mockTx.On("Rollback", mock.Anything).Return(nil).Maybe()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddPatrons(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "commit error")
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})

	t.Run("Should handle empty CSV and return 201 with zero imported", func(t *testing.T) {
		server, mockDB := setupTestServer()
		csvContent := ""
		base64Content := base64.StdEncoding.EncodeToString([]byte(csvContent))

		mockTx := new(MockTx)
		mockDB.On("BeginTx", mock.Anything, pgx.TxOptions{}).Return(mockTx, nil)
		mockTx.On("Commit", mock.Anything).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/patrons/bulk", bytes.NewBufferString(base64Content))
		c.Request.Header.Set("Content-Type", "text/plain")

		server.BulkAddPatrons(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		var response BulkAddResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, 0, response.Imported)
		mockDB.AssertExpectations(t)
		mockTx.AssertExpectations(t)
	})
}
