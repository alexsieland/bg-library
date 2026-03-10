package api

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

func (s Server) AddPatron(c *gin.Context) {
	var jsonObject AddPatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	var errorDetails ErrorDetails
	dbPatron, err := s.insertPatron(c, jsonObject.Name, jsonObject.Barcode, errorDetails, nil)
	if errors.Is(err, errValidation) {
		validationError(c, errorDetails)
	}
	if err != nil {
		log.Printf("Error creating patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, FromPatron(dbPatron))
}

func (s Server) insertPatron(c *gin.Context, name string, barcode *string, errorDetails ErrorDetails, tx *pgx.Tx) (db.Patron, error) {
	errorDetails.ValidateStringLength("name", name, 1, 100)
	if barcode != nil {
		errorDetails.ValidateStringLength("barcode", *barcode, 1, 48)
	}

	if errorDetails.Empty() {
		return db.Patron{}, errValidation
	}

	dbBarcode := pgtype.Text{Valid: false}
	if barcode != nil {
		dbBarcode = pgtype.Text{String: *barcode, Valid: true}
	}
	createPatronParams := db.CreatePatronParams{
		FullName: name,
		Barcode:  dbBarcode,
	}

	if tx != nil {
		return s.queries.WithTx(*tx).CreatePatron(c.Request.Context(), createPatronParams)
	}
	return s.queries.CreatePatron(c.Request.Context(), createPatronParams)
}

func (s Server) BulkAddPatrons(c *gin.Context) {
	decodedReader := base64.NewDecoder(base64.StdEncoding, c.Request.Body)
	csvReader := csv.NewReader(decodedReader)

	// Start a db transaction
	tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		log.Printf("Error creating transaction: %v", err)
		internalError(c, err)
		return
	}

	//defer rollback if there is an error
	defer func() {
		if tx != nil {
			_ = tx.Rollback(c.Request.Context())
		}
	}()

	// Process each row
	var errorDetails ErrorDetails
	recordCount := int32(0)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV: %v", err)
			internalError(c, err)
			return
		}
		if len(record) == 0 {
			continue
		}
		name := record[0]
		var barcode *string
		// Disable the ability to set the barcode on bulk add for now
		// TODO Add this back in once barcode implementation is complete
		//if len(record) > 1 && record[1] != "" {
		//	barcode = &record[1]
		//}

		_, err = s.insertPatron(c, name, barcode, errorDetails, &tx)
		if errors.Is(err, errValidation) {
			continue
		}
		if err != nil {
			log.Printf("Error adding patron: %v", err)
			internalError(c, err)
			return
		}
		recordCount++
	}

	//If there are any validation errors, rollback the transaction
	if errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}

	//If there are no validation errors, commit the transaction
	err = tx.Commit(c.Request.Context())
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		internalError(c, err)
	}
	tx = nil // Prevent deferred rollback after a successful commit

	c.JSON(http.StatusCreated, BulkAddResponse{Imported: recordCount})
}

func (s Server) DeletePatron(c *gin.Context, patronId types.UUID) {
	err := s.queries.DeletePatron(c.Request.Context(), uuidToPgTypeUUID(patronId))
	if err != nil {
		log.Printf("Error deleting patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s Server) GetPatron(c *gin.Context, patronId types.UUID) {
	dbPatron, err := s.queries.GetPatron(c.Request.Context(), uuidToPgTypeUUID(patronId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			notFound(c)
			return
		}
		log.Printf("Error getting patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, FromVwLibraryPatron(dbPatron))
}

func (s Server) GetPatronByBarcode(c *gin.Context, patronBarcode string) {
	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("patronBarcode", patronBarcode, 1, 48)
	if errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}
	var barcode = pgtype.Text{String: patronBarcode, Valid: true}
	dbPatron, err := s.queries.GetPatronByBarcode(c.Request.Context(), barcode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			notFound(c)
			return
		}
		log.Printf("Error getting patron: %v", err)
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, FromVwLibraryPatron(dbPatron))
}

func (s Server) UpdatePatron(c *gin.Context, patronId types.UUID) {
	var jsonObject UpdatePatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}

	var errorDetails ErrorDetails
	errorDetails.ValidateStringLength("name", jsonObject.Name, 1, 100)
	if errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}

	dbBarcode := pgtype.Text{Valid: false}
	if jsonObject.Barcode != nil {
		dbBarcode = pgtype.Text{String: *jsonObject.Barcode, Valid: true}
	}

	err = s.queries.EditPatron(c.Request.Context(), db.EditPatronParams{
		ID:       uuidToPgTypeUUID(patronId),
		FullName: jsonObject.Name,
		Barcode:  dbBarcode,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			notFound(c)
			return
		}
		log.Printf("Error updating patron: %v", err)
		internalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (s Server) ListPatrons(c *gin.Context, params ListPatronsParams) {
	var dbPatronList []db.VwLibraryPatron
	if params.Name == nil {
		var err error
		dbPatronList, err = s.queries.ListPatrons(c.Request.Context(), db.ListPatronsParams{
			Limit:  999,
			Offset: 0,
		})
		if err != nil {
			log.Printf("Error listing patrons: %v", err)
			internalError(c, err)
			return
		}
	} else {
		name := *params.Name
		var err error
		dbPatronList, err = s.queries.SearchPatrons(c.Request.Context(), db.SearchPatronsParams{
			FullName: "%" + name + "%",
			Limit:    999,
			Offset:   0,
		})
		if err != nil {
			log.Printf("Error searching patrons: %v", err)
			internalError(c, err)
			return
		}
	}

	patronList := make([]Patron, len(dbPatronList))
	for i, dbPatron := range dbPatronList {
		patronList[i] = FromVwLibraryPatron(dbPatron)
	}
	c.JSON(http.StatusOK, PatronList{Patrons: patronList})
}
