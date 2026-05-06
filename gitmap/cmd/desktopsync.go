package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
)

// runDesktopSync handles the "desktop-sync" subcommand.
func runDesktopSync() {
	outputDir := constants.DefaultOutputFolder
	jsonPath := filepath.Join(outputDir, constants.DefaultJSONFile)
	validateDesktopSyncPaths(outputDir, jsonPath)
	records := loadDesktopRecords(jsonPath)
	syncToDesktop(records, jsonPath)
}

// validateDesktopSyncPaths checks that the output dir and JSON file exist.
func validateDesktopSyncPaths(outputDir, jsonPath string) {
	info, err := os.Stat(outputDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, constants.MsgNoOutputDir)
		os.Exit(1)
	}
	_, jsonErr := os.Stat(jsonPath)
	if jsonErr != nil {
		fmt.Fprintf(os.Stderr, constants.MsgNoJSONFile, jsonPath)
		os.Exit(1)
	}
}

// loadDesktopRecords reads and parses the JSON file into ScanRecords.
func loadDesktopRecords(path string) []model.ScanRecord {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrDesktopReadFailed, path, err)
		os.Exit(1)
	}
	var records []model.ScanRecord
	err = json.Unmarshal(data, &records)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrDesktopParseFailed, path, err)
		os.Exit(1)
	}

	return records
}

// syncToDesktop registers each repo with GitHub Desktop.
func syncToDesktop(records []model.ScanRecord, source string) {
	_, err := exec.LookPath(constants.GitHubDesktopBin)
	if err == nil {
		fmt.Printf(constants.MsgDesktopSyncStart, source)
		added, skipped, failed := syncAll(records)
		fmt.Printf(constants.MsgDesktopSyncDone, added, skipped, failed)

		return
	}
	fmt.Fprintln(os.Stderr, constants.MsgDesktopNotFound)
	os.Exit(1)
}

// syncAll iterates records and syncs each to GitHub Desktop.
func syncAll(records []model.ScanRecord) (added, skipped, failed int) {
	for _, r := range records {
		result := syncOne(r)
		added, skipped, failed = tallyResult(result, added, skipped, failed)
	}

	return added, skipped, failed
}

// syncResult represents the outcome of syncing one repo.
type syncResult int

const (
	syncAdded syncResult = iota
	syncSkipped
	syncFailed
)

// syncOne attempts to register a single repo with GitHub Desktop.
func syncOne(r model.ScanRecord) syncResult {
	if len(r.AbsolutePath) == 0 {
		fmt.Printf(constants.MsgDesktopSyncFailed, r.RepoName, constants.ErrNoAbsPath)

		return syncFailed
	}

	return syncExistingPath(r)
}

// syncExistingPath checks path existence and registers with Desktop.
func syncExistingPath(r model.ScanRecord) syncResult {
	_, err := os.Stat(r.AbsolutePath)
	if err == nil {
		return registerOne(r.RepoName, r.AbsolutePath)
	}
	fmt.Printf(constants.MsgDesktopSyncSkipped, r.RepoName)

	return syncSkipped
}

// registerOne calls the GitHub Desktop CLI for a single repo.
func registerOne(name, repoPath string) syncResult {
	cmd := exec.Command(constants.GitHubDesktopBin, repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(constants.MsgDesktopSyncFailed, name, fmt.Sprintf("%v: %s", err, output))

		return syncFailed
	}
	fmt.Printf(constants.MsgDesktopSyncAdded, name)

	return syncAdded
}

// tallyResult increments the appropriate counter.
func tallyResult(r syncResult, added, skipped, failed int) (int, int, int) {
	if r == syncAdded {
		return added + 1, skipped, failed
	}
	if r == syncSkipped {
		return added, skipped + 1, failed
	}

	return added, skipped, failed + 1
}
