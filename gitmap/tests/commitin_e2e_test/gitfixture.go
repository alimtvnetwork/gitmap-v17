package commitin_e2e_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// RepoBuilder scaffolds a real git repository under t.TempDir and
// seeds it with deterministic commits whose author + committer dates
// are caller-controlled. Deterministic dates keep SHAs stable across
// runs so dedupe assertions can compare ShaMap rows by value.
//
// The builder is intentionally tiny: it shells out to `git` rather
// than vendoring go-git so the harness reflects what a real user
// sees on their box. Tests skip when `git` is unavailable.
type RepoBuilder struct {
	t       *testing.T
	dir     string
	clock   time.Time
	authorN string
	authorE string
}

// NewRepoBuilder creates an empty repo under t.TempDir, configures
// a fixed author identity, and returns the builder. The starting
// clock is 2024-01-01T00:00:00Z; SeedCommit advances it by one hour
// per call so commit ordering is unambiguous.
func NewRepoBuilder(t *testing.T) *RepoBuilder {
	t.Helper()
	skipIfNoGit(t)
	dir := t.TempDir()
	rb := &RepoBuilder{
		t:       t,
		dir:     dir,
		clock:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		authorN: "Fixture Bot",
		authorE: "fixture@example.test",
	}
	rb.runGit("init", "-q", "-b", "main")
	rb.runGit("config", "user.name", rb.authorN)
	rb.runGit("config", "user.email", rb.authorE)
	return rb
}

// Dir returns the absolute repo path so callers can pass it as the
// commit-in <source> argv token.
func (rb *RepoBuilder) Dir() string { return rb.dir }

// SeedCommit writes file → contents, stages, and commits with the
// builder's deterministic clock (advanced by one hour per call).
// Returns the new HEAD SHA so tests can correlate against ShaMap
// dedupe rows.
func (rb *RepoBuilder) SeedCommit(file, contents, message string) string {
	rb.t.Helper()
	rb.writeFile(file, contents)
	rb.runGit("add", file)
	rb.commitWithFixedDate(message)
	rb.clock = rb.clock.Add(1 * time.Hour)
	return rb.head()
}

// writeFile materializes one file under the repo root, creating any
// missing parent directories. Helper kept under the 15-line cap.
func (rb *RepoBuilder) writeFile(rel, contents string) {
	rb.t.Helper()
	abs := filepath.Join(rb.dir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		rb.t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(contents), 0o644); err != nil {
		rb.t.Fatalf("write %s: %v", abs, err)
	}
}

// commitWithFixedDate runs `git commit` with both AUTHOR and
// COMMITTER dates pinned to rb.clock. Pinning both is critical:
// committer date drives chronological ordering and feeds into the
// commit SHA, so a wall-clock leak would break SHA-stable assertions.
func (rb *RepoBuilder) commitWithFixedDate(message string) {
	rb.t.Helper()
	stamp := rb.clock.Format(time.RFC3339)
	cmd := exec.Command("git", "-C", rb.dir, "commit", "-q", "-m", message)
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_DATE="+stamp,
		"GIT_COMMITTER_DATE="+stamp,
		"GIT_AUTHOR_NAME="+rb.authorN,
		"GIT_AUTHOR_EMAIL="+rb.authorE,
		"GIT_COMMITTER_NAME="+rb.authorN,
		"GIT_COMMITTER_EMAIL="+rb.authorE,
	)
	rb.mustRun(cmd, "git commit")
}

// head returns the current HEAD SHA. Used as the return value of
// SeedCommit so tests can build dedupe assertions without re-shelling.
func (rb *RepoBuilder) head() string {
	rb.t.Helper()
	out := rb.runGit("rev-parse", "HEAD")
	return trimNewline(out)
}

// runGit shells out to `git -C <dir> <args...>` and fatals on error.
// Returns stdout so callers like head() can parse single-line output.
func (rb *RepoBuilder) runGit(args ...string) string {
	rb.t.Helper()
	cmd := exec.Command("git", append([]string{"-C", rb.dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		rb.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// mustRun executes cmd and surfaces stderr-bearing combined output
// in the failure message. Used by code paths that already shaped
// the cmd themselves (e.g. commitWithFixedDate's env-bearing call).
func (rb *RepoBuilder) mustRun(cmd *exec.Cmd, label string) {
	rb.t.Helper()
	out, err := cmd.CombinedOutput()
	if err != nil {
		rb.t.Fatalf("%s: %v\n%s", label, err, out)
	}
}

// skipIfNoGit short-circuits the test cleanly when no git binary is
// on PATH. Hermetic sandboxes (and minimal CI images) routinely lack
// git — failing those runs would block unrelated work.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("skipping: git not on PATH (%v)", err)
	}
}

// trimNewline drops one trailing \n / \r\n, matching the contract
// callers expect from `git rev-parse` style single-line outputs.
func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
