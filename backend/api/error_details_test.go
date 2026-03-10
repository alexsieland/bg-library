package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorDetails(t *testing.T) {
	t.Run("Should validate string length correctly", func(t *testing.T) {
		// Valid
		errorDetails := ErrorDetails{}
		errorDetails.ValidateStringLength("test", "hello", 1, 10)
		assert.Nil(t, errorDetails.Details)

		// Empty
		errorDetails = ErrorDetails{}
		errorDetails.ValidateStringLength("test", "", 1, 10)
		assert.Len(t, errorDetails.Details, 1)
		assert.Equal(t, "Cannot be empty", errorDetails.Details[0].Message)

		// Too short
		errorDetails = ErrorDetails{}
		errorDetails.ValidateStringLength("test", "a", 2, 10)
		assert.Len(t, errorDetails.Details, 1)
		assert.Contains(t, errorDetails.Details[0].Message, "Length must be between 2 and 10")

		// Too long
		errorDetails = ErrorDetails{}
		errorDetails.ValidateStringLength("test", "too long string", 1, 5)
		assert.Len(t, errorDetails.Details, 1)
		assert.Contains(t, errorDetails.Details[0].Message, "Length must be between 1 and 5")
	})
}
