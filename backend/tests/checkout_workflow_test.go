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

func TestCheckoutWorkflow(t *testing.T) {
	ctx := t.Context()

	// 1. Start Postgres container
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

	// 2. Set environment variables for the backend to connect to the container
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "librarydb")
	os.Setenv("GIN_MODE", "release")

	// 3. Initialize backend server
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

	// 4. Create a Patron
	patronName := "John Doe"
	createPatronResp, err := client.AddPatronWithResponse(ctx, api.AddPatronJSONRequestBody{
		Name: patronName,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createPatronResp.StatusCode())
	patron := createPatronResp.JSON201
	assert.NotNil(t, patron)
	assert.Equal(t, patronName, patron.Name)
	patronID := patron.PatronId

	// 5. Create a Game
	gameTitle := "Catan"
	createGameResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{
		Title: gameTitle,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createGameResp.StatusCode())
	game := createGameResp.JSON201
	assert.NotNil(t, game)
	assert.Equal(t, gameTitle, game.Title)
	gameID := game.GameId

	// 6. Check out the Game
	checkoutResp, err := client.CheckOutGameWithResponse(ctx, api.CheckOutGameJSONRequestBody{
		GameId:   gameID,
		PatronId: patronID,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, checkoutResp.StatusCode())
	transaction := checkoutResp.JSON201
	assert.NotNil(t, transaction)
	assert.Equal(t, gameID, transaction.GameId)
	assert.Equal(t, patronID, transaction.PatronId)
	transactionID := transaction.Id

	// 7. Verify game status is checked out
	listGamesResp, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		CheckedOut: func(b bool) *bool { return &b }(true),
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listGamesResp.StatusCode())
	assert.NotNil(t, listGamesResp.JSON200)

	found := false
	for _, gs := range listGamesResp.JSON200.Games {
		if gs.Game.GameId == gameID {
			found = true
			assert.NotNil(t, gs.Patron)
			assert.Equal(t, patronID, gs.Patron.PatronId)
			break
		}
	}
	assert.True(t, found, "Game should be in the checked out games list")

	// 8. Check in the Game
	checkinResp, err := client.CheckInGameWithResponse(ctx, &api.CheckInGameParams{
		TransactionId: transactionID,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, checkinResp.StatusCode())

	// 9. Verify game status is available again
	listGamesResp2, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		CheckedOut: func(b bool) *bool { return &b }(true),
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listGamesResp2.StatusCode())

	found = false
	for _, gs := range listGamesResp2.JSON200.Games {
		if gs.Game.GameId == gameID {
			found = true
			break
		}
	}
	assert.False(t, found, "Game should NOT be in the checked out games list anymore")
}
