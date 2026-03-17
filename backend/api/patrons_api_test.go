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
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		barcode := "P-1001"
		expectedPatron := testDBPatron(uuid.New(), "Jane Doe", &barcode)

		service.On("InsertPatron", ctx, "Jane Doe", &barcode, nil).Return(expectedPatron, nil).Once()

		patron, err := api.AddPatron(ctx, AddPatronJSONRequestBody{Name: "Jane Doe", Barcode: &barcode})

		assert.NoError(t, err)
		assert.Equal(t, FromPatron(expectedPatron), patron)
		service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the request fields are invalid", func(t *testing.T) {
		api := newTestPatronApi(new(mockPatronService), nil, nil)
		tooLongName := strings.Repeat("n", 101)
		tooLongBarcode := strings.Repeat("b", 49)

		patron, err := api.AddPatron(t.Context(), AddPatronJSONRequestBody{Name: tooLongName, Barcode: &tooLongBarcode})

		assert.Equal(t, Patron{}, patron)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{
			{Field: "name", Message: "Length must be between 1 and 100"},
			{Field: "barcode", Message: "Length must be between 1 and 48"},
		}})
	})

	t.Run("Should return the service error when inserting a patron fails", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		expectedErr := errors.New("insert failed")

		service.On("InsertPatron", ctx, "Jane Doe", (*string)(nil), nil).Return(db.Patron{}, expectedErr).Once()

		patron, err := api.AddPatron(ctx, AddPatronJSONRequestBody{Name: "Jane Doe"})

		assert.Equal(t, Patron{}, patron)
		assert.ErrorIs(t, err, expectedErr)
		service.AssertExpectations(t)
	})
}

func TestPatronApiBulkAddPatrons(t *testing.T) {
	t.Run("Should import patrons and commit the transaction when the CSV is valid", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		tx := &stubTx{}
		api := newTestPatronApi(service, tx, nil)
		barcode := "P-2001"

		service.On("InsertPatron", ctx, "Jane Doe", &barcode, mock.Anything).Return(testDBPatron(uuid.New(), "Jane Doe", &barcode), nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})
		service.On("InsertPatron", ctx, "John Smith", (*string)(nil), mock.Anything).Return(testDBPatron(uuid.New(), "John Smith", nil), nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})

		response, err := api.BulkAddPatrons(ctx, encodedCSVBody("name,barcode\nJane Doe,P-2001\nJohn Smith,\n"))

		assert.NoError(t, err)
		assert.Equal(t, BulkAddResponse{Imported: 2}, response)
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 0, tx.rollbackCount)
		service.AssertExpectations(t)
	})

	t.Run("Should return the transaction start error when opening the bulk import transaction fails", func(t *testing.T) {
		api := newTestPatronApi(new(mockPatronService), nil, errors.New("begin failed"))

		response, err := api.BulkAddPatrons(t.Context(), encodedCSVBody("name,barcode\nJane Doe,P-2001\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.EqualError(t, err, "begin failed")
	})

	t.Run("Should return the CSV read error when the request body is not valid base64", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		tx := &stubTx{}
		api := newTestPatronApi(service, tx, nil)

		response, err := api.BulkAddPatrons(ctx, io.NopCloser(strings.NewReader("%%%")))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.Error(t, err)
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		service.AssertNotCalled(t, "InsertPatron", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Should accumulate validation details and skip inserts when any CSV row is invalid", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		tx := &stubTx{}
		api := newTestPatronApi(service, tx, nil)
		tooLongBarcode := strings.Repeat("b", 49)

		response, err := api.BulkAddPatrons(ctx, encodedCSVBody("name,barcode\n,P-2002\nValid Name,"+tooLongBarcode+"\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{
			{Field: "name", Message: "Length must be between 1 and 100"},
			{Field: "barcode", Message: "Length must be between 1 and 48"},
		}})
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		service.AssertNotCalled(t, "InsertPatron", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("Should return the service error and roll back when an insert fails during bulk import", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		tx := &stubTx{}
		api := newTestPatronApi(service, tx, nil)
		expectedErr := errors.New("insert failed")

		service.On("InsertPatron", ctx, "Jane Doe", (*string)(nil), mock.Anything).Return(db.Patron{}, expectedErr).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})

		response, err := api.BulkAddPatrons(ctx, encodedCSVBody("name,barcode\nJane Doe,\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, 0, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		service.AssertExpectations(t)
	})

	t.Run("Should return the commit error and roll back when committing the bulk import transaction fails", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		tx := &stubTx{commitErr: errors.New("commit failed")}
		api := newTestPatronApi(service, tx, nil)

		service.On("InsertPatron", ctx, "Jane Doe", (*string)(nil), mock.Anything).Return(testDBPatron(uuid.New(), "Jane Doe", nil), nil).Once().Run(func(args mock.Arguments) {
			assert.Same(t, tx, args.Get(3))
		})

		response, err := api.BulkAddPatrons(ctx, encodedCSVBody("name,barcode\nJane Doe,\n"))

		assert.Equal(t, BulkAddResponse{}, response)
		assert.EqualError(t, err, "commit failed")
		assert.Equal(t, 1, tx.commitCount)
		assert.Equal(t, 1, tx.rollbackCount)
		service.AssertExpectations(t)
	})
}

func TestPatronApiDeletePatron(t *testing.T) {
	t.Run("Should forward the converted patron identifier when deleting a patron", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		patronID := uuid.New()

		service.On("DeletePatron", ctx, pgtype.UUID{Bytes: patronID, Valid: true}, nil).Return(nil).Once()

		err := api.DeletePatron(ctx, types.UUID(patronID))

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("Should return the service error when deleting a patron fails", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		patronID := uuid.New()
		expectedErr := errors.New("delete failed")

		service.On("DeletePatron", ctx, pgtype.UUID{Bytes: patronID, Valid: true}, nil).Return(expectedErr).Once()

		err := api.DeletePatron(ctx, types.UUID(patronID))

		assert.ErrorIs(t, err, expectedErr)
		service.AssertExpectations(t)
	})
}

func TestPatronApiGetPatron(t *testing.T) {
	t.Run("Should return a converted patron when the service finds one", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		patronID := uuid.New()
		barcode := "P-3001"
		expectedPatron := testDBLibraryPatron(patronID, "Jordan Doe", &barcode)

		service.On("GetPatron", ctx, pgtype.UUID{Bytes: patronID, Valid: true}, nil).Return(expectedPatron, nil).Once()

		patron, err := api.GetPatron(ctx, types.UUID(patronID))

		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryPatron(expectedPatron), patron)
		service.AssertExpectations(t)
	})

	t.Run("Should return the service error when fetching a patron fails", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		patronID := uuid.New()
		expectedErr := errors.New("lookup failed")

		service.On("GetPatron", ctx, pgtype.UUID{Bytes: patronID, Valid: true}, nil).Return(db.VwLibraryPatron{}, expectedErr).Once()

		patron, err := api.GetPatron(ctx, types.UUID(patronID))

		assert.Equal(t, Patron{}, patron)
		assert.ErrorIs(t, err, expectedErr)
		service.AssertExpectations(t)
	})
}

func TestPatronApiGetPatronByBarcode(t *testing.T) {
	t.Run("Should return a converted patron when the barcode is valid", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		barcode := "P-4001"
		expectedPatron := testDBLibraryPatron(uuid.New(), "Taylor Doe", &barcode)

		service.On("GetPatronByBarcode", ctx, barcode, nil).Return(expectedPatron, nil).Once()

		patron, err := api.GetPatronByBarcode(ctx, barcode)

		assert.NoError(t, err)
		assert.Equal(t, FromVwLibraryPatron(expectedPatron), patron)
		service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the barcode is invalid", func(t *testing.T) {
		api := newTestPatronApi(new(mockPatronService), nil, nil)
		tooLongBarcode := strings.Repeat("b", 49)

		patron, err := api.GetPatronByBarcode(t.Context(), tooLongBarcode)

		assert.Equal(t, Patron{}, patron)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "patronBarcode", Message: "Length must be between 1 and 48"}}})
	})
}

func TestPatronApiUpdatePatron(t *testing.T) {
	t.Run("Should forward the update request to the service when the request is valid", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		patronID := uuid.New()
		barcode := "P-5001"

		service.On("UpdatePatron", ctx, pgtype.UUID{Bytes: patronID, Valid: true}, "Updated Name", &barcode, nil).Return(nil).Once()

		err := api.UpdatePatron(ctx, types.UUID(patronID), UpdatePatronJSONRequestBody{Name: "Updated Name", Barcode: &barcode})

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the update request is invalid", func(t *testing.T) {
		api := newTestPatronApi(new(mockPatronService), nil, nil)
		tooLongName := strings.Repeat("n", 101)

		err := api.UpdatePatron(t.Context(), types.UUID(uuid.New()), UpdatePatronJSONRequestBody{Name: tooLongName})

		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "name", Message: "Length must be between 1 and 100"}}})
	})

	t.Run("Should return the service error when updating a patron fails", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		patronID := uuid.New()
		expectedErr := errors.New("update failed")

		service.On("UpdatePatron", ctx, pgtype.UUID{Bytes: patronID, Valid: true}, "Updated Name", (*string)(nil), nil).Return(expectedErr).Once()

		err := api.UpdatePatron(ctx, types.UUID(patronID), UpdatePatronJSONRequestBody{Name: "Updated Name"})

		assert.ErrorIs(t, err, expectedErr)
		service.AssertExpectations(t)
	})
}

func TestPatronApiListPatrons(t *testing.T) {
	t.Run("Should return a converted patron list when the list request is valid", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		searchName := "Doe"
		barcode := "P-6001"
		expectedPatrons := []db.VwLibraryPatron{
			testDBLibraryPatron(uuid.New(), "Casey Doe", &barcode),
			testDBLibraryPatron(uuid.New(), "Logan Doe", nil),
		}

		service.On("ListPatrons", ctx, &searchName, int32(999), int32(0), nil).Return(expectedPatrons, nil).Once()

		patronList, err := api.ListPatrons(ctx, ListPatronsParams{Name: &searchName})

		assert.NoError(t, err)
		assert.Equal(t, PatronList{Patrons: []Patron{
			FromVwLibraryPatron(expectedPatrons[0]),
			FromVwLibraryPatron(expectedPatrons[1]),
		}}, patronList)
		service.AssertExpectations(t)
	})

	t.Run("Should return validation details when the list search name is invalid", func(t *testing.T) {
		api := newTestPatronApi(new(mockPatronService), nil, nil)
		tooLongName := strings.Repeat("n", 101)

		patronList, err := api.ListPatrons(t.Context(), ListPatronsParams{Name: &tooLongName})

		assert.Equal(t, PatronList{}, patronList)
		assertValidationError(t, err, ErrorDetails{Details: []ErrorDetail{{Field: "name", Message: "Length must be between 1 and 100"}}})
	})

	t.Run("Should return the service error when listing patrons fails", func(t *testing.T) {
		ctx := t.Context()
		service := new(mockPatronService)
		api := newTestPatronApi(service, nil, nil)
		expectedErr := errors.New("list failed")

		service.On("ListPatrons", ctx, (*string)(nil), int32(999), int32(0), nil).Return(nil, expectedErr).Once()

		patronList, err := api.ListPatrons(ctx, ListPatronsParams{})

		assert.Equal(t, PatronList{}, patronList)
		assert.ErrorIs(t, err, expectedErr)
		service.AssertExpectations(t)
	})
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
