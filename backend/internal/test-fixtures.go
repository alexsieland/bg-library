package internal

// Centralized test fixtures (exported wrappers).
//
// Many tests currently define make* helpers across multiple *_test.go files.
// This file provides stable, exported wrappers that call the existing helpers
// (so we don't duplicate implementations or need to refactor many tests at
// once). Over time you can move implementations here and remove the old
// helpers from *_test.go files.

import (
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// MakeLibraryGame returns a db.VwLibraryGame fixture.
func MakeLibraryGame(id uuid.UUID, title string, barcode *string) db.VwLibraryGame {
	barcodePg := pgtype.Text{Valid: false}
	if barcode != nil {
		barcodePg = pgtype.Text{String: *barcode, Valid: true}
	}
	return db.VwLibraryGame{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		DisplayTitle:    title,
		Title:           title,
		SanitizedTitle:  SanitizeTitle(title),
		Barcode:         barcodePg,
		PlayToWinGameID: pgtype.UUID{Valid: false},
		CreatedAt:       pgtype.Timestamp{Valid: true},
	}
}

// MakeGame returns a db.Game fixture.
func MakeGame(id uuid.UUID, title string, barcode *string) db.Game {
	barcodePg := pgtype.Text{Valid: false}
	if barcode != nil {
		barcodePg = pgtype.Text{String: *barcode, Valid: true}
	}
	return db.Game{
		ID:             pgtype.UUID{Bytes: id, Valid: true},
		Title:          title,
		DisplayTitle:   pgtype.Text{String: title, Valid: true},
		SanitizedTitle: SanitizeTitle(title),
		CreatedAt:      pgtype.Timestamp{Valid: true},
		DeletedAt:      pgtype.Timestamp{Valid: false},
		Barcode:        barcodePg,
	}
}

// MakeVwPlayToWinGame returns a db.VwPlayToWinGame fixture.
func MakeVwPlayToWinGame(ptwGameId uuid.UUID, libraryGameId uuid.UUID, groupName string) db.VwPlayToWinGame {
	return db.VwPlayToWinGame{
		ID:         pgtype.UUID{Bytes: ptwGameId, Valid: true},
		GameID:     pgtype.UUID{Bytes: libraryGameId, Valid: true},
		PtwGroupID: pgtype.UUID{Valid: false},
		GroupName:  pgtype.Text{String: groupName, Valid: true},
		CreatedAt:  pgtype.Timestamp{Valid: true},
		WinnerID:   pgtype.UUID{Valid: false},
	}
}

// MakeVwPlayToWinGroup returns a db.VwPlayToWinGroup fixture.
func MakeVwPlayToWinGroup(id uuid.UUID, name string) db.VwPlayToWinGroup {
	return db.VwPlayToWinGroup{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Name:      name,
		CreatedAt: pgtype.Timestamp{Valid: true},
	}
}

// MakeVwPlayToWinEntry returns a db.VwPlayToWinEntry fixture.
func MakeVwPlayToWinEntry(id uuid.UUID, sessionId uuid.UUID, groupId uuid.UUID, entrantName string, entrantUniqueID string, createdAt time.Time) db.VwPlayToWinEntry {
	ts := pgtype.Timestamp{Valid: false}
	if !createdAt.IsZero() {
		ts = pgtype.Timestamp{Time: createdAt, Valid: true}
	}
	return db.VwPlayToWinEntry{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		PtwSessionID:    pgtype.UUID{Bytes: sessionId, Valid: true},
		PtwGroupID:      pgtype.UUID{Bytes: groupId, Valid: true},
		EntrantName:     entrantName,
		EntrantUniqueID: entrantUniqueID,
		CreatedAt:       ts,
	}
}

// MakeVwPlayToWinSession returns a db.VwPlayToWinSession fixture.
func MakeVwPlayToWinSession(id uuid.UUID, groupId uuid.UUID, playtimeMinutes *int32, createdAt time.Time) db.VwPlayToWinSession {
	ts := pgtype.Timestamp{Valid: false}
	if !createdAt.IsZero() {
		ts = pgtype.Timestamp{Time: createdAt, Valid: true}
	}
	return db.VwPlayToWinSession{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		PtwGroupID:      pgtype.UUID{Bytes: groupId, Valid: true},
		PlaytimeMinutes: int32ToPgInt4(playtimeMinutes),
		CreatedAt:       ts,
	}
}

// MakeVwPlayToWinGameOverview returns a db.VwPlayToWinGameOverview fixture.
func MakeVwPlayToWinGameOverview(ptwGameId uuid.UUID, gameId uuid.UUID, ptwGroupId uuid.UUID, title string, winnerID *uuid.UUID) db.VwPlayToWinGameOverview {
	winner := pgtype.UUID{Valid: false}
	winnerName := pgtype.Text{Valid: false}
	winnerUnique := pgtype.Text{Valid: false}
	if winnerID != nil {
		winner = pgtype.UUID{Bytes: *winnerID, Valid: true}
		winnerName = pgtype.Text{String: "Winner", Valid: true}
		winnerUnique = pgtype.Text{String: "unique-1", Valid: true}
	}
	return db.VwPlayToWinGameOverview{
		PtwGameID:      pgtype.UUID{Bytes: ptwGameId, Valid: true},
		GameID:         pgtype.UUID{Bytes: gameId, Valid: true},
		PtwGroupID:     pgtype.UUID{Bytes: ptwGroupId, Valid: true},
		GameTitle:      title,
		SanitizedTitle: SanitizeTitle(title),
		CreatedAt:      pgtype.Timestamp{Valid: true},
		WinnerID:       winner,
		WinnerName:     winnerName,
		WinnerUniqueID: winnerUnique,
	}
}

// Note: Several lower-level helpers remain defined in *_test.go files.
// This wrapper file is intentionally non-invasive—it's safe to add now
// and can be expanded or replaced with canonical implementations later.
