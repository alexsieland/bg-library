package api

import (
	"errors"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestConversionUtils(t *testing.T) {
	t.Run("Should correctly convert VwGameStatus when checked out", func(t *testing.T) {
		gameID := uuid.New()
		patronID := uuid.New()
		now := time.Now().UTC()

		dbStatus := db.VwGameStatus{
			GameID:            pgtype.UUID{Bytes: gameID, Valid: true},
			GameTitle:         "Catan",
			PatronID:          pgtype.UUID{Bytes: patronID, Valid: true},
			PatronFullName:    pgtype.Text{String: "John Doe", Valid: true},
			CheckoutTimestamp: pgtype.Timestamp{Time: now, Valid: true},
		}

		status := FromVwGameStatus(dbStatus)

		assert.Equal(t, gameID, status.Game.GameId)
		assert.Equal(t, "Catan", status.Game.Title)
		assert.NotNil(t, status.Patron)
		assert.Equal(t, patronID, status.Patron.PatronId)
		assert.Equal(t, "John Doe", status.Patron.Name)
		assert.NotNil(t, status.CheckedOutAt)
		assert.True(t, now.Equal(*status.CheckedOutAt))
	})

	t.Run("Should correctly convert VwGameStatus when not checked out", func(t *testing.T) {
		gameID := uuid.New()

		dbStatus := db.VwGameStatus{
			GameID:    pgtype.UUID{Bytes: gameID, Valid: true},
			GameTitle: "Catan",
		}

		status := FromVwGameStatus(dbStatus)

		assert.Equal(t, gameID, status.Game.GameId)
		assert.Nil(t, status.Patron)
		assert.Nil(t, status.CheckedOutAt)
	})

	t.Run("Should correctly convert VwLibraryPatron", func(t *testing.T) {
		patronID := uuid.New()
		dbPatron := db.VwLibraryPatron{
			ID:       pgtype.UUID{Bytes: patronID, Valid: true},
			FullName: "John Doe",
		}

		patron := FromVwLibraryPatron(dbPatron)

		assert.Equal(t, patronID, patron.PatronId)
		assert.Equal(t, "John Doe", patron.Name)
	})

	t.Run("Should correctly convert Patron", func(t *testing.T) {
		patronID := uuid.New()
		dbPatron := db.Patron{
			ID:       pgtype.UUID{Bytes: patronID, Valid: true},
			FullName: "John Doe",
		}

		patron := FromPatron(dbPatron)

		assert.Equal(t, patronID, patron.PatronId)
		assert.Equal(t, "John Doe", patron.Name)
	})

	t.Run("Should correctly convert VwLibraryGame", func(t *testing.T) {
		gameID := uuid.New()
		dbGame := db.VwLibraryGame{
			ID:    pgtype.UUID{Bytes: gameID, Valid: true},
			Title: "Catan",
		}

		game := FromVwLibraryGame(dbGame)

		assert.Equal(t, gameID, game.GameId)
		assert.Equal(t, "Catan", game.Title)
	})

	t.Run("Should correctly convert Transaction", func(t *testing.T) {
		gameID := uuid.New()
		patronID := uuid.New()
		transID := uuid.New()
		now := time.Now().UTC()

		dbTrans := db.Transaction{
			ID:                pgtype.UUID{Bytes: transID, Valid: true},
			GameID:            pgtype.UUID{Bytes: gameID, Valid: true},
			PatronID:          pgtype.UUID{Bytes: patronID, Valid: true},
			CheckoutTimestamp: pgtype.Timestamp{Time: now, Valid: true},
		}

		trans := FromTransaction(dbTrans)

		assert.Equal(t, transID, trans.Id)
		assert.Equal(t, gameID, trans.GameId)
		assert.Equal(t, patronID, trans.PatronId)
		assert.True(t, now.Equal(trans.Timestamp))
	})

	t.Run("Should correctly convert Game", func(t *testing.T) {
		gameID := uuid.New()
		dbGame := db.Game{
			ID:    pgtype.UUID{Bytes: gameID, Valid: true},
			Title: "Catan",
		}

		game := FromGame(dbGame, true)

		assert.Equal(t, gameID, game.GameId)
		assert.Equal(t, "Catan", game.Title)
		assert.Equal(t, true, game.IsPlayToWin)
	})

	t.Run("Should correctly convert PlayToWinGameOverview with winner", func(t *testing.T) {
		playToWinID := uuid.New()
		gameID := uuid.New()
		winnerID := uuid.New()

		dbGame := db.VwPlayToWinGameOverview{
			PlayToWinID:    pgtype.UUID{Bytes: playToWinID, Valid: true},
			GameID:         pgtype.UUID{Bytes: gameID, Valid: true},
			GameTitle:      "Heat",
			WinnerID:       pgtype.UUID{Bytes: winnerID, Valid: true},
			WinnerName:     pgtype.Text{String: "Alice", Valid: true},
			WinnerUniqueID: pgtype.Text{String: "alice123", Valid: true},
		}

		game := FromPlayToWinGameOverview(dbGame)

		assert.Equal(t, playToWinID, game.PlayToWinId)
		assert.Equal(t, gameID, game.GameId)
		assert.Equal(t, "Heat", game.Title)
		assert.NotNil(t, game.Winner)
		assert.Equal(t, winnerID, game.Winner.EntryId)
		assert.Equal(t, "Alice", game.Winner.EntrantName)
		assert.Equal(t, "alice123", game.Winner.EntrantUniqueId)
	})

	t.Run("Should correctly convert PlayToWinGameOverview without winner", func(t *testing.T) {
		playToWinID := uuid.New()
		gameID := uuid.New()

		dbGame := db.VwPlayToWinGameOverview{
			PlayToWinID: playToWinIDToPg(playToWinID),
			GameID:      playToWinIDToPg(gameID),
			GameTitle:   "Azul",
			WinnerID:    pgtype.UUID{Valid: false},
		}

		game := FromPlayToWinGameOverview(dbGame)

		assert.Equal(t, playToWinID, game.PlayToWinId)
		assert.Equal(t, gameID, game.GameId)
		assert.Equal(t, "Azul", game.Title)
		assert.Nil(t, game.Winner)
	})

	t.Run("Should correctly convert PlayToWinGameList", func(t *testing.T) {
		playToWinID := uuid.New()
		gameID := uuid.New()

		dbGames := []db.VwPlayToWinGameOverview{
			{
				PlayToWinID: playToWinIDToPg(playToWinID),
				GameID:      playToWinIDToPg(gameID),
				GameTitle:   "Azul",
			},
		}

		games := FromPlayToWinGameList(dbGames)

		assert.Len(t, games.Games, 1)
		assert.Equal(t, playToWinID, games.Games[0].PlayToWinId)
		assert.Equal(t, gameID, games.Games[0].GameId)
		assert.Equal(t, "Azul", games.Games[0].Title)
		assert.Nil(t, games.Games[0].Winner)
	})
}

func TestValidationUtils(t *testing.T) {
	t.Run("Should sanitize title correctly", func(t *testing.T) {
		assert.Equal(t, "catan", SanitizeTitle("Catan"))
		assert.Equal(t, "catan", SanitizeTitle("CATAN"))
		// norm.NFD check (e.g., combined characters)
		assert.Equal(t, "e", SanitizeTitle("\u0065\u0301")) // e + combining acute accent -> e (accents removed)
	})
}

func TestErrorUtils(t *testing.T) {
	t.Run("Should create ErrorResponse with details", func(t *testing.T) {
		details := []ErrorDetail{{Field: "f", Message: "m"}}
		resp := NewErrorResponseWithDetails(VALIDATIONERROR, "msg", details)

		assert.Equal(t, VALIDATIONERROR, resp.Error.Code)
		assert.Equal(t, "msg", resp.Error.Message)
		assert.Equal(t, details, resp.Error.Details)
	})

	t.Run("Should create ErrorResponse without details", func(t *testing.T) {
		resp := NewErrorResponse(NOTFOUND, "msg")

		assert.Equal(t, NOTFOUND, resp.Error.Code)
		assert.Equal(t, "msg", resp.Error.Message)
		assert.Empty(t, resp.Error.Details)
	})

	t.Run("Should create InternalError response", func(t *testing.T) {
		err := errors.New("boom")
		resp := NewInternalError(err)

		assert.Equal(t, INTERNALERROR, resp.Error.Code)
		assert.Equal(t, "boom", resp.Error.Message)
	})
}

// Helper to keep pgtype.UUID test setup compact.
func playToWinIDToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
