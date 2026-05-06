package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// printUsage displays grouped help text for all commands and flags.
func printUsage() {
	fmt.Printf(constants.UsageHeaderFmt, constants.Version)
	fmt.Println(constants.HelpUsage)
	fmt.Println()
	printUsageQuickStart()
	fmt.Println()
	printGroupScanning()
	printGroupCloning()
	printGroupGitOps()
	printGroupNavigation()
	printGroupRelease()
	printGroupReleaseInfo()
	printGroupData()
	printGroupHistory()
	printGroupAmend()
	printGroupProject()
	printGroupSSH()
	printGroupZip()
	printGroupEnvTools()
	printGroupTasks()
	printGroupVisualize()
	printGroupCommitXfer()
	printGroupUtilities()
	fmt.Println()
	printUsageFlagSections()
}

// printUsageQuickStart prints examples and the help hint.
func printUsageQuickStart() {
	fmt.Println(constants.HelpGroupExample)
	fmt.Println(constants.HelpExampleScan)
	fmt.Println(constants.HelpExampleList)
	fmt.Println(constants.HelpExamplePull)
	fmt.Println(constants.HelpExampleCD)
	fmt.Println()
	fmt.Println(constants.HelpGroupHint)
	fmt.Println(constants.HelpCompactHint)
}

// printGroupScanning prints the scanning commands.
func printGroupScanning() {
	fmt.Println()
	fmt.Println(constants.HelpGroupScanning)
	fmt.Println(constants.HelpScan)
	fmt.Println(constants.HelpRescan)
	fmt.Println(constants.HelpList)
}

// printGroupCloning prints the cloning commands.
func printGroupCloning() {
	fmt.Println()
	fmt.Println(constants.HelpGroupCloning)
	fmt.Println(constants.HelpClone)
	fmt.Println(constants.HelpCloneNext)
	fmt.Println(constants.HelpDesktopSync)
	fmt.Println(constants.HelpGitHubDesktop)
}

// printGroupGitOps prints the git operations commands.
func printGroupGitOps() {
	fmt.Println()
	fmt.Println(constants.HelpGroupGitOps)
	fmt.Println(constants.HelpPull)
	fmt.Println(constants.HelpExec)
	fmt.Println(constants.HelpStatus)
	fmt.Println(constants.HelpWatch)
	fmt.Println(constants.HelpHasAnyUpdates)
	fmt.Println(constants.HelpLatestBr)
	fmt.Println(constants.MsgHelpLFSCommon)
}

// printGroupNavigation prints the navigation commands.
func printGroupNavigation() {
	fmt.Println()
	fmt.Println(constants.HelpGroupNavigation)
	fmt.Println(constants.HelpCD)
	fmt.Println(constants.HelpGroup)
	fmt.Println(constants.HelpMultiGroup)
	fmt.Println(constants.HelpSf)
	fmt.Println(constants.HelpAlias)
	fmt.Println(constants.HelpDiffProfiles)
}

// printGroupRelease prints the release workflow commands.
func printGroupRelease() {
	fmt.Println()
	fmt.Println(constants.HelpGroupRelease)
	fmt.Println(constants.HelpRelease)
	fmt.Println(constants.HelpReleasePull)
	fmt.Println(constants.HelpReleaseSelf)
	fmt.Println(constants.HelpReleaseBr)
	fmt.Println(constants.HelpTempRelease)
}

// printGroupReleaseInfo prints the release info commands.
func printGroupReleaseInfo() {
	fmt.Println()
	fmt.Println(constants.HelpGroupReleaseInfo)
	fmt.Println(constants.HelpChangelog)
	fmt.Println(constants.HelpChangelogGen)
	fmt.Println(constants.HelpListVersions)
	fmt.Println(constants.HelpListReleases)
	fmt.Println(constants.HelpReleasePend)
	fmt.Println(constants.HelpRevert)
	fmt.Println(constants.HelpClearReleaseJSON)
	fmt.Println(constants.HelpPrune)
}

// printGroupData prints the data/profile/bookmark commands.
func printGroupData() {
	fmt.Println()
	fmt.Println(constants.HelpGroupData)
	fmt.Println(constants.HelpExport)
	fmt.Println(constants.HelpImport)
	fmt.Println(constants.HelpProfile)
	fmt.Println(constants.HelpBookmark)
	fmt.Println(constants.HelpDBReset)
}

// printGroupHistory prints the history and stats commands.
func printGroupHistory() {
	fmt.Println()
	fmt.Println(constants.HelpGroupHistory)
	fmt.Println(constants.HelpHistory)
	fmt.Println(constants.HelpHistoryReset)
	fmt.Println(constants.HelpVersionHistory)
	fmt.Println(constants.HelpStats)
}

// printGroupAmend prints the amend commands.
func printGroupAmend() {
	fmt.Println()
	fmt.Println(constants.HelpGroupAmendGroup)
	fmt.Println(constants.HelpAmend)
	fmt.Println(constants.HelpAmendList)
}

// printGroupProject prints the project detection commands.
func printGroupProject() {
	fmt.Println()
	fmt.Println(constants.HelpGroupProject)
	fmt.Println(constants.HelpGoRepos)
	fmt.Println(constants.HelpNodeRepos)
	fmt.Println(constants.HelpReactRepos)
	fmt.Println(constants.HelpCppRepos)
	fmt.Println(constants.HelpCsharpRepos)
}

// printGroupSSH prints the SSH key management commands.
func printGroupSSH() {
	fmt.Println()
	fmt.Println(constants.HelpGroupSSH)
	fmt.Println(constants.HelpSSH)
}

// printGroupZip prints the zip group commands.
func printGroupZip() {
	fmt.Println()
	fmt.Println(constants.HelpGroupZip)
	fmt.Println(constants.HelpZipGroup)
}

// printGroupEnvTools prints the env and install commands.
func printGroupEnvTools() {
	fmt.Println()
	fmt.Println(constants.HelpGroupEnvTools)
	fmt.Println(constants.HelpEnv)
	fmt.Println(constants.HelpInstall)
	fmt.Println(constants.HelpUninstall)
}

// printGroupTasks prints the task commands.
func printGroupTasks() {
	fmt.Println()
	fmt.Println(constants.HelpGroupTasks)
	fmt.Println(constants.HelpTask)
	fmt.Println(constants.HelpPending)
	fmt.Println(constants.HelpDoPending)
}

// printGroupVisualize prints the visualization commands.
func printGroupVisualize() {
	fmt.Println()
	fmt.Println(constants.HelpGroupVisualize)
	fmt.Println(constants.HelpDashboard)
}

// printGroupCommitXfer prints the commit-transfer family. Surfacing
// these in `gitmap help` is the primary discovery path — without it
// users have to read spec/01-app/106 or stumble into `help commit-right`
// to learn the aliases (cml / cmr / cmb) even exist.
func printGroupCommitXfer() {
	fmt.Println()
	fmt.Println(constants.HelpGroupCommitXfer)
	fmt.Println(constants.HelpCommitRight)
	fmt.Println(constants.HelpCommitLeft)
	fmt.Println(constants.HelpCommitBoth)
}
