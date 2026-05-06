package commitin_e2e_test

import (
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v16/gitmap/cmd/commitin"
	"github.com/alimtvnetwork/gitmap-v16/gitmap/constants"
)

// TestRepoBuilderProducesDeterministicSHAs is the harness self-test:
// it locks in that two RepoBuilder instances seeded with identical
// inputs yield byte-identical HEAD SHAs. Any future leak of wall-
// clock time, env-driven author identity, or default-branch policy
// will trip this guard before it can corrupt downstream dedupe tests.
func TestRepoBuilderProducesDeterministicSHAs(t *testing.T) {
	first := buildSeededRepo(t)
	second := buildSeededRepo(t)
	if first != second {
		t.Fatalf("HEAD SHAs diverged across RepoBuilder runs:\n  first  = %s\n  second = %s",
			first, second)
	}
	if len(first) != 40 {
		t.Fatalf("HEAD SHA is not a full 40-char SHA-1: %q", first)
	}
}

// TestE2ESmokeOrchestratorRunsAgainstRealRepo proves the harness can
// actually drive orchestrator.Run end-to-end against a fixture repo.
// We assert on the outermost contract only (exit code + a workspace
// DB on disk); per-stage row-level assertions land in Steps 9-12 to
// keep this scaffolding test stable while the matrix expands.
func TestE2ESmokeOrchestratorRunsAgainstRealRepo(t *testing.T) {
	source := NewRepoBuilder(t)
	source.SeedCommit("README.md", "source seed\n", "source: initial")

	input := NewRepoBuilder(t)
	input.SeedCommit("README.md", "input seed\n", "input: initial")
	input.SeedCommit("main.go", "package main\n", "input: add main")

	raw := &commitin.RawArgs{
		Source:     source.Dir(),
		Inputs:     []string{input.Dir()},
		IsDryRun:   true,
		IsNoPrompt: true,
	}
	res := RunHarness(t, raw, source.Dir())

	if res.ExitCode != constants.CommitInExitOk {
		t.Fatalf("orchestrator exit = %d, want %d\n--- stderr\n%s\n--- stdout\n%s",
			res.ExitCode, constants.CommitInExitOk, res.Stderr, res.Stdout)
	}
	assertWorkspaceDBHasRunRow(t, res)
}

// assertWorkspaceDBHasRunRow re-opens the workspace SQLite the
// orchestrator just wrote and confirms exactly one CommitInRun row
// landed. Split out so the smoke test stays under the 15-line cap.
func assertWorkspaceDBHasRunRow(t *testing.T, res RunResult) {
	t.Helper()
	db := OpenWorkspaceDB(t, res.DBPath)
	defer db.Close()
	if got := CountRows(t, db, "CommitInRun"); got != 1 {
		t.Fatalf("CommitInRun rows = %d, want 1\n--- stderr\n%s",
			got, res.Stderr)
	}
	statuses := QueryStrings(t, db, runStatusQuery)
	if len(statuses) != 1 || !isTerminalStatus(statuses[0]) {
		t.Fatalf("CommitInRun final status = %v, want one of %s",
			statuses, strings.Join(terminalStatuses(), ", "))
	}
}

// runStatusQuery joins the run row to its enum mirror so assertions
// read the human-readable status name (Completed / Failed / ...).
const runStatusQuery = `SELECT s.Name
	FROM CommitInRun r
	JOIN RunStatus s ON s.RunStatusId = r.RunStatusId`

// isTerminalStatus returns true when name is any RunStatus the
// orchestrator is allowed to leave behind after Run() returns. Any
// drift here would indicate a leaked half-finished run.
func isTerminalStatus(name string) bool {
	for _, t := range terminalStatuses() {
		if name == t {
			return true
		}
	}
	return false
}

// terminalStatuses lists the post-Run RunStatus enum values per
// finalRunStatus() in orchestrator/run.go. Centralized so a future
// status (e.g. CancelledByUser) is added in exactly one place.
func terminalStatuses() []string {
	return []string{
		constants.CommitInRunStatusCompleted,
		constants.CommitInRunStatusFailed,
		constants.CommitInRunStatusPartiallyFailed,
	}
}

// buildSeededRepo creates a one-off RepoBuilder with the canonical
// two-commit seed and returns the HEAD SHA. Used by the determinism
// self-test so each call site stays under the file's line budget.
func buildSeededRepo(t *testing.T) string {
	t.Helper()
	rb := NewRepoBuilder(t)
	rb.SeedCommit("README.md", "hello\n", "initial commit")
	return rb.SeedCommit("main.go", "package main\n", "add main")
}
