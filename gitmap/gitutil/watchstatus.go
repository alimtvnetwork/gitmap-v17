package gitutil

import (
	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// FetchAll runs git fetch --all --prune for a repo (best effort).
func FetchAll(repoPath string) {
	_, _ = runGit(repoPath, constants.GitFetch, constants.GitArgAll, constants.GitArgPrune)
}
