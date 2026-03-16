package internal

import (
	"github.com/alexsieland/bg-library/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PatronService struct {
	libService *LibraryService
}

func (s PatronService) InsertPatron(c *gin.Context, name string, barcode *string, optTx *pgx.Tx) (db.Patron, error) {
	patron, err := WithinTx(s.libService, optTx, func(tx pgx.Tx) (*db.Patron, error) {
		dbBarcode := pgtype.Text{Valid: false}
		if barcode != nil {
			dbBarcode = pgtype.Text{String: *barcode, Valid: true}
		}
		createPatronParams := db.CreatePatronParams{
			FullName: name,
			Barcode:  dbBarcode,
		}

		dbPatron, err := s.libService.queries.WithTx(tx).CreatePatron(c.Request.Context(), createPatronParams)
		return &dbPatron, err
	})
	if err != nil || patron == nil {
		return db.Patron{}, wrapDatabaseError(err)
	}
	return *patron, nil
}

func (s PatronService) DeletePatron(c *gin.Context, patronId pgtype.UUID, optTx *pgx.Tx) error {
	_, err := WithinTx(s.libService, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libService.queries.WithTx(tx).DeletePatron(c.Request.Context(), patronId)
		return nil, err
	})
	return wrapDatabaseError(err)
}

func (s PatronService) GetPatron(c *gin.Context, patronId pgtype.UUID) (db.VwLibraryPatron, error) {
	patron, err := WithinTx(s.libService, nil, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		dbPatron, err := s.libService.queries.WithTx(tx).GetPatron(c.Request.Context(), patronId)
		if err != nil {
			return nil, err
		}
		return &dbPatron, nil
	})
	if err != nil || patron == nil {
		return db.VwLibraryPatron{}, wrapDatabaseError(err)
	}
	return *patron, nil
}

func (s PatronService) GetPatronByBarcode(c *gin.Context, patronBarcode string) (db.VwLibraryPatron, error) {
	patron, err := WithinTx(s.libService, nil, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		var barcode = pgtype.Text{String: patronBarcode, Valid: true}
		dbPatron, err := s.libService.queries.WithTx(tx).GetPatronByBarcode(c.Request.Context(), barcode)
		if err != nil {
			return nil, wrapDatabaseError(err)
		}
		return &dbPatron, nil
	})
	if err != nil || patron == nil {
		return db.VwLibraryPatron{}, wrapDatabaseError(err)
	}
	return *patron, nil
}

func (s PatronService) UpdatePatron(c *gin.Context, patronId pgtype.UUID, fullName string, barcode *string) error {
	_, err := WithinTx(s.libService, nil, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		dbBarcode := pgtype.Text{Valid: false}
		if barcode != nil {
			dbBarcode = pgtype.Text{String: *barcode, Valid: true}
		}

		err := s.libService.queries.WithTx(tx).EditPatron(c.Request.Context(), db.EditPatronParams{
			ID:       patronId,
			FullName: fullName,
			Barcode:  dbBarcode,
		})
		return nil, err
	})
	if err != nil {
		return wrapDatabaseError(err)
	}
	return nil
}

func (s PatronService) ListPatrons(c *gin.Context, fullName *string, limit int32, offset int32) ([]db.VwLibraryPatron, error) {
	patrons, err := WithinTx(s.libService, nil, func(tx pgx.Tx) (*[]db.VwLibraryPatron, error) {
		var (
			dbPatronList []db.VwLibraryPatron
			err          error
		)
		if fullName == nil {
			dbPatronList, err = s.libService.queries.WithTx(tx).ListPatrons(c.Request.Context(), db.ListPatronsParams{
				Limit:  limit,
				Offset: offset,
			})
		} else {
			dbPatronList, err = s.libService.queries.WithTx(tx).SearchPatrons(c.Request.Context(), db.SearchPatronsParams{
				FullName: "%" + *fullName + "%",
				Limit:    999,
				Offset:   0,
			})
		}
		if err != nil {
			return nil, err
		}
		return &dbPatronList, nil
	})
	if err != nil || patrons == nil {
		return []db.VwLibraryPatron{}, wrapDatabaseError(err)
	}
	return *patrons, nil
}
