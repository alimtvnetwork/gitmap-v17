- ✅ **Step 6 — Marker, CHANGELOG, helptext audit (2026-05-06).**
  Verified `gitmap/constants/constants_cli.go` line 3 already
  carries the file-wide `// gitmap:cmd top-level` marker, so
  `CmdCommitIn` + `CmdCommitInAlias` (lines 159-160, no `skip` tag)
  are auto-discovered by the completion generator — no edit needed.
  `gitmap/cmd/rootcore.go` line 42 already routes both tokens to
  `runCommitIn(argsTail())`. Helptext at `gitmap/helptext/commit-in.md`
  is 105 lines, under the 120-line cap. Added a comprehensive
  v4.18.0 entry to `CHANGELOG.md` covering: command surface,
  source-resolution rules, input keywords (`all` / `-N`), profile
  schema + load order + save semantics + `--set-default` atomicity,
  conflict modes (ForceMerge vs Prompt), `--dry-run` banner,
  function-intel, message-pipeline order, exclusions, on-disk
  layout under `<source>/.gitmap/`, and the advisory file lock.
  `Version` constant in `constants.go` is already at `4.18.0`.
  `go vet ./...` + `go build ./...` clean.
- ✅ **Step 5 — Conflict-mode wiring + dry-run banner (2026-05-06).**
  Added `replay/clobber.go` (`DetectClobbers` compares
  `<sha>:path` blob hashes via `git rev-parse` on both source and
  target HEAD; missing-on-target = add, not clobber; empty HEAD =
  empty repo = nothing to clobber). New `orchestrator/conflict.go`
  introduces a package-private `errConflictAborted` sentinel and a
  `conflictCheck(ctx, plan, c, stdout) → (abort, skip)` helper that
  consults `finalize.Resolve(ConflictMode, sourceSha, stdout)`:
  ForceMerge logs the clobber count and proceeds; Prompt aborts the
  whole run by flipping `runContext.aborted`. `commit.go`
  short-circuits replay through `conflictCheck` BEFORE
  `replay.ApplyCommit` so an aborted run records exactly one Failed
  row and writes nothing. `pipeline.go` polls `ctx.aborted` after
  every commit + every input and `executePipeline` returns
  `CommitInExitConflictAborted` so the top-level `Run` propagates
  the spec §2.7 exit code. `finalize.PrintDryRunBanner` appends a
  one-line "DRY RUN — no commits were created" notice when
  `--dry-run` is set, called from `Run` after `PrintSummary` so
  CI grep cannot mistake a dry-run for a zero-commit real run.
  `go build ./...` + `go test ./cmd/commitin/... ./store/...` green.

# Plan: Install System Overhaul + README Redesign

> Status legend: ✅ Done · 🔄 In Progress · ⏳ Pending · 🚫 Blocked

## v3.39.0 Release-Version Script (spec 105) — 2026-04-21
- ✅ Authored `spec/01-app/105-release-version-script.md` (full contract: URL flow, generic + snapshot artefacts, flags, exit codes, missing-version interactive flow)
- ✅ Authored `.lovable/memory/features/release-version-script.md` and indexed
- ⏳ Implement `gitmap/scripts/release-version.ps1` (generic, parameterized, embedded via `go:embed`)
- ⏳ Implement `gitmap/scripts/release-version.sh` (bash counterpart, identical contract)
- ⏳ Add `constants/constants_install.go` entries: `ScriptReleaseVersionPS1`, `ScriptReleaseVersionSh`, snapshot filename format `release-version-vX.Y.Z.{ps1,sh}`
- ⏳ Wire snapshot generation into `cmd/release.go` release pipeline (copy generic, prepend `$Version = '<tag>'`, upload as release asset alongside binaries + checksums)
- ⏳ Update `src/pages/Release.tsx` to render TWO install boxes per `/release/:version` page: pinned (snapshot URL) and generic (`-Version` parameter form)
- ⏳ Confirm front-page `install.ps1` (latest-resolving) is **untouched** — out of scope
- ⏳ Add Vitest coverage for snapshot generator: input version → script body has `$Version = 'vX.Y.Z'` at line 1
- ⏳ Add Go test for missing-version interactive flow (mock GitHub API 404 + simulated TTY)
- ⏳ Add CHANGELOG v3.39.0 entry + bump `Version` constant
- 🚫 Decision needed before implementation: confirm GitHub release asset upload is automated (release pipeline) or manual (release notes checklist)


## v3.12.1 Session Snapshot (2026-04-20)
- ✅ Migrated all stale `Draft`/`PreRelease` field references to `IsDraft`/`IsPreRelease` (`release/metadata_test.go`, `tests/release_test/skipmeta_test.go`)
- ✅ Fixed `cmd/probe.go` `go vet` non-constant format string error
- ✅ Implemented `TestTopLevelCmdRegistryMatchesAST` AST parity test
- ✅ Cross-linked uniqueness CI guard from `spec/01-app/02-cli-interface.md` and `38-command-help.md`
- ✅ Bumped `Version` constant → `3.12.1`; added CHANGELOG v3.12.1 entry
- ✅ v15 legacy compat shim audit — KEEP through v3.x, remove in v4.0.0 (`mem://02-v15-legacy-compat-audit`)
- ✅ Generated fresh 28-table ERD `spec/01-app/gitmap-database-erd-v3.12.1.mmd`
- ⏳ Run `.\run.ps1` then `go test ./...` end-to-end build/test sweep
- ⏳ Tag and publish v3.12.1 GitHub release
- ⏳ Author `spec/01-app/v4-breaking-change-matrix.md`
- ⏳ Audit `migrate_v15phase4.go` for v4.0 removal schedule
- ⏳ Promote new ERD to canonical (delete stale ERDs, rename to `gitmap-database-erd.mmd`)
- ⏳ Add CI test for ERD ↔ `SQLCreate*` parity

## v3.0.0 Session Snapshot (2026-04-19)
- ✅ `as` / `release-alias` / `release-alias-pull` shipped with auto-stash + label-match pop
- ✅ `db-migrate` shipped + auto-invoked from `gitmap update`
- ✅ Marker-comment generator refactor (`// gitmap:cmd top-level` / `// gitmap:cmd skip`)
- ✅ CI `generate-check` drift detection
- ✅ Spec `spec/01-app/98-as-and-release-alias.md` authored (matches 97-move-and-merge format)
- ✅ CHANGELOG v3.0.0 entry + Migration guide block for constants contributors
- ✅ Docs layout shows `v3.0.0` badge (`src/components/docs/DocsLayout.tsx`)
- ⏳ Centralize `VERSION` constant in `src/constants/index.ts`
- ⏳ Add version badge to `Index.tsx` landing page hero
- ⏳ Add `## Migration guide` link to docs sidebar
- ⏳ Lint rule for missing `// gitmap:cmd top-level` markers in `constants/*.go`
- ⏳ Integration test for `release-alias` auto-stash round-trip

## Guardrail: Go Refactor Validation
- After any Go file split or refactor, run `go test ./<affected-package>` before marking the work done.
- Treat unused imports and stale references as blocking regressions, not cleanup for later.
- For install-flow changes under `gitmap/cmd`, verify `go test ./cmd` and `go vet ./cmd` before finalizing.

## Guardrail: Installer Output Contract
- Every installer flow must end with a visible summary showing installed version, binary path, install directory, and PATH target/status.
- Unix installers must print which shell/profile file received the PATH entry and how to reload it.
- Unix installers must explicitly warn that OTHER shells (sh, bash, fish) will NOT have gitmap unless the user manually adds the PATH line to those shells' profiles too.
- Windows installers must print whether User PATH was updated or already present.
- PowerShell installers must show the installed version and binary path.

## Part A: README Redesign (styled after scripts-fixer-v5)
1. **Center-aligned header** with badges, tagline, and horizontal rules
2. **Quick Start** section at the top (one-liner install + first scan)
3. **Clean grouped tables** with consistent formatting (ID-based like scripts-fixer-v5)
4. **Installation section** with all variants (one-liner, pinned version, custom dir, Linux/macOS)
5. **Project Structure** tree view section

---

## Part B: Expand Supported Tools (from scripts-fixer-v5)

### New tools to add to `gitmap install`:

**Core Tools (already have):** vscode, node, yarn, bun, pnpm, python, go, git, git-lfs, gh, github-desktop, cpp, php, powershell

**New tools to add:**
| Tool | Keyword | Choco Package | Winget Package | Apt Package | Brew Package | Snap Package |
|------|---------|---------------|----------------|-------------|-------------|-------------|
| MySQL | `mysql` | `mysql` | — | `mysql-server` | `mysql` | — |
| MariaDB | `mariadb` | `mariadb` | — | `mariadb-server` | `mariadb` | — |
| PostgreSQL | `postgresql` | `postgresql` | — | `postgresql` | `postgresql` | — |
| SQLite | `sqlite` | `sqlite` | — | `sqlite3` | `sqlite` | — |
| MongoDB | `mongodb` | `mongodb` | — | `mongod` | `mongodb-community` | — |
| CouchDB | `couchdb` | `couchdb` | — | `couchdb` | `couchdb` | `couchdb` |
| Redis | `redis` | `redis-64` | — | `redis-server` | `redis` | `redis` |
| Cassandra | `cassandra` | — | — | `cassandra` | `cassandra` | — |
| Neo4j | `neo4j` | `neo4j-community` | — | — | `neo4j` | — |
| Elasticsearch | `elasticsearch` | `elasticsearch` | — | `elasticsearch` | `elasticsearch` | — |
| DuckDB | `duckdb` | `duckdb` | — | — | `duckdb` | — |
| Chocolatey | `chocolatey` | (self) | — | — | — | — |
| Winget | `winget` | — | (self) | — | — | — |

---

## Part C: SQLite Installation Tracking (New DB Table)

### 1. New `InstalledTools` table schema:
```sql
CREATE TABLE IF NOT EXISTS InstalledTools (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    Tool TEXT NOT NULL,
    VersionMajor INTEGER NOT NULL DEFAULT 0,
    VersionMinor INTEGER NOT NULL DEFAULT 0,
    VersionPatch INTEGER NOT NULL DEFAULT 0,
    VersionBuild INTEGER NOT NULL DEFAULT 0,
    VersionString TEXT NOT NULL DEFAULT '',
    PackageManager TEXT NOT NULL DEFAULT '',
    InstalledAt TEXT NOT NULL DEFAULT '',
    UpdatedAt TEXT NOT NULL DEFAULT '',
    InstallPath TEXT NOT NULL DEFAULT '',
    UNIQUE(Tool)
);
```

### 2. New model: `model/installedtool.go`
- `InstalledTool` struct with all fields
- `ParseVersion(versionStr string) (major, minor, patch, build int)` — parse version strings like `20.11.1`, `3.12.4`, `1.23.5`
- `CompileVersionString(major, minor, patch, build int) string` — build `"1.2.3.4"` from parts
- `CompareVersions(a, b InstalledTool) int` — compare two versions (-1, 0, 1)

### 3. Store operations: `store/installedtools.go`
- `SaveInstalledTool(tool InstalledTool) error` — INSERT OR REPLACE
- `GetInstalledTool(name string) (InstalledTool, error)`
- `ListInstalledTools() ([]InstalledTool, error)`
- `RemoveInstalledTool(name string) error`
- `IsInstalled(name string) bool`

### 4. Post-install recording
After successful `installTool()`, detect the installed version and save a record to the DB with parsed version components.

---

## Part D: Multi-Platform Package Manager Resolution

### 1. Config-based default manager (`config.json`):
```json
{
  "install": {
    "defaultManager": "choco",
    "managers": {
      "windows": "choco",
      "darwin": "brew",
      "linux": "apt"
    }
  }
}
```

### 2. Resolution priority:
1. `--manager` CLI flag (explicit override)
2. `install.defaultManager` from config.json
3. Platform auto-detect:
   - **Windows** → Chocolatey (fallback: Winget)
   - **macOS** → Homebrew
   - **Linux** → apt (fallback: snap, dnf)

### 3. Add Snap package manager support:
- New `PkgMgrSnap = "snap"` constant
- `buildSnapCommand(pkg string) []string` → `["sudo", "snap", "install", pkg]`
- Snap package name mappings for databases (redis, couchdb, etc.)

### 4. Expand package name mappings:
- `resolveAptPackage(tool) string` — Ubuntu/Debian package names
- `resolveBrewPackage(tool) string` — Homebrew package/cask names  
- `resolveSnapPackage(tool) string` — Snap package names
- Each function has a complete mapping for all ~27 tools

---

## Part E: Uninstall Support

### 1. New `gitmap uninstall <tool>` command:
- Check if tool exists in `InstalledTools` DB
- Build uninstall command based on the package manager that was used to install
- Remove the DB record after successful uninstall

### 2. Uninstall command builders:
- `buildChocoUninstallCommand(pkg) []string` → `["choco", "uninstall", pkg, "-y"]`
- `buildWingetUninstallCommand(pkg) []string` → `["winget", "uninstall", pkg]`
- `buildAptUninstallCommand(pkg) []string` → `["sudo", "apt", "remove", "-y", pkg]`
- `buildBrewUninstallCommand(pkg) []string` → `["brew", "uninstall", pkg]`
- `buildSnapUninstallCommand(pkg) []string` → `["sudo", "snap", "remove", pkg]`

### 3. Flags:
- `--dry-run` — show command without executing
- `--force` — skip confirmation
- `--purge` — remove config files too (apt: `purge`, choco: `-x`)

---

## Part F: Install List/Status Enhancements

### 1. `gitmap install --list` improvements:
- Group tools by category (Core, Databases, Utilities)
- Show installed status from DB (✓/✗ indicator)
- Show installed version from DB

### 2. `gitmap install --status` (new flag):
- Show all tools from DB with version, manager, install date
- Highlight outdated packages (compare DB version vs detected version)

### 3. `gitmap install --upgrade <tool>` (new flag):
- Re-run install for an already-installed tool to upgrade it
- Update the DB record with new version

---

## Execution Order

| Phase | Steps | Files Changed |
|-------|-------|---------------|
| **Phase 1** | README redesign (centered badges, clean structure) | `README.md` |
| **Phase 2** | Add new database tool constants + package mappings | `constants_install.go`, `installtools.go` |
| **Phase 3** | Add `InstalledTools` DB table + model + store CRUD | `store/`, `model/`, migration |
| **Phase 4** | Wire post-install DB recording + version parsing | `cmd/install.go`, `cmd/installtools.go` |
| **Phase 5** | Add config-based manager resolution | `config.json` schema, `cmd/installtools.go` |
| **Phase 6** | Add Snap package manager support | `constants_install.go`, `installtools.go` |
| **Phase 7** | Add uninstall command | `cmd/uninstall.go`, constants, helptext |
| **Phase 8** | Enhanced `--list`, `--status`, `--upgrade` flags | `cmd/install.go` |
| **Phase 9** | Completion support for install/uninstall tool names | Shell scripts, completion handler |

Each phase is independently shippable and testable.

---

## Part G: Pending Task Workflow (Task-Based Deletion)

Spec: `spec/01-app/83-pending-task-workflow.md`
Prevention: `spec/02-app-issues/21-pending-task-durability.md`

### Rule
Every `os.Remove` / `os.RemoveAll` must be preceded by a `PendingTask` insert.
No silent loss of delete intent is acceptable.

### Phase 1 — Database Layer
| Step | Files |
|------|-------|
| Add `TaskType`, `PendingTask`, `CompletedTask` SQL to constants | `constants/constants_pending_task.go` |
| Add model structs | `model/pendingtask.go`, `model/tasktype.go` |
| Add store CRUD (insert, list, complete, fail, find) | `store/pendingtask.go`, `store/tasktype.go` |
| Add seed logic for TaskType (Delete, Remove) | `store/store.go` (Migrate) |
| Add create/drop to migration + reset | `store/store.go` |
| Run `go test ./store/...` | — |

### Phase 2 — Delete Workflow Integration
| Step | Files |
|------|-------|
| Wrap `clone-next --delete` removal in task flow | `cmd/clonenext.go` |
| Create helpers: `CreateTask`, `CompleteTask`, `FailTask` | `cmd/pendingtaskhelper.go` |
| Duplicate prevention (same type + path) | `store/pendingtask.go` |
| Run `go vet ./cmd` + `go test ./cmd` | — |

### Phase 3 — CLI Commands
| Step | Files |
|------|-------|
| Add `pending` command (list all pending tasks) | `cmd/pending.go` |
| Add `do-pending` / `dp` command (retry all) | `cmd/dopending.go` |
| Add `do-pending <id>` (retry single) | `cmd/dopending.go` |
| Route in dispatcher | `cmd/roottooling.go` |
| Add constants (commands, messages, errors) | `constants/constants_cli.go`, `constants/constants_pending_task.go` |

### Phase 4 — Help Integration
| Step | Files |
|------|-------|
| Create `helptext/pending.md` | `helptext/pending.md` |
| Create `helptext/do-pending.md` | `helptext/do-pending.md` |
| Add to root usage output | `cmd/rootusage.go` |
| Add to UI commands data | `src/data/commands.ts` |
| Update documentation site help page | `src/pages/` |

### Phase 5 — Validation & Edge Cases
| Step | Files |
|------|-------|
| Test missing folder retry | tests |
| Test permission failure | tests |
| Test duplicate prevention | tests |
| Test completed-task transactional move | tests |
| Run full `golangci-lint` | — |

---

## v3.153.0 Clone-Pick (spec 100) — 2026-04-27
- ✅ Authored `spec/01-app/100-clone-pick.md` (full contract: sparse-checkout pipeline, --ask picker, CloneInteractiveSelection schema, --replay rules)
- ✅ Authored `.lovable/memory/features/clone-pick.md` and indexed in `index.md` Core + Memories
- ⏳ Implement `gitmap/constants/constants_clonepick.go` (command IDs, flags, messages, autoExclude defaults)
- ⏳ Implement `gitmap/store/cloneinteractiveselection.go` + add `SQLCreateCloneInteractiveSelection` to constants_store.go and to `Migrate()` statements list
- ⏳ Implement `gitmap/clonepick/` package: `parse.go`, `plan.go`, `sparse.go`, `picker.go` (bubbletea), `persist.go`, `render.go`
- ⏳ Implement `gitmap/cmd/clonepick.go` dispatcher entry + register in `coreDispatchEntries()` in `rootcore.go`
- ⏳ Add `// gitmap:cmd top-level` marker on `CmdClonePick`/`CmdClonePickAlias` const block (drift CI)
- ⏳ Author `gitmap/helptext/clone-pick.md` (≤120 lines, 5 examples per spec §9)
- ⏳ Tests: parse, plan cone-detection, store insert/lookup, cmd dry-run + replay-not-found + missing-args
- ⏳ Bump `Version` constant → `3.153.0`; add CHANGELOG v3.153.0 entry
- ⏳ Verify: `go vet ./...` and `go test ./clonepick/... ./cmd/... ./store/...`

## v3.154.0 rescan-subtree — 2026-04-27
- ✅ Added `gitmap rescan-subtree <absolutePath>` (alias `rss`) — thin wrapper over `runScan` that validates the directory and injects `--max-depth 8` (constants.RescanSubtreeDefaultMaxDepth) when the user does not supply one
- ✅ Constants: `CmdRescanSubtree` / `CmdRescanSubtreeAlias` / `RescanSubtreeDefaultMaxDepth` in `constants_cli.go`; registered in `cmd_constants_test.go` for the uniqueness/parity test
- ✅ Dispatcher: registered in `roottooling.go`; compact help string in `constants_helpgroups.go`; LLM docs entry in `llmdocsgroups.go`
- ✅ Helptext: `gitmap/helptext/rescan-subtree.md` (workflow, behavior, examples, exit codes), auto-discovered by `helptext/coverage_test.go`
- ✅ Tests: `cmd/rescansubtree_test.go` covers arg-splitting (path-only / path+flags / flags+path / inline-value / errors), `--max-depth` injection (default + space override + inline override), banner extraction, and a guardrail asserting the rescan default is deeper than the scan default

## commit-in / cin — 2026-05-06 (SPEC ONLY — DO NOT IMPLEMENT until user says `next`)

Spec authored under `spec/03-commit-in/` (7 files, AI-blind ready). DB
follows project convention: PascalCase tables/columns/JSON keys/JSON
values; every PK is `INTEGER PRIMARY KEY AUTOINCREMENT` named
`<TableName>Id`; every classifier (Type/Status/Kind/Mode/Reason/
Outcome/Stage/Source) is an enum mirrored to a `(Id, Name UNIQUE)`
join table. Source-repo auto-init rule encoded: URL → clone; existing
repo → reuse; existing non-repo folder → `git init` in place; missing
path → `mkdir -p && git init`. No prompt, no flag.

### Phased implementation (gated — execute one phase per `next`)

- ✅ **Phase 1 — Constants & enums (2026-05-06).** Added `CmdCommitIn`
  / `CmdCommitInAlias` to `constants_cli.go` and registered both in
  `cmd_constants_test.go` registry. New `gitmap/constants/constants_commitin.go`
  owns all flag names, descriptions, exit codes, enum tokens, defaults,
  filesystem paths, URL prefixes, phase banners, and error formats.
  New `gitmap/cmd/commitin/enums.go` provides typed `uint8` enums
  (`ConflictMode`, `InputKind`, `RunStatus`, `CommitOutcome`,
  `SkipReason`, `ExclusionKind`, `MessageRuleKind`, `FunctionIntelLanguage`)
  with `String()` returning the constants tokens and an `AllX()` slice
  per enum. `enums_test.go` locks every enum's spec ↔ constants ↔ Go
  member list, asserts PascalCase shape, exit-code uniqueness, and
  flag-name kebab-case uniqueness. `go build ./...` and
  `go test ./cmd/commitin/... ./constants/...` both pass.
- ✅ **Phase 2 — DB migrations (2026-05-06).** New
  `gitmap/constants/constants_commitin_sql.go` defines all 18
  commit-in tables (8 enum mirrors + Profile/2 children + 7 run/commit
  tables) + indexes + `INSERT OR IGNORE` seeds. New
  `gitmap/store/migrate_commitin.go` orders the DDL (mirrors before
  Profile before run-chain) and runs it inside the existing
  `Migrate()` pipeline, BEFORE the `SchemaVersionCurrent` stamp so
  any failure forces a retry. `SchemaVersionCurrent` bumped 23 → 24.
  `migrate_commitin_test.go` locks: (a) every spec §4 table is
  present, (b) every enum-mirror seed exactly matches the spec
  member set, (c) re-running `migrateCommitIn()` is a no-op (no
  duplicate seed rows). `go build ./...` clean,
  `go test ./store/... ./cmd/commitin/... ./constants/...` green.
- ✅ **Phase 3 — CLI parsing (2026-05-06).** Pure parser under
  `gitmap/cmd/commitin/parse*.go` (5 files, all <200 lines, every
  func <15 lines). `Parse(args []string) (*RawArgs, *ParseError)`
  with zero git, zero filesystem, zero DB. Splits responsibilities:
  `parse_types.go` (RawArgs + ParseError), `parse_helpers.go`
  (separator/quote split, `-N` keyword classifier, CSV split),
  `parse_validate.go` (enum validators using AllX() from enums.go,
  message-rule shape, author-pair, source/inputs presence,
  KEYWORD-alone), `parse_flags.go` (flag.NewFlagSet registration +
  bool/string/csv groups + bool-set for reorderer), `parse.go`
  (orchestration, positional split, flag re-ordering, `-N` tail
  keyword recognised as positional). Tests: AC #1 separator
  equivalence (6 forms produce identical inputs), AC #4 keyword
  recognition + `-0` rejection, missing positionals, author-pair
  rule, conflict/function-intel/languages enum validators,
  `--message-exclude` shape (Kind:Value), flags-after-positionals
  reordering, `-d` short alias, language error lists supported set.
  `go build ./...` clean; `go test ./cmd/commitin/...` green.
- ✅ **Phase 4 — Workspace + source resolution (2026-05-06).** New
  `gitmap/cmd/commitin/workspace/` package with one helper per file:
  `paths.go` (`EnsureWorkspace` → idempotent `<source>/.gitmap/{,commit-in/{,profiles},temp}` layout),
  `lock.go` (`AcquireLock` / `LockHandle.Release`, stale-PID reclaim,
  spec §2.7 `CommitInExitLockBusy` message),
  `source.go` (`EnsureSource` four-case rule: clone URL / reuse repo /
  init-in-place / mkdir+init), `expand.go` (explicit token classifier
  + `all` / `-N` sibling discovery via `-vN` regex, ascending sort),
  `clone.go` (`CloneInputs` stages every input under
  `<TempRoot>/<runId>/<idx>-<basename>`; local folders reused in
  place), `runner.go` (swappable `gitRunner` + `SetGitRunnerForTest`
  hook). All funcs ≤15 lines, all files <200 lines. Hermetic test
  suite (`workspace_test.go`) with 9 cases covers idempotency, lock
  collision/release, all four `EnsureSource` branches, sibling sort
  (`demo`/`-v1`/`-v3`/`-v10`), `-N` truncation, mixed
  URL+folder classification, three-kind staging.
- ✅ **Phase 5 — Walk + dedupe + replay (2026-05-06).** Four small
  packages under `gitmap/cmd/commitin/`: `walk/` (first-parent
  oldest→newest via `git rev-list --first-parent --reverse`, hydrates
  each commit with metadata + file list using `\x1f`-delimited
  `git show` format; empty-repo path returns nil w/o error),
  `dedupe/` (`Lookup` against `ShaMap` returns `Verdict{IsHit,
  PreviousRewrittenId}`; miss is success-not-error),
  `replay/` (uses `git cat-file blob` → `git hash-object -w` →
  `git update-index --add --cacheinfo` → `git write-tree` →
  `git commit-tree -p HEAD` → `git update-ref HEAD`; pins BOTH dates
  via `GIT_AUTHOR_DATE` / `GIT_COMMITTER_DATE` env vars in RFC3339;
  `dryRun=true` short-circuits with zero side effects),
  `runlog/` (`StartRun` / `FinishRun` / `InsertInputRepo` /
  `InsertSourceCommit` (tx-wrapped commit + files) /
  `RecordRewritten` (also seeds `ShaMap` on `Created`) / `RecordSkip`;
  enum-mirror `Name → Id` lookups centralized in `lookup.go`). All
  packages expose swappable hooks (`SetGitRunnerForTest`, `SetTestHooks`)
  so tests run hermetically. Test suites: walk (2 cases — happy path
  + empty repo), dedupe (3 — miss/hit/empty-input guard), replay
  (3 — dry-run zero-call, full pipeline call sequence, both-dates
  env-var pin), runlog (4 — start/finish flow, atomic source+files
  insert, Created → ShaMap upsert, skip log persistence).
- ✅ **Phase 6 — Profiles + message pipeline (2026-05-06).** Three
  small packages under `gitmap/cmd/commitin/`:
  `profile/` (`types.go` Profile/Author/Exclusion/MessageRule/FunctionIntel +
  LoadError; `json.go` strict `Decode` with `DisallowUnknownFields` +
  SchemaVersion gate, canonical `Encode` with fixed §5.2 key order +
  trailing newline for stable diffs; `io.go` atomic `LoadFromDisk` /
  `SaveToDisk` with `--save-profile-overwrite` refusal; `resolve.go`
  four-layer precedence `defaults < profile < CLI` via
  `Resolve(*CliOverrides, *Profile) Resolved`),
  `message/` (`types.go` Inputs/Result; `strip.go` §6.1 step-1
  StartsWith/EndsWith/Contains line-strip + blank-line collapse;
  `weak.go` §6.2 first-word lowercased + punctuation-stripped
  matcher; `affix.go` title-only first-line affix + body pool
  random-pick wrap; `pipeline.go` `Build()` runs the 6 stages in
  spec order and sets IsEmpty for the EmptyAfterMessageRules skip),
  `prompt/` (`Asker` with swappable `In`/`Out`; `AskString` /
  `AskEnum` honour `--no-prompt` by emitting the standardized
  `commit-in: --no-prompt set but %s is unset` line + returning
  `ErrNoPrompt` for exit-code mapping). All funcs ≤15 lines, all
  files <200 lines. Test suites: profile (7 — encode/decode round
  trip, unknown-field rejection, SchemaVersion gate, overwrite
  refusal + override, missing-file load, ProfilePath layout, auto-
  mkdir, four-layer precedence with CLI winning over profile and
  defaults filling unset weak-words), message (7 — strip+collapse,
  weak-word matrix incl UPPERCASE+colon+empty, override-only-weak
  fires for weak/skips for strong, title affix scoped to first line,
  body affix wraps, function-intel block append, empty-after-strip
  flag, pipeline ordering keeps strip before override gate),
  prompt (4 — no-prompt error path with field name in stderr,
  fallback on empty input, typed answer pass-through, enum retry
  loop until valid).
- ✅ **Phase 7 — Function-intel + finalize (2026-05-06).** Two new
  packages plus dispatcher wire-up:
  `funcintel/` — `types.go` (Detector interface + FileChange),
  `registry.go` (extension→language map + `Get` / `register` /
  `EnabledLanguages`), `diff.go` (shared `addedNames` set-diff +
  line regex extractor), one self-registering file per language
  (`lang_go.go`, `lang_js.go`, `lang_ts.go` reusing JS regex,
  `lang_rust.go`, `lang_python.go`, `lang_php.go`, `lang_java.go`
  shared with C#), `render.go` (§6.3 per-file block, sorted
  ascending, includes newly-added files even when empty).
  `finalize/` — `finalize.go` (`Counters`, `Outcome` mapping
  `Failed>0` to `CommitInExitPartiallyFailed`, `PrintSummary` using
  `CommitInMsgSummaryLine`, `CleanupTemp` honouring `--keep-temp`),
  `conflict.go` (`Resolve(mode, sha, out)` returning
  `ConflictDecisionTakeTheirs` for `ForceMerge` vs `Abort` for
  `Prompt`, prints standardized banner via `CommitInErrConflictAborted`).
  Dispatcher: `gitmap/cmd/commitin.go` adds `runCommitIn` (parses
  argv, exits `CommitInExitBadArgs` on parse error, prints a
  "orchestration loop pending" stub for now); `gitmap/cmd/rootcore.go`
  registers `{CmdCommitIn, CmdCommitInAlias} → runCommitIn`.
  Helptext: `gitmap/helptext/commit-in.md` (105 lines, 5 realistic
  examples, full flag table, exit-code table). Version bumped
  4.17.0 → 4.18.0. All files <200 lines, all funcs ≤15 lines. Tests
  (8 funcintel + 5 finalize): extension dispatch, Go/TS/Python/Java
  detector matrices, EnabledLanguages filter, render output format,
  unchanged-decl noise rejection, Outcome partition, summary format
  uses constants, ForceMerge silent / Prompt prints+aborts,
  CleanupTemp honours keep flag.

## Deferred follow-ups (not blocking the gated 7-phase plan)
- ✅ **Step 2 — End-to-end orchestration (2026-05-06).** New
  `gitmap/cmd/commitin/orchestrator/` package wires the full pipeline:
  `run.go` (top-level Run + setUp/finishSetUp + finalRunStatus),
  `setup.go` (resolveSource → ensureWorkspace → acquireLock →
  openAndMigrate via new `store.OpenAt` → loadProfile/pickProfile),
  `cli_overrides.go` (RawArgs → profile.CliOverrides projection),
  `pipeline.go` (expandAndStage + per-input walk loop with PRNG
  picker for affix randomization), `commit.go` (per-commit dedupe →
  message build → replay → record with skip/fail/created paths),
  `context.go` (runContext bundle + idempotent Cleanup), and
  `input_cache.go` (one InputRepo row per staged input via OrderIndex
  cache). `runCommitIn` in `gitmap/cmd/commitin.go` now delegates to
  `orchestrator.Run` and propagates its exit code. All files
  <200 lines, all funcs ≤15 lines. `go build ./...` clean,
  `go test ./cmd/commitin/... ./store/...` green.
- ✅ **Step 3 — Per-commit pipeline polish (2026-05-06).** Added
  `orchestrator/exclude.go` (applyExclusions, PathFile exact match +
  PathFolder prefix/segment match, POSIX-normalized) and
  `orchestrator/funcintel_block.go` (renderFunctionIntel: `git show
  <sha>:path` + `<sha>^:path` per file, dispatched via
  funcintel.LanguageForPath + EnabledLanguages, "" on failure per
  spec §6.3 best-effort rule). `commit.go` now filters Files through
  exclusions, emits `SkipReasonExcludedAllFiles` when filter empties
  a non-empty list, and threads the §6.3 block into `message.Build`.
  Added `exclude_test.go` (4 cases). All builds + tests green.
- ✅ **Step 4 — Profile save + interactive shim (2026-05-06).**
  Added `profile/build.go` (`BuildFromResolved` materializes a
  byte-stable `Profile` from layered `Resolved` settings) and
  `profile/default.go` (`ClearOtherDefaults` sweeps sibling
  `*.json` profiles bound to the same `SourceRepoPath` and flips
  `IsDefault=false` so only the freshly-saved profile holds the
  flag, satisfying spec §5.5 atomicity intent on the disk-only
  layer). New `orchestrator/save_profile.go` runs once between
  `setUp` and `executePipeline`, gated on `--save-profile <name>`:
  honors `--save-profile-overwrite`, splits "exists" (→
  `CommitInExitBadArgs` with `CommitInErrSaveProfileExists`) from
  generic IO (→ `CommitInExitDbFailed`), and records a `Failed`
  RunStatus when the save aborts before any commits. Interactive
  prompt remains a no-op for the implemented surface — every
  required setting has a built-in default per §02 — so `--no-prompt`
  never produces a `MissingAnswer` exit on this code path. All
  `go build ./...` + `go test ./cmd/commitin/... ./store/...` green.
- `// gitmap:cmd top-level` marker on the `CmdCommitIn` const block in
  `constants_cli.go` (drift-CI catches this on next `generate-check`).
- CHANGELOG v4.18.0 entry documenting the new command surface +
  spec/03-commit-in/ link.

### Guardrails (must hold across every phase)

- No file content, no file hash, no diff payload in SQLite. Only
  `RelativePath` strings.
- Never rewrite an existing source-repo commit; only append.
- Replicate BOTH `AuthorDate` AND `CommitterDate` byte-for-byte.
- Profiles bind by absolute symlink-resolved `<source>` path, never
  by `origin` URL.
- Every error path logs to `os.Stderr` with the standardized format
  (`commit-in: <stage>: <message>`); zero swallow.
