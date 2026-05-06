// Package formatter — clonescript.go generates a clone.ps1 PowerShell script.
package formatter

import (
	"io"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/model"
)

// WriteCloneScript writes a self-contained PowerShell clone script
// using the embedded clone.ps1.tmpl template.
func WriteCloneScript(w io.Writer, records []model.ScanRecord) error {
	tmpl, err := loadTemplate("clone.ps1.tmpl")
	if err != nil {
		return err
	}

	data := CloneData{
		Repos: buildRepoEntries(records),
	}

	return tmpl.Execute(w, data)
}

// buildRepoEntries converts ScanRecords into template-friendly RepoEntry slices.
func buildRepoEntries(records []model.ScanRecord) []RepoEntry {
	entries := make([]RepoEntry, 0, len(records))
	for _, r := range records {
		entries = append(entries, RepoEntry{
			Name:   r.RepoName,
			Branch: r.Branch,
			URL:    cloneURL(r),
			Path:   backslashPath(r.RelativePath),
		})
	}

	return entries
}

// cloneURL picks the best URL from a record.
func cloneURL(r model.ScanRecord) string {
	if len(r.HTTPSUrl) > 0 {
		return r.HTTPSUrl
	}

	return r.SSHUrl
}
