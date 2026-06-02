# Verification Report

**Change**: fix-injection-pipeline  
**Date**: 2026-06-01  
**Mode**: Strict TDD  
**Verdict**: PASS WITH WARNINGS

---

## Build Evidence

| Command | Result |
|---------|--------|
| `go build ./...` | ✅ PASS — zero errors |

---

## Test Evidence

| Command | Result |
|---------|--------|
| `go test ./... -count=1` | ✅ PASS — 13 packages, 0 failures |

```
?   github.com/KevG1t/SpecAI/cmd/specai          [no test files]
ok  github.com/KevG1t/SpecAI/internal/backup     1.816s
ok  github.com/KevG1t/SpecAI/internal/catalog    0.589s
ok  github.com/KevG1t/SpecAI/internal/model      0.589s
ok  github.com/KevG1t/SpecAI/internal/pipeline   0.600s
ok  github.com/KevG1t/SpecAI/internal/planner    0.600s
ok  github.com/KevG1t/SpecAI/internal/state      0.823s
ok  github.com/KevG1t/SpecAI/internal/steps      1.081s
ok  github.com/KevG1t/SpecAI/internal/system     1.431s
ok  github.com/KevG1t/SpecAI/internal/tui        1.461s
ok  github.com/KevG1t/SpecAI/internal/tui/screens 0.707s
ok  github.com/KevG1t/SpecAI/internal/update     3.011s
ok  github.com/KevG1t/SpecAI/internal/verify     0.598s
ok  github.com/KevG1t/SpecAI/scripts             0.652s
```

---

## Task Completeness

| Task | Status |
|------|--------|
| 1.1–1.4 Rename 4 asset files | ✅ Complete |
| 1.5 Replace engram text in root `assets/` | ⚠️ 2 residual occurrences — see Issues |
| 1.6 Update Go call-sites to new paths | ✅ Complete — no old paths in Go source |
| 1.7 Verify grep returns 0 matches | ⚠️ FAIL — 2 matches remain |
| 2.1 RED — mock stubs for AgentID/AssetFolder | ✅ Complete |
| 2.2 GREEN — Interface 8 methods | ✅ Complete |
| 2.3 AgentID + AssetFolder on 3 existing adapters | ✅ Complete |
| 2.4 12 new adapter structs | ✅ Complete |
| 2.5 GetAdapters() returns 15 | ✅ Complete |
| 2.6 Table-driven tests for all 15 adapters | ✅ Complete |
| 2.7 Verify go test ./internal/system/... | ✅ PASS |
| 3.1 RED — failing inject test | ✅ Complete |
| 3.2 GREEN — per-agent InjectAgentFolder | ✅ Complete |
| 3.3 Zero-adapters test | ✅ Complete |
| 3.4 Unselected adapter skipped gracefully | ✅ Complete |
| 3.5 Verify go test ./internal/steps/... | ✅ PASS |
| 4.1 RED — persona fallback tests | ✅ Complete |
| 4.2 GREEN — 3-tier persona resolution | ✅ Complete |
| 4.3 Verify go test ./internal/steps/... | ✅ PASS |
| 5.1 RED — AgentSelectModel test | ✅ Complete |
| 5.2 GREEN — agent_select.go | ✅ Complete |
| 5.3 RED — model AgentsSelectedMsg test | ✅ Complete |
| 5.4 GREEN — model.go handler + no dummy IDs | ✅ Complete |
| 5.5 ScreenAgentSelect + DetectedAgentIDs | ✅ Complete |
| 5.6 Verify go test ./internal/tui/... | ✅ PASS |
| 6.1 go build ./... | ✅ PASS |
| 6.2 go test ./... all green | ✅ PASS |
| 6.3 Manual smoke test | ⏳ Not automated (known, tracked) |
| 6.4 Verify no old engram filenames in source | ⚠️ `internal/steps/assets/_shared/engram-convention.md` still exists |

---

## Spec Compliance Matrix

### Spec: engram-to-sdd-memory-rename

| Requirement | Scenario | Status | Covering Test / Evidence |
|-------------|----------|--------|--------------------------|
| Four files renamed | Renamed files accessible under new paths | ✅ PASS | Confirmed: sdd-memory-protocol.md, sdd-memory-compact-prompt.md, sdd-memory-instructions.md, sdd-memory-convention.md exist in root assets/ |
| Text replacement across 25 files | No residual engram references in assets | ⚠️ WARNING | 2 residual occurrences found (see Issues) |
| Section tags updated | Tag renamed in protocol file | ✅ PASS | Root asset files confirmed; not embedded by steps package |
| Internal references updated | MustRead call uses new path | ✅ PASS | No Go source references old paths |

### Spec: ide-adapter (delta)

| Requirement | Scenario | Status | Covering Test |
|-------------|----------|--------|---------------|
| AgentID method on IDEAdapter | AgentID returns correct value | ✅ PASS | TestIDEAdapter_AgentID — ide_test.go:50 |
| AssetFolder method on IDEAdapter | AssetFolder returns "claude" | ✅ PASS | TestIDEAdapter_AssetFolder — ide_test.go:72 |
| Generic adapters return "generic" | openclaw/pi/trae/vscode-copilot | ✅ PASS | TestGenericAdapters_ReturnGenericAssetFolder — ide_test.go:85 |
| 15 adapters registered | GetAdapters() returns exactly 15 | ✅ PASS | TestGetAdapters_Returns15 — ide_test.go:34 |
| Mocks implement new methods | Mock satisfies interface | ✅ PASS | stubIDE in sync_test.go updated; build passes |
| IDEAdapter interface = 8 methods | Exactly 8 methods | ✅ PASS | ide.go:12-21 |

### Spec: persona-injection (delta)

| Requirement | Scenario | Status | Covering Test |
|-------------|----------|--------|---------------|
| Tier 1: agent-specific persona | Agent-specific persona used | ✅ PASS | TestResolvePersona_AgentHasDedicatedPersona — install_test.go:39 |
| Tier 2: generic gentleman fallback | Generic gentleman when tier 1 absent | ✅ PASS | TestResolvePersona_FallbackToGenericGentleman — install_test.go:55 |
| Tier 3: neutral fallback | Neutral when tiers 1+2 absent | ✅ PASS | TestResolvePersona_FallbackToNeutralPersona — install_test.go:72 |
| Error when none found | Returns descriptive error | ✅ PASS | TestResolvePersona_ErrorWhenNoPersonaFound — install_test.go:89 |
| Integration: writes rule file | StepInstallGlobalRules writes file | ✅ PASS | TestStepInstallGlobalRules_WritesRuleFile — install_test.go:101 |

### Spec: agent-aware-injection

| Requirement | Scenario | Status | Covering Test |
|-------------|----------|--------|---------------|
| Per-agent asset injection | Agent-specific assets delivered | ✅ PASS | TestInjectAssets_PerAgentWalk — inject_assets_test.go:52 |
| Missing agent folder skipped | Graceful skip on missing folder | ✅ PASS | TestInjectAssets_MissingAgentFolderSkipped — inject_assets_test.go:121 |
| Shared skills injected once (3 adapters) | Skills not duplicated | ✅ PASS | TestInjectAssets_SharedSkillsOnce — inject_assets_test.go:94 |
| Zero adapters → skills still injected | No adapters = shared still runs | ✅ PASS | TestInjectAssets_NoAdapters — inject_assets_test.go:140 |
| Plan-driven adapter scope | Unselected agent receives no assets | ✅ PASS | TestInjectAssets_PerAgentWalk confirms only ctx.IDEs iterated |
| Go embed limitation | Per-agent content absent from embed; graceful skip | ⚠️ WARNING | Known deviation — user-confirmed follow-up |

### Spec: agent-selection-tui

| Requirement | Scenario | Status | Covering Test |
|-------------|----------|--------|---------------|
| Full catalog displayed (15 agents) | All 15 agents listed | ✅ PASS | TestAgentSelectModel_AllAgentsListed — agent_select_test.go:11 |
| Pre-selection of detected agents | Detected agents pre-checked | ✅ PASS | TestAgentSelectModel_DetectedAgentsPreChecked — agent_select_test.go:20 |
| Enter emits AgentsSelectedMsg | User confirms, message emitted | ✅ PASS | TestAgentSelectModel_EnterWithSelectionEmitsMsg — agent_select_test.go:40 |
| At-least-one guard | Empty selection shows error | ✅ PASS | TestAgentSelectModel_EnterWithNoSelectionShowsError — agent_select_test.go:76 |
| Selection wired to IDEs in context | AgentsSelectedMsg populates installCtx.IDEs | ✅ PASS | TestTUIModel_AgentsSelectedMsg_PopulatesIDEs — model_test.go:53 |
| Screen sequence: Welcome → AgentSelect → Pipeline | Correct screen order | ✅ PASS | TestTUIAsyncUpdates — model_test.go:10 |
| No hardcoded dummy agent list | model.go uses system.DetectedAgentIDs() | ✅ PASS | Confirmed by code review: model.go:122 |

---

## Design Coherence

| Decision | Implemented | Notes |
|----------|-------------|-------|
| IDEAdapter extended with AgentID() + AssetFolder() | ✅ Yes | Backward-compatible; existing mocks updated |
| AssetInjector interface split from step | ✅ Yes | Allows mock injection in tests |
| personaReadFileFS injectable for persona resolution | ✅ Yes | Clean inversion; tests use testPersonaFS |
| AgentsSelectedMsg defined in screens package | ✅ Yes | Avoids circular import |
| DetectedAgentIDs() added as convenience wrapper | ✅ Yes | Hides DetectInstalledIDEs error from TUI |
| installCtx field on MainModel | ✅ Yes | Clean state management across screen transitions |
| InjectSharedSkills uses embed root not assets/skills/ | ⚠️ Deviation | Embed root IS the skills content — correct behavior, semantic mismatch |

---

## Issues

### CRITICAL

*None.*

---

### WARNING

**W1 — Two residual "engram" text occurrences in root `assets/`**

Spec requirement: `grep -r "engram" assets/` MUST return zero matches (engram-to-sdd-memory-rename spec, Scenario: No residual engram references in assets).

Found via PowerShell `Select-String`:

1. `assets/opencode/plugins/background-agents.ts:16` — `Exported as BackgroundAgents (matching the Engram plugin convention)` — Capitalized `Engram` as a proper noun in a code comment; survived bulk replace because the replace targeted lowercase. Addressable.

2. `assets/skills/_shared/sdd-memory-convention.md:122` — `The --project flag and ENGRAM_PROJECT env var can override detection` — `ENGRAM_PROJECT` is an upstream sdd-memory MCP server environment variable name. Renaming it would break the MCP server's own ENV contract (third-party). This is arguably out-of-scope for SpecAI's rename operation.

Task 1.7 (`grep ... = 0 matches`) is formally **INCOMPLETE**. A follow-up decision is needed: either replace the two occurrences or document them as accepted exceptions in the spec.

---

**W2 — `internal/steps/assets/_shared/engram-convention.md` not renamed**

Task 1.4 renamed `assets/skills/_shared/engram-convention.md` → `sdd-memory-convention.md` in **root** `assets/`. However, `internal/steps/assets/_shared/engram-convention.md` is a **separate copy** inside the Go embed directory and was NOT renamed.

This file is embedded into the binary via `//go:embed assets/*` in `inject_assets.go` and delivered to users via `InjectSharedSkills`. Users who install will receive the old-named file (`engram-convention.md`) with old content (`# Engram Artifact Convention` header) in their skills directory.

**Impact**: Functional gap — injected content is stale relative to the root `assets/` tree.

---

**W3 — Go embed limitation: per-agent content never physically injected**

Confirmed: `inject_assets.go` embeds from `internal/steps/assets/` which contains only skill files (no `claude/`, `codex/`, `gemini/` etc. subdirectories). `InjectAgentFolder` gracefully skips missing folders (spec-compliant for the skip scenario), but no per-agent content is ever injected — not a logic bug, the embedded FS simply has no agent-specific directories.

Evaluated as **WARNING** (not CRITICAL) because:
- Graceful-skip test passes (spec scenario covered at the behavior level)
- User explicitly acknowledged and accepted this state
- Follow-up change planned (align embed FS with root assets/)

---

**W4 — `InjectSharedSkills` copies entire embed root, not `assets/skills/` subfolder**

Spec: inject `assets/skills/` into a shared location exactly once. Implementation calls `walkAndCopy("assets", targetDir)` — copies the entire `internal/steps/assets/` tree including `_shared/`, `sdd-apply/`, `sdd-tasks/`, etc.

In practice, the embed root IS the skills content tree, so the final paths are correct (e.g., `~/.specai/skills/sdd-tasks/SKILL.md`). However:
- The spec's `assets/skills/` path does not physically exist in the embed FS
- The stale `_shared/engram-convention.md` (W2) is also injected

No functional bug if the embed root is intentionally structured as a skills tree, but the semantic mismatch with spec phrasing warrants clarification in the follow-up.

---

### SUGGESTION

**S1 — `TestAgentSelectModel_ViewNotEmpty` is near-tautological**

`agent_select_test.go:119` asserts only that `View()` returns a non-empty string. This would pass even if `View()` returned `" "`. A stronger triangulation test (checking that known agent names appear in the rendered output) would meaningfully cover the rendering behavior.

**S2 — No E2E test for AgentsSelectedMsg → StepInjectAssets pipeline wiring**

`TestTUIModel_AgentsSelectedMsg_PopulatesIDEs` confirms `installCtx.IDEs` is populated. No test verifies that these IDEs flow through to `StepInjectAssets` in a pipeline execution. Acceptable for now; worth adding in the embed-alignment follow-up.

**S3 — Task 6.3 manual smoke test not automated**

A headless BubbleTea test with input injection would close this gap without manual effort.

---

## TDD Evidence Review

| Task | TDD Cycle | RED Evidence | GREEN Evidence | Assertion Quality |
|------|-----------|--------------|----------------|-------------------|
| Phase 2: IDEAdapter | ✅ Full | ide_test.go compile-fails before interface extended | 15 adapters + 8-method interface | Table-driven per-field — solid |
| Phase 3: Inject assets | ✅ Full | TestInjectAssets_* fails before AssetInjector split | AssetInjector interface + walkAndCopy | 4 test cases: per-agent walk, shared-once, missing-skip, zero-adapters |
| Phase 4: Persona fallback | ✅ Full | resolvePersona undefined → compile fail | 3-tier chain + defaultPersonaFS bridge | 4 test cases (one per tier + error case) |
| Phase 5: AgentSelectModel | ✅ Full | NewAgentSelectModel undefined → fail | agent_select.go created | 6 tests: catalog size, pre-check, enter-with-sel, enter-empty, space-toggle, view-not-empty |
| Phase 5: model.go wiring | ✅ Full | TestTUIModel_AgentsSelectedMsg fails before handler | Handler + installCtx field added | 2 tests (flow sequence + IDEs population) |

**Banned pattern check**:
- No tautologies (all assertions check meaningful state or exact values)
- TestAgentSelectModel_ViewNotEmpty is borderline (see S1) but not strictly banned
- No empty collection assertions without context
- No type-only assertions — all type checks are paired with content verification

---

## Checklist Summary

| Check | Result |
|-------|--------|
| `go build ./...` clean | ✅ |
| `go test ./...` all green | ✅ |
| `grep -r "engram" assets/` = 0 matches | ❌ 2 matches remain |
| IDEAdapter has AgentID() + AssetFolder() added | ✅ |
| GetAdapters() returns 15 | ✅ |
| inject_assets.go does per-adapter walk | ✅ |
| 3-tier persona fallback in install.go | ✅ |
| AgentSelectModel exists with at-least-one guard | ✅ |
| model.go has no hardcoded dummy agent list | ✅ |
| All spec scenarios covered by passing tests | ✅ (embed deviation — user-accepted) |
| TDD cycle evidence present in apply-progress | ✅ |

---

## Next Recommended

1. **Follow-up change (high priority)**: Rename `internal/steps/assets/_shared/engram-convention.md` → `sdd-memory-convention.md` and update its content (closes W2).
2. **Engram text residuals (medium)**: Decide on `ENGRAM_PROJECT` env var reference (third-party, may be documented exception) and `Engram plugin convention` comment in `background-agents.ts` (addressable rename) to formally close task 1.7 / W1.
3. **Embed alignment follow-up (high, user-planned)**: Align `internal/steps/assets/` with root `assets/` structure to enable actual per-agent content injection (closes W3 + W4).
4. **Manual smoke test**: Run TUI interactively to verify agent select → pipeline flow (task 6.3).
