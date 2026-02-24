package tests

import (
	"testing"
)

func TestPatronLifecycleWorkflow(t *testing.T) {
	// 1. Start Postgres container and Backend server

	// 2. Create a Patron

	// 3. Update the Patron's name via UpdatePatron

	// 4. Verify the name change via GetPatron

	// 5. Delete the Patron via DeletePatron

	// 6. Verify the Patron no longer appears in ListPatrons

	// 7. Verify that searching for the deleted Patron's name via ListPatrons returns no results
}
