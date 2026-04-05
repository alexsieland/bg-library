package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToPgText_Internal(t *testing.T) {
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

func TestInt32ToPgInt4_Internal(t *testing.T) {
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
