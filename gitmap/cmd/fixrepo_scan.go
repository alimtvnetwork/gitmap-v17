package cmd

// File enumeration + binary-detection helpers for `gitmap fix-repo`.
// Mirrors scripts/fix-repo/File-Scan.ps1: list tracked files via
// `git ls-files`, skip reparse points / oversized / binary-extension
// / NUL-byte-prefixed files.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// fixRepoBinaryExts is the suffix set we treat as opaque assets. The
// values match the PowerShell + Bash scripts so all three engines
// agree on which files are scanned.
var fixRepoBinaryExts = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".webp": {},
	".ico": {}, ".pdf": {}, ".zip": {}, ".tar": {}, ".gz": {},
	".tgz": {}, ".bz2": {}, ".xz": {}, ".7z": {}, ".rar": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".otf": {}, ".eot": {},
	".mp3": {}, ".mp4": {}, ".mov": {}, ".wav": {}, ".ogg": {},
	".webm": {}, ".class": {}, ".jar": {}, ".so": {}, ".dylib": {},
	".dll": {}, ".exe": {}, ".pyc": {},
}

// fixRepoSweepResult aggregates one full sweep's counts.
type fixRepoSweepResult struct {
	scanned      int
	changed      int
	replacements int
	failed       bool
	goFiles      []string // absolute paths of modified .go files (for gofmt)
}

// runFixRepoSweep enumerates tracked files and rewrites each.
func runFixRepoSweep(identity fixRepoIdentity, targets []int, opts fixRepoOptions) fixRepoSweepResult {
	files := listTrackedFiles(identity.root)
	result := fixRepoSweepResult{}
	for _, rel := range files {
		processFixRepoFile(rel, identity, targets, opts, &result)
	}

	return result
}

// processFixRepoFile is the per-file branch extracted from the sweep
// loop so runFixRepoSweep stays under the 15-line cap.
func processFixRepoFile(rel string, identity fixRepoIdentity, targets []int,
	opts fixRepoOptions, result *fixRepoSweepResult,
) {
	full := filepath.Join(identity.root, rel)
	if isFixRepoIgnoredPath(rel) {
		return
	}
	if !isFixRepoScannable(full) {
		return
	}
	result.scanned++
	reps, err := rewriteFixRepoFile(full, identity.base, identity.current, targets, opts.isDryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.FixRepoErrWriteFmt, rel, err)
		result.failed = true

		return
	}
	if reps > 0 {
		result.changed++
		result.replacements += reps
		if isGoSourceFile(rel) {
			result.goFiles = append(result.goFiles, full)
		}
		if opts.isVerbose {
			fmt.Printf(constants.FixRepoMsgModified, rel, reps)
		}
	}
}

// listTrackedFiles runs `git ls-files` in repoRoot. Failures yield
// an empty list (the caller logs nothing because git already wrote
// to stderr) and an empty sweep is the natural no-op.
func listTrackedFiles(repoRoot string) []string {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = repoRoot
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	files := make([]string, 0, len(lines))
	for _, l := range lines {
		if l != "" {
			files = append(files, l)
		}
	}

	return files
}

// isFixRepoScannable composes the per-file skip checks: reparse,
// oversize, binary-extension, NUL-byte prefix.
func isFixRepoScannable(fullPath string) bool {
	if isFixRepoSkippablePath(fullPath) {
		return false
	}
	if isFixRepoBinaryExt(fullPath) {
		return false
	}
	if hasFixRepoNullByte(fullPath) {
		return false
	}

	return true
}

// isFixRepoSkippablePath flags reparse points and >5 MiB files.
func isFixRepoSkippablePath(fullPath string) bool {
	info, err := os.Lstat(fullPath)
	if err != nil {
		return true
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	if info.Size() > constants.FixRepoMaxFileBytes {
		return true
	}

	return false
}

// isFixRepoBinaryExt reports whether fullPath has a known binary extension.
func isFixRepoBinaryExt(fullPath string) bool {
	ext := strings.ToLower(filepath.Ext(fullPath))
	_, ok := fixRepoBinaryExts[ext]

	return ok
}

// hasFixRepoNullByte checks the first 8 KiB for a NUL byte. A NUL
// in early bytes is the standard "this is binary" heuristic used by
// git, grep, etc., and matches the PowerShell script's behavior.
func hasFixRepoNullByte(fullPath string) bool {
	f, err := os.Open(fullPath)
	if err != nil {
		return true
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, constants.FixRepoBinarySniffMax)
	n, _ := f.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}
