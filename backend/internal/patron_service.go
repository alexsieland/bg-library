package internal

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
)

type PatronService struct {
	Database db.DB
	queries  *db.Queries
}

func (s PatronService) insertPatron(c *gin.Context, name string, barcode *string, optTx *pgx.Tx) (db.Patron, error) {
	var tx pgx.Tx
	if optTx != nil {
		tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
		if err != nil {
			return db.Patron{}, err
		}
		defer func() {
			_ = tx.Rollback(c.Request.Context())
		}()
	}

	validationError := ValidationError{}
	validationError.ValidateStringLength("name", name, 1, 100)
	if barcode != nil {
		validationError.ValidateStringLength("barcode", *barcode, 1, 48)
	}

	if !validationError.Empty() {
		return db.Patron{}, &validationError
	}

	dbBarcode := pgtype.Text{Valid: false}
	if barcode != nil {
		dbBarcode = pgtype.Text{String: *barcode, Valid: true}
	}
	createPatronParams := db.CreatePatronParams{
		FullName: name,
		Barcode:  dbBarcode,
	}

	patron, err := s.queries.WithTx(tx).CreatePatron(c.Request.Context(), createPatronParams)
	if err != nil {
		println("Error creating patron: %v", err)
		return wrapDatabaseError(db.Patron{}, err)
	}
	if optTx == nil {
		_ = tx.Commit(c.Request.Context())
	}
	return patron, nil
}

func (s PatronService) DeletePatron(c *gin.Context, patronId pgtype.UUID, optTx *pgx.Tx) error {
	var tx pgx.Tx
	if optTx != nil {
		tx, err := s.Database.BeginTx(c.Request.Context(), pgx.TxOptions{})
		if err != nil {
			return db.Patron{}, err
		}
		defer func() {
			_ = tx.Rollback(c.Request.Context())
		}()
	}

	err := s.queries.DeletePatron(c.Request.Context(), patronId)
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
		if isNotFound(err) {
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
	if !errorDetails.Empty() {
		validationError(c, errorDetails)
		return
	}
	var barcode = pgtype.Text{String: patronBarcode, Valid: true}
	dbPatron, err := s.queries.GetPatronByBarcode(c.Request.Context(), barcode)
	if err != nil {
		if isNotFound(err) {
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
	if !errorDetails.Empty() {
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
		if isNotFound(err) {
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
