package api

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"io"
	"log"

	"github.com/alexsieland/bg-library/db"
	"github.com/alexsieland/bg-library/internal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

type patronService interface {
	InsertPatron(ctx context.Context, name string, barcode *string, optTx pgx.Tx) (db.Patron, error)
	DeletePatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) error
	GetPatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryPatron, error)
	GetPatronByBarcode(ctx context.Context, patronBarcode string, optTx pgx.Tx) (db.VwLibraryPatron, error)
	UpdatePatron(ctx context.Context, patronId pgtype.UUID, fullName string, barcode *string, optTx pgx.Tx) error
	ListPatrons(ctx context.Context, fullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryPatron, error)
}

type PatronApi struct {
	libraryService *internal.LibraryService
	service        patronService
}

func NewPatronApi(libService *internal.LibraryService, patronSrv *internal.PatronService) *PatronApi {
	return &PatronApi{
		libraryService: libService,
		service:        patronSrv,
	}
}

func (api *PatronApi) AddPatron(ctx context.Context, request AddPatronJSONRequestBody) (Patron, error) {
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("name", request.Name, 1, 100)
	if request.Barcode != nil {
		errorDetails.ValidateStringLength("barcode", *request.Barcode, 1, 48)
	}
	if !errorDetails.Empty() {
		return Patron{}, errorDetails
	}

	dbPatron, err := api.service.InsertPatron(ctx, request.Name, request.Barcode, nil)
	if err != nil {
		log.Printf("Error adding patron: %v", err)
		return Patron{}, err
	}

	return FromPatron(dbPatron), nil
}

func (api *PatronApi) BulkAddPatrons(ctx context.Context, requestBody io.ReadCloser) (BulkAddResponse, error) {
	decodedReader := base64.NewDecoder(base64.StdEncoding, requestBody)
	csvReader := csv.NewReader(decodedReader)

	// Start a db transaction
	tx, err := api.libraryService.BeginTx(ctx)
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return BulkAddResponse{}, err
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Process each row
	var errorDetails ErrorDetails
	recordCount := int32(0)
	firstRow := true
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV: %v", err)
			return BulkAddResponse{}, err
		}
		if firstRow {
			firstRow = false
			continue
		}
		if len(record) == 0 {
			continue
		}

		// Validate each row
		name := record[0]
		errorDetails.ValidateStringLength("name", name, 1, 100)

		var barcode *string
		if len(record) > 1 && record[1] != "" {
			barcode = &record[1]
			errorDetails.ValidateStringLength("barcode", *barcode, 1, 48)
		}

		// If there have been any validation errors, check remaining records for validation errors but skip database inserts
		if !errorDetails.Empty() {
			continue
		}

		// If there are no validation errors, attempt to insert patron into database
		_, err = api.service.InsertPatron(ctx, name, barcode, tx)
		if err != nil {
			log.Printf("Error adding patron: %v", err)
			return BulkAddResponse{}, err
		}
		recordCount++
	}

	if !errorDetails.Empty() {
		return BulkAddResponse{}, errorDetails
	}

	//If there are no validation errors, commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return BulkAddResponse{}, err
	}

	tx = nil // Prevent deferred rollback after a successful commit
	return BulkAddResponse{Imported: recordCount}, nil
}

func (api *PatronApi) DeletePatron(ctx context.Context, patronId types.UUID) error {
	err := api.service.DeletePatron(ctx, uuidToPgTypeUUID(patronId), nil)
	if err != nil {
		log.Printf("Error deleting patron: %v", err)
		return err
	}
	return nil
}

func (api *PatronApi) GetPatron(ctx context.Context, patronId types.UUID) (Patron, error) {
	dbPatron, err := api.service.GetPatron(ctx, uuidToPgTypeUUID(patronId), nil)
	if err != nil {
		log.Printf("Error getting patron: %v", err)
		return Patron{}, err
	}
	return FromVwLibraryPatron(dbPatron), nil
}

func (api *PatronApi) GetPatronByBarcode(ctx context.Context, patronBarcode string) (Patron, error) {
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("patronBarcode", patronBarcode, 1, 48)
	if !errorDetails.Empty() {
		return Patron{}, errorDetails
	}
	dbPatron, err := api.service.GetPatronByBarcode(ctx, patronBarcode, nil)
	if err != nil {
		log.Printf("Error getting patron: %v", err)
		return Patron{}, err
	}
	return FromVwLibraryPatron(dbPatron), nil
}

func (api *PatronApi) UpdatePatron(ctx context.Context, patronId types.UUID, request UpdatePatronJSONRequestBody) error {
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("name", request.Name, 1, 100)
	if request.Barcode != nil {
		errorDetails.ValidateStringLength("barcode", *request.Barcode, 1, 48)
	}
	if !errorDetails.Empty() {
		return errorDetails
	}

	err := api.service.UpdatePatron(ctx, uuidToPgTypeUUID(patronId), request.Name, request.Barcode, nil)
	if err != nil {
		log.Printf("Error updating patron: %v", err)
		return err
	}

	return nil
}

func (api *PatronApi) ListPatrons(ctx context.Context, params ListPatronsParams) (PatronList, error) {
	var errorDetails ErrorDetails
	if params.Name != nil {
		errorDetails.ValidateStringLength("name", *params.Name, 1, 100)
	}
	if !errorDetails.Empty() {
		return PatronList{}, errorDetails
	}

	dbPatronList, err := api.service.ListPatrons(ctx, params.Name, 999, 0, nil)
	if err != nil {
		log.Printf("Error listing patrons: %v", err)
		return PatronList{}, err
	}

	patronList := make([]Patron, len(dbPatronList))
	for i, dbPatron := range dbPatronList {
		patronList[i] = FromVwLibraryPatron(dbPatron)
	}

	return PatronList{Patrons: patronList}, nil
}
