// Package commitin_e2e_test is the end-to-end test harness for the
// `gitmap commit-in` pipeline. Unlike the per-package unit tests
// under cmd/commitin/**, this suite drives orchestrator.Run against
// a real on-disk git repository created by gitfixture.go and asserts
// the visible side effects (run rows, rewritten commits, dedupe
// behavior) the way a downstream user would observe them.
//
// Layout (Step 8 of the e2e plan; Steps 9-12 add more cases):
//
//	gitfixture.go    — RepoBuilder: scaffolds repo, seeds commits with
//	                   controlled author/committer dates so SHAs are
//	                   stable across runs and platforms.
//	harness.go       — RunHarness: invokes orchestrator.Run with a
//	                   captured stdout/stderr and exposes the on-disk
//	                   SQLite for assertions.
//	happy_path_test.go — Step 9 happy-path / dedupe smoke test.
//
// All tests skip cleanly when `git` is not on PATH so the suite is
// safe to run inside hermetic sandboxes that lack a git binary.
package commitin_e2e_test
