package api

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func isForeignKeyConstraintViolation(err error) bool {
	var pgErr *pgconn.PgError
	// 23503 is the error code for a foreign key constraint violation
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return true
	}
	return false
}

func isUniqueConstraintViolation(err error) bool {
	var pgErr *pgconn.PgError
	// 23505 is the error code for a unique constraint violation
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {

		return true
	}
	return false
}
