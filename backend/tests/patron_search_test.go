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

func TestPatronSearchWorkflow(t *testing.T) {
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

	// 2. Create multiple patrons: "Alice Smith", "Bob Smith", "Charlie Brown"
	patronsToCreate := []string{"Alice Smith", "Bob Smith", "Charlie Brown"}
	for _, name := range patronsToCreate {
		resp, err := client.AddPatronWithResponse(ctx, api.AddPatronJSONRequestBody{Name: name})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode())
	}

	// 3. Search for "Smith" via ListPatrons(name="Smith")
	smithSearch := "Smith"
	listResp, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{
		Name: &smithSearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	// 4. Verify that exactly "Alice Smith" and "Bob Smith" are returned
	assert.NotNil(t, listResp.JSON200)
	assert.Len(t, listResp.JSON200.Patrons, 2)

	names := make(map[string]bool)
	for _, p := range listResp.JSON200.Patrons {
		names[p.Name] = true
	}
	assert.True(t, names["Alice Smith"])
	assert.True(t, names["Bob Smith"])

	// 5. Search for "Charlie" via ListPatrons(name="Charlie")
	charlieSearch := "Charlie"
	listResp2, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{
		Name: &charlieSearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp2.StatusCode())

	// 6. Verify that exactly "Charlie Brown" is returned
	assert.NotNil(t, listResp2.JSON200)
	assert.Len(t, listResp2.JSON200.Patrons, 1)
	assert.Equal(t, "Charlie Brown", listResp2.JSON200.Patrons[0].Name)

	// 7. Search for "Zoe" via ListPatrons(name="Zoe") (Expected: Empty list)
	zoeSearch := "Zoe"
	listResp3, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{
		Name: &zoeSearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp3.StatusCode())

	// 8. Verify that an empty list is returned
	assert.NotNil(t, listResp3.JSON200)
	assert.Len(t, listResp3.JSON200.Patrons, 0)
}
