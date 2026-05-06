package cmd

import (
	"os"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/gitutil"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/helptext"
)

// checkHelp prints embedded help and exits if --help or -h is present.
// Honors --pretty / --no-pretty so users can force-enable rendering for
// pagers (`gitmap foo --help --pretty | less -R`) or strip ANSI for
// scripting (`gitmap foo --help --no-pretty > help.txt`).
func checkHelp(command string, args []string) {
	if !hasHelpFlag(args) {
		return
	}
	_, mode := ParsePrettyFlag(args)
	helptext.PrintWithMode(command, mode)
	os.Exit(0)
}

// hasHelpFlag scans args for the standard help triggers.
func hasHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}

	return false
}

// requireOnline checks network connectivity and exits if offline.
func requireOnline() {
	if gitutil.IsOnline() {
		return
	}

	gitutil.PrintOfflineWarning()
	os.Exit(1)
}
