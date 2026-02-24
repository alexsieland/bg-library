package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s Server) AddPatron(c *gin.Context) {
	var jsonObject AddPatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}
	errorDetails := ValidateStringLength("name", jsonObject.Name, 1, 100, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	dbPatron, err := s.queries.CreatePatron(c.Request.Context(), jsonObject.Name)
	if err != nil {
		log.Printf("Error creating patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, FromPatron(dbPatron))
}

func (s Server) DeletePatron(c *gin.Context, patronId string) {
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", patronId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	err := s.queries.DeletePatron(c.Request.Context(), patronUUID)
	if err != nil {
		log.Printf("Error deleting patron: %v", err)
		internalError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (s Server) GetPatron(c *gin.Context, patronId string) {
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", patronId, []ErrorDetail{})
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	dbPatron, err := s.queries.GetPatron(c.Request.Context(), patronUUID)
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

func (s Server) UpdatePatron(c *gin.Context, patronId string) {
	var jsonObject UpdatePatronJSONRequestBody
	err := c.ShouldBindBodyWithJSON(&jsonObject)
	if err != nil {
		malformedJson(c)
		return
	}
	errorDetails := ValidateStringLength("name", jsonObject.Name, 1, 100, []ErrorDetail{})
	patronUUID, errorDetails := ConvertToPgTypeUUID("PatronId", patronId, errorDetails)
	if len(errorDetails) > 0 {
		validationError(c, errorDetails)
		return
	}
	err = s.queries.EditPatron(c.Request.Context(), db.EditPatronParams{
		ID:       patronUUID,
		FullName: jsonObject.Name,
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
			FullName: name,
			Limit:    999,
			Offset:   0,
		})
		if err != nil {
			log.Printf("Error saerching patrons: %v", err)
			internalError(c, err)
			return
		}
	}

	patronList := make([]Patron, len(dbPatronList))
	for i, dbPatron := range dbPatronList {
		patronList[i] = FromVwLibraryPatron(dbPatron)
	}
	c.JSON(http.StatusOK, patronList)
}
