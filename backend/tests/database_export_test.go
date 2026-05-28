package tests

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/alexsieland/bg-library/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const pgAnonymizerBinPath = "/opt/pg-anonymizer/node_modules/.bin/pg-anonymizer"

func TestDatabaseExportSmoke(t *testing.T) {
	// Skip when runtime dependencies are not present (e.g. local dev without Docker).
	if _, err := exec.LookPath("pg_dump"); err != nil {
		t.Skip("pg_dump not in PATH — skipping database export smoke test")
	}
	if _, err := os.Stat(pgAnonymizerBinPath); err != nil {
		t.Skip("pg-anonymizer binary not present — skipping database export smoke test")
	}

	ctx := t.Context()

	schemaPath, err := filepath.Abs("../db/schema.sql")
	assert.NoError(t, err)
	viewsPath, err := filepath.Abs("../db/views.sql")
	assert.NoError(t, err)

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("librarydb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		postgres.WithInitScripts(schemaPath, viewsPath),
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
	err = server.LibService.Start()
	assert.NoError(t, err)
	defer server.LibService.Stop()

	r := gin.New()
	api.RegisterHandlers(r, server)
	ts := httptest.NewServer(r)
	defer ts.Close()

	client, err := api.NewClientWithResponses(ts.URL)
	assert.NoError(t, err)

	t.Run("Should return 200 with a SQL file attachment when called", func(t *testing.T) {
		resp, err := client.ExportDatabaseWithResponse(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())

		contentDisposition := resp.HTTPResponse.Header.Get("Content-Disposition")
		assert.Contains(t, contentDisposition, "attachment")
		assert.Contains(t, contentDisposition, ".sql")
	})

	t.Run("Should return a filename matching the db_export_<timestamp>.sql pattern", func(t *testing.T) {
		resp, err := client.ExportDatabaseWithResponse(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())

		contentDisposition := resp.HTTPResponse.Header.Get("Content-Disposition")
		pattern := regexp.MustCompile(`db_export_\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z\.sql`)
		assert.True(t, pattern.MatchString(contentDisposition),
			"Content-Disposition should contain db_export_<timestamp>.sql, got: %s", contentDisposition)
	})
}
