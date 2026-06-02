# Tasks: embed-alignment

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~150 + file moves |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

## Phase 1: File Consolidation
- [x] 1.1 Create `internal/assets` directory.
- [x] 1.2 Move `assets/*` (from project root) to `internal/assets/`.
- [x] 1.3 Move `internal/steps/assets/*` to `internal/assets/`.
- [x] 1.4 Remove old directories if empty.

## Phase 2: Centralize embed.FS
- [x] 2.1 Create `internal/assets/assets.go` matching `gentle-ai` with `var FS embed.FS` and `MustRead`/`Read` helpers.

## Phase 3: Update Consumers
- [x] 3.1 `internal/steps/inject_assets.go`: Remove local `embed.FS` and update `walkAndCopy` to use `assets.FS` from the new package.
- [x] 3.2 `internal/steps/inject_assets_test.go`: Update paths and logic to pass tests against the new `AssetInjector` implementation.
- [x] 3.3 `internal/tui/screens/install.go`: Update persona reading logic to use `assets.Read()`.
- [x] 3.4 `internal/tui/screens/install_test.go`: Update tests for the new persona resolution.

## Phase 4: Maintenance Scripts
- [x] 4.1 `scripts/extract_assets.go`: Update `destDir` to `internal/assets/`.
- [x] 4.2 Run `go test ./...` to ensure the build passes and tests are green.
