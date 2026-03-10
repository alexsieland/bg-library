package tests

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAdminBulkUploadWorkflow(t *testing.T) {
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

	// 2. Admin prepares CSV files for bulk upload
	// CSV of board games for the library
	gamesCSV := strings.Join([]string{
		"Catan",
		"Ticket to Ride",
		"Wingspan",
		"Azul",
		"Pandemic",
		"7 Wonders",
	}, "\n")

	// CSV of convention attendees (patrons)
	patronsCSV := strings.Join([]string{
		"Alice Johnson",
		"Bob Smith",
		"Carol Williams",
		"David Brown",
		"Eve Davis",
	}, "\n")

	// 3. Upload games CSV
	gamesBase64 := base64.StdEncoding.EncodeToString([]byte(gamesCSV))
	bulkGamesResp, err := client.BulkAddGamesWithTextBodyWithResponse(ctx, []byte(gamesBase64))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, bulkGamesResp.StatusCode())
	assert.NotNil(t, bulkGamesResp.JSON201)
	assert.Equal(t, int32(6), bulkGamesResp.JSON201.Imported, "Should import 6 games")

	// 4. Upload patrons CSV
	patronsBase64 := base64.StdEncoding.EncodeToString([]byte(patronsCSV))
	bulkPatronsResp, err := client.BulkAddPatronsWithTextBodyWithResponse(ctx, []byte(patronsBase64))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, bulkPatronsResp.StatusCode())
	assert.NotNil(t, bulkPatronsResp.JSON201)
	assert.Equal(t, int32(5), bulkPatronsResp.JSON201.Imported, "Should import 5 patrons")

	// 5. Verify games are in the library by listing all games
	listGamesResp, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listGamesResp.StatusCode())
	assert.NotNil(t, listGamesResp.JSON200)
	assert.Len(t, listGamesResp.JSON200.Games, 6, "Should have 6 games in the library")

	// Verify all uploaded games are present
	expectedGameTitles := map[string]bool{
		"Catan":          false,
		"Ticket to Ride": false,
		"Wingspan":       false,
		"Azul":           false,
		"Pandemic":       false,
		"7 Wonders":      false,
	}

	for _, gameStatus := range listGamesResp.JSON200.Games {
		if _, exists := expectedGameTitles[gameStatus.Game.Title]; exists {
			expectedGameTitles[gameStatus.Game.Title] = true
		}
	}

	for title, found := range expectedGameTitles {
		assert.True(t, found, "Game '%s' should be in the library", title)
	}

	// 6. Verify all games are available (not checked out)
	for _, gameStatus := range listGamesResp.JSON200.Games {
		assert.Nil(t, gameStatus.Patron, "Game '%s' should not be checked out", gameStatus.Game.Title)
	}

	// 7. Verify patrons are in the system by listing all patrons
	listPatronsResp, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listPatronsResp.StatusCode())
	assert.NotNil(t, listPatronsResp.JSON200)
	assert.Len(t, listPatronsResp.JSON200.Patrons, 5, "Should have 5 patrons in the system")

	// Verify all uploaded patrons are present
	expectedPatronNames := map[string]bool{
		"Alice Johnson":  false,
		"Bob Smith":      false,
		"Carol Williams": false,
		"David Brown":    false,
		"Eve Davis":      false,
	}

	for _, patron := range listPatronsResp.JSON200.Patrons {
		if _, exists := expectedPatronNames[patron.Name]; exists {
			expectedPatronNames[patron.Name] = true
		}
	}

	for name, found := range expectedPatronNames {
		assert.True(t, found, "Patron '%s' should be in the system", name)
	}

	// 8. Search for a specific game to verify search functionality works with bulk uploaded data
	searchTitle := "Wingspan"
	searchGamesResp, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{
		Title: &searchTitle,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, searchGamesResp.StatusCode())
	assert.NotNil(t, searchGamesResp.JSON200)
	assert.Greater(t, len(searchGamesResp.JSON200.Games), 0, "Should find 'Wingspan' in search results")

	// Verify the searched game is in the results
	foundWingspan := false
	for _, gameStatus := range searchGamesResp.JSON200.Games {
		if gameStatus.Game.Title == "Wingspan" {
			foundWingspan = true
			break
		}
	}
	assert.True(t, foundWingspan, "Should find 'Wingspan' in search results")

	// 9. Search for a specific patron to verify search functionality works with bulk uploaded data
	searchName := "Bob"
	searchPatronsResp, err := client.ListPatronsWithResponse(ctx, &api.ListPatronsParams{
		Name: &searchName,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, searchPatronsResp.StatusCode())
	assert.NotNil(t, searchPatronsResp.JSON200)
	assert.Greater(t, len(searchPatronsResp.JSON200.Patrons), 0, "Should find patrons matching 'Bob'")

	// Verify Bob Smith is in the search results
	foundBob := false
	for _, patron := range searchPatronsResp.JSON200.Patrons {
		if patron.Name == "Bob Smith" {
			foundBob = true
			break
		}
	}
	assert.True(t, foundBob, "Should find 'Bob Smith' in search results")

	// 10. Verify that we can check out one of the bulk-uploaded games to one of the bulk-uploaded patrons
	// Get the first game and first patron
	gameToCheckOut := listGamesResp.JSON200.Games[0].Game.GameId
	patronToCheckOut := listPatronsResp.JSON200.Patrons[0].PatronId

	checkoutResp, err := client.CheckOutGameWithResponse(ctx, api.CheckOutGameJSONRequestBody{
		GameId:   gameToCheckOut,
		PatronId: patronToCheckOut,
	})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, checkoutResp.StatusCode())
	assert.NotNil(t, checkoutResp.JSON201)

	// 11. Verify the game now shows as checked out
	getGameResp, err := client.GetGameWithResponse(ctx, gameToCheckOut)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getGameResp.StatusCode())

	// Verify in the full game list that the game is now checked out
	listGamesResp2, err := client.ListGamesWithResponse(ctx, &api.ListGamesParams{})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, listGamesResp2.StatusCode())

	checkedOutCount := 0
	for _, gameStatus := range listGamesResp2.JSON200.Games {
		if gameStatus.Game.GameId == gameToCheckOut {
			assert.NotNil(t, gameStatus.Patron, "Game should be checked out")
			assert.Equal(t, patronToCheckOut, gameStatus.Patron.PatronId, "Game should be checked out by the correct patron")
			checkedOutCount++
		}
	}
	assert.Equal(t, 1, checkedOutCount, "Should find exactly one checked out game")
}
