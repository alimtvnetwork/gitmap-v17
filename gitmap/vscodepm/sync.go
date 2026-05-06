package vscodepm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// Pair is one (rootPath, name, paths, tags) tuple to upsert into
// projects.json. Both Paths (multi-root, v3.39.0+) and Tags (auto-derived,
// v3.40.0+) are UNIONed with whatever the user already has on disk so
// gitmap never silently removes a user-added path or tag.
type Pair struct {
	RootPath string
	Name     string
	Paths    []string
	Tags     []string
}

// Sync reconciles projects.json with the supplied DB-side pairs.
//
// Behavior:
//   - New rootPath -> append a default Entry with Paths = pair.Paths.
//   - Existing rootPath -> update Name. Paths becomes UNION(existing, pair.Paths).
//     Tags / Enabled / Profile are preserved (so user edits in the VS Code UI
//     survive untouched).
//   - Foreign entries (rootPath not in pairs) -> preserved verbatim.
//
// Writes are atomic: temp file in the same directory then os.Rename.
// Returns ErrUserDataMissing / ErrExtensionMissing when the path
// cannot be resolved — callers should treat those as soft skips.
func Sync(pairs []Pair) (SyncSummary, error) {
	path, err := ProjectsJSONPath()
	if err != nil {
		return SyncSummary{}, err
	}

	existing, err := readEntries(path)
	if err != nil {
		return SyncSummary{}, err
	}

	merged, summary := mergePairs(existing, pairs)

	if err := writeEntriesAtomic(path, merged); err != nil {
		return summary, err
	}

	summary.Total = len(merged)

	return summary, nil
}

// RenameByPath updates the Name field of the entry whose rootPath matches.
// Paths / Tags / Enabled / Profile are intentionally left alone.
// Returns true when an entry was actually renamed (false = no-op).
func RenameByPath(rootPath, newName string) (bool, error) {
	path, err := ProjectsJSONPath()
	if err != nil {
		return false, err
	}

	entries, err := readEntries(path)
	if err != nil {
		return false, err
	}

	changed := false

	for i := range entries {
		if pathsEqual(entries[i].RootPath, rootPath) && entries[i].Name != newName {
			entries[i].Name = newName
			changed = true
		}
	}

	if !changed {
		return false, nil
	}

	return true, writeEntriesAtomic(path, entries)
}

// readEntries returns the parsed entries. Missing file -> empty slice.
func readEntries(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, nil
		}

		return nil, fmt.Errorf(constants.ErrVSCodePMReadFailed, path, err)
	}

	if len(data) == 0 {
		return []Entry{}, nil
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf(constants.ErrVSCodePMParseFailed, path, err)
	}

	for i := range entries {
		entries[i] = ensureSlices(entries[i])
	}

	return entries, nil
}

// mergePairs upserts pairs into existing by rootPath. Returns the new
// slice plus an Added/Updated/Unchanged summary (Total set by caller).
//
// "Updated" counts when EITHER Name OR the union'd Paths set actually
// changes vs. what's currently on disk. Pure no-ops increment Unchanged.
func mergePairs(existing []Entry, pairs []Pair) ([]Entry, SyncSummary) {
	indexByPath := make(map[string]int, len(existing))
	for i, e := range existing {
		indexByPath[normalizePath(e.RootPath)] = i
	}

	summary := SyncSummary{}

	for _, p := range pairs {
		key := normalizePath(p.RootPath)

		idx, found := indexByPath[key]
		if !found {
			existing = append(existing, newEntry(p.RootPath, p.Name, p.Paths, p.Tags))
			indexByPath[key] = len(existing) - 1
			summary.Added++

			continue
		}

		mergedPaths := unionPaths(existing[idx].Paths, p.Paths)
		mergedTags := unionTags(existing[idx].Tags, p.Tags)
		nameChanged := existing[idx].Name != p.Name
		pathsChanged := len(mergedPaths) != len(existing[idx].Paths)
		tagsChanged := len(mergedTags) != len(existing[idx].Tags)

		if !nameChanged && !pathsChanged && !tagsChanged {
			summary.Unchanged++

			continue
		}

		existing[idx].Name = p.Name
		existing[idx].Paths = mergedPaths
		existing[idx].Tags = mergedTags
		summary.Updated++
	}

	return existing, summary
}

// writeEntriesAtomic encodes entries to a sibling .tmp then renames.
// On Windows, os.Rename overwrites the destination; on Unix it does too,
// but we explicitly remove the destination first if rename fails to keep
// behavior consistent.
func writeEntriesAtomic(path string, entries []Entry) error {
	tmpPath := path + constants.VSCodePMProjectsTempSuffix

	if err := writeEntriesToFile(tmpPath, entries); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)

		return fmt.Errorf(constants.ErrVSCodePMRenameFailed,
			filepath.Base(path), err)
	}

	return nil
}

// writeEntriesToFile serializes entries with tab indent + trailing newline.
func writeEntriesToFile(path string, entries []Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), constants.DirPermission); err != nil {
		return fmt.Errorf(constants.ErrVSCodePMWriteTempFailed, path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf(constants.ErrVSCodePMWriteTempFailed, path, err)
	}

	if err := encodeEntries(file, entries); err != nil {
		_ = file.Close()
		_ = os.Remove(path)

		return fmt.Errorf(constants.ErrVSCodePMWriteTempFailed, path, err)
	}

	return file.Close()
}

// encodeEntries writes entries as pretty JSON with a trailing newline.
func encodeEntries(w io.Writer, entries []Entry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", constants.VSCodePMJSONIndent)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(entries); err != nil {
		return err
	}

	return nil
}

// normalizePath returns the canonical key used for rootPath comparisons.
// Case-insensitive on Windows, case-sensitive elsewhere.
func normalizePath(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(filepath.Clean(p))
	}

	return filepath.Clean(p)
}

// pathsEqual compares two paths using normalizePath.
func pathsEqual(a, b string) bool {
	return normalizePath(a) == normalizePath(b)
}
