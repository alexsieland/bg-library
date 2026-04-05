package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationUtils(t *testing.T) {
	t.Run("Should sanitize title correctly", func(t *testing.T) {
		assert.Equal(t, "catan", SanitizeTitle("Catan"))
		assert.Equal(t, "catan", SanitizeTitle("CATAN"))
		// norm.NFD check (e.g., combined characters)
		assert.Equal(t, "e", SanitizeTitle("\u0065\u0301")) // e + combining acute accent -> e (accents removed)
	})
}
