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

func TestGameLifecycleWorkflow(t *testing.T) {
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

	// 2. Create a Game
	originalTitle := "Wingspan"
	createResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{
		Title: originalTitle,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createResp.StatusCode())
	gameID := createResp.JSON201.GameId

	// 3. Update the Game's title via UpdateGame
	updatedTitle := "Wingspan (Asia Expansion)"
	updateResp, err := client.UpdateGameWithResponse(ctx, gameID, api.UpdateGameJSONRequestBody{
		Title: updatedTitle,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, updateResp.StatusCode())

	// 4. Verify the title change via GetGame
	getResp, err := client.GetGameWithResponse(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode())
	assert.Equal(t, updatedTitle, getResp.JSON200.Title)

	// 5. Delete the Game via DeleteGame
	deleteResp, err := client.DeleteGameWithResponse(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode())

	// 6. Verify the Game no longer appears in ListGames
	listResp, err := client.ListGamesWithResponse(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())

	found := false
	for _, gs := range listResp.JSON200.Games {
		if gs.Game.GameId == gameID {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted game should not be in the list")

	// 7. Verify that attempting to GetGame by the deleted Game's ID returns 404 Not Found
	getResp2, err := client.GetGameWithResponse(ctx, gameID)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, getResp2.StatusCode())
}
