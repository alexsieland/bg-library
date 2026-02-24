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

func TestDuplicateCheckoutWorkflow(t *testing.T) {
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

	// 2. Create two Patrons (A and B) and one Game
	patronAResp, err := client.AddPatronWithResponse(ctx, api.AddPatronJSONRequestBody{Name: "Patron A"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, patronAResp.StatusCode())
	patronA := patronAResp.JSON201

	patronBResp, err := client.AddPatronWithResponse(ctx, api.AddPatronJSONRequestBody{Name: "Patron B"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, patronBResp.StatusCode())
	patronB := patronBResp.JSON201

	gameResp, err := client.AddGameWithResponse(ctx, api.AddGameJSONRequestBody{Title: "Everdell"})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, gameResp.StatusCode())
	game := gameResp.JSON201

	// 3. Check out the Game to Patron A (Expect Success 201)
	checkoutAResp, err := client.CheckOutGameWithResponse(ctx, api.CheckOutGameJSONRequestBody{
		GameId:   game.GameId,
		PatronId: patronA.PatronId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, checkoutAResp.StatusCode())
	transactionA := checkoutAResp.JSON201

	// 4. Attempt to check out the same Game to Patron B (Expect Conflict 409)
	checkoutBResp, err := client.CheckOutGameWithResponse(ctx, api.CheckOutGameJSONRequestBody{
		GameId:   game.GameId,
		PatronId: patronB.PatronId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, checkoutBResp.StatusCode())

	// 5. Verify the error response contains "already checked out" message
	assert.NotNil(t, checkoutBResp.JSON409)
	assert.Contains(t, checkoutBResp.JSON409.Error.Message, "already checked out")

	// 6. Check in the Game from Patron A (Expect Success 204)
	checkinResp, err := client.CheckInGameWithResponse(ctx, &api.CheckInGameParams{
		TransactionId: transactionA.Id.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, checkinResp.StatusCode())

	// 7. Check out the Game to Patron B (Expect Success 201)
	checkoutB2Resp, err := client.CheckOutGameWithResponse(ctx, api.CheckOutGameJSONRequestBody{
		GameId:   game.GameId,
		PatronId: patronB.PatronId,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, checkoutB2Resp.StatusCode())
}
