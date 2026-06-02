# Tasks: Fix Injection Pipeline

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~600 (code) + 25 asset files text-only |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: asset renames · PR 2: IDEAdapter foundation · PR 3: injection + TUI logic |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main/pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main|pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| A | Rename 4 assets + replace 25 files | PR 1 | Zero code risk; independent |
| B | IDEAdapter interface + 12 adapters + mock fixes | PR 2 | Foundation; all other code depends on this |
| C | inject_assets + persona + AgentSelectModel + model.go wiring | PR 3 | Main logic; depends on PR 2 |

---

## Phase 1: Asset Rename (PR 1 — Unit A)

- [x] 1.1 Rename `assets/claude/engram-protocol.md` → `assets/claude/sdd-memory-protocol.md`
- [x] 1.2 Rename `assets/codex/engram-compact-prompt.md` → `assets/codex/sdd-memory-compact-prompt.md`
- [x] 1.3 Rename `assets/codex/engram-instructions.md` → `assets/codex/sdd-memory-instructions.md`
- [x] 1.4 Rename `assets/skills/_shared/engram-convention.md` → `assets/skills/_shared/sdd-memory-convention.md`
- [x] 1.5 Run `grep -r "engram" assets/` — replace every occurrence (`engram`→`sdd-memory`, `Engram`→`sdd-memory`) across all 25 affected files, including `<!-- gentle-ai:engram-* -->` section tags
- [x] 1.6 Grep source code for `assets.MustRead` or any Go string referencing old filenames; update any call-sites to new paths
- [x] 1.7 **Verify**: `grep -r "engram" assets/` must return zero matches (spec scenario: *No residual engram references in assets*)

## Phase 2: IDEAdapter Foundation (PR 2 — Unit B)

> TDD order: update mocks first (compile fails = RED), then extend interface (GREEN).

- [x] 2.1 **RED** — In `internal/system/ide_test.go` (create if missing): add `AgentID() model.AgentID` and `AssetFolder() string` stubs to every mock `IDEAdapter`; confirm `go build ./...` fails
- [x] 2.2 **GREEN** — In `internal/system/ide.go`: add `AgentID() model.AgentID` and `AssetFolder() string` to the `IDEAdapter` interface (8 methods total)
- [x] 2.3 Add `AgentID()` + `AssetFolder()` to the 3 existing adapter structs (`CursorAdapter`, `WindsurfAdapter`, `CodexAdapter`) with correct values per the asset-folder mapping table
- [x] 2.4 Add 12 new adapter structs: `ClaudeCodeAdapter`, `GeminiCLIAdapter`, `AntigravityAdapter`, `OpenCodeAdapter`, `KiroIDEAdapter`, `KimiCodeAdapter`, `QwenCodeAdapter`, `KiloAdapter`, `OpenClawAdapter`, `PiAdapter`, `TraeAdapter`, `VSCodeCopilotAdapter` — each implementing all 8 interface methods
- [x] 2.5 Update `GetAdapters()` to return all 15 adapters (verify slice length = 15)
- [x] 2.6 Write table-driven tests in `internal/system/ide_test.go`: for each of the 15 adapters assert `AgentID()` and `AssetFolder()` return the exact values from the mapping table
- [x] 2.7 **Verify**: `go test ./internal/system/...` passes green

## Phase 3: Injection Fix (PR 3 — Unit C, part 1)

- [x] 3.1 **RED** — In `internal/steps/inject_assets_test.go`: write failing test for per-agent walk and shared-skills-once test
- [x] 3.2 **GREEN** — Rewrite `Inject()` in `internal/steps/inject_assets.go`: loop over `ctx.IDEs`, copy `assets/{adapter.AssetFolder()}/` to `adapter.GlobalSkillsDir()`; skip gracefully if folder absent; after loop copy `assets/skills/` once
- [x] 3.3 Write test: zero adapters in plan → skills still injected, no error
- [x] 3.4 Write test: unselected adapter folder NOT present in any target dir (graceful skip behavior)
- [x] 3.5 **Verify**: `go test ./internal/steps/...` passes green

## Phase 4: Persona Fix (PR 3 — Unit C, part 2)

- [x] 4.1 **RED** — In `internal/steps/install_test.go`: write tests for three-tier persona fallback chain and error case
- [x] 4.2 **GREEN** — In `internal/steps/install.go`: update `StepInstallGlobalRules` to resolve persona via priority chain: agent-specific → generic gentleman → neutral; error if none found
- [x] 4.3 **Verify**: `go test ./internal/steps/...` passes green

## Phase 5: Agent Selection TUI (PR 3 — Unit C, part 3)

- [x] 5.1 **RED** — In `internal/tui/screens/agent_select_test.go` (new): write failing tests
- [x] 5.2 **GREEN** — Create `internal/tui/screens/agent_select.go`: implement `AgentSelectModel`
- [x] 5.3 **RED** — In `internal/tui/model_test.go`: write failing test for `AgentsSelectedMsg` handling
- [x] 5.4 **GREEN** — In `internal/tui/model.go`: add `AgentsSelectedMsg` handler; remove hardcoded agent IDs; route Install through AgentSelect
- [x] 5.5 Add `ScreenAgentSelect` constant to router.go; add `DetectedAgentIDs()` to system/ide.go
- [x] 5.6 **Verify**: `go test ./internal/tui/...` passes green

## Phase 6: Integration & Cleanup

- [x] 6.1 Run `go build ./...` — zero compile errors
- [x] 6.2 Run `go test ./...` — all tests green; no regressions
- [ ] 6.3 Manual smoke: launch TUI, dismiss welcome, verify agent select screen shows 15 agents, select ≥2, confirm pipeline runs without panic
- [x] 6.4 Verify zero occurrences of old engram filenames in source
