package internal

import (
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Test fixtures / builders for PlayToWin related DB view/models used by unit tests.

func makeVwPlayToWinGame(ptwGameId uuid.UUID, libraryGameId uuid.UUID, groupName string) db.VwPlayToWinGame {
	return db.VwPlayToWinGame{
		ID:         pgtype.UUID{Bytes: ptwGameId, Valid: true},
		GameID:     pgtype.UUID{Bytes: libraryGameId, Valid: true},
		PtwGroupID: pgtype.UUID{Valid: false},
		GroupName:  pgtype.Text{String: groupName, Valid: true},
		CreatedAt:  pgtype.Timestamp{Valid: true},
		WinnerID:   pgtype.UUID{Valid: false},
	}
}

func makeVwPlayToWinGroup(id uuid.UUID, name string) db.VwPlayToWinGroup {
	return db.VwPlayToWinGroup{
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		Name:      name,
		CreatedAt: pgtype.Timestamp{Valid: true},
	}
}

func makeVwPlayToWinEntry(id uuid.UUID, sessionId uuid.UUID, groupId uuid.UUID, entrantName string, entrantUniqueID string, createdAt time.Time) db.VwPlayToWinEntry {
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

func makeVwPlayToWinSession(id uuid.UUID, groupId uuid.UUID, playtimeMinutes *int32, createdAt time.Time) db.VwPlayToWinSession {
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

func makeVwPlayToWinGameOverview(ptwGameId uuid.UUID, gameId uuid.UUID, ptwGroupId uuid.UUID, title string, winnerID *uuid.UUID) db.VwPlayToWinGameOverview {
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
