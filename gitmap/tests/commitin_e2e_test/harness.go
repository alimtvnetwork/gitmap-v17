package commitin_e2e_test

import (
	"bytes"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/gitmap-v16/gitmap/cmd/commitin"
	"github.com/alimtvnetwork/gitmap-v16/gitmap/cmd/commitin/orchestrator"
	"github.com/alimtvnetwork/gitmap-v16/gitmap/constants"

	_ "modernc.org/sqlite" // commit-in opens the workspace DB via this driver
)

// RunResult captures everything a test needs to assert about one
// orchestrator.Run invocation: the exit code, both captured streams,
// and the on-disk SQLite path so the test can re-open it for row-
// level introspection (RewrittenCommit, ShaMap, SkipLog).
type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	DBPath   string
}

// RunHarness wires a RawArgs through orchestrator.Run with captured
// stdout/stderr buffers, then locates the workspace DB the orchestrator
// just provisioned so callers can run assertion queries against it.
//
// We deliberately do NOT keep the *sql.DB open across the call: the
// orchestrator owns the connection during Run() and closes it on exit.
// The harness re-opens read-only via OpenWorkspaceDB for assertions.
func RunHarness(t *testing.T, raw *commitin.RawArgs, repoDir string) RunResult {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := orchestrator.Run(raw, &stdout, &stderr)
	return RunResult{
		ExitCode: code,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		DBPath:   workspaceDBPath(repoDir),
	}
}

// workspaceDBPath mirrors workspace.buildPaths so assertion code can
// reach the SQLite file without importing the workspace package
// (which would create a cycle for fixtures that live under that
// package's test taxonomy). Layout: <source>/.gitmap/<DBFile>.
func workspaceDBPath(repoDir string) string {
	return filepath.Join(repoDir, constants.GitMapDir, constants.DBFile)
}

// OpenWorkspaceDB opens the on-disk SQLite the orchestrator wrote to.
// Read-only by convention — assertion helpers should never mutate it.
// Caller must defer db.Close().
func OpenWorkspaceDB(t *testing.T, path string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open workspace db %s: %v", path, err)
	}
	return db
}

// CountRows runs `SELECT COUNT(*) FROM <table>` and fatals on any
// error. Tiny helper kept here so per-case asserts read naturally.
func CountRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// QueryStrings runs a single-column SELECT and returns its rows as
// strings. Used to pull the list of outcome / skip-reason names from
// the run for outcome-shape assertions.
func QueryStrings(t *testing.T, db *sql.DB, query string, args ...any) []string {
	t.Helper()
	rows, err := db.Query(query, args...)
	if err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("scan %q: %v", query, err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err %q: %v", query, err)
	}
	return out
}
