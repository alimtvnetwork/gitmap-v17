package cmd

import (
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// dispatchCore routes scan, clone, pull, and status commands.
func dispatchCore(command string) bool {
	return runDispatchTable(command, coreDispatchEntries())
}

// coreDispatchEntries returns the routing table for core commands.
func coreDispatchEntries() []dispatchEntry {
	return []dispatchEntry{
		{[]string{constants.CmdScan, constants.CmdScanAlias}, func() { runScan(argsTail()) }},
		{[]string{constants.CmdClone, constants.CmdCloneAlias}, func() { runClone(argsTail()) }},
		{[]string{constants.CmdPull, constants.CmdPullAlias}, func() { runPull(argsTail()) }},
		{[]string{constants.CmdStatus, constants.CmdStatusAlias}, func() { runStatus(argsTail()) }},
		{[]string{constants.CmdExec, constants.CmdExecAlias}, func() { runExec(argsTail()) }},
		{
			[]string{
				constants.CmdHasAnyUpdates, constants.CmdHasAnyUpdatesAlias,
				constants.CmdHasAnyChanges, constants.CmdHasAnyChangesAlias,
			},
			func() { runHasAnyUpdates(argsTail()) },
		},
		{[]string{constants.CmdHasChange, constants.CmdHasChangeAlias}, func() { runHasChange(argsTail()) }},
		{[]string{constants.CmdCloneNext, constants.CmdCloneNextAlias}, func() { runCloneNext(argsTail()) }},
		{[]string{constants.CmdAs, constants.CmdAsAlias}, func() { runAs(argsTail()) }},
		{[]string{constants.CmdCode}, func() { runCode(argsTail()) }},
		{[]string{constants.CmdInject, constants.CmdInjectAlias}, func() { runInject(argsTail()) }},
		{[]string{constants.CmdCloneFrom, constants.CmdCloneFromAlias}, func() { runCloneFrom(argsTail()) }},
		{
			[]string{
				constants.CmdCloneReclone, constants.CmdCloneRecloneAlias,
				constants.CmdCloneNow, constants.CmdCloneNowAlias,
				constants.CmdCloneRel, constants.CmdCloneRelAlias,
			},
			func() { runCloneNow(argsTail()) },
		},
		{[]string{constants.CmdClonePick, constants.CmdClonePickAlias}, func() { runClonePick(argsTail()) }},
		{[]string{constants.CmdCommitIn, constants.CmdCommitInAlias}, func() { runCommitIn(argsTail()) }},
		{[]string{constants.CmdMakePublic}, func() { runMakePublic(argsTail()) }},
		{[]string{constants.CmdMakePrivate}, func() { runMakePrivate(argsTail()) }},
		{
			[]string{constants.CmdCloneFixRepo, constants.CmdCloneFixRepoAlias},
			func() { runCloneFixRepo(argsTail()) },
		},
		{
			[]string{constants.CmdCloneFixRepoPub, constants.CmdCloneFixRepoPubAlias},
			func() { runCloneFixRepoPub(argsTail()) },
		},
	}
}
