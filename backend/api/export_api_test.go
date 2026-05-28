package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// overrideExportVars replaces all package-level export vars with test-safe
// values and registers t.Cleanup to restore the originals afterwards.
//
// fakeBinPath and fakeConfigPath must be paths to real files (e.g. created
// with os.CreateTemp) so that exportDependencyCheck's os.Stat calls succeed.
func overrideExportVars(t *testing.T, fakeBinPath, fakeConfigPath string) {
	t.Helper()
	origBin := pgAnonymizerBin
	origConfig := anonymizerConfig
	origTimeout := exportTimeout
	origLookup := pgDumpLookup
	origCmd := newExportCommand

	pgAnonymizerBin = fakeBinPath
	anonymizerConfig = fakeConfigPath
	exportTimeout = 5 * time.Second
	pgDumpLookup = func() bool { return true }

	t.Cleanup(func() {
		pgAnonymizerBin = origBin
		anonymizerConfig = origConfig
		exportTimeout = origTimeout
		pgDumpLookup = origLookup
		newExportCommand = origCmd
	})
}

// createFakeDeps creates temporary files standing in for the pg-anonymizer
// binary and the anonymizer config, registers their removal with t.Cleanup,
// and returns their paths.
func createFakeDeps(t *testing.T) (binPath, configPath string) {
	t.Helper()

	binFile, err := os.CreateTemp("", "fake-pg-anonymizer-*")
	assert.NoError(t, err)
	binFile.Close()

	configFile, err := os.CreateTemp("", "fake-anonymizer-config-*")
	assert.NoError(t, err)
	configFile.Close()

	t.Cleanup(func() {
		os.Remove(binFile.Name())
		os.Remove(configFile.Name())
	})

	return binFile.Name(), configFile.Name()
}

// newWritingCommand returns a newExportCommand override that writes dummy SQL
// content to outputPath and exits 0. As a side effect it also stores the
// outputPath into *capturedPath so callers can inspect the temp directory.
func newWritingCommand(capturedPath *string) func(ctx context.Context, outputPath, connStr string) *exec.Cmd {
	return func(ctx context.Context, outputPath, connStr string) *exec.Cmd {
		if capturedPath != nil {
			*capturedPath = outputPath
		}
		script := fmt.Sprintf("echo '-- dummy export' > %q", outputPath)
		return exec.CommandContext(ctx, "sh", "-c", script)
	}
}

func TestExportDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server, _, _ := setupTestServer()

	t.Run("Should return 503 when pg-anonymizer binary is missing", func(t *testing.T) {
		_, configPath := createFakeDeps(t)
		overrideExportVars(t, "/nonexistent/pg-anonymizer", configPath)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "SERVICE_UNAVAILABLE")
	})

	t.Run("Should return 503 when anonymizer config file is missing", func(t *testing.T) {
		binPath, _ := createFakeDeps(t)
		overrideExportVars(t, binPath, "/nonexistent/anonymizer-config.txt")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "SERVICE_UNAVAILABLE")
	})

	t.Run("Should return 503 when pg_dump is not in PATH", func(t *testing.T) {
		binPath, configPath := createFakeDeps(t)
		overrideExportVars(t, binPath, configPath)
		pgDumpLookup = func() bool { return false }

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "SERVICE_UNAVAILABLE")
	})

	t.Run("Should return 500 when export subprocess exits non-zero", func(t *testing.T) {
		binPath, configPath := createFakeDeps(t)
		overrideExportVars(t, binPath, configPath)
		newExportCommand = func(ctx context.Context, outputPath, connStr string) *exec.Cmd {
			return exec.CommandContext(ctx, "sh", "-c", "exit 1")
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
	})

	t.Run("Should return 504 when export subprocess exceeds timeout", func(t *testing.T) {
		binPath, configPath := createFakeDeps(t)
		overrideExportVars(t, binPath, configPath)
		exportTimeout = 50 * time.Millisecond
		newExportCommand = func(ctx context.Context, outputPath, connStr string) *exec.Cmd {
			return exec.CommandContext(ctx, "sleep", "10")
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusGatewayTimeout, w.Code)
		assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
	})

	t.Run("Should return 200 with correct Content-Disposition filename on success", func(t *testing.T) {
		binPath, configPath := createFakeDeps(t)
		overrideExportVars(t, binPath, configPath)
		newExportCommand = newWritingCommand(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusOK, w.Code)
		contentDisposition := w.Header().Get("Content-Disposition")
		assert.True(t, strings.HasPrefix(contentDisposition, "attachment; filename="),
			"Content-Disposition should be an attachment, got: %s", contentDisposition)
		assert.Contains(t, contentDisposition, "db_export_")
		assert.Contains(t, contentDisposition, ".sql")
		assert.Contains(t, w.Body.String(), "-- dummy export")
	})

	t.Run("Should clean up temp directory after successful export", func(t *testing.T) {
		binPath, configPath := createFakeDeps(t)
		overrideExportVars(t, binPath, configPath)

		var capturedOutputPath string
		newExportCommand = newWritingCommand(&capturedOutputPath)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, capturedOutputPath, "command should have been called with an output path")
		tmpDir := filepath.Dir(capturedOutputPath)
		_, statErr := os.Stat(tmpDir)
		assert.True(t, os.IsNotExist(statErr), "temp directory should have been removed after successful export")
	})

	t.Run("Should clean up temp directory after failed export", func(t *testing.T) {
		binPath, configPath := createFakeDeps(t)
		overrideExportVars(t, binPath, configPath)

		var capturedOutputPath string
		newExportCommand = func(ctx context.Context, outputPath, connStr string) *exec.Cmd {
			capturedOutputPath = outputPath
			return exec.CommandContext(ctx, "sh", "-c", "exit 1")
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/export/database", nil)

		server.ExportDatabase(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.NotEmpty(t, capturedOutputPath, "command should have been called with an output path")
		tmpDir := filepath.Dir(capturedOutputPath)
		_, statErr := os.Stat(tmpDir)
		assert.True(t, os.IsNotExist(statErr), "temp directory should have been removed after failed export")
	})
}
