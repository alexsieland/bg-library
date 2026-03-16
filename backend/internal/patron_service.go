package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PatronService struct {
	LibraryService *LibraryService
}

func NewPatronService(libService *LibraryService) *PatronService {
	return &PatronService{LibraryService: libService}
}

func (s PatronService) InsertPatron(ctx context.Context, name string, barcode *string, optTx pgx.Tx) (db.Patron, error) {
	patron, err := WithinTx(s.LibraryService, ctx, optTx, func(tx pgx.Tx) (*db.Patron, error) {
		dbBarcode := pgtype.Text{Valid: false}
		if barcode != nil {
			dbBarcode = pgtype.Text{String: *barcode, Valid: true}
		}
		createPatronParams := db.CreatePatronParams{
			FullName: name,
			Barcode:  dbBarcode,
		}

		dbPatron, err := s.LibraryService.queries.WithTx(tx).CreatePatron(ctx, createPatronParams)
		return &dbPatron, err
	})

	if err != nil || patron == nil {
		return db.Patron{}, wrapDatabaseError(err)
	}
	return *patron, nil
}

func (s PatronService) DeletePatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) error {
	_, err := WithinTx(s.LibraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.LibraryService.queries.WithTx(tx).DeletePatron(ctx, patronId)
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s PatronService) GetPatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryPatron, error) {
	patron, err := WithinTx(s.LibraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		dbPatron, err := s.LibraryService.queries.WithTx(tx).GetPatron(ctx, patronId)
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

func (s PatronService) GetPatronByBarcode(ctx context.Context, patronBarcode string, optTx pgx.Tx) (db.VwLibraryPatron, error) {
	var (
		dbPatron db.VwLibraryPatron
		err      error
	)

	barcode := pgtype.Text{String: patronBarcode, Valid: true}
	if optTx != nil {
		dbPatron, err = s.LibraryService.queries.WithTx(optTx).GetPatronByBarcode(ctx, barcode)
	} else {
		dbPatron, err = s.LibraryService.queries.GetPatronByBarcode(ctx, barcode)
	}

	if err != nil {
		return dbPatron, wrapDatabaseError(err)
	}
	return dbPatron, nil
}

func (s PatronService) UpdatePatron(ctx context.Context, patronId pgtype.UUID, fullName string, barcode *string, optTx pgx.Tx) error {
	_, err := WithinTx(s.LibraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		dbBarcode := pgtype.Text{Valid: false}
		if barcode != nil {
			dbBarcode = pgtype.Text{String: *barcode, Valid: true}
		}

		err := s.LibraryService.queries.WithTx(tx).EditPatron(ctx, db.EditPatronParams{
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

func (s PatronService) ListPatrons(ctx context.Context, fullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryPatron, error) {
	var (
		dbPatronList []db.VwLibraryPatron
		err          error
	)

	if fullName == nil {
		params := db.ListPatronsParams{
			Limit:  limit,
			Offset: offset,
		}
		if optTx != nil {
			dbPatronList, err = s.LibraryService.queries.WithTx(optTx).ListPatrons(ctx, params)
		} else {
			dbPatronList, err = s.LibraryService.queries.ListPatrons(ctx, params)
		}
	} else {
		params := db.SearchPatronsParams{
			FullName: "%" + *fullName + "%",
			Limit:    limit,
			Offset:   offset,
		}
		if optTx != nil {
			dbPatronList, err = s.LibraryService.queries.WithTx(optTx).SearchPatrons(ctx, params)
		} else {
			dbPatronList, err = s.LibraryService.queries.SearchPatrons(ctx, params)
		}
	}

	if err != nil {
		return dbPatronList, wrapDatabaseError(err)
	}
	return dbPatronList, nil
}
