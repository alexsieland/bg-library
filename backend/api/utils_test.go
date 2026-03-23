package api

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Helper to keep pgtype.UUID test setup compact.
func ptwGameIdToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
