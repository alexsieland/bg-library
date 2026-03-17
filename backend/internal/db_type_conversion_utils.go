package internal

import (
	"github.com/alexsieland/bg-library/db"
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

func playToWinGameDeletionReason(deletionReason *string) (db.NullPlayToWinGameDeletionType, error) {
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
			return db.NullPlayToWinGameDeletionType{}, ErrInvalidInput
		}
	}

	return nullableReason, nil
}
