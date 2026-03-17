package api

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPatronApiAddPatron(t *testing.T) {
	t.Run("Should return a patron when the request is valid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		expectedPatron := testDBPatron(uuid.New(), "Jane Doe", ptr("P-1001"))

		fixture.service.On("InsertPatron", fixture.ctx, "Jane Doe", ptr("P-1001"), nil).Return(expectedPatron, nil).Once()

		patron, err := fixture.api.AddPatron(fixture.ctx, AddPatronJSONRequestBody{Name: "Jane Doe", Barcode: ptr("P-1001")})

		assert.NoError(t, err)
		assert.Equal(t, FromPatron(expectedPatron), patron)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the request fields are invalid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		tooLongName := strings.Repeat("n", 101)
		tooLongBarcode := strings.Repeat("b", 49)

		patron, err := fixture.api.AddPatron(fixture.ctx, AddPatronJSONRequestBody{Name: tooLongName, Barcode: &tooLongBarcode})

		assert.Equal(t, Patron{}, patron)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{
			{Field: "name", Message: "Length must be between 1 and 100"},
			{Field: "barcode", Message: "Length must be between 1 and 48"},
		}})
	})

	t.Run("Should return the service error when inserting a patron fails", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		expectedErr := errors.New("insert failed")

		fixture.service.On("InsertPatron", fixture.ctx, "Jane Doe", (*string)(nil), nil).Return(db.Patron{}, expectedErr).Once()

		patron, err := fixture.api.AddPatron(fixture.ctx, AddPatronJSONRequestBody{Name: "Jane Doe"})

		assert.Equal(t, Patron{}, patron)
		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

func TestPatronApiBulkAddPatrons(t *testing.T) {
	t.Run("Should import patrons and commit the transaction when the CSV is valid", func(t *testing.T) {
		tx := &stubTx{}
		fixture := newPatronApiTestFixture(t).withTx(tx).build()

		fixture.service.On("InsertPatron", fixture.ctx, "Jane Doe", ptr("P-2001"), mock.Anything).Return(testDBPatron(uuid.New(), "Jane Doe", ptr("P-2001")), nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})
		fixture.service.On("InsertPatron", fixture.ctx, "John Smith", (*string)(nil), mock.Anything).Return(testDBPatron(uuid.New(), "John Smith", nil), nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})

		response, err := fixture.api.BulkAddPatrons(fixture.ctx, encodedCSVBody("name,barcode\nJane Doe,P-2001\nJohn Smith,\n"))

		assert.NoError(t, err)
		assert.Equal(t, BulkAddResponse{Imported: 2}, response)
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 0, tx.rollbackCount)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return the transaction start error when opening the bulk import transaction fails", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).withTxError(errors.New("begin failed")).build()

		response, err := fixture.api.BulkAddPatrons(fixture.ctx, encodedCSVBody("name,barcode\nJane Doe,P-2001\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.EqualError(t, err, "begin failed")
	})

	t.Run("Should return the CSV read error when the request body is not valid base64", func(t *testing.T) {
		tx := &stubTx{}
		fixture := newPatronApiTestFixture(t).withTx(tx).build()

		response, err := fixture.api.BulkAddPatrons(fixture.ctx, io.NopCloser(strings.NewReader("%%%")))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.Error(t, err)
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		fixture.service.AssertNotCalled(t, "InsertPatron", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Should accumulate validation details and skip inserts when any CSV row is invalid", func(t *testing.T) {
		tx := &stubTx{}
		fixture := newPatronApiTestFixture(t).withTx(tx).build()
		tooLongBarcode := strings.Repeat("b", 49)

		response, err := fixture.api.BulkAddPatrons(fixture.ctx, encodedCSVBody("name,barcode\n,P-2002\nValid Name,"+tooLongBarcode+"\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{
			{Field: "name", Message: "Length must be between 1 and 100"},
			{Field: "barcode", Message: "Length must be between 1 and 48"},
		}})
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		fixture.service.AssertNotCalled(t, "InsertPatron", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Should return the service error and roll back when an insert fails during bulk import", func(t *testing.T) {
		tx := &stubTx{}
		fixture := newPatronApiTestFixture(t).withTx(tx).build()
		expectedErr := errors.New("insert failed")

		fixture.service.On("InsertPatron", fixture.ctx, "Jane Doe", (*string)(nil), mock.Anything).Return(db.Patron{}, expectedErr).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})

		response, err := fixture.api.BulkAddPatrons(fixture.ctx, encodedCSVBody("name,barcode\nJane Doe,\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return the commit error and roll back when committing the bulk import transaction fails", func(t *testing.T) {
		tx := &stubTx{commitErr: errors.New("commit failed")}
		fixture := newPatronApiTestFixture(t).withTx(tx).build()

		fixture.service.On("InsertPatron", fixture.ctx, "Jane Doe", (*string)(nil), mock.Anything).Return(testDBPatron(uuid.New(), "Jane Doe", nil), nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})

		response, err := fixture.api.BulkAddPatrons(fixture.ctx, encodedCSVBody("name,barcode\nJane Doe,\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.EqualError(t, err, "commit failed")
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		fixture.service.AssertExpectations(t)
	})
}

func TestPatronApiDeletePatron(t *testing.T) {
	t.Run("Should forward the converted patron identifier when deleting a patron", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		patronID := uuid.New()

		fixture.service.On("DeletePatron", fixture.ctx, testUUID(patronID), nil).Return(nil).Once()

		err := fixture.api.DeletePatron(fixture.ctx, types.UUID(patronID))

		assert.NoError(t, err)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return the service error when deleting a patron fails", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		patronID := uuid.New()
		expectedErr := errors.New("delete failed")

		fixture.service.On("DeletePatron", fixture.ctx, testUUID(patronID), nil).Return(expectedErr).Once()

		err := fixture.api.DeletePatron(fixture.ctx, types.UUID(patronID))

		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

func TestPatronApiGetPatron(t *testing.T) {
	t.Run("Should return a converted patron when the service finds one", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		patronID := uuid.New()
		barcode := "P-3001"
		expectedPatron := testDBLibraryPatron(patronID, "Jordan Doe", ptr(barcode))

		fixture.service.On("GetPatron", fixture.ctx, testUUID(patronID), nil).Return(expectedPatron, nil).Once()

		patron, err := fixture.api.GetPatron(fixture.ctx, types.UUID(patronID))

		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryPatron(expectedPatron), patron)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return the service error when fetching a patron fails", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		patronID := uuid.New()
		expectedErr := errors.New("lookup failed")

		fixture.service.On("GetPatron", fixture.ctx, testUUID(patronID), nil).Return(db.VwLibraryPatron{}, expectedErr).Once()

		patron, err := fixture.api.GetPatron(fixture.ctx, types.UUID(patronID))

		assert.Equal(t, Patron{}, patron)
		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

func TestPatronApiGetPatronByBarcode(t *testing.T) {
	t.Run("Should return a converted patron when the barcode is valid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		barcode := "P-4001"
		expectedPatron := testDBLibraryPatron(uuid.New(), "Taylor Doe", ptr(barcode))

		fixture.service.On("GetPatronByBarcode", fixture.ctx, barcode, nil).Return(expectedPatron, nil).Once()

		patron, err := fixture.api.GetPatronByBarcode(fixture.ctx, barcode)

		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryPatron(expectedPatron), patron)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the barcode is invalid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		tooLongBarcode := strings.Repeat("b", 49)

		patron, err := fixture.api.GetPatronByBarcode(fixture.ctx, tooLongBarcode)

		assert.Equal(t, Patron{}, patron)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "patronBarcode", Message: "Length must be between 1 and 48"}}})
	})
}

func TestPatronApiUpdatePatron(t *testing.T) {
	t.Run("Should forward the update request to the service when the request is valid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		patronID := uuid.New()
		barcode := "P-5001"

		fixture.service.On("UpdatePatron", fixture.ctx, testUUID(patronID), "Updated Name", ptr(barcode), nil).Return(nil).Once()

		err := fixture.api.UpdatePatron(fixture.ctx, types.UUID(patronID), UpdatePatronJSONRequestBody{Name: "Updated Name", Barcode: ptr(barcode)})

		assert.NoError(t, err)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the update request is invalid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		tooLongName := strings.Repeat("n", 101)

		err := fixture.api.UpdatePatron(fixture.ctx, types.UUID(uuid.New()), UpdatePatronJSONRequestBody{Name: tooLongName})

		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "name", Message: "Length must be between 1 and 100"}}})
	})

	t.Run("Should return the service error when updating a patron fails", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		patronID := uuid.New()
		expectedErr := errors.New("update failed")

		fixture.service.On("UpdatePatron", fixture.ctx, testUUID(patronID), "Updated Name", (*string)(nil), nil).Return(expectedErr).Once()

		err := fixture.api.UpdatePatron(fixture.ctx, types.UUID(patronID), UpdatePatronJSONRequestBody{Name: "Updated Name"})

		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

func TestPatronApiListPatrons(t *testing.T) {
	t.Run("Should return a converted patron list when the list request is valid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		searchName := "Doe"
		barcode := "P-6001"
		expectedPatrons := []db.VwLibraryPatron{
			testDBLibraryPatron(uuid.New(), "Casey Doe", ptr(barcode)),
			testDBLibraryPatron(uuid.New(), "Logan Doe", nil),
		}

		fixture.service.On("ListPatrons", fixture.ctx, ptr(searchName), int32(999), int32(0), nil).Return(expectedPatrons, nil).Once()

		patronList, err := fixture.api.ListPatrons(fixture.ctx, ListPatronsParams{Name: ptr(searchName)})

		assert.NoError(t, err)
		assert.Equal(t, PatronList{Patrons: []Patron{
			FromVwLibraryPatron(expectedPatrons[0]),
			FromVwLibraryPatron(expectedPatrons[1]),
		}}, patronList)
		fixture.service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the list search name is invalid", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		tooLongName := strings.Repeat("n", 101)

		patronList, err := fixture.api.ListPatrons(fixture.ctx, ListPatronsParams{Name: &tooLongName})

		assert.Equal(t, PatronList{}, patronList)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "name", Message: "Length must be between 1 and 100"}}})
	})

	t.Run("Should return the service error when listing patrons fails", func(t *testing.T) {
		fixture := newPatronApiTestFixture(t).build()
		expectedErr := errors.New("list failed")

		fixture.service.On("ListPatrons", fixture.ctx, (*string)(nil), int32(999), int32(0), nil).Return(nil, expectedErr).Once()

		patronList, err := fixture.api.ListPatrons(fixture.ctx, ListPatronsParams{})

		assert.Equal(t, PatronList{}, patronList)
		assert.ErrorIs(t, err, expectedErr)
		fixture.service.AssertExpectations(t)
	})
}

type patronApiTestFixture struct {
	ctx     context.Context
	service *mockPatronService
	api     *PatronApi
	tx      pgx.Tx
	txErr   error
}

func newPatronApiTestFixture(t *testing.T) *patronApiTestFixture {
	return &patronApiTestFixture{
		ctx:     t.Context(),
		service: new(mockPatronService),
	}
}

func (f *patronApiTestFixture) withTx(tx pgx.Tx) *patronApiTestFixture {
	f.tx = tx
	return f
}

func (f *patronApiTestFixture) withTxError(err error) *patronApiTestFixture {
	f.txErr = err
	return f
}

func (f *patronApiTestFixture) build() *patronApiTestFixture {
	f.api = newTestPatronApi(f.service, f.tx, f.txErr)
	return f
}

// ptr creates a pointer to a value, useful for creating pointers to literals or constants
func ptr[T any](v T) *T {
	return &v
}

// testUUID converts a uuid.UUID to pgtype.UUID for testing
func testUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

type mockPatronService struct {
	mock.Mock
}

func (m *mockPatronService) InsertPatron(ctx context.Context, name string, barcode *string, optTx pgx.Tx) (db.Patron, error) {
	args := m.Called(ctx, name, barcode, optTx)
	if args.Get(0) == nil {
		return db.Patron{}, args.Error(1)
	}
	return args.Get(0).(db.Patron), args.Error(1)
}

func (m *mockPatronService) DeletePatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) error {
	args := m.Called(ctx, patronId, optTx)
	return args.Error(0)
}

func (m *mockPatronService) GetPatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryPatron, error) {
	args := m.Called(ctx, patronId, optTx)
	if args.Get(0) == nil {
		return db.VwLibraryPatron{}, args.Error(1)
	}
	return args.Get(0).(db.VwLibraryPatron), args.Error(1)
}

func (m *mockPatronService) GetPatronByBarcode(ctx context.Context, patronBarcode string, optTx pgx.Tx) (db.VwLibraryPatron, error) {
	args := m.Called(ctx, patronBarcode, optTx)
	if args.Get(0) == nil {
		return db.VwLibraryPatron{}, args.Error(1)
	}
	return args.Get(0).(db.VwLibraryPatron), args.Error(1)
}

func (m *mockPatronService) UpdatePatron(ctx context.Context, patronId pgtype.UUID, fullName string, barcode *string, optTx pgx.Tx) error {
	args := m.Called(ctx, patronId, fullName, barcode, optTx)
	return args.Error(0)
}

func (m *mockPatronService) ListPatrons(ctx context.Context, fullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryPatron, error) {
	args := m.Called(ctx, fullName, limit, offset, optTx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.VwLibraryPatron), args.Error(1)
}

func newTestPatronApi(service patronService, tx pgx.Tx, beginErr error) *PatronApi {
	return &PatronApi{
		service: service,
		beginTx: func(context.Context) (pgx.Tx, error) {
			return tx, beginErr
		},
	}
}

func encodedCSVBody(csvContent string) io.ReadCloser {
	encoded := base64.StdEncoding.EncodeToString([]byte(csvContent))
	return io.NopCloser(strings.NewReader(encoded))
}

func testDBPatron(id uuid.UUID, fullName string, barcode *string) db.Patron {
	return db.Patron{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		FullName:  fullName,
		Barcode:   testDBBarcode(barcode),
		CreatedAt: pgtype.Timestamp{Valid: true},
		DeletedAt: pgtype.Timestamp{Valid: false},
	}
}

func testDBLibraryPatron(id uuid.UUID, fullName string, barcode *string) db.VwLibraryPatron {
	return db.VwLibraryPatron{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		FullName:  fullName,
		Barcode:   testDBBarcode(barcode),
		CreatedAt: pgtype.Timestamp{Valid: true},
	}
}

func testDBBarcode(barcode *string) pgtype.Text {
	if barcode == nil {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{String: *barcode, Valid: true}
}

func assertValidationError(t *testing.T, err error, expected ErrorDetails) {
	t.Helper()
	details, ok := err.(ErrorDetails)
	if !assert.True(t, ok, "expected ErrorDetails, got %T", err) {
		return
	}
	assert.Equal(t, expected, details)
}
