# Proposal: Fix Injection Pipeline

## Intent
The asset injection pipeline is broken end-to-end: the injector dumps all agent assets into every IDE, adapters for 12 of 15 agents are missing, personas are always generic, and the TUI never wires user selections into the plan. The `assets/` layout is already correct — only the consuming layers are broken.

## Scope

### In Scope
- Extend `IDEAdapter` interface: add `AgentID() model.AgentID` and `AssetFolder() string`
- Add 12 missing adapter structs to `internal/system/ide.go`
- Rewrite `inject_assets.go` to inject `assets/{agent}/` per-adapter + `assets/skills/` once (shared)
- Fix `steps/install.go` persona lookup: per-agent → generic fallback
- Add `AgentSelectModel` TUI screen between welcome and install pipeline
- Wire `AgentsSelectedMsg` → `model.go` → `ResolvedPlan.Agents` → injector

### Out of Scope
- `StepSetupLocalRules` persona fix (tracked as follow-up risk)
- Creating missing `assets/windsurf/` or other empty agent folders
- Merging the two embed.FS (`inject_assets.go` vs `templates/embed.go`)
- E2E TUI tests (none exist; model-level tests only)
- Any changes to the catalog, planner, or model packages
- MCP config injection (separate change: `mcp-config-injection`)

## Capabilities

### New Capabilities
- `agent-aware-injection`: Per-agent, path-aware asset injection using `IDEAdapter.AssetFolder()`; `assets/skills/` injected once (shared), not once per agent
- `agent-selection-tui`: New `AgentSelectModel` TUI screen; checkbox list built from `catalog.AllAgents()`; emits `AgentsSelectedMsg` consumed by `model.go` before pipeline fires
- `engram-to-sdd-memory-rename`: All references to "engram" / "Engram" in `assets/` replaced with "sdd-memory"; 4 files renamed (`engram-*.md` → `sdd-memory-*.md`); section tags updated accordingly

### Modified Capabilities
- `ide-adapter`: Extend `IDEAdapter` interface with `AgentID() model.AgentID` and `AssetFolder() string`; add 12 missing adapter structs; `GetAdapters()` returns all 15 agents
- `persona-injection`: `StepInstallGlobalRules` resolves persona as: `assets/{agent}/persona-gentleman.md` → `assets/generic/persona-gentleman.md` → `assets/generic/persona-neutral.md`

## Approach
Surgical extension of the existing `IDEAdapter` pattern. Each adapter gains identity (`AgentID`) and asset path (`AssetFolder`). The injector becomes a simple loop: for each adapter in `ctx.IDEs`, copy `assets/{folder}/` to that adapter's `GlobalSkillsDir()`; then copy `assets/skills/` once. The TUI gets one new screen (`AgentSelectModel`) inserted between welcome and the pipeline launch — same pattern as existing screens.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/system/ide.go` | Modified | Add `AgentID()` + `AssetFolder()` to interface; add 12 adapter structs; update `GetAdapters()` |
| `internal/steps/inject_assets.go` | Modified | Rewrite to walk `assets/{assetFolder}/` per adapter; inject `assets/skills/` once |
| `internal/steps/install.go` | Modified | Per-agent persona resolution with generic fallback |
| `internal/tui/model.go` | Modified | Handle `AgentsSelectedMsg`; replace hardcoded agent IDs; wire selection to `ResolvedPlan` |
| `internal/tui/screens/install.go` | Modified | Add agent selection step before firing `StartPipelineMsg` |
| `internal/tui/screens/agent_select.go` | New | `AgentSelectModel` screen with multi-select checkbox list |
| `assets/claude/engram-protocol.md` | Renamed | → `assets/claude/sdd-memory-protocol.md`; all "engram"/"Engram" text replaced with "sdd-memory" |
| `assets/codex/engram-compact-prompt.md` | Renamed | → `assets/codex/sdd-memory-compact-prompt.md` |
| `assets/codex/engram-instructions.md` | Renamed | → `assets/codex/sdd-memory-instructions.md` |
| `assets/skills/_shared/engram-convention.md` | Renamed | → `assets/skills/_shared/sdd-memory-convention.md` |
| All 25 files in `assets/` containing "engram"/"Engram" | Modified | Text replacement: `engram` → `sdd-memory`, `Engram` → `sdd-memory` |

## Asset Folder Mapping (encoded in each adapter's `AssetFolder()`)

| Adapter | `AgentID()` | `AssetFolder()` |
|---------|-------------|-----------------|
| ClaudeCode | `claude-code` | `claude` |
| GeminiCLI | `gemini-cli` | `gemini` |
| Antigravity | `antigravity` | `antigravity` |
| OpenCode | `opencode` | `opencode` |
| KiroIDE | `kiro-ide` | `kiro` |
| KimiCode | `kimi-code` | `kimi` |
| QwenCode | `qwen-code` | `qwen` |
| Kilo | `kilo` | `kilo` |
| OpenClaw | `openclaw` | `generic` |
| Pi | `pi` | `generic` |
| Trae | `trae` | `generic` |
| VSCodeCopilot | `vscode-copilot` | `generic` (no dedicated assets) |
| Cursor _(existing)_ | `cursor` | `cursor` |
| Windsurf _(existing)_ | `windsurf` | `windsurf` |
| Codex _(existing)_ | `codex` | `codex` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Interface change breaks 3 test mocks | High | Update all mocks as part of this change (tracked as success criterion) |
| `GetAdapters()` fallback returns all 15 when none detected | Med | Preserve current behavior; document explicitly in code comment |
| Some agent asset folders missing (e.g. `assets/windsurf/`) | Med | Injector skips gracefully if folder not found in embed.FS; verify before spec |
| `StepSetupLocalRules` still uses `base/persona.md` | Low | Tracked separately; not a regression since it's broken today |
| Dual embed.FS collision | Low | Keep separate; do not merge `inject_assets.go` embed with `templates/embed.go` |

## Rollback Plan
All changes are isolated to 6 files. Revert via `git revert` of the single commit (or branch). No schema/data migrations. Adapter additions to `GetAdapters()` can be feature-flagged per adapter if partial rollback is needed.

## Dependencies
- `catalog.AllAgents()` must remain stable (no changes to catalog package)
- `assets/` folder structure must not be reorganized during this change
- `config_scan.go#knownAgentConfigDirs` is the source of truth for adapter config paths

## Success Criteria
- [ ] `IDEAdapter` interface has `AgentID()` and `AssetFolder()` methods
- [ ] `GetAdapters()` returns all 15 adapters
- [ ] All 3 existing test mocks updated to implement new interface methods
- [ ] `inject_assets.go` injects only `assets/{agentFolder}/` per adapter (not all of `assets/`)
- [ ] `assets/skills/` is injected exactly once per install (not once per agent)
- [ ] `StepInstallGlobalRules` uses per-agent persona file with generic fallback
- [ ] `AgentSelectModel` TUI screen renders agent list from `catalog.AllAgents()`
- [ ] `model.go` replaces hardcoded agent IDs with selection from `AgentsSelectedMsg`
- [ ] `ResolvedPlan.Agents` drives which adapters are included in `ctx.IDEs`
- [ ] Install runs end-to-end with at least 2 agents selected in TUI without panicking
- [ ] Zero occurrences of "engram" or "Engram" remain in `assets/` (verified by grep)
- [ ] All 4 files with `engram-` prefix renamed to `sdd-memory-` prefix
- [ ] Any internal references to renamed files updated (e.g. Jinja `include` paths, `assets.MustRead()` calls)
