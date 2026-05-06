package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/clonenext"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/desktop"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/gitutil"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/lockcheck"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/verbose"
)

// runCloneNext handles the "clone-next" subcommand.
//
// Form 1 — `gitmap cn vX.Y.Z`         : operates on the current repo.
// Form 2 — `gitmap cn <repo> vX.Y.Z`  : cross-dir — chdir into <repo>, run
//
//	clone-next, chdir back. See
//	`clonenextcrossdir.go`.
func runCloneNext(args []string) {
	// v3.117.0: folder-arg dispatch runs FIRST so path-shaped tokens
	// win over the release-alias resolver. Order matters — see
	// spec/01-app/111-cn-folder-arg.md §Disambiguation.
	if tryFolderArgCloneNext(args) {
		return
	}
	if tryCrossDirCloneNext(args) {
		return
	}
	checkHelp("clone-next", args)
	cnFlags := parseCloneNextFlags(args)
	setCmdFaithfulVerify(cnFlags.VerifyCmdFaithful)
	setCmdFaithfulExitOnMismatch(cnFlags.VerifyCmdFaithfulExitOnMismatch)
	setCmdPrintArgv(cnFlags.PrintCloneArgv)

	if cnFlags.Verbose {
		log, err := verbose.Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, constants.WarnVerboseLogFailed, err)
		} else {
			defer log.Close()
		}
	}

	// Batch mode: --csv, --all, or implicit (cwd is not a git repo
	// but has git subdirs one level down) triggers the multi-repo
	// dispatcher. See shouldRunBatch for the priority order.
	if shouldRunBatch(cnFlags, currentWorkingDir()) {
		if cnFlags.DryRun {
			previewDryRunBatch(cnFlags.CSVPath, cnFlags.All)

			return
		}
		runCloneNextBatch(cnFlags.CSVPath, cnFlags.All, cnFlags.MaxConcurrency, cnFlags.NoProgress, cnFlags.ReportErrors)
		maybeExitOnCmdFaithfulMismatch()

		return
	}

	if len(cnFlags.VersionArg) == 0 {
		fmt.Fprintln(os.Stderr, constants.ErrCloneNextUsage)
		exitWith(1)
	}

	requireOnline()
	applySSHKey(cnFlags.SSHKeyName)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextCwd, err)
		exitWith(1)
	}

	remoteURL, err := gitutil.RemoteURL(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextNoRemote, err)
		exitWith(1)
	}

	currentFolder := filepath.Base(cwd)
	parentDir := filepath.Dir(cwd)

	// Strip .git suffix from remote URL for repo name extraction.
	repoName := extractRepoName(remoteURL)

	parsed := clonenext.ParseRepoName(repoName)
	targetVersion, err := clonenext.ResolveTarget(parsed, cnFlags.VersionArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextBadVersion, err)
		exitWith(1)
	}

	targetName := clonenext.TargetRepoName(parsed.BaseName, targetVersion)
	targetURL := clonenext.ReplaceRepoInURL(remoteURL, repoName, targetName)

	// Flatten by default: clone into base name folder (no version suffix).
	flattenedFolder := parsed.BaseName
	targetPath := filepath.Join(parentDir, flattenedFolder)

	// --output terminal: print the standardized RepoTermBlock so
	// the cn pre-clone summary matches the shape used by scan,
	// clone-from, and probe. Emitted BEFORE the legacy stage
	// banners so it's always the first thing the user sees.
	maybePrintCloneNextTermBlock(cnFlags, targetName, currentBranch(cwd),
		remoteURL, targetURL, targetPath)

	// Stage 1/3 banner — only emitted in -f mode where the multi-stage
	// nature actually helps. Default mode keeps the legacy terse output
	// so we don't break existing screenshots / scripts that grep it.
	if cnFlags.Force {
		fmt.Printf(constants.MsgCNStagePrepare, currentFolder, flattenedFolder)
	}

	// Force-flatten pre-step: if the user passed -f / --force AND their
	// shell cwd is exactly the target folder (the "already flattened"
	// case from a previous cn run), Windows holds an open handle on
	// the cwd that prevents os.RemoveAll. Chdir-to-parent here releases
	// that handle BEFORE the existence check below tries to remove it.
	// Linux/macOS don't strictly need this, but doing it unconditionally
	// keeps the code path simple and gives the same UX everywhere.
	if cnFlags.Force && samePath(cwd, targetPath) {
		fmt.Printf(constants.MsgForceReleasing, cwd)
		if chErr := os.Chdir(parentDir); chErr != nil {
			fmt.Fprintf(os.Stderr, constants.ErrCloneNextForceFailed,
				flattenedFolder, chErr, flattenedFolder)
			exitWith(1)
		}
	}

	// If the flattened folder already exists, try to remove it for a fresh clone.
	// On Windows, the current shell's working directory is locked and cannot be
	// removed by this process. In that case, fall back to a versioned folder name
	// (e.g. scripts-fixer-v2) and warn — UNLESS -f was passed, which refuses the
	// fallback and aborts so the user gets either a flat layout or a clear error.
	if _, statErr := os.Stat(targetPath); statErr == nil {
		fmt.Printf(constants.MsgFlattenRemoving, flattenedFolder)
		if removeErr := os.RemoveAll(targetPath); removeErr != nil {
			if cnFlags.Force {
				// Strict force contract: do NOT silently rename to
				// macro-ahk-v22/. The whole point of -f is "I want flat
				// or nothing" — degrading would be a footgun.
				fmt.Fprintf(os.Stderr, constants.ErrCloneNextForceFailed,
					flattenedFolder, removeErr, flattenedFolder)
				exitWith(1)
			}
			fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, flattenedFolder, removeErr)
			fallbackFolder := targetName
			fallbackPath := filepath.Join(parentDir, fallbackFolder)
			fmt.Printf(constants.MsgFlattenFallback, fallbackFolder)
			fmt.Printf(constants.MsgFlattenLockedHint, flattenedFolder)
			// If the versioned fallback also exists, attempt to remove it; if that
			// fails too, warn but continue — git clone will surface a clear error.
			if _, fbStat := os.Stat(fallbackPath); fbStat == nil {
				if fbErr := os.RemoveAll(fallbackPath); fbErr != nil {
					fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, fallbackFolder, fbErr)
				}
			}
			flattenedFolder = fallbackFolder
			targetPath = fallbackPath
		}
	}

	// Optionally check and create the target GitHub repo when --create-remote is set.
	if cnFlags.CreateRemote {
		owner, _, parseErr := clonenext.ParseOwnerRepo(remoteURL)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, constants.ErrCloneNextRemoteParse, parseErr)
			exitWith(1)
		}

		exists, checkErr := clonenext.RepoExists(owner, targetName)
		if checkErr != nil {
			fmt.Fprintf(os.Stderr, constants.ErrCloneNextRepoCheck, checkErr)
			exitWith(1)
		}

		if !exists {
			fmt.Printf(constants.MsgCloneNextCreating, targetName)
			createErr := clonenext.CreateRepo(owner, targetName, true)
			if createErr != nil {
				fmt.Fprintf(os.Stderr, constants.ErrCloneNextRepoCreate, targetName, createErr)
				exitWith(1)
			}
			fmt.Printf(constants.MsgCloneNextCreated, targetName)
		}
	}

	// Dry-run gate: print the planned clone command and exit BEFORE
	// any side effect (clone, removal, DB write, GH Desktop, VS Code,
	// shell handoff). Placed after target/folder resolution so the
	// previewed url+dest match exactly what a real run would invoke.
	if cnFlags.DryRun {
		printCloneNextDryRun(targetURL, targetPath)
	}

	if cnFlags.Force {
		fmt.Printf(constants.MsgCNStageClone, targetName)
	}
	fmt.Printf(constants.MsgFlattenCloning, targetName, flattenedFolder)
	cloneResult := runGitClone(targetURL, targetPath)
	if !cloneResult {
		fmt.Fprintf(os.Stderr, constants.ErrCloneNextFailed, targetName)
		exitWith(1)
	}
	fmt.Printf(constants.MsgFlattenDone, targetName, flattenedFolder)

	if cnFlags.Force {
		fmt.Printf(constants.MsgCNStageFinalize)
	}

	// Record version history in DB.
	recordVersionHistory(targetPath, parsed.CurrentVersion, targetVersion, flattenedFolder)

	if !cnFlags.NoDesktop {
		registerCloneNextDesktop(targetName, targetPath)
	}

	// Handle removal of the old versioned folder (only if different from flattened path).
	// With -f / --force the user has already opted into a flat layout, so we
	// auto-skip the "Remove current folder?" prompt and the lock-detector loop
	// that follows it. Without -f, behavior is unchanged.
	if currentFolder != flattenedFolder {
		keep := cnFlags.Keep || cnFlags.Force
		handleCloneNextRemoval(currentFolder, cwd, targetPath, cnFlags.Delete, keep)
	}

	// Shell handoff: write target path to the wrapper's sentinel file
	// so the parent shell can cd into the new flattened folder.
	WriteShellHandoff(targetPath)

	// Open in VS Code if available.
	openInVSCode(targetPath)

	// VS Code Project Manager: register the freshly-cloned (and
	// flattened) repo so it appears in the sidebar without a
	// separate `gitmap code` step. Soft-fails when VS Code or the
	// extension is missing.
	syncSingleClonedRepoToVSCodePM(targetPath, flattenedFolder, cnFlags.NoVSCodeSync)

	if cnFlags.Force {
		fmt.Printf(constants.MsgCNDone, flattenedFolder)
	}
	maybeExitOnCmdFaithfulMismatch()
}

// extractRepoName extracts the repository name from a remote URL.
func extractRepoName(remoteURL string) string {
	name := remoteURL
	// Remove trailing .git
	name = strings.TrimSuffix(name, ".git")
	// Get last path segment
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	if idx := strings.LastIndex(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}

	return name
}

// runGitClone executes git clone and returns success status.
func runGitClone(url, dest string) bool {
	cmd := exec.Command(constants.GitBin, constants.GitClone, url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() == nil
}

// registerCloneNextDesktop registers the cloned repo with GitHub Desktop.
func registerCloneNextDesktop(name, absPath string) {
	records := []model.ScanRecord{{
		RepoName:     name,
		AbsolutePath: absPath,
	}}
	result := desktop.AddRepos(records)
	if result.Added > 0 {
		fmt.Printf(constants.MsgCloneNextDesktop, name)
	}
}

// handleCloneNextRemoval manages removal of the current version folder.
// It changes to the parent directory first to release file locks on Windows.
func handleCloneNextRemoval(folderName, fullPath, targetPath string, deleteFlag, keepFlag bool) {
	if keepFlag {
		return
	}

	// Move out of the folder before attempting removal to avoid Windows file locks.
	parentDir := filepath.Dir(fullPath)
	if chErr := os.Chdir(parentDir); chErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not cd to %s: %v\n", parentDir, chErr)
	}

	removed := false
	var shouldRemove bool

	if deleteFlag {
		shouldRemove = true
	} else {
		// Prompt
		fmt.Printf(constants.MsgCloneNextRemovePrompt, folderName)
		var answer string
		_, _ = fmt.Scanln(&answer)
		shouldRemove = strings.ToLower(strings.TrimSpace(answer)) == "y"
	}

	if shouldRemove {
		removed = removeFolderWithLockCheck(folderName, fullPath)
	}

	// After removing the old folder, move into the newly cloned directory.
	if removed {
		if chErr := os.Chdir(targetPath); chErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not cd to %s: %v\n", targetPath, chErr)
		} else {
			fmt.Printf(constants.MsgCloneNextMovedTo, filepath.Base(targetPath))
		}
	}
}

// removeFolderWithLockCheck attempts to remove a directory, and if it fails,
// scans for locking processes and offers to terminate them before retrying.
// All removal attempts are tracked as pending tasks in the database.
func removeFolderWithLockCheck(name, path string) bool {
	// Record the delete intent as a pending task before any OS operation.
	taskID, db := createPendingTask(constants.TaskTypeDelete, path, "", constants.CmdCloneNext, "")
	if db != nil {
		defer db.Close()
	}

	// First attempt.
	err := os.RemoveAll(path)
	if err == nil {
		fmt.Printf(constants.MsgCloneNextRemoved, name)
		completePendingTask(db, taskID)

		return true
	}

	// Removal failed — scan for locking processes.
	fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, name, err)
	fmt.Printf(constants.MsgLockCheckScanning, name)

	procs, scanErr := lockcheck.FindLockingProcesses(path)
	if scanErr != nil {
		fmt.Fprintf(os.Stderr, constants.WarnLockCheckScanFailed, scanErr)
		failPendingTask(db, taskID, fmt.Sprintf(constants.ReasonLockScanFailed, scanErr))

		return false
	}

	if len(procs) == 0 {
		fmt.Print(constants.MsgLockCheckNoneFound)
		failPendingTask(db, taskID, fmt.Sprintf(constants.ReasonNoLockingProcs, err))

		return false
	}

	// Show locking processes and prompt to kill.
	fmt.Printf(constants.MsgLockCheckFound, lockcheck.FormatProcessList(procs))
	fmt.Print(constants.MsgLockCheckKillPrompt)

	var answer string
	_, _ = fmt.Scanln(&answer)
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		failPendingTask(db, taskID, constants.ReasonUserDeclined)

		return false
	}

	// Terminate each process.
	for _, p := range procs {
		fmt.Printf(constants.MsgLockCheckKilling, p.Name, p.PID)
		killErr := lockcheck.KillProcess(p.PID)
		if killErr != nil {
			fmt.Fprintf(os.Stderr, constants.WarnLockCheckKillFailed, p.Name, p.PID, killErr)
		} else {
			fmt.Printf(constants.MsgLockCheckKilled, p.Name)
		}
	}

	// Brief pause to let OS release handles.
	time.Sleep(500 * time.Millisecond)

	// Retry removal.
	fmt.Print(constants.MsgLockCheckRetrying)
	retryErr := os.RemoveAll(path)
	if retryErr != nil {
		fmt.Fprintf(os.Stderr, constants.WarnCloneNextRemoveFailed, name, retryErr)
		failPendingTask(db, taskID, fmt.Sprintf(constants.ReasonRetryFailed, retryErr))

		return false
	}

	fmt.Printf(constants.MsgCloneNextRemoved, name)
	completePendingTask(db, taskID)

	return true
}
