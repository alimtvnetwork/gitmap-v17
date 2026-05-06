package cmd

import (
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// dispatchTooling routes dev tooling and maintenance commands.
func dispatchTooling(command string) bool {
	return runDispatchTable(command, toolingDispatchEntries())
}

// toolingDispatchEntries returns the routing table for tooling commands.
func toolingDispatchEntries() []dispatchEntry {
	return []dispatchEntry{
		{
			[]string{constants.CmdDesktopSync, constants.CmdDesktopSyncAlias},
			func() { checkHelp("desktop-sync", argsTail()); runDesktopSync() },
		},
		{[]string{constants.CmdGitHubDesktop, constants.CmdGitHubDesktopAlias}, func() { runGitHubDesktop(argsTail()) }},
		{
			[]string{constants.CmdRescan, constants.CmdRescanAlias},
			func() { checkHelp("rescan", argsTail()); runRescan() },
		},
		{
			[]string{constants.CmdRescanSubtree, constants.CmdRescanSubtreeAlias},
			func() { runRescanSubtree(argsTail()) },
		},
		{[]string{constants.CmdSetup}, func() { runSetup(argsTail()) }},
		{[]string{constants.CmdDoctor}, func() { checkHelp("doctor", argsTail()); runDoctor() }},
		{[]string{constants.CmdLatestBranch, constants.CmdLatestBranchAlias}, func() { runLatestBranch(argsTail()) }},
		{[]string{constants.CmdBranch, constants.CmdBranchAlias}, func() { runBranch(argsTail()) }},
		{[]string{constants.CmdListVersions, constants.CmdListVersionsAlias}, func() { runListVersions(argsTail()) }},
		{
			[]string{constants.CmdListReleases, constants.CmdListReleasesAlias, constants.CmdReleases},
			func() { runListReleases(argsTail()) },
		},
		{[]string{constants.CmdSEOWrite, constants.CmdSEOWriteAlias}, func() { runSEOWrite(argsTail()) }},
		{[]string{constants.CmdGoMod, constants.CmdGoModAlias}, func() { runGoMod(argsTail()) }},
		{[]string{constants.CmdCompletion, constants.CmdCompletionAlias}, func() { runCompletion(argsTail()) }},
		{[]string{constants.CmdZipGroup, constants.CmdZipGroupShort}, func() { runZipGroup(argsTail()) }},
		{[]string{constants.CmdAlias, constants.CmdAliasShort}, func() { runAlias(argsTail()) }},
		{[]string{constants.CmdSSH}, func() { runSSH(argsTail()) }},
		{[]string{constants.CmdPrune, constants.CmdPruneAlias}, func() { runPrune(argsTail()) }},
		{[]string{constants.CmdTempRelease, constants.CmdTempReleaseShort}, func() { runTempRelease(argsTail()) }},
		{[]string{constants.CmdTask, constants.CmdTaskAlias}, func() { runTask(argsTail()) }},
		{[]string{constants.CmdEnv, constants.CmdEnvAlias}, func() { runEnv(argsTail()) }},
		{[]string{constants.CmdInstall, constants.CmdInstallAlias}, func() { runInstall(argsTail()) }},
		{[]string{constants.CmdUninstall, constants.CmdUninstallAlias}, func() { runUninstall(argsTail()) }},
		{[]string{constants.CmdStartupAdd, constants.CmdStartupAddAlias}, func() { runStartupAdd(argsTail()) }},
		{[]string{constants.CmdStartupList, constants.CmdStartupListAlias}, func() { runStartupList(argsTail()) }},
		{[]string{constants.CmdStartupRemove, constants.CmdStartupRemoveAlias}, func() { runStartupRemove(argsTail()) }},
		{[]string{constants.CmdSelfInstall}, func() { runSelfInstall(argsTail()) }},
		{[]string{constants.CmdSelfUninstall}, func() { runSelfUninstall(argsTail()) }},
		{[]string{constants.CmdSelfUninstallRunner}, func() { runSelfUninstallRunner() }},
		{[]string{constants.CmdPending}, func() { runPending() }},
		{[]string{constants.CmdDoPending, constants.CmdDoPendingAlias}, func() { runDoPending(argsTail()) }},
		{
			[]string{constants.CmdDownloaderConfig, constants.CmdDownloaderConfigAlias},
			func() { runDownloaderConfig(argsTail()) },
		},
		{
			[]string{constants.CmdUnzipCompact, constants.CmdUnzipCompactAlias},
			func() { runUnzipCompact(argsTail()) },
		},
		{[]string{constants.CmdZip}, func() { runZip(argsTail()) }},
		{[]string{constants.CmdReplace, constants.CmdReplaceAlias}, func() { runReplace(argsTail()) }},
		{[]string{constants.CmdRegoldens, constants.CmdRegoldensAlias}, func() { runRegoldens(argsTail()) }},
		{
			[]string{constants.CmdAuditLegacy, constants.CmdAuditLegacyAlias, constants.CmdAuditLegacyAlias2},
			func() { runAuditLegacy(argsTail()) },
		},
		{[]string{constants.CmdFixRepo, constants.CmdFixRepoAlias}, func() { runFixRepo(argsTail()) }},
		{
			[]string{constants.CmdHistoryPurge, constants.CmdHistoryPurgeAlias},
			func() { runHistoryPurge(argsTail()) },
		},
		{
			[]string{constants.CmdHistoryPin, constants.CmdHistoryPinAlias},
			func() { runHistoryPin(argsTail()) },
		},
	}
}
