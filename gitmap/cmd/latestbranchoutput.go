// Package cmd — latest-branch output formatters.
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/gitutil"
)

// latestBranchJSON is the JSON output structure.
// CONTRACT: pinned by gitmap/cmd/latestbranchjson_contract_test.go
// (field set, tag names, declaration order; both top-present and
// top-absent states). Changes need fixture regen + changelog bump.
type latestBranchJSON struct {
	Branch     []string              `json:"branch"`
	Remote     string                `json:"remote"`
	Sha        string                `json:"sha"`
	CommitDate string                `json:"commitDate"`
	Subject    string                `json:"subject"`
	Ref        string                `json:"ref"`
	Top        []latestBranchTopItem `json:"top,omitempty"`
}

// latestBranchTopItem is a single entry in the top-N list.
// CONTRACT: same pinning as latestBranchJSON.
type latestBranchTopItem struct {
	Branch     string `json:"branch"`
	Sha        string `json:"sha"`
	CommitDate string `json:"commitDate"`
	Subject    string `json:"subject"`
}

// dispatchLatestOutput routes to the correct output formatter.
func dispatchLatestOutput(result latestBranchResult, items []gitutil.RemoteBranchInfo, cfg latestBranchConfig) {
	if cfg.format == constants.OutputJSON {
		printLatestJSON(result, items, cfg.top)

		return
	}
	if cfg.format == constants.OutputCSV {
		printLatestCSV(items, result.selectedRemote, cfg.top)

		return
	}
	printLatestTerminal(result, items, cfg.top)
}

// printLatestJSON outputs the latest branch result as JSON.
func printLatestJSON(result latestBranchResult, items []gitutil.RemoteBranchInfo, top int) {
	if err := encodeLatestBranchJSON(os.Stdout, result, items, top); err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ Failed to encode latest branch JSON: %v\n", err)
	}
}

// encodeLatestBranchJSON builds the on-the-wire struct and writes it
// to w with the project-standard 2-space indent. Split out from
// printLatestJSON so contract tests can capture the bytes into a
// buffer instead of stdout.
func encodeLatestBranchJSON(
	w io.Writer, result latestBranchResult,
	items []gitutil.RemoteBranchInfo, top int,
) error {
	out := buildLatestJSON(result)
	if top > 0 {
		out.Top = buildTopItems(items, top)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", constants.JSONIndent)

	return enc.Encode(out)
}

// buildLatestJSON constructs the base JSON output struct.
func buildLatestJSON(result latestBranchResult) latestBranchJSON {

	return latestBranchJSON{
		Branch:     result.branchNames,
		Remote:     result.selectedRemote,
		Sha:        result.shortSha,
		CommitDate: result.commitDate,
		Subject:    result.latest.Subject,
		Ref:        result.latest.RemoteRef,
	}
}

// buildTopItems constructs the top-N list for JSON output.
func buildTopItems(items []gitutil.RemoteBranchInfo, top int) []latestBranchTopItem {
	count := top
	if count > len(items) {
		count = len(items)
	}
	topItems := make([]latestBranchTopItem, 0, count)
	for _, item := range items[:count] {
		topItems = append(topItems, latestBranchTopItem{
			Branch:     gitutil.StripRemotePrefix(item.RemoteRef),
			Sha:        gitutil.TruncSha(item.Sha),
			CommitDate: gitutil.FormatDisplayDate(item.CommitDate),
			Subject:    item.Subject,
		})
	}

	return topItems
}

// CSV output lives in latestbranchcsv.go (file-size budget split).

// printLatestTerminal outputs the latest branch result as text.
func printLatestTerminal(result latestBranchResult, items []gitutil.RemoteBranchInfo, top int) {
	fmt.Println()
	printTerminalHeader(result)
	if top > 0 {
		printTerminalTopTable(items, result.selectedRemote, top)
	}
	fmt.Println()
}

// printTerminalHeader prints the main latest-branch info block.
func printTerminalHeader(result latestBranchResult) {
	fmt.Printf(constants.LBTermLatestFmt, strings.Join(result.branchNames, ", "))
	fmt.Printf(constants.LBTermRemoteFmt, result.selectedRemote)
	fmt.Printf(constants.LBTermSHAFmt, result.shortSha)
	fmt.Printf(constants.LBTermDateFmt, result.commitDate)
	fmt.Printf(constants.LBTermSubjectFmt, result.latest.Subject)
	fmt.Printf(constants.LBTermRefFmt, result.latest.RemoteRef)
}

// printTerminalTopTable prints the top-N branches table.
func printTerminalTopTable(items []gitutil.RemoteBranchInfo, remote string, top int) {
	count := resolveTopCount(top, len(items))
	fmt.Println()
	fmt.Printf(constants.LBTermTopHdrFmt, count, remote)
	printTerminalTopHeader()
	for _, item := range items[:count] {
		printTerminalTopRow(item)
	}
}

// printTerminalTopHeader prints the table column headers.
func printTerminalTopHeader() {
	fmt.Printf(constants.LBTermRowFmt,
		constants.LatestBranchTableColumns[0], constants.LatestBranchTableColumns[1],
		constants.LatestBranchTableColumns[2], constants.LatestBranchTableColumns[3])
}

// printTerminalTopRow prints a single branch row.
func printTerminalTopRow(item gitutil.RemoteBranchInfo) {
	fmt.Printf(constants.LBTermRowFmt,
		gitutil.FormatDisplayDate(item.CommitDate),
		gitutil.StripRemotePrefix(item.RemoteRef),
		gitutil.TruncSha(item.Sha),
		item.Subject)
}
