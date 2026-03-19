package internal

import (
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func SanitizeTitle(title string) string {
	t := norm.NFD.String(strings.ToLower(title))
	var result strings.Builder
	for _, r := range t {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == ' ' || r == ':' || r == '-' || r == '%' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GenerateDBRegexString Formats a string to be used in a LIKE query in the database.
func GenerateDBRegexString(s string) string {
	if s == "" {
		return ""
	}
	// trim spaces and %
	trimmed := strings.TrimFunc(s, func(r rune) bool {
		return r == ' ' || r == '%'
	})
	return fmt.Sprintf("%%%s%%", trimmed)
}
