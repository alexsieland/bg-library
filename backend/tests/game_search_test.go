package tests

import (
	"testing"
)

func TestGameSearchWorkflow(t *testing.T) {
	// 1. Start Postgres container and Backend server

	// 2. Create multiple games: "Catan", "Catan: Seafarers", "Gloomhaven"

	// 3. Search for "Catan" via ListGames(title="Catan")

	// 4. Verify that exactly "Catan" and "Catan: Seafarers" are returned

	// 5. Search for "Gloom" via ListGames(title="Gloom")

	// 6. Verify that exactly "Gloomhaven" is returned

	// 7. Search for "Monopoly" via ListGames(title="Monopoly")

	// 8. Verify that an empty list is returned
}
