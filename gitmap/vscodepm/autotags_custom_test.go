package vscodepm

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// makeRepoFixture writes a few marker files so DetectTagsCustom has
// something concrete to inspect. Returns the temp dir path.
func makeRepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{".git", "go.mod", "Gemfile"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatalf("seed %q: %v", name, err)
		}
	}

	return root
}

// resetTagEnv unsets every tag env var so cases run in isolation.
func resetTagEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		constants.EnvVSCodeTagAdd,
		constants.EnvVSCodeTagSkip,
		constants.EnvVSCodeTagMarker,
	} {
		os.Unsetenv(k)
	}
}

// TestDetectTagsCustomNoEnvMatchesBuiltin asserts that with no env
// overrides, DetectTagsCustom returns exactly what DetectTags does.
func TestDetectTagsCustomNoEnvMatchesBuiltin(t *testing.T) {
	resetTagEnv(t)
	root := makeRepoFixture(t)

	got := DetectTagsCustom(root)
	want := DetectTags(root)
	if !reflect.DeepEqual(sortCopy(got), sortCopy(want)) {
		t.Errorf("custom = %v, builtin = %v (want equal)", got, want)
	}
}

// TestDetectTagsCustomSkipsAndAdds covers the skip + add overlay,
// including the documented case where the same tag appears in both
// (always-add wins because it runs after skip).
func TestDetectTagsCustomSkipsAndAdds(t *testing.T) {
	resetTagEnv(t)
	defer resetTagEnv(t)
	root := makeRepoFixture(t)

	os.Setenv(constants.EnvVSCodeTagSkip, "git")
	os.Setenv(constants.EnvVSCodeTagAdd, "work"+constants.EnvVSCodeTagSeparator+"git")

	got := DetectTagsCustom(root)
	// Detected portion = everything except the trailing always-add slice.
	addCount := 2 // "work" and "git"
	cutoff := len(got) - addCount
	if cutoff < 0 {
		cutoff = 0
	}
	if containsString(got[:cutoff], "git") {
		t.Errorf("git should be skipped from detected portion, got %v", got)
	}
	if !containsString(got, "work") || !containsString(got, "git") {
		t.Errorf("always-add tags missing, got %v", got)
	}
}

// TestDetectTagsCustomCustomMarker registers a Gemfile→ruby rule and
// verifies it shows up after the canonical built-in tags.
func TestDetectTagsCustomCustomMarker(t *testing.T) {
	resetTagEnv(t)
	defer resetTagEnv(t)
	root := makeRepoFixture(t)

	os.Setenv(constants.EnvVSCodeTagMarker, "Gemfile=ruby")

	got := DetectTagsCustom(root)
	if !containsString(got, "ruby") {
		t.Errorf("custom marker tag missing, got %v", got)
	}
}

func containsString(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}

	return false
}

func sortCopy(s []string) []string {
	c := append([]string{}, s...)
	sort.Strings(c)

	return c
}
