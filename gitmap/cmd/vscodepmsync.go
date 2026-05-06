package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/store"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/vscodepm"
)

// syncRecordsToVSCodePM upserts every scanned record into the VSCodeProject
// table and reconciles the alefragnani.project-manager projects.json file.
//
// When noAutoTags is false (default), each pair is enriched with auto-detected
// tags (git/node/go/...) based on the rootPath's top-level files. Tags are
// UNIONed with whatever the user already set in the VS Code UI — gitmap
// never silently removes a user-added tag.
//
// Soft-fails: when the user-data root or the extension dir is missing, the
// function reports a one-line note to stdout and returns without error so
// `gitmap scan` keeps working on machines without VS Code installed.
func syncRecordsToVSCodePM(records []model.ScanRecord, skip, noAutoTags bool) {
	if skip {
		fmt.Print(constants.MsgVSCodePMSyncSkipped)

		return
	}

	if len(records) == 0 {
		return
	}

	if err := upsertVSCodePMRecords(records); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())

		return
	}

	pairs := make([]vscodepm.Pair, 0, len(records))
	for _, r := range records {
		pairs = append(pairs, vscodepm.Pair{
			RootPath: r.AbsolutePath,
			Name:     r.RepoName,
			Tags:     autoTagsFor(r.AbsolutePath, noAutoTags),
		})
	}

	summary, err := vscodepm.Sync(pairs)
	if err != nil {
		reportVSCodePMSoftError(err)

		return
	}

	fmt.Printf(constants.MsgVSCodePMSyncSummary,
		summary.Added, summary.Updated, summary.Unchanged, summary.Total)
}

// autoTagsFor returns the auto-detected tags for rootPath, or nil when
// auto-tagging has been disabled via --no-auto-tags.
func autoTagsFor(rootPath string, disabled bool) []string {
	if disabled {
		return nil
	}

	return vscodepm.DetectTagsCustom(rootPath)
}

// upsertVSCodePMRecords pushes every record into the DB. Errors are
// returned to the caller for centralized stderr reporting.
func upsertVSCodePMRecords(records []model.ScanRecord) error {
	db, err := store.OpenDefault()
	if err != nil {
		return fmt.Errorf(constants.MsgDBUpsertFailed, err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		return fmt.Errorf(constants.MsgDBUpsertFailed, err)
	}

	for _, r := range records {
		if err := db.UpsertVSCodeProject(r.AbsolutePath, r.RepoName); err != nil {
			return err
		}
	}

	return nil
}

// reportVSCodePMSoftError prints a friendly note for the two recoverable
// errors (no VS Code, no extension) and the standardized stderr message
// for anything else.
func reportVSCodePMSoftError(err error) {
	switch {
	case errors.Is(err, vscodepm.ErrUserDataMissing):
		fmt.Printf(constants.MsgVSCodePMSectionHeader,
			"VS Code not detected — sync skipped")
	case errors.Is(err, vscodepm.ErrExtensionMissing):
		fmt.Printf(constants.MsgVSCodePMSectionHeader,
			"alefragnani.project-manager extension not installed — sync skipped")
	default:
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

// renameVSCodePMByPath mirrors a name change to projects.json AND the
// VSCodeProject table. Used by `gitmap as` and any other rename path.
// Soft-fails (no error returned) when VS Code or the extension is missing.
func renameVSCodePMByPath(rootPath, newName string) {
	db, err := store.OpenDefault()
	if err == nil {
		defer db.Close()

		if err := db.Migrate(); err == nil {
			_, _ = db.RenameVSCodeProjectByPath(rootPath, newName)
		}
	}

	changed, err := vscodepm.RenameByPath(rootPath, newName)
	if err != nil {
		reportVSCodePMSoftError(err)

		return
	}

	if changed {
		fmt.Printf(constants.MsgVSCodePMRenamed, rootPath, newName)

		return
	}

	fmt.Printf(constants.MsgVSCodePMRenameNoMatch, rootPath)
}
