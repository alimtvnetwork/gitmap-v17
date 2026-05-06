package render

// adapters.go — convenience builders that turn the gitmap data
// types most-likely-to-be-rendered (model.ScanRecord, etc.) into
// RepoTermBlock instances. Lives in the render package so callers
// don't have to know the field-name mapping; producers in
// cmd/clone-from/clone-next/probe import this helper to stay DRY.

import (
	"fmt"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
)

// FromScanRecord builds a RepoTermBlock from a model.ScanRecord.
// idx is the 1-based row number to surface to the user.
//
// CloneCommand defaults to the record's CloneInstruction, which
// the mapper has already shaped to "git clone [-b BRANCH] URL PATH".
// When the instruction is empty we synthesize a minimal command
// from the picked URL so the block is always populated.
func FromScanRecord(idx int, r model.ScanRecord) RepoTermBlock {
	original := preferHTTPS(r.HTTPSUrl, r.SSHUrl)
	target := original
	cmd := strings.TrimSpace(r.CloneInstruction)
	if len(cmd) == 0 && len(target) > 0 {
		cmd = fmt.Sprintf("git clone %s", target)
	}

	return RepoTermBlock{
		Index:        idx,
		Name:         r.RepoName,
		Branch:       r.Branch,
		BranchSource: r.BranchSource,
		OriginalURL:  original,
		TargetURL:    target,
		CloneCommand: cmd,
	}
}

// FromScanRecords maps a slice with 1-based indexing.
func FromScanRecords(records []model.ScanRecord) []RepoTermBlock {
	out := make([]RepoTermBlock, 0, len(records))
	for i, r := range records {
		out = append(out, FromScanRecord(i+1, r))
	}

	return out
}

// preferHTTPS returns the HTTPS URL when present, otherwise the SSH
// URL. Mirrors formatter.cloneURL — duplicated here to avoid an
// import cycle (formatter → render → formatter).
func preferHTTPS(https, ssh string) string {
	if len(strings.TrimSpace(https)) > 0 {
		return https
	}

	return ssh
}
