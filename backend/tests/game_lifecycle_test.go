package tests

import (
	"testing"
)

func TestGameLifecycleWorkflow(t *testing.T) {
	// 1. Start Postgres container and Backend server

	// 2. Create a Game

	// 3. Update the Game's title via UpdateGame

	// 4. Verify the title change via GetGame

	// 5. Delete the Game via DeleteGame

	// 6. Verify the Game no longer appears in ListGames

	// 7. Verify that attempting to GetGame by the deleted Game's ID returns 404 Not Found
}
