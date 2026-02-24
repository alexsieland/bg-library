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

func TestGameSearchWorkflow(t *testing.T) {
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

	// 2. Create multiple games: "Catan", "Catan: Seafarers", "Gloomhaven"
	gamesToCreate := []string{"Catan", "Catan: Seafarers", "Gloomhaven"}
	for _, title := range gamesToCreate {
		resp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: title})
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode())
	}

	// 3. Search for "Catan" via ListGames(title="Catan")
	catanSearch := "Catan"
	listResp, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		Title: &catanSearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	// 4. Verify that exactly "Catan" and "Catan: Seafarers" are returned
	assert.NotNil(t, listResp.JSON200)
	assert.Len(t, listResp.JSON200.Games, 2)

	titles := make(map[string]bool)
	for _, gs := range listResp.JSON200.Games {
		titles[gs.Game.Title] = true
	}
	assert.True(t, titles["Catan"])
	assert.True(t, titles["Catan: Seafarers"])

	// 5. Search for "Gloom" via ListGames(title="Gloom")
	gloomSearch := "Gloom"
	listResp2, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		Title: &gloomSearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp2.StatusCode())

	// 6. Verify that exactly "Gloomhaven" is returned
	assert.NotNil(t, listResp2.JSON200)
	assert.Len(t, listResp2.JSON200.Games, 1)
	assert.Equal(t, "Gloomhaven", listResp2.JSON200.Games[0].Game.Title)

	// 7. Search for "Monopoly" via ListGames(title="Monopoly")
	monopolySearch := "Monopoly"
	listResp3, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		Title: &monopolySearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp3.StatusCode())

	// 8. Verify that an empty list is returned
	assert.NotNil(t, listResp3.JSON200)
	assert.Len(t, listResp3.JSON200.Games, 0)

	// 9. Test accent folding: Create "Bärenpark", search for "barenpark"
	barenparkTitle := "Bärenpark"
	respBP, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: barenparkTitle})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, respBP.StatusCode())

	bpSearch := "barenpark"
	listRespBP, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		Title: &bpSearch,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listRespBP.StatusCode())
	assert.NotNil(t, listRespBP.JSON200)
	assert.Len(t, listRespBP.JSON200.Games, 1)
	assert.Equal(t, barenparkTitle, listRespBP.JSON200.Games[0].Game.Title)
}
