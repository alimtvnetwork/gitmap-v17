// Package cmd implements the CLI commands for gitmap.
package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// runDoctor handles the 'doctor' command.
func runDoctor() {
	fixPath := parseDoctorFlags(os.Args[2:])

	if fixPath {
		runFixPath()

		return
	}

	fmt.Printf(constants.DoctorBannerFmt, constants.Version)
	fmt.Println(constants.DoctorBannerRule)
	issues := runDoctorChecks()
	printDoctorSummary(issues)
}

// runDoctorChecks executes all diagnostic checks and returns total issues.
func runDoctorChecks() int {
	issues := 0
	issues += checkRepoPath()
	issues += checkActiveBinary()
	issues += checkDuplicateBinaries()
	issues += checkDeployedBinary()
	issues += checkVersionMismatch()
	issues += checkGit()
	issues += checkGo()
	issues += checkChangelogFile()
	issues += checkConfigFile()
	issues += checkSetupConfig()
	issues += checkShellWrapper()
	issues += checkDatabase()
	issues += checkReleaseRepoIntegrity()
	issues += checkLockFile()
	issues += checkNetwork()
	issues += checkLegacyDirs()
	issues += checkSignature()
	issues += checkVSCodeProjectManager()

	return issues
}

// printDoctorSummary prints final summary based on issue count.
func printDoctorSummary(issues int) {
	fmt.Println()
	if issues > 0 {
		fmt.Printf(constants.DoctorIssuesFmt, issues)
		fmt.Printf(constants.DoctorFixPathTip)

		return
	}

	fmt.Println(constants.DoctorAllPassed)
}

// parseDoctorFlags parses flags for the doctor command.
func parseDoctorFlags(args []string) (fixPath bool) {
	fs := flag.NewFlagSet(constants.CmdDoctor, flag.ExitOnError)
	fixPathFlag := fs.Bool("fix-path", false, constants.DoctorFixFlagDesc)
	fs.Parse(args)

	return *fixPathFlag
}
