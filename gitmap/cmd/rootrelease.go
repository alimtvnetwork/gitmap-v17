package cmd

import (
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// dispatchRelease routes release-related commands.
func dispatchRelease(command string) bool {
	return runDispatchTable(command, releaseDispatchEntries())
}

// releaseDispatchEntries returns the routing table for release commands.
func releaseDispatchEntries() []dispatchEntry {
	return []dispatchEntry{
		{[]string{constants.CmdRelease, constants.CmdReleaseShort}, func() { runRelease(argsTail()) }},
		{[]string{constants.CmdReleasePull, constants.CmdReleasePullAlias}, func() { runReleasePull(argsTail()) }},
		{
			[]string{constants.CmdReleaseSelf, constants.CmdReleaseSelfAlias, constants.CmdReleaseSelfAlias2},
			func() { runReleaseSelf(argsTail()) },
		},
		{[]string{constants.CmdReleaseBranch, constants.CmdReleaseBranchAlias}, func() { runReleaseBranch(argsTail()) }},
		{[]string{constants.CmdReleasePending, constants.CmdReleasePendingAlias}, func() { runReleasePending(argsTail()) }},
		{[]string{constants.CmdChangelog, constants.CmdChangelogAlias}, func() { runChangelog(argsTail()) }},
		{[]string{constants.CmdChangelogMD}, func() { runChangelog([]string{constants.FlagOpenValue}) }},
		{[]string{constants.CmdClearReleaseJSON, constants.CmdClearReleaseJSONAlias}, func() { runClearReleaseJSON(argsTail()) }},
		{[]string{constants.CmdChangelogGen, constants.CmdChangelogGenAlias}, func() { runChangelogGen(argsTail()) }},
		{[]string{constants.CmdReleaseAlias, constants.CmdReleaseAliasShort}, func() { runReleaseAlias(argsTail(), false) }},
		{[]string{constants.CmdReleaseAliasPull, constants.CmdReleaseAliasPullShort}, func() { runReleaseAlias(argsTail(), true) }},
	}
}
