package cmd

import (
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// runSSH handles the "ssh" subcommand and routes to sub-handlers.
func runSSH(args []string) {
	checkHelp("ssh", args)
	if len(args) == 0 {
		runSSHGenerate(args)

		return
	}
	dispatchSSH(args[0], args[1:])
}

// dispatchSSH routes SSH subcommands to their handlers.
func dispatchSSH(sub string, args []string) {
	if sub == constants.SubCmdSSHCat {
		runSSHCat(args)

		return
	}
	if sub == constants.SubCmdSSHList || sub == constants.SubCmdSSHListS {
		runSSHList(args...)

		return
	}
	if sub == constants.SubCmdSSHDelete || sub == constants.SubCmdSSHDeleteS {
		runSSHDelete(args)

		return
	}
	if sub == constants.SubCmdSSHConfig {
		runSSHConfig()

		return
	}

	// Not a subcommand — treat all args as flags for generate.
	runSSHGenerate(append([]string{sub}, args...))
}
