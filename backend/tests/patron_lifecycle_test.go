package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPatronLifecycleWorkflow(t *testing.T) {
	ctx := t.Context()

	// 1. Start Postgres container and Backend server
	schemaPath, err := filepath.Abs("../db/schema.sql")
	assert.NoError(t, err)

	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("librarydb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(schemaPath),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, pgContainer.Terminate(ctx))
	}()

	host, err := pgContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "librarydb")
	os.Setenv("GIN_MODE", "release")

	server := api.NewServer()
	err = server.Database.Connect()
	assert.NoError(t, err)
	defer server.Database.Close()

	r := gin.New()
	api.RegisterHandlers(r, server)
	ts := httptest.NewServer(r)
	defer ts.Close()

	client, err := api.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)

	// 2. Create a Patron
	name := "Original Name"
	createResp, err := client.AddPatronWithResponse(ctx, api.AddPatronJSONRequestBody{
		Name: name,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	patron := createResp.JSON201
	patronID := patron.PatronId

	// 3. Update the Patron's name via UpdatePatron
	newName := "Updated Name"
	updateResp, err := client.UpdatePatronWithResponse(ctx, patronID.String(), api.UpdatePatronJSONRequestBody{
		Name: newName,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, updateResp.StatusCode())

	// 4. Verify the name change via GetPatron
	getResp, err := client.GetPatronWithResponse(ctx, patronID.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.Equal(t, newName, getResp.JSON200.Name)

	// 5. Delete the Patron via DeletePatron
	deleteResp, err := client.DeletePatronWithResponse(ctx, patronID.String())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode())

	// 6. Verify the Patron no longer appears in ListPatrons
	listResp, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	found := false
	for _, p := range listResp.JSON200.Patrons {
		if p.PatronId == patronID {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted patron should not be in the list")

	// 7. Verify that searching for the deleted Patron's name via ListPatrons returns no results
	listRespSearch, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{
		Name: &newName,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listRespSearch.StatusCode())
	assert.Len(t, listRespSearch.JSON200.Patrons, 0, "Search should return no results for deleted patron")
}
