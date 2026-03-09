package api

import (
	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func stringToPgText(str *string) pgtype.Text {
	if str == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *str, Valid: true}
}

func stringToPgTypeUUID(fieldName string, str string, errorDetails []ErrorDetail) (pgtype.UUID, []ErrorDetail) {
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

func uuidToPgTypeUUID(uuid uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}
}

func playToWinGameDeletionReason(deletionReason *string, errorDetails []ErrorDetail) (db.NullPlayToWinGameDeletionType, []ErrorDetail) {
	nullableReason := db.NullPlayToWinGameDeletionType{Valid: false}
	if deletionReason != nil {
		reason := db.PlayToWinGameDeletionType(*deletionReason)
		switch reason {
		case db.PlayToWinGameDeletionTypeClaimed, db.PlayToWinGameDeletionTypeMistake, db.PlayToWinGameDeletionTypeOther:
			nullableReason = db.NullPlayToWinGameDeletionType{
				PlayToWinGameDeletionType: reason,
				Valid:                     true,
			}
		default:
			errorDetails = append(errorDetails, ErrorDetail{
				Field:   "deletionReason",
				Message: "Must be one of: claimed, mistake, other",
			})
		}
	}

	return nullableReason, errorDetails
}
