package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestIsForeignKeyConstraintViolation(t *testing.T) {
	t.Run("Should return true when error is a foreign key constraint violation", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.True(t, isForeignKeyConstraintViolation(err))
	})

	t.Run("Should return true when foreign key constraint violation is wrapped", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23503"})
		assert.True(t, isForeignKeyConstraintViolation(err))
	})

	t.Run("Should return false when error is a different pg error code", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		assert.False(t, isForeignKeyConstraintViolation(err))
	})

	t.Run("Should return false when error is not a pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.False(t, isForeignKeyConstraintViolation(err))
	})
}

func TestIsUniqueConstraintViolation(t *testing.T) {
	t.Run("Should return true when error is a unique constraint violation", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505"}
		assert.True(t, isUniqueConstraintViolation(err))
	})

	t.Run("Should return true when unique constraint violation is wrapped", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", &pgconn.PgError{Code: "23505"})
		assert.True(t, isUniqueConstraintViolation(err))
	})

	t.Run("Should return false when error is a different pg error code", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23503"}
		assert.False(t, isUniqueConstraintViolation(err))
	})

	t.Run("Should return false when error is not a pg error", func(t *testing.T) {
		err := errors.New("generic error")
		assert.False(t, isUniqueConstraintViolation(err))
	})
}
