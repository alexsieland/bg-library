package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorUtils(t *testing.T) {
	t.Run("Should create ErrorResponse with details", func(t *testing.T) {
		details := []ErrorDetail{{Field: "f", Message: "m"}}
		resp := NewErrorResponseWithDetails(VALIDATIONERROR, "msg", details)

		assert.Equal(t, VALIDATIONERROR, resp.Error.Code)
		assert.Equal(t, "msg", resp.Error.Message)
		assert.Equal(t, details, resp.Error.Details)
	})

	t.Run("Should create ErrorResponse without details", func(t *testing.T) {
		resp := NewErrorResponse(NOTFOUND, "msg")

		assert.Equal(t, NOTFOUND, resp.Error.Code)
		assert.Equal(t, "msg", resp.Error.Message)
		assert.Empty(t, resp.Error.Details)
	})

	t.Run("Should create InternalError response", func(t *testing.T) {
		err := errors.New("boom")
		resp := NewInternalError(err)

		assert.Equal(t, INTERNALERROR, resp.Error.Code)
		assert.Equal(t, "boom", resp.Error.Message)
	})
}
