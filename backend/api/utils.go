package api

import (
	"strings"
	"time"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/unicode/norm"
)

//Conversion Utils

func FromVwGameStatus(dbGameStatus db.VwGameStatus) GameStatus {
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

func FromVwLibraryPatron(dbPatron db.VwLibraryPatron) Patron {
	return Patron{
		PatronId: uuid.MustParse(dbPatron.ID.String()),
		Name:     dbPatron.FullName,
	}
}

func FromPatron(dbPatron db.Patron) Patron {
	return Patron{
		PatronId: uuid.MustParse(dbPatron.ID.String()),
		Name:     dbPatron.FullName,
	}
}

func FromVwLibraryGame(dbGame db.VwLibraryGame) Game {
	return Game{
		GameId: uuid.MustParse(dbGame.ID.String()),
		Title:  dbGame.Title,
	}
}

func FromTransaction(dbTransaction db.Transaction) LibraryTransaction {
	return LibraryTransaction{
		GameId:    uuid.MustParse(dbTransaction.GameID.String()),
		Id:        uuid.MustParse(dbTransaction.ID.String()),
		PatronId:  uuid.MustParse(dbTransaction.PatronID.String()),
		Timestamp: dbTransaction.CheckoutTimestamp.Time,
	}
}

func FromGame(dbGame db.Game) Game {
	return Game{
		GameId: uuid.MustParse(dbGame.ID.String()),
		Title:  dbGame.Title,
	}
}

// Validation Utils

func ConvertToPgTypeUUID(fieldName string, str string, errorDetails []ErrorDetail) (pgtype.UUID, []ErrorDetail) {
	dbUuid, err := uuid.Parse(str)
	if err != nil {
		return pgtype.UUID{}, append(errorDetails, ErrorDetail{
			Field:   fieldName,
			Message: "Invalid UUID format",
		})
	}
	return pgtype.UUID{
		Bytes: dbUuid,
		Valid: true,
	}, errorDetails
}

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
			Message: "Length must be between " + string(minLength) + " and " + string(maxLength),
		})
	}
	return nil
}

func SanitizeTitle(title string) string {
	return norm.NFC.String(strings.ToLower(title))
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
