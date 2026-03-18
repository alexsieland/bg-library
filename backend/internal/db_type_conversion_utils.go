package internal

import (
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
