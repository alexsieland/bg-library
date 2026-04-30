package api

import (
	"testing"
	time "time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromVwLibraryGames_ListAndEmpty(t *testing.T) {
	t.Run("empty slice returns empty list", func(t *testing.T) {
		out := FromVwLibraryGames([]db.VwLibraryGame{})
		require.NotNil(t, out.Games)
		assert.Len(t, out.Games, 0)
	})

	t.Run("maps multiple items one-to-one", func(t *testing.T) {
		id1 := uuid.New()
		id2 := uuid.New()

		dbGames := []db.VwLibraryGame{
			{
				ID:              pgtype.UUID{Bytes: id1, Valid: true},
				Title:           "G1",
				Barcode:         pgtype.Text{String: "B1", Valid: true},
				PlayToWinGameID: pgtype.UUID{Bytes: uuid.Nil, Valid: false},
			},
			{
				ID:              pgtype.UUID{Bytes: id2, Valid: true},
				Title:           "G2",
				Barcode:         pgtype.Text{Valid: false},
				PlayToWinGameID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			},
		}

		out := FromVwLibraryGames(dbGames)
		assert.Len(t, out.Games, 2)
		assert.Equal(t, id1, out.Games[0].GameId)
		assert.Equal(t, "G1", out.Games[0].Title)
		require.NotNil(t, out.Games[0].Barcode)
		assert.Equal(t, "B1", *out.Games[0].Barcode)

		assert.Equal(t, id2, out.Games[1].GameId)
		assert.Equal(t, "G2", out.Games[1].Title)
		assert.Nil(t, out.Games[1].Barcode)
		assert.True(t, out.Games[1].IsPlayToWin)
	})
}

func TestFromVwLibraryGame_BarcodeAndPlayToWin(t *testing.T) {
	id := uuid.New()
	dbGameWithBarcode := db.VwLibraryGame{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		Title:           "HasBarcode",
		Barcode:         pgtype.Text{String: "123", Valid: true},
		PlayToWinGameID: pgtype.UUID{Valid: false},
	}

	g1 := FromVwLibraryGame(dbGameWithBarcode)
	require.NotNil(t, g1.Barcode)
	assert.Equal(t, "123", *g1.Barcode)
	assert.False(t, g1.IsPlayToWin)

	dbGameNoBarcode := db.VwLibraryGame{
		ID:              pgtype.UUID{Bytes: id, Valid: true},
		Title:           "NoBarcode",
		Barcode:         pgtype.Text{Valid: false},
		PlayToWinGameID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
	}
	g2 := FromVwLibraryGame(dbGameNoBarcode)
	assert.Nil(t, g2.Barcode)
	assert.True(t, g2.IsPlayToWin)
}

func TestFromGame_BarcodeAndIsPlayToWin(t *testing.T) {
	id := uuid.New()
	dbGame := db.Game{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		Title:   "G",
		Barcode: pgtype.Text{String: "abc", Valid: true},
	}

	g := FromGame(dbGame, true)
	require.NotNil(t, g.Barcode)
	assert.Equal(t, "abc", *g.Barcode)
	assert.True(t, g.IsPlayToWin)

	dbGame.Barcode = pgtype.Text{Valid: false}
	g2 := FromGame(dbGame, false)
	assert.Nil(t, g2.Barcode)
	assert.False(t, g2.IsPlayToWin)
}

func TestDbGameStatusToOpenAPIGame_PlayToWinToggle(t *testing.T) {
	id := uuid.New()
	dbgs := db.VwGameStatus{
		GameID:    pgtype.UUID{Bytes: id, Valid: true},
		GameTitle: "T",
		PtwGameID: pgtype.UUID{Valid: false},
	}
	g := dbGameStatusToOpenAPIGame(dbgs)
	assert.False(t, g.IsPlayToWin)

	dbgs.PtwGameID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	g2 := dbGameStatusToOpenAPIGame(dbgs)
	assert.True(t, g2.IsPlayToWin)
}

func TestFromVwGameStatus_CheckedInAndOut(t *testing.T) {
	t.Run("checked-out scenario", func(t *testing.T) {
		gameID := uuid.New()
		patronID := uuid.New()
		now := requireTimeNow(t)

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

	t.Run("checked-in scenario returns nils", func(t *testing.T) {
		gameID := uuid.New()

		dbStatus := db.VwGameStatus{
			GameID:           pgtype.UUID{Bytes: gameID, Valid: true},
			GameTitle:        "Catan",
			CheckinTimestamp: pgtype.Timestamp{Valid: true},
		}

		status := FromVwGameStatus(dbStatus)

		assert.Equal(t, gameID, status.Game.GameId)
		assert.Nil(t, status.Patron)
		assert.Nil(t, status.CheckedOutAt)
		assert.Nil(t, status.TransactionId)
	})
}

func TestFromVwLibraryPatronAndPatronConversions(t *testing.T) {
	t.Run("FromVwLibraryPatron barcode handling", func(t *testing.T) {
		patronID := uuid.New()
		dbPatron := db.VwLibraryPatron{
			ID:       pgtype.UUID{Bytes: patronID, Valid: true},
			FullName: "John Doe",
			Barcode:  pgtype.Text{String: "B-1", Valid: true},
		}

		patron := FromVwLibraryPatron(dbPatron)
		assert.Equal(t, patronID, patron.PatronId)
		assert.Equal(t, "John Doe", patron.Name)
		requireNotNilStr(t, patron.Barcode)
		assert.Equal(t, "B-1", *patron.Barcode)
	})

	t.Run("FromPatron barcode handling", func(t *testing.T) {
		patronID := uuid.New()
		dbPatron := db.Patron{
			ID:       pgtype.UUID{Bytes: patronID, Valid: true},
			FullName: "Jane",
			Barcode:  pgtype.Text{Valid: false},
		}

		patron := FromPatron(dbPatron)
		assert.Equal(t, patronID, patron.PatronId)
		assert.Equal(t, "Jane", patron.Name)
		assert.Nil(t, patron.Barcode)
	})
}

func TestFromTransactionConversion(t *testing.T) {
	gameID := uuid.New()
	patronID := uuid.New()
	transID := uuid.New()
	now := requireTimeNow(t)

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
}

func TestFromPlayToWinGameOverview_WinnerMapping(t *testing.T) {
	wid := uuid.New()
	dbPTW := db.VwPlayToWinGameOverview{
		WinnerID:       pgtype.UUID{Bytes: wid, Valid: true},
		WinnerName:     pgtype.Text{String: "Alice", Valid: true},
		WinnerUniqueID: pgtype.Text{String: "A-1", Valid: true},
		PtwGameID:      pgtype.UUID{Bytes: uuid.New(), Valid: true},
		GameID:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
		GameTitle:      "G",
	}
	out := FromPlayToWinGameOverview(dbPTW)
	require.NotNil(t, out.Winner)
	assert.Equal(t, wid, out.Winner.EntryId)
	assert.Equal(t, "Alice", out.Winner.EntrantName)
}

func TestFromPlayToWinGameOverview_NoWinner(t *testing.T) {
	ptwGameId := uuid.New()
	gameID := uuid.New()

	dbGame := db.VwPlayToWinGameOverview{
		PtwGameID: ptwGameIdToPg(ptwGameId),
		GameID:    ptwGameIdToPg(gameID),
		GameTitle: "Azul",
		WinnerID:  pgtype.UUID{Valid: false},
	}

	game := FromPlayToWinGameOverview(dbGame)

	assert.Equal(t, ptwGameId, game.PlayToWinId)
	assert.Equal(t, gameID, game.GameId)
	assert.Equal(t, "Azul", game.Title)
	assert.Nil(t, game.Winner)
}

func TestFromPlayToWinGameList(t *testing.T) {
	ptwGameId := uuid.New()
	ptwGroupID := uuid.New()
	gameID := uuid.New()

	dbGames := []db.VwPlayToWinGameOverview{
		{
			PtwGameID:  ptwGameIdToPg(ptwGameId),
			PtwGroupID: ptwGameIdToPg(ptwGroupID),
			GameID:     ptwGameIdToPg(gameID),
			GameTitle:  "Azul",
		},
	}

	games := FromPlayToWinGameList(dbGames)

	assert.Len(t, games.Games, 1)
	assert.Equal(t, ptwGameId, games.Games[0].PlayToWinId)
	assert.Equal(t, gameID, games.Games[0].GameId)
	assert.Equal(t, "Azul", games.Games[0].Title)
	assert.Nil(t, games.Games[0].Winner)
}

func TestFromDeletedPlayToWinGameOverview_NoWinner(t *testing.T) {
	ptwID := uuid.New()
	gameID := uuid.New()

	dbGame := db.VwDeletedPlayToWinGameOverview{
		PtwGameID: pgtype.UUID{Bytes: ptwID, Valid: true},
		GameID:    pgtype.UUID{Bytes: gameID, Valid: true},
		GameTitle: "Deleted Game",
		WinnerID:  pgtype.UUID{Valid: false},
	}

	game := FromDeletedPlayToWinGameOverview(dbGame)

	assert.Equal(t, ptwID, game.PlayToWinId)
	assert.Equal(t, gameID, game.GameId)
	assert.Equal(t, "Deleted Game", game.Title)
	assert.Nil(t, game.Winner)
}

func TestFromDeletedPlayToWinGameOverview_WithWinner(t *testing.T) {
	ptwID := uuid.New()
	gameID := uuid.New()
	winnerID := uuid.New()

	dbGame := db.VwDeletedPlayToWinGameOverview{
		PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
		GameID:         pgtype.UUID{Bytes: gameID, Valid: true},
		GameTitle:      "Claimed Game",
		WinnerID:       pgtype.UUID{Bytes: winnerID, Valid: true},
		WinnerName:     pgtype.Text{String: "Bob", Valid: true},
		WinnerUniqueID: pgtype.Text{String: "bob-1", Valid: true},
	}

	game := FromDeletedPlayToWinGameOverview(dbGame)

	assert.Equal(t, ptwID, game.PlayToWinId)
	assert.Equal(t, gameID, game.GameId)
	assert.Equal(t, "Claimed Game", game.Title)
	require.NotNil(t, game.Winner)
	assert.Equal(t, winnerID, game.Winner.EntryId)
	assert.Equal(t, "Bob", game.Winner.EntrantName)
	assert.Equal(t, "bob-1", game.Winner.EntrantUniqueId)
}

func TestFromDeletedPlayToWinGameOverview_WinnerWithNullNameFields(t *testing.T) {
	ptwID := uuid.New()
	winnerID := uuid.New()

	dbGame := db.VwDeletedPlayToWinGameOverview{
		PtwGameID:      pgtype.UUID{Bytes: ptwID, Valid: true},
		GameID:         pgtype.UUID{Valid: true},
		GameTitle:      "Mystery",
		WinnerID:       pgtype.UUID{Bytes: winnerID, Valid: true},
		WinnerName:     pgtype.Text{Valid: false},
		WinnerUniqueID: pgtype.Text{Valid: false},
	}

	game := FromDeletedPlayToWinGameOverview(dbGame)

	require.NotNil(t, game.Winner)
	assert.Equal(t, "", game.Winner.EntrantName)
	assert.Equal(t, "", game.Winner.EntrantUniqueId)
}

func TestFromDeletedPlayToWinGameList(t *testing.T) {
	ptwID1, ptwID2 := uuid.New(), uuid.New()
	gameID1, gameID2 := uuid.New(), uuid.New()

	dbGames := []db.VwDeletedPlayToWinGameOverview{
		{
			PtwGameID: pgtype.UUID{Bytes: ptwID1, Valid: true},
			GameID:    pgtype.UUID{Bytes: gameID1, Valid: true},
			GameTitle: "Alpha",
			WinnerID:  pgtype.UUID{Valid: false},
		},
		{
			PtwGameID: pgtype.UUID{Bytes: ptwID2, Valid: true},
			GameID:    pgtype.UUID{Bytes: gameID2, Valid: true},
			GameTitle: "Beta",
			WinnerID:  pgtype.UUID{Valid: false},
		},
	}

	list := FromDeletedPlayToWinGameList(dbGames)

	assert.Len(t, list.Games, 2)
	assert.Equal(t, ptwID1, list.Games[0].PlayToWinId)
	assert.Equal(t, "Alpha", list.Games[0].Title)
	assert.Equal(t, ptwID2, list.Games[1].PlayToWinId)
	assert.Equal(t, "Beta", list.Games[1].Title)
}

// requireTimeNow returns a time suitable for tests and fails the test on error.
func requireTimeNow(t *testing.T) (nowTime time.Time) {
	t.Helper()
	nowTime = time.Now().UTC()
	return
}

// requireNotNilStr asserts a pointer to string is non-nil for clearer test code.
func requireNotNilStr(t *testing.T, s *string) {
	t.Helper()
	require.NotNil(t, s)
}
