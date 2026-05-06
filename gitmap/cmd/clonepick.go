package cmd

// CLI entry point for `gitmap clone-pick <repo-url> <paths>` (spec
// 100, v3.153.0+). Sparse-checkout a subset of a git repo into the
// current working directory (or --dest), and auto-save the selection
// to the CloneInteractiveSelection table.
//
// Exit codes:
//
//   0   -- dry-run rendered OR clone succeeded
//   1   -- runtime failure (git, fs, db)
//   2   -- bad CLI usage (missing args, invalid flag value)
//   130 -- user canceled the picker (reserved for --ask v2)
//
// The picker (--ask) and --replay are scaffolded as stubs in v1: the
// flag is accepted and the value flows to the Plan, but the picker
// UI lands in a follow-up patch (tracked in .lovable/plan.md).

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/cliexit"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/clonepick"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// runClonePick is the dispatcher entry registered in rootcore.go.
func runClonePick(args []string) {
	checkHelp("clone-pick", args)

	parsed := parseClonePickFlags(args)
	setCmdFaithfulVerify(parsed.VerifyCmdFaithful)
	setCmdFaithfulExitOnMismatch(parsed.VerifyCmdFaithfulExitOnMismatch)
	setCmdPrintArgv(parsed.PrintCloneArgv)
	plan, err := clonepick.ParseArgs(parsed.RawURL, parsed.RawPaths, parsed.Flags)
	if err != nil {
		cliexit.Fail(constants.CmdClonePick, "parse-args", parsed.RawURL, err, 2)
	}

	if plan.DryRun {
		// `--output terminal`: emit the standardized block instead
		// of the legacy clonepick.Render output. Keeps the per-repo
		// summary shape consistent across every clone command.
		if parsed.Output == constants.OutputTerminal {
			printClonePickTermBlock(plan)
			maybeExitOnCmdFaithfulMismatch()

			return
		}
		if err := clonepick.Render(os.Stdout, plan); err != nil {
			cliexit.Fail(constants.CmdClonePick, "render-dry-run", parsed.RawURL, err, 1)
		}
		maybeExitOnCmdFaithfulMismatch()

		return
	}

	if parsed.Output == constants.OutputTerminal {
		printClonePickTermBlock(plan)
	}
	runClonePickExecute(plan, parsed.NoVSCodeSync)
}

// clonePickParsed bundles every output of parseClonePickFlags so a
// new audit/debug toggle can be added without churning the call
// site signature each time. Fields are exported because the struct
// itself stays unexported (cmd-package-internal).
type clonePickParsed struct {
	RawURL                          string
	RawPaths                        string
	Flags                           clonepick.Flags
	Output                          string
	VerifyCmdFaithful               bool
	VerifyCmdFaithfulExitOnMismatch bool
	PrintCloneArgv                  bool
	// NoVSCodeSync suppresses the post-clone update of the
	// alefragnani.project-manager projects.json file. Mirrors
	// `gitmap scan --no-vscode-sync`. Default false. See
	// spec/01-vscode-project-manager-sync/02-clone-sync.md.
	NoVSCodeSync bool
}

// parseClonePickFlags binds every clone-pick flag and extracts the
// two positional args. Validation that needs cross-flag knowledge
// happens in clonepick.ParseArgs so this stays focused on flag
// binding.
func parseClonePickFlags(args []string) clonePickParsed {
	defaults := clonepick.DefaultFlags()
	flags := defaults
	fs := flag.NewFlagSet("clone-pick", flag.ExitOnError)
	fs.BoolVar(&flags.Ask, constants.FlagClonePickAsk, defaults.Ask,
		constants.FlagDescClonePickAsk)
	fs.StringVar(&flags.Name, constants.FlagClonePickName, defaults.Name,
		constants.FlagDescClonePickName)
	fs.StringVar(&flags.Mode, constants.FlagClonePickMode, defaults.Mode,
		constants.FlagDescClonePickMode)
	fs.StringVar(&flags.Branch, constants.FlagClonePickBranch, defaults.Branch,
		constants.FlagDescClonePickBranch)
	fs.IntVar(&flags.Depth, constants.FlagClonePickDepth, defaults.Depth,
		constants.FlagDescClonePickDepth)
	fs.BoolVar(&flags.Cone, constants.FlagClonePickCone, defaults.Cone,
		constants.FlagDescClonePickCone)
	fs.StringVar(&flags.Dest, constants.FlagClonePickDest, defaults.Dest,
		constants.FlagDescClonePickDest)
	fs.BoolVar(&flags.KeepGit, constants.FlagClonePickKeepGit, defaults.KeepGit,
		constants.FlagDescClonePickKeepGit)
	fs.BoolVar(&flags.DryRun, constants.FlagClonePickDryRun, defaults.DryRun,
		constants.FlagDescClonePickDryRun)
	fs.BoolVar(&flags.Quiet, constants.FlagClonePickQuiet, defaults.Quiet,
		constants.FlagDescClonePickQuiet)
	fs.BoolVar(&flags.Force, constants.FlagClonePickForce, defaults.Force,
		constants.FlagDescClonePickForce)
	output := fs.String(constants.FlagCloneTermOutput, "",
		constants.FlagDescCloneTermOutput)
	verify := fs.Bool(constants.FlagCloneVerifyCmdFaithful, false,
		constants.FlagDescCloneVerifyCmdFaithful)
	verifyExit := fs.Bool(constants.FlagCloneVerifyCmdFaithfulExitOnMismatch,
		false, constants.FlagDescCloneVerifyCmdFaithfulExitOnMismatch)
	printArgv := fs.Bool(constants.FlagClonePrintArgv, false,
		constants.FlagDescClonePrintArgv)
	noVSCodeSync := fs.Bool(constants.FlagNoVSCodeSync, false,
		constants.FlagDescNoVSCodeSync)

	reordered := reorderFlagsBeforeArgs(args)
	fs.Parse(reordered)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, constants.MsgClonePickMissingURL)
		os.Exit(2)
	}
	rawPaths := ""
	if fs.NArg() >= 2 {
		rawPaths = fs.Arg(1)
	}

	return clonePickParsed{
		RawURL:                          fs.Arg(0),
		RawPaths:                        rawPaths,
		Flags:                           flags,
		Output:                          *output,
		VerifyCmdFaithful:               *verify,
		VerifyCmdFaithfulExitOnMismatch: *verifyExit,
		PrintCloneArgv:                  *printArgv,
		NoVSCodeSync:                    *noVSCodeSync,
	}
}

// runClonePickExecute opens the DB (best-effort), runs the
// sparse-checkout, and translates the Result to an exit code.
func runClonePickExecute(plan clonepick.Plan, noVSCodeSync bool) {
	progress := io.Writer(os.Stderr)
	if plan.Quiet {
		progress = io.Discard
	}

	db, dbErr := openDB()
	if dbErr != nil {
		// DB open failure is non-fatal -- clone still proceeds, just
		// without persistence. Per the zero-swallow policy we surface
		// the error to stderr so it isn't silently dropped.
		fmt.Fprintln(os.Stderr, dbErr)
	}

	result := clonepick.Execute(plan, db, progress)
	if result.SelectionId > 0 {
		name := plan.Name
		if len(name) == 0 {
			name = "(unnamed)"
		}
		fmt.Fprintf(os.Stderr, constants.MsgClonePickSaved,
			result.SelectionId, plan.RepoCanonicalId, name)
	}

	if result.Status == clonepick.StatusFailed {
		maybeExitOnCmdFaithfulMismatch()
		os.Exit(1)
	}

	// VS Code Project Manager sync. Result.Detail carries the
	// resolved destination path on success; fall back to the plan's
	// DestDir when empty.
	syncClonePickResultToVSCodePM(plan, result, noVSCodeSync)

	if plan.DestDir != "." && plan.DestDir != "" {
		WriteShellHandoff(result.Detail)
	}
	maybeExitOnCmdFaithfulMismatch()
}

// syncClonePickResultToVSCodePM registers a successful sparse-checkout
// dest in projects.json. The pick name (when set) wins over the folder
// basename so users see their alias in the Project Manager sidebar.
func syncClonePickResultToVSCodePM(plan clonepick.Plan, result clonepick.Result, skip bool) {
	dest := result.Detail
	if dest == "" {
		dest = plan.DestDir
	}
	abs, err := filepath.Abs(dest)
	if err != nil {
		abs = dest
	}
	name := plan.Name
	if name == "" {
		name = filepath.Base(abs)
	}
	syncSingleClonedRepoToVSCodePM(abs, name, skip)
}
