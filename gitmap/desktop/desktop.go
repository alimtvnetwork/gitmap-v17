// Package desktop integrates with GitHub Desktop application.
package desktop

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
)

// AddRepos registers discovered repositories with GitHub Desktop.
func AddRepos(records []model.ScanRecord) DesktopSummary {
	summary := DesktopSummary{}
	if isInstalled() {
		return addAll(records, summary)
	}
	fmt.Println(constants.MsgDesktopNotFound)

	return summary
}

// isInstalled checks if GitHub Desktop CLI is available.
func isInstalled() bool {
	_, err := exec.LookPath(constants.GitHubDesktopBin)

	return err == nil
}

// addAll iterates records and adds each to GitHub Desktop.
func addAll(records []model.ScanRecord, summary DesktopSummary) DesktopSummary {
	for _, rec := range records {
		err := addOne(rec.AbsolutePath)
		summary = updateSummary(summary, rec.RepoName, err)
	}

	return summary
}

// addOne opens a single repo in GitHub Desktop.
func addOne(repoPath string) error {
	cmd := buildCommand(repoPath)
	_, err := cmd.Output()

	return err
}

// buildCommand creates the platform-appropriate command.
func buildCommand(repoPath string) *exec.Cmd {
	if runtime.GOOS == constants.OSWindows {
		return exec.Command(constants.GitHubDesktopBin, repoPath)
	}

	return exec.Command(constants.GitHubDesktopBin, repoPath)
}

// updateSummary tracks success/failure for each repo.
func updateSummary(s DesktopSummary, name string, err error) DesktopSummary {
	if err == nil {
		s.Added++
		fmt.Printf(constants.MsgDesktopAdded, name)

		return s
	}
	s.Failed++
	fmt.Printf(constants.MsgDesktopFailed, name, err)

	return s
}

// DesktopSummary tracks GitHub Desktop registration results.
type DesktopSummary struct {
	Added  int
	Failed int
}
