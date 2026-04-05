package api

import (
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUUIDToPgTypeUUID(t *testing.T) {
	t.Run("Should return valid pg uuid when input is provided", func(t *testing.T) {
		id := uuid.New()

		result := uuidToPgTypeUUID(id)

		assert.True(t, result.Valid)
		assert.Equal(t, [16]byte(id), result.Bytes)
	})
}

func TestPlayToWinGameDeletionReason(t *testing.T) {
	t.Run("Should return invalid nullable reason and no errors when reason is nil", func(t *testing.T) {
		var errors ErrorDetails
		result := errors.playToWinGameDeletionReason(nil)

		assert.False(t, result.Valid)
		assert.True(t, errors.Empty())
	})

	t.Run("Should return valid claimed reason when reason is claimed", func(t *testing.T) {
		reason := "claimed"

		var errors ErrorDetails
		result := errors.playToWinGameDeletionReason(&reason)

		assert.True(t, result.Valid)
		assert.Equal(t, db.PlayToWinGameDeletionTypeClaimed, result.PlayToWinGameDeletionType)
		assert.True(t, errors.Empty())
	})

	t.Run("Should return valid mistake reason when reason is mistake", func(t *testing.T) {
		reason := "mistake"

		var errors ErrorDetails
		result := errors.playToWinGameDeletionReason(&reason)

		assert.True(t, result.Valid)
		assert.Equal(t, db.PlayToWinGameDeletionTypeMistake, result.PlayToWinGameDeletionType)
		assert.True(t, errors.Empty())
	})

	t.Run("Should return valid other reason when reason is other", func(t *testing.T) {
		reason := "other"

		var errors ErrorDetails
		result := errors.playToWinGameDeletionReason(&reason)

		assert.True(t, result.Valid)
		assert.Equal(t, db.PlayToWinGameDeletionTypeOther, result.PlayToWinGameDeletionType)
		assert.True(t, errors.Empty())
	})

	t.Run("Should append validation error when reason is invalid", func(t *testing.T) {
		reason := "invalid_reason"

		errors := ErrorDetails{[]ErrorDetail{{Field: "existingField", Message: "existing message"}}}
		result := errors.playToWinGameDeletionReason(&reason)

		assert.False(t, result.Valid)
		assert.Len(t, errors.Details, 2)
		assert.Equal(t, "existingField", errors.Details[0].Field)
		assert.Equal(t, "existing message", errors.Details[0].Message)
		assert.Equal(t, "deletionReason", errors.Details[1].Field)
		assert.Equal(t, "Invalid play to win game deletion reason", errors.Details[1].Message)
	})
}
