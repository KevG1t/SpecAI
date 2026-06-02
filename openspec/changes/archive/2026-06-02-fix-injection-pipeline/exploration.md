# Exploration: fix-injection-pipeline

## Current State

### Asset layout (already correct)
`assets/` is perfectly structured per-agent:
```
assets/
  claude/           persona-gentleman.md, sdd-orchestrator.md, agents/, commands/
  antigravity/      sdd-orchestrator.md
  gemini/           sdd-orchestrator.md
  cursor/           sdd-orchestrator.md, agents/
  windsurf/         (implied)
  kimi/             KIMI.md, agents/, persona-gentleman.md, ...
  opencode/         sdd-orchestrator.md, persona-gentleman.md, ...
  kiro/             sdd-orchestrator.md, persona-gentleman.md, agents/
  gga/              AGENTS.md, gga.ps1, pr_mode.sh
  generic/          persona-gentleman.md, persona-neutral.md, sdd-orchestrator.md
  skills/           22 skill subdirectories + _shared/
  codex/            (present)
  qwen/             (present)
```

### Injector (`internal/steps/inject_assets.go`)
- Has its own `//go:embed assets/*` → separate embed.FS from `templates.FS`
- `Inject(targetDir, plan)` walks the ENTIRE `assets/` tree flat into `targetDir`
- `plan planner.ResolvedPlan` parameter is accepted but **never used**
- Result: everything (claude/, antigravity/, skills/, gga/, kimi/…) gets dumped into `.cursor/skills/`, `.codex/skills/`, etc.

### IDE adapters (`internal/system/ide.go`)
- Only 3 adapters: `CursorAdapter`, `WindsurfAdapter`, `CodexAdapter`
- `GetAdapters()` returns only these 3 — completely disconnected from the 15-agent catalog
- `DetectInstalledIDEs()` calls `ScanConfigs()` (which knows all 15 agents) but then only loops over `GetAdapters()` — so even if Claude Code is detected, it's silently dropped
- `IDEAdapter` interface: `Name()`, `ConfigDir()`, `GlobalRulesDir()`, `GlobalSkillsDir()`, `LocalRulesFile()`, `LocalSkillsDir()` — **no `AssetFolder()` method**

### Install steps (`internal/steps/install.go`)
- `StepInstallGlobalRules.Run()`: reads `templates.MustRead("base/persona.md")` — the generic fallback — for **every** IDE, ignoring per-agent persona files in `assets/`
- `StepSetupLocalRules.Run()`: same — `templates.MustRead("base/persona.md")` and copies `base/skills/` (generic) regardless of agent

### TUI model (`internal/tui/model.go` L136-L144)
- `StartPipelineMsg{Action: "Install"}` handler hardcodes:
  ```go
  Agents: []model.AgentID{"claude", "gpt", "gemini", "cursor", "copilot"}
  ```
  These are **invalid AgentIDs** — the real constants are `"claude-code"`, `"gemini-cli"`, etc.
- `resolvedPlan` is passed to `NewStepInjectAssets()` but the injector ignores it
- `StepScanGlobalIDEs` runs and populates `ctx.IDEs`, but the scan only returns from `GetAdapters()` (3 adapters), ignoring what was actually detected

### TUI install screen (`internal/tui/screens/install.go`)
- No agent selection step — pressing Enter immediately fires `StartPipelineMsg{Action: "Install"}`
- `InstallModel` has no field for user's agent selection

### Catalog (`internal/catalog/agents.go`)
- Correctly defines 15 agents with proper IDs, names, tiers, config paths
- `AllAgents()` / `MVPAgents()` functions ready
- **Not used** by the IDE adapter layer at all

### Model types (`internal/model/`)
- All 15 `AgentID` constants are correctly defined (`"claude-code"`, `"antigravity"`, etc.)
- `Selection.Agents []AgentID` and `ResolvedPlan.Agents []AgentID` exist and are correctly structured

---

## Affected Areas

| File | Why Affected |
|------|-------------|
| `internal/system/ide.go` | Missing 12 adapters; must add all agents; must add `AgentID() model.AgentID` method to bridge to catalog |
| `internal/steps/inject_assets.go` | Must use `plan.Agents` to filter which asset subfolders to inject, and per-agent target dir |
| `internal/steps/install.go` | `StepInstallGlobalRules` must read per-agent persona file from `assets/{agent}/` not `base/persona.md` |
| `internal/tui/model.go` | Must use real AgentIDs from user selection (not hardcoded dummies); must wire `SelectedAgents` from TUI into `Selection.Agents` |
| `internal/tui/screens/install.go` | Must add agent selection step before firing pipeline |
| `internal/tui/screens/` | New `AgentSelectModel` screen or integrated multi-select within install screen |

---

## Approaches

### Approach 1 — Add `AgentID()` to IDEAdapter + filter injector by plan (Recommended)

Extend `IDEAdapter` interface with one new method:
```go
AgentID() model.AgentID   // maps adapter → catalog ID for asset lookup
AssetFolder() string      // maps adapter → assets subfolder name (e.g. "cursor")
```

**Injector** becomes agent-aware: instead of walking all of `assets/`, it walks `assets/{adapter.AssetFolder()}/` for each IDE in `ctx.IDEs`, and also injects `assets/skills/` (shared, once).

**`StepInstallGlobalRules`** reads `assets/{agent}/persona-gentleman.md` (fallback to `assets/generic/persona-gentleman.md`) instead of `base/persona.md`.

**TUI** gets a multi-select agent list screen before install, populating `model.Selection.Agents` with real catalog IDs. `model.go` uses `catalog.AllAgents()` to build the selection and passes it to `resolver.Resolve()`. The resolved agents drive which adapters are injected.

- **Pros**: Clean separation of concerns; adapter knows its identity; injector stays simple; catalog stays the source of truth; works with existing IDEAdapter pattern
- **Cons**: Interface change breaks existing mock/test implementations of IDEAdapter (3 places); need to add 12 new adapters
- **Effort**: Medium

---

### Approach 2 — Keep IDEAdapter as-is; add an external `AgentID → assetFolder` map

Add a `map[model.AgentID]string` in `inject_assets.go` that maps each known agent to its asset subfolder. The injector receives a list of `(agentID, targetDir)` pairs instead of using `IDEAdapter` directly.

- **Pros**: Zero interface changes; isolated change
- **Cons**: Duplicates the adapter ↔ catalog mapping in two places; `IDEAdapter` still doesn't know its own identity; doesn't fix the `GetAdapters()` gap
- **Effort**: Low initially, high maintenance debt

---

### Approach 3 — Replace IDEAdapter with catalog.Agent as the central type

Merge `IDEAdapter` into `catalog.Agent`, adding path methods directly to the catalog struct. `GetAdapters()` becomes `catalog.AllAgents()`.

- **Pros**: Single source of truth
- **Cons**: Breaking refactor across all pipeline steps, sync, upgrade, uninstall; very high blast radius for this change
- **Effort**: High

---

## Comparison Table

| Criterion | Approach 1 (Recommended) | Approach 2 | Approach 3 |
|-----------|--------------------------|------------|------------|
| Fixes adapter gap (12 missing) | ✅ | ❌ (partial) | ✅ |
| Injector uses plan | ✅ | ✅ | ✅ |
| Per-agent persona | ✅ | ✅ | ✅ |
| TUI agent selection | ✅ | ✅ | ✅ |
| Interface breakage | Minimal (3 mocks) | None | High |
| Maintenance debt | Low | High | Low |
| Effort | Medium | Low | High |

---

## Recommendation

**Approach 1**. Adding `AgentID() model.AgentID` and `AssetFolder() string` to `IDEAdapter` is the minimal surgical change that:

1. Lets the injector walk `assets/{agent}/` instead of all of `assets/`
2. Lets `StepInstallGlobalRules` look up `assets/{agent}/persona-gentleman.md` with `assets/generic/` fallback
3. Bridges the adapter ↔ catalog gap without a full merge
4. Enables `DetectInstalledIDEs()` to return adapters for all 15 agents (not just 3)

The 12 missing adapters should be added to `ide.go` using the config paths already enumerated in `config_scan.go#knownAgentConfigDirs()` as the source of truth.

For the **TUI**, agent selection should be a **new screen** (`AgentSelectModel`) inserted between the welcome confirmation and the pipeline launch — not crammed into `install.go`. This keeps each screen's responsibility atomic and follows the existing pattern (each action has its own `*Model`).

`ResolvedPlan` SHOULD drive injection. The TUI selects agents → resolver validates them → `ctx.IDEs` is populated from the resolved agent list (not just filesystem detection) → injector walks only the relevant asset subfolders.

---

## Key Implementation Notes

### `IDEAdapter` interface additions
```go
AgentID() model.AgentID  // e.g. model.AgentCursor
AssetFolder() string     // e.g. "cursor" — matches assets/ subfolder name
```

### Injector rewrite sketch
```go
func (a *assetInjector) Inject(targetDir string, agentFolder string, plan planner.ResolvedPlan) error {
    // 1. Inject agent-specific assets: assets/{agentFolder}/
    // 2. Inject shared skills: assets/skills/ → targetDir/skills/
}
```
Called per-adapter from `StepInjectAssets.Run()`.

### Persona resolution order
1. `assets/{agent}/persona-gentleman.md`
2. `assets/generic/persona-gentleman.md`
3. `assets/generic/persona-neutral.md` (last fallback)

### TUI flow (new)
```
WelcomeScreen → [Install] → AgentSelectScreen → [Enter/Confirm] → InstallRunScreen
```
`AgentSelectScreen` uses a checkbox list built from `catalog.AllAgents()`. On confirm it emits `AgentsSelectedMsg{Agents: []model.AgentID{...}}` which `model.go` captures before firing `StartPipelineMsg`.

### AgentID ↔ AssetFolder mapping
The asset subfolder names already match the catalog IDs with minor normalization:
- `"claude-code"` → `"claude"` (only exception)
- `"gemini-cli"` → `"gemini"`
- `"kiro-ide"` → `"kiro"`
- `"qwen-code"` → `"qwen"`
- `"antigravity"` → `"antigravity"`
- All others: strip suffix, use first segment

This should be encoded in each adapter's `AssetFolder()` method, not derived generically.

---

## Risks

1. **`//go:embed assets/*` in `inject_assets.go` vs `//go:embed all:base` in `templates/embed.go`** — two separate embedded FSes. The injector must stay on its own embed or be refactored to use `templates.FS`. Currently they're separate; confirm no double-embedding before splitting.
2. **`GetAdapters()` is called by `DetectInstalledIDEs()` AND `GetAdapterByName()`** — adding 12 adapters means detection and lookup both grow. The fallback path (`if len(installed) == 0 → return GetAdapters()`) could return 15 adapters to a user with none installed; this may be the desired behavior, but confirm.
3. **`StepSetupLocalRules`** also hardcodes `base/persona.md` and `base/skills/` — same fix applies to local setup, not just global install.
4. **Missing asset subfolders for some agents** — e.g. `assets/windsurf/` may need to be created if it only has config-based injection (verify before implementing).
5. **`skills/` is a shared resource** — injecting it once into each agent's `GlobalSkillsDir()` is correct, but some agents (Antigravity, Gemini CLI) may have platform-specific skills dirs that don't follow `~/{agent}/skills/`. Verify per-agent paths.
6. **Interface change tests** — `IDEAdapter` is tested in `ide.go` and potentially mocked in step tests. Adding methods will require updating test doubles.

---

## Ready for Proposal

**Yes.** The exploration surfaces a clear recommended path with well-scoped changes across 6 files. The implementation is bounded: no new packages needed, no external dependencies. The proposal phase should define the exact interface contract for `IDEAdapter`, the 12 new adapter structs, the injector rewrite, and the `AgentSelectModel` TUI screen spec.
