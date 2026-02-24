package tests

import (
	"testing"
)

func TestDuplicateCheckoutWorkflow(t *testing.T) {
	// 1. Start Postgres container and Backend server

	// 2. Create two Patrons (A and B) and one Game

	// 3. Check out the Game to Patron A (Expect Success 201)

	// 4. Attempt to check out the same Game to Patron B (Expect Conflict 409)

	// 5. Verify the error response contains "already checked out" message

	// 6. Check in the Game from Patron A (Expect Success 204)

	// 7. Check out the Game to Patron B (Expect Success 201)
}
