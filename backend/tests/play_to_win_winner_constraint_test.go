package tests

import (
	"net/http"
	"net/http/httptest"
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

func TestPlayToWinWinnerMustBelongToSameGame(t *testing.T) {
	ctx := t.Context()

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

	t.Setenv("DB_HOST", host)
	t.Setenv("DB_PORT", port.Port())
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "postgres")
	t.Setenv("DB_NAME", "librarydb")
	t.Setenv("GIN_MODE", "release")

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

	createAlphaResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: "Alpha"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createAlphaResp.StatusCode())
	assert.NotNil(t, createAlphaResp.JSON201)

	createBetaResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: "Beta"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createBetaResp.StatusCode())
	assert.NotNil(t, createBetaResp.JSON201)

	_, err = client.AddPlayToWinGameByGameIdWithResponse(ctx, createAlphaResp.JSON201.GameId)
	assert.NoError(t, err)
	_, err = client.AddPlayToWinGameByGameIdWithResponse(ctx, createBetaResp.JSON201.GameId)
	assert.NoError(t, err)

	listResp, err := client.ListPlayToWinGamesWithResponse(ctx, &api.ListPlayToWinGamesParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)
	assert.Len(t, listResp.JSON200.Games, 2)

	ptwByTitle := make(map[string]api.PlayToWinGame)
	for _, game := range listResp.JSON200.Games {
		ptwByTitle[game.Title] = game
	}

	alpha := ptwByTitle["Alpha"]
	beta := ptwByTitle["Beta"]

	alphaSessionResp, err := client.AddPlayToWinSessionWithResponse(ctx, api.AddPlayToWinSessionJSONRequestBody{
		PlayToWinId: alpha.PlayToWinId,
		Entries: []struct {
			EntrantName     string `json:"entrantName"`
			EntrantUniqueId string `json:"entrantUniqueId"`
		}{
			{EntrantName: "Alice", EntrantUniqueId: "A-1"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, alphaSessionResp.StatusCode())
	assert.NotNil(t, alphaSessionResp.JSON201)
	assert.Len(t, alphaSessionResp.JSON201.PlayToWinEntries, 1)

	betaSessionResp, err := client.AddPlayToWinSessionWithResponse(ctx, api.AddPlayToWinSessionJSONRequestBody{
		PlayToWinId: beta.PlayToWinId,
		Entries: []struct {
			EntrantName     string `json:"entrantName"`
			EntrantUniqueId string `json:"entrantUniqueId"`
		}{
			{EntrantName: "Bob", EntrantUniqueId: "B-1"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, betaSessionResp.StatusCode())
	assert.NotNil(t, betaSessionResp.JSON201)

	crossGameWinnerResp, err := client.UpdatePlayToWinGameWithResponse(ctx, beta.PlayToWinId, api.UpdatePlayToWinGameJSONRequestBody{
		WinnerId: &alphaSessionResp.JSON201.PlayToWinEntries[0].EntryId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, crossGameWinnerResp.StatusCode())
	assert.NotNil(t, crossGameWinnerResp.JSON400)
	assert.NotEmpty(t, crossGameWinnerResp.JSON400.Error.Details)
	assert.Equal(t, "winnerId", crossGameWinnerResp.JSON400.Error.Details[0].Field)

	validWinnerResp, err := client.UpdatePlayToWinGameWithResponse(ctx, alpha.PlayToWinId, api.UpdatePlayToWinGameJSONRequestBody{
		WinnerId: &alphaSessionResp.JSON201.PlayToWinEntries[0].EntryId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, validWinnerResp.StatusCode())

	getAlphaResp, err := client.GetPlayToWinGameWithResponse(ctx, alpha.PlayToWinId)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getAlphaResp.StatusCode())
	assert.NotNil(t, getAlphaResp.JSON200)
	assert.NotNil(t, getAlphaResp.JSON200.Winner)
	assert.Equal(t, alphaSessionResp.JSON201.PlayToWinEntries[0].EntryId, getAlphaResp.JSON200.Winner.EntryId)

	getBetaResp, err := client.GetPlayToWinGameWithResponse(ctx, beta.PlayToWinId)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getBetaResp.StatusCode())
	assert.NotNil(t, getBetaResp.JSON200)
	assert.Nil(t, getBetaResp.JSON200.Winner)
}
