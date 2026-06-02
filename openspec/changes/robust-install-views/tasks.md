# Tasks: robust-install-views

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~350-450 + embedded assets |
| 400-line budget risk | High |
| Chained PRs recommended | No |
| Suggested split | Single PR (size exception for assets) |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full feature implementation | PR 1 | Due to `single-pr` strategy, size exception is required before apply. |

## Phase 1: Foundation (Assets Extraction)

- [x] 1.1 `scripts/extract_assets_test.go` (RED): Write failing test for `engram` string replacement and basic copy behavior.
- [x] 1.2 `scripts/extract_assets.go` (GREEN): Implement script to pull `gentle-ai` assets, rewrite strings, and output to `assets/`.
- [x] 1.3 `assets/...` (GREEN): Run the script to generate the `assets/` directory.
- [x] 1.4 `scripts/extract_assets.go` (REFACTOR): Clean up file handling and paths.

## Phase 2: Core Implementation (Asset Injection Step)

- [x] 2.1 `internal/steps/inject_assets_test.go` (RED): Write failing test for `AssetInjector.Inject` using `afero` memory filesystem.
- [x] 2.2 `internal/steps/inject_assets.go` (GREEN): Implement `AssetInjector` using `go:embed` to unpack assets to target.
- [x] 2.3 `internal/steps/inject_assets.go` (REFACTOR): Clean up interface implementation and file permission logic.

## Phase 3: Integration (Planner & TUI Wiring)

- [x] 3.1 `internal/planner/` (RED): Write failing test for progress reporting mechanism (`pipeline.ProgressEvent`).
- [x] 3.2 `internal/planner/` (GREEN): Implement `pipeline.ProgressEvent` channels during `StagePlan` execution.
- [x] 3.3 `internal/tui/screens/install_test.go` (RED): Add failing test for dynamic pipeline rendering.
- [x] 3.4 `internal/tui/model.go` & `internal/tui/screens/install.go` (GREEN): Replace 3-state enum. Update `InstallModel` to consume `planner.ResolvedPlan` and render progress channels.
- [x] 3.5 `internal/tui/screens/install.go` (REFACTOR): Extract BubbleTea view components for pipeline steps.

## Phase 4: Cleanup

- [x] 4.1 `internal/steps/global_skills.go`: Delete file (replaced by modular asset injector).
- [x] 4.2 Fix any resulting build errors from removed dependencies.
