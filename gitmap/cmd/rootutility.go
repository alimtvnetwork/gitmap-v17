package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/helptext"
)

// isFlagToken returns true when arg looks like a CLI flag (-x or --xx).
func isFlagToken(arg string) bool {
	return strings.HasPrefix(arg, "-")
}

// dispatchUtility routes setup, update, doctor, and other utility commands.
func dispatchUtility(command string) bool {
	return runDispatchTable(command, utilityDispatchEntries())
}

// utilityDispatchEntries returns the routing table for utility commands.
func utilityDispatchEntries() []dispatchEntry {
	return []dispatchEntry{
		{[]string{constants.CmdUpdate}, func() { checkHelp("update", argsTail()); runUpdate() }},
		{[]string{constants.CmdUpdateRunner}, func() { runUpdateRunner() }},
		{[]string{constants.CmdUpdateCleanup}, func() { runUpdateCleanup() }},
		{
			[]string{constants.CmdInstalledDir, constants.CmdInstalledDirAlias},
			func() { checkHelp("installed-dir", argsTail()); runInstalledDir() },
		},
		{[]string{constants.CmdRevert}, func() { runRevert(argsTail()) }},
		{[]string{constants.CmdRevertRunner}, func() { runRevertRunner() }},
		{
			[]string{constants.CmdVersion, constants.CmdVersionAlias},
			func() { checkHelp("version", argsTail()); fmt.Printf(constants.MsgVersionFmt, constants.Version) },
		},
		{[]string{constants.CmdHelp}, runHelpDispatch},
		{[]string{constants.CmdDocs, constants.CmdDocsAlias}, func() { runDocs(argsTail()) }},
		{[]string{constants.CmdHelpDashboard, constants.CmdHelpDashboardAlias}, func() { runHelpDashboard(argsTail()) }},
		{[]string{constants.CmdLLMDocs, constants.CmdLLMDocsAlias}, func() { runLLMDocs(argsTail()) }},
		{[]string{constants.CmdSetSourceRepo}, func() { runSetSourceRepo() }},
		{[]string{constants.CmdSf}, func() { runSf(argsTail()) }},
		{[]string{constants.CmdProbe}, func() { runProbe(argsTail()) }},
		{[]string{constants.CmdFindNext, constants.CmdFindNextAlias}, func() { runFindNext(argsTail()) }},
		{[]string{constants.CmdVSCodePMPath, constants.CmdVSCodePMPathAlias}, func() { runVSCodePMPath(argsTail()) }},
		{[]string{constants.CmdVSCodeWorkspace, constants.CmdVSCodeWorkspaceAlias}, func() { runVSCodeWorkspace(argsTail()) }},
		{[]string{constants.CmdLFSCommon, constants.CmdLFSCommonAlias}, func() { runLFSCommon(argsTail()) }},
		{[]string{constants.CmdReinstall}, func() { runReinstall(argsTail()) }},
	}
}

// runHelpDispatch handles the `help` subcommand including topic
// help, --groups, --compact, and the default usage screen.
func runHelpDispatch() {
	if len(os.Args) >= 3 && !isFlagToken(os.Args[2]) {
		_, mode := ParsePrettyFlag(os.Args[3:])
		helptext.PrintWithMode(os.Args[2], mode)

		return
	}
	if hasFlag(constants.FlagGroups) {
		printHelpGroups()

		return
	}
	if hasFlag(constants.FlagCompact) {
		printUsageCompact()

		return
	}
	printUsage()
}
