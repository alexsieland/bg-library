package api

import (
	"testing"

	"github.com/alexsieland/bg-library/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestStringToPgText(t *testing.T) {
	t.Run("Should return invalid pg text when input is nil", func(t *testing.T) {
		result := stringToPgText(nil)

		assert.False(t, result.Valid)
		assert.Empty(t, result.String)
	})

	t.Run("Should return valid pg text when input has value", func(t *testing.T) {
		value := "hello"

		result := stringToPgText(&value)

		assert.True(t, result.Valid)
		assert.Equal(t, value, result.String)
	})
}

func TestInt32ToPgInt4(t *testing.T) {
	t.Run("Should return invalid pg int4 when input is nil", func(t *testing.T) {
		result := int32ToPgInt4(nil)

		assert.False(t, result.Valid)
		assert.Equal(t, int32(0), result.Int32)
	})

	t.Run("Should return valid pg int4 when input has value", func(t *testing.T) {
		value := int32(42)

		result := int32ToPgInt4(&value)

		assert.True(t, result.Valid)
		assert.Equal(t, value, result.Int32)
	})
}

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
		result, errors := playToWinGameDeletionReason(nil, []ErrorDetail{})

		assert.False(t, result.Valid)
		assert.Empty(t, errors)
	})

	t.Run("Should return valid claimed reason when reason is claimed", func(t *testing.T) {
		reason := "claimed"

		result, errors := playToWinGameDeletionReason(&reason, []ErrorDetail{})

		assert.True(t, result.Valid)
		assert.Equal(t, db.PlayToWinGameDeletionTypeClaimed, result.PlayToWinGameDeletionType)
		assert.Empty(t, errors)
	})

	t.Run("Should return valid mistake reason when reason is mistake", func(t *testing.T) {
		reason := "mistake"

		result, errors := playToWinGameDeletionReason(&reason, []ErrorDetail{})

		assert.True(t, result.Valid)
		assert.Equal(t, db.PlayToWinGameDeletionTypeMistake, result.PlayToWinGameDeletionType)
		assert.Empty(t, errors)
	})

	t.Run("Should return valid other reason when reason is other", func(t *testing.T) {
		reason := "other"

		result, errors := playToWinGameDeletionReason(&reason, []ErrorDetail{})

		assert.True(t, result.Valid)
		assert.Equal(t, db.PlayToWinGameDeletionTypeOther, result.PlayToWinGameDeletionType)
		assert.Empty(t, errors)
	})

	t.Run("Should append validation error when reason is invalid", func(t *testing.T) {
		reason := "invalid_reason"
		existingErrors := []ErrorDetail{{Field: "existingField", Message: "existing message"}}

		result, errors := playToWinGameDeletionReason(&reason, existingErrors)

		assert.False(t, result.Valid)
		assert.Len(t, errors, 2)
		assert.Equal(t, "existingField", errors[0].Field)
		assert.Equal(t, "existing message", errors[0].Message)
		assert.Equal(t, "deletionReason", errors[1].Field)
		assert.Equal(t, "Must be one of: claimed, mistake, other", errors[1].Message)
	})
}
