// Package cmd — clonefixrepo.go: entry points for `gitmap clone-fix-repo`
// (alias `cfr`) and `gitmap clone-fix-repo-pub` (alias `cfrp`).
//
// These are convenience pipelines that chain three existing commands
// in one shot:
//
//	cfr  : clone <url>  →  cd <folder>  →  fix-repo --all
//	cfrp : clone <url>  →  cd <folder>  →  fix-repo --all  →  make-public --yes
//
// Implementation strategy: the chained commands (runFixRepo,
// runMakePublic) all call os.Exit at the end, which would terminate
// our parent process before the next step runs. To stay decoupled
// and side-effect-clean, we shell out to our own binary (resolved
// via os.Executable) for the fix-repo and make-public steps after
// invoking executeDirectClone in-process. This also keeps each
// step's exit code, stdout, and stderr semantics intact.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/clonenext"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// runCloneFixRepo implements `gitmap clone-fix-repo` (alias cfr).
func runCloneFixRepo(args []string) {
	checkHelp(constants.CmdCloneFixRepo, args)
	runCloneFixRepoPipeline(args, false)
}

// runCloneFixRepoPub implements `gitmap clone-fix-repo-pub` (alias cfrp).
func runCloneFixRepoPub(args []string) {
	checkHelp(constants.CmdCloneFixRepoPub, args)
	runCloneFixRepoPipeline(args, true)
}

// runCloneFixRepoPipeline is the shared core. `makePublic` controls
// whether the optional 3rd step (visibility flip) runs.
func runCloneFixRepoPipeline(args []string, makePublic bool) {
	url, folderName, noVSCodeSync := parseCloneFixRepoArgs(args)
	if len(url) == 0 {
		fmt.Fprint(os.Stderr, constants.ErrCloneFixRepoUsage)
		os.Exit(constants.ExitCloneFixRepoBadFlag)
	}

	absPath := resolveCloneTargetFolder(url, folderName)
	requireOnline()
	executeDirectClone(url, folderName, true, false, "", noVSCodeSync)

	if err := os.Chdir(absPath); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneFixRepoChdirFmt, absPath, err)
		os.Exit(constants.ExitCloneFixRepoChdir)
	}

	runChainedGitmapStep([]string{constants.CmdFixRepo, "--" + constants.FixRepoFlagAll})
	if makePublic {
		runChainedGitmapStep([]string{constants.CmdMakePublic, "--" + constants.FlagVisYes})
	}
	fmt.Printf(constants.MsgCloneFixRepoDone, absPath)
}

// parseCloneFixRepoArgs returns (url, folderName). The first
// non-flag arg is the URL; an optional second non-flag arg is the
// destination folder. Unknown flags are ignored at this layer
// because clone itself accepts a wide flag surface.
//
// The cfr/cfrp pipelines also recognize `--no-vscode-sync` so the
// projects.json update at the end of the inner clone step can be
// suppressed in CI / headless runs without VS Code installed.
func parseCloneFixRepoArgs(args []string) (string, string, bool) {
	positional := make([]string, 0, len(args))
	noVSCodeSync := false
	syncFlag := "--" + constants.FlagNoVSCodeSync
	for _, a := range args {
		if a == syncFlag {
			noVSCodeSync = true

			continue
		}
		if len(a) > 0 && a[0] != '-' {
			positional = append(positional, a)
		}
	}
	url := ""
	folder := ""
	if len(positional) > 0 {
		url = positional[0]
	}
	if len(positional) > 1 {
		folder = positional[1]
	}

	return url, folder, noVSCodeSync
}

// resolveCloneTargetFolder mirrors the folder-naming logic in
// executeDirectClone so we know which directory to cd into after
// the clone step finishes. Versioned URLs auto-flatten to BaseName.
func resolveCloneTargetFolder(url, folderName string) string {
	if len(folderName) == 0 {
		repoName := repoNameFromURL(url)
		parsed := clonenext.ParseRepoName(repoName)
		if parsed.HasVersion {
			folderName = parsed.BaseName
		} else {
			folderName = repoName
		}
	}
	abs, err := filepath.Abs(folderName)
	if err != nil {
		return folderName
	}

	return abs
}

// runChainedGitmapStep re-execs the current gitmap binary with the
// given args, streaming stdin/stdout/stderr through. Any non-zero
// exit propagates immediately so the pipeline halts on first failure.
func runChainedGitmapStep(args []string) {
	bin, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneFixRepoExecFmt, err)
		os.Exit(constants.ExitCloneFixRepoChainFailed)
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, constants.ErrCloneFixRepoExecFmt, runErr)
		os.Exit(constants.ExitCloneFixRepoChainFailed)
	}
}
