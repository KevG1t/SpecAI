# agent-aware-injection Specification

## Purpose
Per-agent, path-aware asset injection: each IDE adapter receives only its own agent-specific assets; shared skills are injected exactly once across all adapters.

## Requirements

### Requirement: Per-Agent Asset Injection
The system MUST inject only `assets/{adapter.AssetFolder()}/` into each adapter's `GlobalSkillsDir()`, not the entire `assets/` tree.

#### Scenario: Agent-specific assets delivered
- GIVEN two adapters with different `AssetFolder()` values (e.g. `claude`, `gemini`)
- WHEN the inject step runs
- THEN each adapter's `GlobalSkillsDir()` receives only the files from its own `assets/{folder}/` directory
- AND files from the other agent's folder are NOT present in either target directory

#### Scenario: Missing agent folder skipped gracefully
- GIVEN an adapter whose `AssetFolder()` maps to a folder that does not exist in the embedded FS
- WHEN the inject step runs
- THEN the system MUST skip that folder without error
- AND injection for all other adapters MUST complete successfully

### Requirement: Shared Skills Injected Once
The system MUST inject `assets/skills/` into a shared location exactly once per install, regardless of how many adapters are present.

#### Scenario: Skills not duplicated across agents
- GIVEN three adapters are selected
- WHEN the inject step runs
- THEN `assets/skills/` content is written to `GlobalSkillsDir()` exactly one time (not three times)

#### Scenario: No adapters selected
- GIVEN zero adapters are in the resolved plan
- WHEN the inject step runs
- THEN `assets/skills/` is still injected once and the step completes without error

### Requirement: Plan-Driven Adapter Scope
The system MUST use only the adapters listed in `ResolvedPlan.Agents` (via `ctx.IDEs`) when determining which per-agent folders to inject.

#### Scenario: Unselected agent receives no assets
- GIVEN a user selected only `claude-code` in the TUI
- WHEN the inject step runs
- THEN assets for `gemini`, `codex`, or any other unselected agent are NOT injected into any directory
