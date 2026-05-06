package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
)

// runExport handles the "export" subcommand.
func runExport(args []string) {
	checkHelp("export", args)
	outFile := resolveExportFile(args)
	export := loadExportData()

	writeExportFile(outFile, export)
	printExportSummary(outFile, export)
}

// resolveExportFile determines the output file from args or default.
func resolveExportFile(args []string) string {
	if len(args) > 0 && args[0] != "" && args[0][0] != '-' {
		return args[0]
	}

	return constants.DefaultExportFile
}

// loadExportData fetches the full database export.
func loadExportData() model.DatabaseExport {
	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgExportFailed, err)
		os.Exit(1)
	}
	defer db.Close()

	export, err := db.ExportAll()
	if err != nil {
		if isLegacyDataError(err) {
			fmt.Fprint(os.Stderr, constants.MsgLegacyProjectData)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, constants.MsgExportFailed, err)
		os.Exit(1)
	}

	return export
}

// writeExportFile marshals the export data to a JSON file.
func writeExportFile(path string, export model.DatabaseExport) {
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgExportFailed, err)
		os.Exit(1)
	}

	err = os.WriteFile(path, data, constants.DirPermission)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.MsgExportFailed, err)
		os.Exit(1)
	}
}

// printExportSummary prints the export result summary.
func printExportSummary(path string, e model.DatabaseExport) {
	fmt.Printf(constants.MsgExportDone, path,
		len(e.Repos), len(e.Groups), len(e.Releases),
		len(e.History), len(e.Bookmarks))
}
