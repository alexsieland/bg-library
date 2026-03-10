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

func int32ToPgInt4(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func uuidToPgTypeUUID(uuid uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid,
		Valid: true,
	}
}

func playToWinGameDeletionReason(deletionReason *string, errorDetails *ErrorDetails) db.NullPlayToWinGameDeletionType {
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
			errorDetails.AddErrorDetail("deletionReason", "Must be one of: claimed, mistake, other")
		}
	}

	return nullableReason
}
