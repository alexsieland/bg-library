package api

import (
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
)

func ConvertToOpenAPIGameStatus(dbGameStatus db.VwGameStatus) GameStatus {
	var checkedOutAt *time.Time
	if dbGameStatus.CheckoutTimestamp.Valid {
		checkedOutAt = &dbGameStatus.CheckoutTimestamp.Time
	}

	return GameStatus{
		CheckedOutAt: checkedOutAt,
		Game:         dbGameStatusToOpenAPIGame(dbGameStatus),
		Patron:       dbGameStatusToOpenAPIPatron(dbGameStatus),
	}
}

func dbGameStatusToOpenAPIGame(dbGameStatus db.VwGameStatus) Game {
	return Game{
		GameId: uuid.MustParse(dbGameStatus.GameID.String()),
		Title:  dbGameStatus.GameTitle,
	}
}

func dbGameStatusToOpenAPIPatron(dbGameStatus db.VwGameStatus) *Patron {
	if dbGameStatus.PatronID.Valid && dbGameStatus.PatronFullName.Valid {
		return &Patron{
			PatronId: uuid.MustParse(dbGameStatus.GameID.String()),
			Name:     dbGameStatus.PatronFullName.String,
		}
	}
	return nil
}
