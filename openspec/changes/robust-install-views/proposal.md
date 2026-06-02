# Proposal: robust-install-views

## Intent
The SpecAI TUI currently uses a static 3-state install flow that ignores the underlying planner and blindly injects a limited set of 3 skills. We need a robust, interactive TUI driven by modular component steps that accurately reflects pipeline execution, along with a comprehensive backport of all 21+ skills and configurations from the `gentle-ai` project.

## Scope

### In Scope
- Extract and literal copy/paste of ALL assets from `gentle-ai` (all 21+ skills, persona configs, and MCP configs like context7, engram, sdd-memory).
- Implement modular component steps for the install pipeline (Strict TDD mode, Install Plan, Review, Confirm).
- Refactor TUI to be interactive and dynamically report progress from pipeline execution.
- Wire the TUI views to the existing `planner` logic instead of bypassing it.

### Out of Scope
- Modifications to the core `gentle-ai` skills or configs (they must be exact copies).
- Changing the core logic of the `planner` outside of exposing progress hooks.

## Capabilities

### New Capabilities
- `tui-pipeline-integration`: Wiring the interactive TUI to report actual pipeline execution progress.
- `asset-injection-engine`: The system responsible for exactly copying and injecting all 21+ `gentle-ai` assets.

### Modified Capabilities
- None

## Approach
Replace the hardcoded 3-state TUI steps with dynamic view components driven by the `planner` pipeline. Create an asset extractor that pulls skills, personas, and MCP configurations from `C:\Users\kevin\dev\gentle-ai` and injects them precisely during the setup phase. Utilize Go channels or callbacks to feed execution progress directly to Bubble Tea views.

## Affected Areas
| Area | Impact | Description |
|------|--------|-------------|
| `pkg/tui/` | Modified | Replace static states with dynamic pipeline views |
| `pkg/planner/` | Modified | Add progress reporting mechanisms to execution |
| `assets/` | New | Complete directory of 21+ gentle-ai skills and configs |

## Risks
| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Asset path mismatches from gentle-ai | Med | Use absolute source extraction or exact bundled embedding. |
| TUI lag during synchronous pipeline work | Low | Run planner asynchronously and update UI via standard message events. |

## Rollback Plan
Revert the `pkg/tui` commit and restore the static 3-state implementation. Drop the new `assets` directory containing the gentle-ai payload.

## Dependencies
- Local access to `C:\Users\kevin\dev\gentle-ai` or pre-bundled assets.

## Success Criteria
- [ ] The TUI displays real-time progress for all pipeline steps (Strict TDD, Install Plan, etc.).
- [ ] All 21+ skills, personas, and MCP configurations are successfully injected exactly as they exist in `gentle-ai`.
- [ ] The TUI uses modular views rather than the static 3-state loop.
