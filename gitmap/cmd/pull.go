package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/cloner"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/store"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/verbose"
)

// pullOptions holds parsed pull flags.
type pullOptions struct {
	slug          string
	group         string
	all           bool
	verbose       bool
	stopOnFail    bool
	parallel      int
	onlyAvailable bool
}

// runPull handles the "pull" subcommand.
func runPull(args []string) {
	checkHelp("pull", args)
	requireOnline()
	opts := parsePullFlags(args)
	if opts.verbose {
		initVerboseLog()
	}
	records := resolvePullTargets(opts.slug, opts.group, opts.all)
	if opts.onlyAvailable {
		records = filterByAvailableUpdates(records)
		if len(records) == 0 {
			fmt.Print(constants.MsgPullNoAvailable)
			return
		}
	}

	taskID, taskDB := beginPullTask(records)
	if taskDB != nil {
		defer taskDB.Close()
	}

	prog := cloner.NewBatchProgress(len(records), "Pull", false)
	prog.SetStopOnFail(opts.stopOnFail)
	executePull(records, prog, opts)
	prog.PrintSummary()
	prog.PrintFailureReport()

	if code := prog.ExitCodeForBatch(); code != 0 {
		failPendingTask(taskDB, taskID, fmt.Sprintf("pull batch failed with exit code %d", code))
		exitWith(code)
	}

	completePendingTask(taskDB, taskID)
}

// beginPullTask records the pending task entry for this pull batch.
func beginPullTask(records []model.ScanRecord) (int64, *store.DB) {
	workDir, wdErr := os.Getwd()
	if wdErr != nil {
		fmt.Fprintf(os.Stderr, "  ⚠ Could not determine working directory: %v\n", wdErr)
	}
	cmdArgs := buildCommandArgs(append([]string{"pull"}, os.Args[2:]...))
	targetPath := workDir
	if len(records) == 1 {
		targetPath = records[0].AbsolutePath
	}

	return createPendingTask(constants.TaskTypePull, targetPath, workDir, "pull", cmdArgs)
}

// executePull dispatches to either the serial or parallel runner.
func executePull(records []model.ScanRecord, prog *cloner.BatchProgress, opts pullOptions) {
	if opts.parallel > 1 {
		runPullParallel(records, prog, opts.parallel, opts.stopOnFail)
		return
	}
	for _, rec := range records {
		if prog.Stopped() {
			break
		}
		prog.BeginItem(rec.RepoName)
		pullOneRepoTracked(rec, prog)
	}
}

// parsePullFlags parses flags for the pull command.
func parsePullFlags(args []string) pullOptions {
	fs := flag.NewFlagSet(constants.CmdPull, flag.ExitOnError)
	vFlag := fs.Bool("verbose", false, constants.FlagDescVerbose)
	gFlag := fs.String("group", "", constants.FlagDescGroup)
	fs.StringVar(gFlag, "g", "", constants.FlagDescGroup)
	aFlag := fs.Bool("all", false, constants.FlagDescAll)
	sFlag := fs.Bool(constants.FlagStopOnFail, false, constants.FlagDescStopOnFail)
	pFlag := fs.Int("parallel", 1, constants.FlagDescPullParallel)
	oFlag := fs.Bool("only-available", false, constants.FlagDescPullOnlyAvailable)
	fs.Parse(args)

	opts := pullOptions{
		group: *gFlag, all: *aFlag, verbose: *vFlag, stopOnFail: *sFlag,
		parallel: *pFlag, onlyAvailable: *oFlag,
	}
	if fs.NArg() > 0 {
		opts.slug = fs.Arg(0)
	}

	return opts
}

// initVerboseLog sets up verbose logging, warning on failure.
func initVerboseLog() {
	log, err := verbose.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.WarnVerboseLogFailed, err)

		return
	}
	log.Close()
}

// resolvePullTargets returns records based on alias, group, all, or slug lookup.
func resolvePullTargets(slug, groupName string, all bool) []model.ScanRecord {
	if HasAlias() {
		return []model.ScanRecord{{
			RepoName:     GetAliasSlug(),
			Slug:         GetAliasSlug(),
			AbsolutePath: GetAliasPath(),
		}}
	}
	if len(groupName) > 0 {
		return loadRecordsByGroup(groupName)
	}
	if all {
		return loadAllRecordsDB()
	}
	if len(slug) == 0 {
		fmt.Fprintln(os.Stderr, constants.ErrPullSlugRequired)
		fmt.Fprintln(os.Stderr, constants.ErrPullUsage)
		os.Exit(1)
	}

	return lookupBySlugDBFirst(slug)
}

// lookupBySlugDBFirst tries the database first, then falls back to JSON.
func lookupBySlugDBFirst(slug string) []model.ScanRecord {
	db, err := openDB()
	if err == nil {
		defer db.Close()
		repos, dbErr := db.FindBySlug(strings.ToLower(slug))
		if dbErr == nil && len(repos) > 0 {
			return repos
		}
	}

	return lookupBySlugJSON(slug)
}

// lookupBySlugJSON loads gitmap.json and matches by repo name.
func lookupBySlugJSON(slug string) []model.ScanRecord {
	jsonPath := filepath.Join(constants.DefaultOutputFolder, constants.DefaultJSONFile)
	records, err := loadJSONRecords(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrPullLoadFailed, jsonPath, err)
		os.Exit(1)
	}

	return findBySlug(records, slug)
}

// loadJSONRecords reads ScanRecords from a JSON file.
func loadJSONRecords(path string) ([]model.ScanRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []model.ScanRecord
	err = json.NewDecoder(file).Decode(&records)

	return records, err
}

// findBySlug finds records matching the slug (case-insensitive, partial match).
func findBySlug(records []model.ScanRecord, slug string) []model.ScanRecord {
	slugLower := strings.ToLower(slug)
	exact, partial := partitionBySlug(records, slugLower)

	if len(exact) > 0 {
		return exact
	}

	return partial
}

// partitionBySlug separates records into exact and partial matches.
func partitionBySlug(records []model.ScanRecord, slugLower string) ([]model.ScanRecord, []model.ScanRecord) {
	var exact, partial []model.ScanRecord

	for _, r := range records {
		nameLower := strings.ToLower(r.RepoName)
		if nameLower == slugLower {
			exact = append(exact, r)
		} else if strings.Contains(nameLower, slugLower) {
			partial = append(partial, r)
		}
	}

	return exact, partial
}

// pullOneRepo runs safe-pull on a single repo using its absolute path.
func pullOneRepo(rec model.ScanRecord) {
	fmt.Printf(constants.MsgPullStarting, rec.RepoName, rec.AbsolutePath)

	if cloner.IsMissingRepo(rec.AbsolutePath) {
		fmt.Fprintf(os.Stderr, constants.ErrPullNotRepo, rec.AbsolutePath)

		return
	}

	result := cloner.SafePullOne(rec, rec.AbsolutePath)
	if result.Success {
		fmt.Printf(constants.MsgPullSuccess, rec.RepoName)
	} else {
		fmt.Fprintf(os.Stderr, constants.MsgPullFailed, rec.RepoName, result.Error)
	}
}

// pullOneRepoTracked runs safe-pull with progress tracking.
func pullOneRepoTracked(rec model.ScanRecord, prog *cloner.BatchProgress) {
	if cloner.IsMissingRepo(rec.AbsolutePath) {
		prog.Skip()

		return
	}

	result := cloner.SafePullOne(rec, rec.AbsolutePath)
	if result.Success {
		prog.Succeed()
	} else {
		prog.FailWithError(rec.RepoName, result.Error)
	}
}
