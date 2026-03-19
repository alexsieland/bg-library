package internal

import (
	"context"

	"github.com/alexsieland/bg-library/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PatronService struct {
	libraryService *LibraryService
}

func NewPatronService(libService *LibraryService) *PatronService {
	return &PatronService{libraryService: libService}
}

func (s *PatronService) InsertPatron(ctx context.Context, name string, barcode *string, optTx pgx.Tx) (db.Patron, error) {
	patron, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.Patron, error) {
		dbBarcode := pgtype.Text{Valid: false}
		if barcode != nil {
			dbBarcode = pgtype.Text{String: *barcode, Valid: true}
		}
		createPatronParams := db.CreatePatronParams{
			FullName: name,
			Barcode:  dbBarcode,
		}

		dbPatron, err := s.libraryService.queries.WithTx(tx).CreatePatron(ctx, createPatronParams)
		return &dbPatron, err
	})

	return wrapErrorOrReturn(patron, db.Patron{}, err)
}

func (s *PatronService) DeletePatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) error {
	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*struct{}, error) {
		err := s.libraryService.queries.WithTx(tx).DeletePatron(ctx, patronId)
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s *PatronService) GetPatron(ctx context.Context, patronId pgtype.UUID, optTx pgx.Tx) (db.VwLibraryPatron, error) {
	patron, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		dbPatron, err := s.libraryService.queries.WithTx(tx).GetPatron(ctx, patronId)
		if err != nil {
			return nil, err
		}
		return &dbPatron, nil
	})

	return wrapErrorOrReturn(patron, db.VwLibraryPatron{}, err)
}

func (s *PatronService) GetPatronByBarcode(ctx context.Context, patronBarcode string, optTx pgx.Tx) (db.VwLibraryPatron, error) {
	var (
		dbPatron db.VwLibraryPatron
		err      error
	)

	barcode := pgtype.Text{String: patronBarcode, Valid: true}
	if optTx != nil {
		dbPatron, err = s.libraryService.queries.WithTx(optTx).GetPatronByBarcode(ctx, barcode)
	} else {
		dbPatron, err = s.libraryService.queries.GetPatronByBarcode(ctx, barcode)
	}

	if err != nil {
		return dbPatron, wrapDatabaseError(err)
	}
	return dbPatron, nil
}

func (s *PatronService) UpdatePatron(ctx context.Context, patronId pgtype.UUID, fullName string, barcode *string, optTx pgx.Tx) error {
	_, err := WithinTx(s.libraryService, ctx, optTx, func(tx pgx.Tx) (*db.VwLibraryPatron, error) {
		dbBarcode := pgtype.Text{Valid: false}
		if barcode != nil {
			dbBarcode = pgtype.Text{String: *barcode, Valid: true}
		}

		err := s.libraryService.queries.WithTx(tx).EditPatron(ctx, db.EditPatronParams{
			ID:       patronId,
			FullName: fullName,
			Barcode:  dbBarcode,
		})
		return nil, err
	})

	return wrapDatabaseError(err)
}

func (s *PatronService) listPatrons(ctx context.Context, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryPatron, error) {
	params := db.ListPatronsParams{
		Limit:  limit,
		Offset: offset,
	}

	if optTx != nil {
		return s.libraryService.queries.WithTx(optTx).ListPatrons(ctx, params)
	}
	return s.libraryService.queries.ListPatrons(ctx, params)
}

func (s *PatronService) searchPatrons(ctx context.Context, fullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryPatron, error) {
	if fullName == nil || *fullName == "" {
		return s.listPatrons(ctx, limit, offset, optTx)
	}

	params := db.SearchPatronsParams{
		FullName: GenerateDBRegexString(*fullName),
		Limit:    limit,
		Offset:   offset,
	}

	if optTx != nil {
		return s.libraryService.queries.WithTx(optTx).SearchPatrons(ctx, params)
	}
	return s.libraryService.queries.SearchPatrons(ctx, params)
}

func (s *PatronService) ListPatrons(ctx context.Context, fullName *string, limit int32, offset int32, optTx pgx.Tx) ([]db.VwLibraryPatron, error) {
	dbPatronList, err := s.searchPatrons(ctx, fullName, limit, offset, optTx)
	if err != nil {
		// For callers that expect a nil slice on error, return nil alongside the error
		return nil, wrapDatabaseError(err)
	}
	return dbPatronList, nil
}
