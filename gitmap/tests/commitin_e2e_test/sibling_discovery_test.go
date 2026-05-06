package commitin_e2e_test

import (
	"path/filepath"
	"testing"

	"github.com/alimtvnetwork/gitmap-v16/gitmap/cmd/commitin"
	"github.com/alimtvnetwork/gitmap-v16/gitmap/constants"
)

// TestSiblingDiscovery_AllKeywordPicksEveryVersionedSibling proves
// the `all` keyword finds every `<base>` and `<base>-vN` directory
// under the source's parent (excluding source itself), in ascending
// version order, when driven end-to-end through orchestrator.Run.
//
// Layout under tmpRoot:
//
//	gitmap/         -> source (version 0; excluded from siblings)
//	gitmap-v2/      -> sibling, version 2
//	gitmap-v10/     -> sibling, version 10 (width-crossing guard)
//	unrelated/      -> NOT a sibling (no `-vN` suffix on `gitmap` base)
//
// Acceptance:
//   - orchestrator exits 0 (siblings found, dry-run completes)
//   - exactly one CommitInRun row with a terminal status
//   - the run's ShaMap is empty (dry-run never persists rewrites)
func TestSiblingDiscovery_AllKeywordPicksEveryVersionedSibling(t *testing.T) {
	skipIfNoGit(t)
	source, _ := buildSiblingTree(t)

	raw := &commitin.RawArgs{
		Source:     source,
		Keyword:    constants.CommitInInputKeywordAll,
		IsDryRun:   true,
		IsNoPrompt: true,
	}
	res := RunHarness(t, raw, source)

	requireExitOk(t, res)
	requireOneTerminalRun(t, res)
	requireDryRunPersistedNoShaMap(t, res)
}

// TestSiblingDiscovery_TailKeywordTruncatesToLastN proves the `-N`
// keyword truncates the sibling list to its tail. With siblings
// {v2, v10}, `-1` must keep only v10 — the highest-versioned tail —
// and orchestrator.Run must still complete cleanly against it.
func TestSiblingDiscovery_TailKeywordTruncatesToLastN(t *testing.T) {
	skipIfNoGit(t)
	source, _ := buildSiblingTree(t)

	raw := &commitin.RawArgs{
		Source:      source,
		Keyword:     constants.CommitInInputKeywordTailDash + "1",
		KeywordTail: 1,
		IsDryRun:    true,
		IsNoPrompt:  true,
	}
	res := RunHarness(t, raw, source)

	requireExitOk(t, res)
	requireOneTerminalRun(t, res)
}

// TestSiblingDiscovery_ZeroSiblingsFailsInputUnusable proves the
// negative path: when the source has no versioned neighbors, `all`
// must exit with CommitInExitInputUnusable (the orchestrator maps
// the keyword-expansion failure to the input-stage exit code) AND
// the stderr/stdout must surface the spec §2.4 "matched zero
// siblings" message so the user can act on it.
func TestSiblingDiscovery_ZeroSiblingsFailsInputUnusable(t *testing.T) {
	source := NewRepoBuilder(t)
	source.SeedCommit("README.md", "lonely\n", "lonely: initial")

	raw := &commitin.RawArgs{
		Source:     source.Dir(),
		Keyword:    constants.CommitInInputKeywordAll,
		IsDryRun:   true,
		IsNoPrompt: true,
	}
	res := RunHarness(t, raw, source.Dir())

	if res.ExitCode != constants.CommitInExitInputUnusable {
		t.Fatalf("exit = %d, want %d (InputUnusable)\n--- stderr\n%s\n--- stdout\n%s",
			res.ExitCode, constants.CommitInExitInputUnusable, res.Stderr, res.Stdout)
	}
	combined := res.Stdout + res.Stderr
	if !contains(combined, "matched zero siblings") {
		t.Fatalf("missing zero-siblings diagnostic in output:\n%s", combined)
	}
}

// contains is a tiny strings.Contains alias so the assertion above
// reads naturally without forcing every test file to import strings
// just for one call site.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// buildSiblingTree materializes a real on-disk tree of one source repo
// plus two versioned siblings (v2 + v10) plus one unrelated directory.
// Returns (sourceDir, parentDir). Each repo gets one seeded commit so
// downstream commit-in stages have something to walk.
func buildSiblingTree(t *testing.T) (string, string) {
	t.Helper()
	parent := t.TempDir()
	source := makeNamedRepo(t, parent, "gitmap", "src.txt", "source\n")
	makeNamedRepo(t, parent, "gitmap-v2", "v2.txt", "v2\n")
	makeNamedRepo(t, parent, "gitmap-v10", "v10.txt", "v10\n")
	makeNamedRepo(t, parent, "unrelated", "noise.txt", "noise\n")
	return source, parent
}

// makeNamedRepo creates a named repo under parent and seeds one
// commit. Centralizing this keeps buildSiblingTree under the 15-line
// helper cap and gives every per-version repo identical date/identity
// settings so the tree is fully deterministic across runs.
func makeNamedRepo(t *testing.T, parent, name, file, contents string) string {
	t.Helper()
	rb := NewRepoBuilderAt(t, filepath.Join(parent, name))
	rb.SeedCommit(file, contents, "seed: "+name)
	return rb.Dir()
}

// requireExitOk fatals when the orchestrator did not return exit 0,
// dumping both streams so a regression here is debuggable from the
// CI log alone (no replay required).
func requireExitOk(t *testing.T, res RunResult) {
	t.Helper()
	if res.ExitCode != constants.CommitInExitOk {
		t.Fatalf("exit = %d, want %d (Ok)\n--- stderr\n%s\n--- stdout\n%s",
			res.ExitCode, constants.CommitInExitOk, res.Stderr, res.Stdout)
	}
}

// requireOneTerminalRun re-opens the workspace DB and asserts exactly
// one CommitInRun row exists with a terminal RunStatus enum value.
// Reuses the helpers landed in scaffold_test.go (Step 8).
func requireOneTerminalRun(t *testing.T, res RunResult) {
	t.Helper()
	db := OpenWorkspaceDB(t, res.DBPath)
	defer db.Close()
	if got := CountRows(t, db, "CommitInRun"); got != 1 {
		t.Fatalf("CommitInRun rows = %d, want 1\n--- stderr\n%s",
			got, res.Stderr)
	}
	statuses := QueryStrings(t, db, runStatusQuery)
	if len(statuses) != 1 || !isTerminalStatus(statuses[0]) {
		t.Fatalf("CommitInRun final status = %v, want one terminal value", statuses)
	}
}

// requireDryRunPersistedNoShaMap locks in the spec §2.5 contract that
// a dry-run NEVER mutates ShaMap — the dedupe table only fills on a
// real Created outcome. Any future leak would silently break dedupe
// across subsequent real runs.
func requireDryRunPersistedNoShaMap(t *testing.T, res RunResult) {
	t.Helper()
	db := OpenWorkspaceDB(t, res.DBPath)
	defer db.Close()
	if got := CountRows(t, db, "ShaMap"); got != 0 {
		t.Fatalf("ShaMap rows after dry-run = %d, want 0", got)
	}
}
