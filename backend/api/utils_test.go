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
}

func TestValidationUtils(t *testing.T) {
	t.Run("Should validate string length correctly", func(t *testing.T) {
		// Valid
		errors := ValidateStringLength("test", "hello", 1, 10, nil)
		assert.Nil(t, errors)

		// Empty
		errors = ValidateStringLength("test", "", 1, 10, nil)
		assert.Len(t, errors, 1)
		assert.Equal(t, "Cannot be empty", errors[0].Message)

		// Too short
		errors = ValidateStringLength("test", "a", 2, 10, nil)
		assert.Len(t, errors, 1)
		assert.Contains(t, errors[0].Message, "Length must be between 2 and 10")

		// Too long
		errors = ValidateStringLength("test", "too long string", 1, 5, nil)
		assert.Len(t, errors, 1)
		assert.Contains(t, errors[0].Message, "Length must be between 1 and 5")
	})

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
