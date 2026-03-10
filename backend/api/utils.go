package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oapi-codegen/runtime/types"
	"golang.org/x/text/unicode/norm"
)

//Conversion Utils

func pgUUIDToUUID(pgUUID pgtype.UUID) uuid.UUID {
	return pgUUID.Bytes
}

func FromVwGameStatus(dbGameStatus db.VwGameStatus) GameStatus {
	var checkedOutAt *time.Time
	if dbGameStatus.CheckoutTimestamp.Valid {
		checkedOutAt = &dbGameStatus.CheckoutTimestamp.Time
	}

	var transactionId *types.UUID
	if dbGameStatus.TransactionID.Valid {
		id := types.UUID(dbGameStatus.TransactionID.Bytes)
		transactionId = &id
	}

	// If the game is currently checked in, we return the game status with a null checkedOutAt, patron, and transaction ID.
	// This is because the game is no longer checked out.
	if dbGameStatus.CheckinTimestamp.Valid {
		return GameStatus{
			CheckedOutAt:  nil,
			Game:          dbGameStatusToOpenAPIGame(dbGameStatus),
			Patron:        nil,
			TransactionId: nil,
		}
	}

	return GameStatus{
		CheckedOutAt:  checkedOutAt,
		Game:          dbGameStatusToOpenAPIGame(dbGameStatus),
		Patron:        dbGameStatusToOpenAPIPatron(dbGameStatus),
		TransactionId: transactionId,
	}
}

func dbGameStatusToOpenAPIGame(dbGameStatus db.VwGameStatus) Game {
	return Game{
		GameId:      pgUUIDToUUID(dbGameStatus.GameID),
		Title:       dbGameStatus.GameTitle,
		IsPlayToWin: dbGameStatus.PlayToWinGameID.Valid,
	}
}

func dbGameStatusToOpenAPIPatron(dbGameStatus db.VwGameStatus) *Patron {
	if dbGameStatus.PatronID.Valid && dbGameStatus.PatronFullName.Valid {
		return &Patron{
			PatronId: pgUUIDToUUID(dbGameStatus.PatronID),
			Name:     dbGameStatus.PatronFullName.String,
		}
	}
	return nil
}

func FromVwLibraryPatron(dbPatron db.VwLibraryPatron) Patron {
	patron := Patron{
		PatronId: pgUUIDToUUID(dbPatron.ID),
		Name:     dbPatron.FullName,
	}
	if dbPatron.Barcode.Valid {
		patron.Barcode = &dbPatron.Barcode.String
	}
	return patron
}

func FromPatron(dbPatron db.Patron) Patron {
	patron := Patron{
		PatronId: pgUUIDToUUID(dbPatron.ID),
		Name:     dbPatron.FullName,
	}
	if dbPatron.Barcode.Valid {
		patron.Barcode = &dbPatron.Barcode.String
	}
	return patron
}

func FromVwLibraryGame(dbGame db.VwLibraryGame) Game {
	game := Game{
		GameId: pgUUIDToUUID(dbGame.ID),
		Title:  dbGame.Title,
	}
	if dbGame.Barcode.Valid {
		game.Barcode = &dbGame.Barcode.String
	}
	return game
}

func FromVwLibraryGames(dbGame []db.VwLibraryGame) GameList {
	gameList := make([]Game, len(dbGame))
	for i, dbGame := range dbGame {
		gameList[i] = FromVwLibraryGame(dbGame)
	}
	return GameList{Games: gameList}
}

func FromTransaction(dbTransaction db.Transaction) LibraryTransaction {
	return LibraryTransaction{
		GameId:    pgUUIDToUUID(dbTransaction.GameID),
		Id:        pgUUIDToUUID(dbTransaction.ID),
		PatronId:  pgUUIDToUUID(dbTransaction.PatronID),
		Timestamp: dbTransaction.CheckoutTimestamp.Time,
	}
}

func FromGame(dbGame db.Game, isPlayToWin bool) Game {
	game := Game{
		GameId:      pgUUIDToUUID(dbGame.ID),
		Title:       dbGame.Title,
		IsPlayToWin: isPlayToWin,
	}
	if dbGame.Barcode.Valid {
		game.Barcode = &dbGame.Barcode.String
	}
	return game
}

// Validation Utils

func ValidateStringLength(fieldName string, str string, minLength int, maxLength int, errorDetails []ErrorDetail) []ErrorDetail {
	if minLength > 0 && str == "" {
		return append(errorDetails, ErrorDetail{
			Field:   fieldName,
			Message: "Cannot be empty",
		})
	}
	if len(str) < minLength || len(str) > maxLength {
		return append(errorDetails, ErrorDetail{
			Field:   fieldName,
			Message: "Length must be between " + strconv.Itoa(minLength) + " and " + strconv.Itoa(maxLength),
		})
	}
	return nil
}

func ValidateIntMin(fieldName string, i int32, minVal int32, errorDetails []ErrorDetail) []ErrorDetail {
	if i < minVal {
		return append(errorDetails, ErrorDetail{
			Field:   fieldName,
			Message: "Must be greater than or equal to " + strconv.Itoa(int(minVal)),
		})
	}
	return errorDetails
}

func ValidateIntMax(fieldName string, i int32, maxVal int32, errorDetails []ErrorDetail) []ErrorDetail {
	if i > maxVal {
		return append(errorDetails, ErrorDetail{
			Field:   fieldName,
			Message: "Must be less than or equal to " + strconv.Itoa(int(maxVal)),
		})
	}
	return errorDetails
}

func SanitizeTitle(title string) string {
	t := norm.NFD.String(strings.ToLower(title))
	var result strings.Builder
	for _, r := range t {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == ' ' || r == ':' || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Error Utils

func NewErrorResponseWithDetails(errorCode ErrorResponseErrorCode, message string, details []ErrorDetail) ErrorResponse {
	resp := ErrorResponse{}
	resp.Error.Code = errorCode
	resp.Error.Message = message
	resp.Error.Details = details
	return resp
}

func NewErrorResponse(errorCode ErrorResponseErrorCode, message string) ErrorResponse {
	return NewErrorResponseWithDetails(errorCode, message, []ErrorDetail{})
}

func NewInternalError(err error) ErrorResponse {
	return NewErrorResponse(INTERNALERROR, err.Error())
}
