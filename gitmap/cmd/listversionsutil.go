package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v17/gitmap/release"
)

// collectVersionTags reads git tags, parses, sorts descending.
func collectVersionTags() []release.Version {
	cmd := exec.Command(constants.GitBin, constants.GitTag,
		constants.GitTagListFlag, constants.GitTagGlob)
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, constants.ErrListVersionsNoTags)
		os.Exit(1)
	}

	versions := parseVersionTags(strings.TrimSpace(string(out)))
	if len(versions) == 0 {
		fmt.Fprintln(os.Stderr, constants.ErrListVersionsNoTags)
		os.Exit(1)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GreaterThan(versions[j])
	})

	return versions
}

// parseVersionTags parses lines into valid versions.
func parseVersionTags(output string) []release.Version {
	lines := strings.Split(output, "\n")
	var versions []release.Version

	for _, line := range lines {
		tag := strings.TrimSpace(line)
		if len(tag) == 0 {
			continue
		}
		v, err := release.Parse(tag)
		if err != nil {
			continue
		}
		versions = append(versions, v)
	}

	return versions
}

// loadChangelogMap reads CHANGELOG.md into a version→notes map.
func loadChangelogMap() map[string][]string {
	entries, err := release.ReadChangelog()
	if err != nil {
		return map[string][]string{}
	}

	m := make(map[string][]string, len(entries))
	for _, e := range entries {
		m[e.Version] = e.Notes
	}

	return m
}

// printVersionEntriesTerminal prints versions with source and changelog sub-points.
func printVersionEntriesTerminal(entries []versionEntry) {
	for _, e := range entries {
		if e.Source != "" {
			fmt.Printf("%s  [%s]\n", e.Version.String(), e.Source)
		} else {
			fmt.Println(e.Version.String())
		}
		for _, note := range e.Notes {
			fmt.Printf("  - %s\n", note)
		}
	}
}

// lvJSONEntry is the JSON output shape for list-versions.
type lvJSONEntry struct {
	Version   string   `json:"version"`
	Source    string   `json:"source,omitempty"`
	Changelog []string `json:"changelog,omitempty"`
}

// printVersionEntriesJSON prints versions with source and changelog as JSON.
func printVersionEntriesJSON(entries []versionEntry) {
	out := make([]lvJSONEntry, len(entries))
	for i, e := range entries {
		out[i] = lvJSONEntry{Version: e.Version.String(), Source: e.Source, Changelog: e.Notes}
	}

	data, err := json.MarshalIndent(out, "", constants.JSONIndent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ✗ Failed to marshal versions to JSON: %v\n", err)

		return
	}
	fmt.Println(string(data))
}
