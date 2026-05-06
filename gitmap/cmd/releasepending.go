// Package cmd implements the CLI commands for gitmap.
package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/release"
)

// runReleasePending handles the 'release-pending' command.
func runReleasePending(args []string) {
	checkHelp("release-pending", args)
	assets, notes, draft, dryRun, verbose, noCommit, yes := parseReleasePendingFlags(args)
	_ = verbose

	err := release.ExecutePending(assets, notes, draft, dryRun, noCommit, yes)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrBareFmt, err)
		os.Exit(1)
	}
}

// parseReleasePendingFlags parses flags for the release-pending command.
func parseReleasePendingFlags(args []string) (assets, notes string, draft, dryRun, verbose, noCommit, yes bool) {
	fs := flag.NewFlagSet(constants.CmdReleasePending, flag.ExitOnError)
	assetsFlag := fs.String("assets", "", constants.FlagDescAssets)
	notesFlag := fs.String("notes", "", constants.FlagDescNotes)
	draftFlag := fs.Bool("draft", false, constants.FlagDescDraft)
	dryRunFlag := fs.Bool("dry-run", false, constants.FlagDescDryRun)
	verboseFlag := fs.Bool("verbose", false, constants.FlagDescVerbose)
	noCommitFlag := fs.Bool("no-commit", false, constants.FlagDescNoCommit)
	yesFlag := fs.Bool("yes", false, constants.FlagDescYes)

	// Register -N as shorthand for --notes, -y as shorthand for --yes.
	fs.StringVar(notesFlag, "N", "", constants.FlagDescNotes)
	fs.BoolVar(yesFlag, "y", false, constants.FlagDescYes)

	fs.Parse(args)

	return *assetsFlag, *notesFlag, *draftFlag, *dryRunFlag, *verboseFlag, *noCommitFlag, *yesFlag
}
