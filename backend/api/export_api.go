package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// These are declared as vars (not consts) so that tests can override them
// without requiring real binaries, config files, or a live database.
var (
	pgAnonymizerBin  = "/opt/pg-anonymizer/node_modules/.bin/pg-anonymizer"
	anonymizerConfig = "/opt/anonymizer-config.txt"
	exportTimeout    = 2 * time.Minute

	// pgDumpLookup checks whether pg_dump is available in PATH.
	// Declared as a var so tests can substitute it.
	pgDumpLookup = func() bool {
		_, err := exec.LookPath("pg_dump")
		return err == nil
	}

	// newExportCommand builds the pg-anonymizer subprocess.
	// Declared as a var so tests can substitute a fake command that does not
	// require real binaries or a live database.
	newExportCommand = func(ctx context.Context, outputPath, connStr string) *exec.Cmd {
		// Connection string and pg_dump flags are positional PGARG arguments.
		// "--" separates pg-anonymizer flags from PGARG so --data-only is not
		// misinterpreted by the OCLIF argument parser.
		return exec.CommandContext(ctx, pgAnonymizerBin,
			"--config="+anonymizerConfig,
			"-o", outputPath,
			"--",
			connStr,
			"--data-only",
			"--exclude-table-data=public.transaction_events",
		)
	}
)

// ExportDatabase shells out to pg-anonymizer to produce a data-only, anonymized
// SQL dump of the library database and streams the resulting .sql file to the
// client. Patron full names are anonymized per the anonymizer config.
//
// DB credentials are resolved from ~/.pgpass, which is written by
// entrypoint.sh at container startup. The password never appears in Go code or
// subprocess argument lists.
//
// This endpoint requires pg_dump, nodejs, and the pg-anonymizer binary to be
// present; it is expected to run inside the Docker image. Outside Docker,
// requests will receive a 503 if the dependencies are absent.
func (s *Server) ExportDatabase(c *gin.Context) {
	if missing := exportDependencyCheck(); missing != "" {
		log.Printf("Database export unavailable: %s", missing)
		safeAbortWithStatusJSON(c, http.StatusServiceUnavailable,
			NewErrorResponse(SERVICEUNAVAILABLE, "Database export is not available in this environment: "+missing))
		return
	}

	// Password is intentionally omitted — pg_dump reads it from ~/.pgpass.
	connStr := fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=disable",
		getEnvWithDefault("DB_USER", "postgres"),
		getEnvWithDefault("DB_HOST", "localhost"),
		getEnvWithDefault("DB_PORT", "5432"),
		getEnvWithDefault("DB_NAME", "librarydb"),
	)

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05Z")
	filename := fmt.Sprintf("db_export_%s.sql", timestamp)

	// MkdirTemp gives a uniquely-named directory; the download filename is used
	// as the file name inside it so the artifact on disk is human-readable.
	// RemoveAll cleans up the directory and file atomically once the response is
	// sent — c.FileAttachment streams synchronously before the defer runs.
	tmpDir, err := os.MkdirTemp("", "db-export-*")
	if err != nil {
		log.Printf("Error creating temp export directory: %v", err)
		internalError(c, err)
		return
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	tmpPath := filepath.Join(tmpDir, filename)

	ctx, cancel := context.WithTimeout(c.Request.Context(), exportTimeout)
	defer cancel()

	var stderr bytes.Buffer
	cmd := newExportCommand(ctx, tmpPath, connStr)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Error running pg-anonymizer (stderr: %s): %v", stderr.String(), err)
		if ctx.Err() == context.DeadlineExceeded {
			safeAbortWithStatusJSON(c, http.StatusGatewayTimeout,
				NewErrorResponse(INTERNALERROR, "Database export timed out"))
			return
		}
		safeAbortWithStatusJSON(c, http.StatusInternalServerError,
			NewErrorResponse(INTERNALERROR, "Database export failed"))
		return
	}

	// pg-anonymizer wraps the dump with \restrict / \unrestrict directives that
	// are not valid psql meta-commands. Strip them so the file loads cleanly.
	if err := stripAnonymizerDirectives(tmpPath); err != nil {
		log.Printf("Error stripping anonymizer directives from export: %v", err)
		internalError(c, err)
		return
	}

	c.FileAttachment(tmpPath, filename)
}

// exportDependencyCheck verifies that all runtime dependencies for the export
// are present. Returns a non-empty description of the first missing dependency,
// or an empty string if all checks pass.
func exportDependencyCheck() string {
	if _, err := os.Stat(pgAnonymizerBin); os.IsNotExist(err) {
		return "pg-anonymizer binary not found at " + pgAnonymizerBin
	}
	if _, err := os.Stat(anonymizerConfig); os.IsNotExist(err) {
		return "anonymizer config not found at " + anonymizerConfig
	}
	if !pgDumpLookup() {
		return "pg_dump not found in PATH"
	}
	return ""
}

// getEnvWithDefault returns the value of key, or defaultValue if unset or empty.
func getEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// stripAnonymizerDirectives removes pg-anonymizer's \restrict and \unrestrict
// lines from a SQL dump file in-place. These directives are not valid psql
// meta-commands and will cause errors when the file is loaded with psql -f.
func stripAnonymizerDirectives(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	var buf bytes.Buffer
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, `\restrict`) || strings.HasPrefix(line, `\unrestrict`) {
			continue
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	in.Close()

	return os.WriteFile(path, buf.Bytes(), 0600)
}
