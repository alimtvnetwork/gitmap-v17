package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
)

// runHistory handles the "history" subcommand.
func runHistory(args []string) {
	checkHelp("history", args)
	detail, cmdFilter, limit, jsonOut := parseHistoryFlags(args)
	records := loadHistory(cmdFilter)
	records = applyHistoryLimit(records, limit)

	if jsonOut {
		printHistoryJSON(records)

		return
	}

	printHistoryTerminal(records, detail)
}

// parseHistoryFlags parses --detail, --command, --limit, --json flags.
func parseHistoryFlags(args []string) (string, string, int, bool) {
	fs := flag.NewFlagSet(constants.CmdHistory, flag.ExitOnError)
	detail := fs.String("detail", constants.DetailStandard, constants.FlagDescDetail)
	command := fs.String("command", "", constants.FlagDescCommand)
	limit := fs.Int("limit", 0, constants.FlagDescLimit)
	jsonFlag := fs.Bool("json", false, constants.FlagDescLBJSON)
	fs.Parse(args)

	return *detail, *command, *limit, *jsonFlag
}

// loadHistory fetches history from the database.
func loadHistory(cmdFilter string) []model.CommandHistoryRecord {
	db, err := openDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrHistoryQuery+"\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if cmdFilter != "" {
		records, err := db.ListHistoryByCommand(cmdFilter)
		if err != nil {
			if isLegacyDataError(err) {
				fmt.Fprint(os.Stderr, constants.MsgLegacyProjectData)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, constants.ErrHistoryQuery+"\n", err)
			os.Exit(1)
		}

		return records
	}

	records, err := db.ListHistory()
	if err != nil {
		if isLegacyDataError(err) {
			fmt.Fprint(os.Stderr, constants.MsgLegacyProjectData)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, constants.ErrHistoryQuery+"\n", err)
		os.Exit(1)
	}

	return records
}

// applyHistoryLimit truncates results to the given limit.
func applyHistoryLimit(records []model.CommandHistoryRecord, limit int) []model.CommandHistoryRecord {
	if limit > 0 && limit < len(records) {
		return records[:limit]
	}

	return records
}

// printHistoryTerminal prints history in table format based on detail level.
func printHistoryTerminal(records []model.CommandHistoryRecord, detail string) {
	if len(records) == 0 {
		fmt.Print(constants.MsgHistoryEmpty)

		return
	}

	printHistoryHeader(detail)
	for _, r := range records {
		printHistoryRow(r, detail)
	}
}

// printHistoryHeader prints the column header for the chosen detail level.
func printHistoryHeader(detail string) {
	if detail == constants.DetailBasic {
		fmt.Println(constants.MsgHistoryColumnsBasic)

		return
	}
	if detail == constants.DetailDetailed {
		fmt.Println(constants.MsgHistoryColumnsDetailed)

		return
	}

	fmt.Println(constants.MsgHistoryColumnsStandard)
}

// printHistoryRow prints a single history row at the chosen detail level.
func printHistoryRow(r model.CommandHistoryRecord, detail string) {
	status := resolveStatus(r.ExitCode)

	if detail == constants.DetailBasic {
		fmt.Printf(constants.MsgHistoryRowBasicFmt, r.Command, r.StartedAt, status)

		return
	}
	if detail == constants.DetailDetailed {
		dur := strconv.FormatInt(r.DurationMs, 10) + "ms"
		repos := strconv.Itoa(r.RepoCount)
		fmt.Printf(constants.MsgHistoryRowDetailFmt, r.Command, r.StartedAt, r.Args, r.Flags, status, dur, repos, r.Summary)

		return
	}

	dur := strconv.FormatInt(r.DurationMs, 10) + "ms"
	fmt.Printf(constants.MsgHistoryRowStdFmt, r.Command, r.StartedAt, r.Flags, status, dur)
}

// resolveStatus returns OK or FAIL based on exit code.
func resolveStatus(code int) string {
	if code == 0 {
		return constants.MsgHistoryStatusOK
	}

	return constants.MsgHistoryStatusFail
}

// printHistoryJSON outputs history as JSON.
func printHistoryJSON(records []model.CommandHistoryRecord) {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ Failed to marshal history to JSON: %v\n", err)

		return
	}
	fmt.Println(string(data))
}
