package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/store"
)

// runSSHCat displays the public key for a named SSH key.
func runSSHCat(args []string) {
	fs := flag.NewFlagSet("ssh-cat", flag.ExitOnError)
	nameFlag := fs.String("name", constants.DefaultSSHKeyName, "Key name")
	fs.StringVar(nameFlag, "n", constants.DefaultSSHKeyName, "Key name (short)")
	fs.Parse(args)

	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSSHQuery, err)
		os.Exit(1)
	}
	defer db.Close()

	key, err := db.FindSSHKeyByName(*nameFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrSSHNotFound, *nameFlag)
		printAvailableKeys(db)
		os.Exit(1)
	}

	fmt.Println(strings.TrimSpace(key.PublicKey))
}

// printAvailableKeys prints available SSH key names to stderr.
func printAvailableKeys(db *store.DB) {
	names, err := db.SSHKeyNames()
	if err != nil || len(names) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, constants.ErrSSHAvailable, strings.Join(names, ", "))
}
