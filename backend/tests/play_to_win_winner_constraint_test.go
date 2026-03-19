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

func TestPlayToWinWinnerMustBelongToSourceOrDuplicateGame(t *testing.T) {
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
	err = server.LibService.Start()
	assert.NoError(t, err)
	defer server.LibService.Stop()

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

	// Same title as Alpha to force creation of a duplicate PTW game (ref_id points to first Alpha PTW)
	createAlphaDuplicateResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: "Alpha"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createAlphaDuplicateResp.StatusCode())
	assert.NotNil(t, createAlphaDuplicateResp.JSON201)

	createBetaResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: "Beta"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createBetaResp.StatusCode())
	assert.NotNil(t, createBetaResp.JSON201)

	_, err = client.AddPlayToWinGameByGameIdWithResponse(ctx, createAlphaResp.JSON201.GameId)
	assert.NoError(t, err)
	_, err = client.AddPlayToWinGameByGameIdWithResponse(ctx, createAlphaDuplicateResp.JSON201.GameId)
	assert.NoError(t, err)
	_, err = client.AddPlayToWinGameByGameIdWithResponse(ctx, createBetaResp.JSON201.GameId)
	assert.NoError(t, err)

	listResp, err := client.ListPlayToWinGamesWithResponse(ctx, &api.ListPlayToWinGamesParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listResp.StatusCode())
	assert.NotNil(t, listResp.JSON200)
	assert.Len(t, listResp.JSON200.Games, 3)

	ptwByGameID := make(map[string]api.PlayToWinGame)
	for _, game := range listResp.JSON200.Games {
		ptwByGameID[game.GameId.String()] = game
	}

	alpha := ptwByGameID[createAlphaResp.JSON201.GameId.String()]
	alphaDuplicate := ptwByGameID[createAlphaDuplicateResp.JSON201.GameId.String()]
	beta := ptwByGameID[createBetaResp.JSON201.GameId.String()]
	assert.NotEqual(t, alpha.PlayToWinId, alphaDuplicate.PlayToWinId)

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

	alphaDuplicateSessionResp, err := client.AddPlayToWinSessionWithResponse(ctx, api.AddPlayToWinSessionJSONRequestBody{
		PlayToWinId: alphaDuplicate.PlayToWinId,
		Entries: []struct {
			EntrantName     string `json:"entrantName"`
			EntrantUniqueId string `json:"entrantUniqueId"`
		}{
			{EntrantName: "Bob", EntrantUniqueId: "B-1"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, alphaDuplicateSessionResp.StatusCode())
	assert.NotNil(t, alphaDuplicateSessionResp.JSON201)
	assert.Len(t, alphaDuplicateSessionResp.JSON201.PlayToWinEntries, 1)

	duplicateWinnerForParentResp, err := client.UpdatePlayToWinGameWithResponse(ctx, alpha.PlayToWinId, api.UpdatePlayToWinGameJSONRequestBody{
		WinnerId: &alphaDuplicateSessionResp.JSON201.PlayToWinEntries[0].EntryId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, duplicateWinnerForParentResp.StatusCode())

	unrelatedWinnerResp, err := client.UpdatePlayToWinGameWithResponse(ctx, beta.PlayToWinId, api.UpdatePlayToWinGameJSONRequestBody{
		WinnerId: &alphaSessionResp.JSON201.PlayToWinEntries[0].EntryId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, unrelatedWinnerResp.StatusCode())
	if assert.NotNil(t, unrelatedWinnerResp.JSON400) {
		assert.NotEmpty(t, unrelatedWinnerResp.JSON400.Error.Details)
		assert.Equal(t, "winnerId", unrelatedWinnerResp.JSON400.Error.Details[0].Field)
	}

	getAlphaResp, err := client.GetPlayToWinGameWithResponse(ctx, alpha.PlayToWinId)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getAlphaResp.StatusCode())
	assert.NotNil(t, getAlphaResp.JSON200)
	assert.NotNil(t, getAlphaResp.JSON200.Winner)
	assert.Equal(t, alphaDuplicateSessionResp.JSON201.PlayToWinEntries[0].EntryId, getAlphaResp.JSON200.Winner.EntryId)

	getBetaResp, err := client.GetPlayToWinGameWithResponse(ctx, beta.PlayToWinId)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getBetaResp.StatusCode())
	assert.NotNil(t, getBetaResp.JSON200)
	assert.Nil(t, getBetaResp.JSON200.Winner)
}
