## Exploration: robust-install-views

### Current State
**SpecAI** currently has an `internal/planner` package containing the logic to resolve dependencies and build a `ReviewPayload` (handling `Persona`, `Preset`, `Components`, `Skills`, `StrictTDD`, `HasSDD`). However, the TUI bypasses this entirely. The "Install" flow jumps straight to a hardcoded 3-state screen (`Confirm`, `Running`, `Result`) with a basic spinner. 
The installation logic (`internal/steps/install.go`) blindly copies a single `base/persona.md` and a `base/skills` directory that only contains 3 skills (`sdd-apply`, `sdd-explore`, `sdd-verify`). It does not inject the other 18 MVP skills, nor does it support MCPs like `Context7` or `Engram/sdd-memory`.

**Gentle-AI**, by contrast, wires its planner to interactive BubbleTea screens:
- A `review.go` screen showing exactly what will be installed (Selected Agents, Persona, Preset, Components with auto-dependency badges, individual Skills, Strict TDD mode).
- An `installing.go` screen showing step-by-step progress, a percentage bar, item statuses (running, succeeded, failed), and tailing logs.
- Modular installation logic where each component (like `Context7`, `Engram`) has dedicated steps and templates.

### Affected Areas
- `internal/tui/screens/review.go` (new) — Needs to be created based on the `gentle-ai` review screen, consuming `planner.ReviewPayload`.
- `internal/tui/screens/install.go` — Needs to be refactored to an `installing` model that renders progress, steps, and logs (like `gentle-ai`).
- `internal/tui/router.go` / `tui.go` — Update routing to flow: `Welcome` -> `Selection` (if needed) -> `Review` -> `Installing` -> `Result`.
- `internal/steps/install.go` — Must be broken down into modular steps (e.g., `StepInjectSkills`, `StepInjectContext7`, `StepInjectEngram`) that read from `planner.ResolvedPlan`.
- `internal/templates/base/skills/` — Missing 18 MVP skills need their templates added to match `catalog.MVPSkills()`.
- `internal/templates/base/mcps/` (new) — Needs `context7` and `engram` configuration templates.

### Approaches
1. **Modular Component Steps & Interactive TUI (Recommended)**
   - **Description**: Wire the existing `planner` into the TUI. Introduce `screens/review.go` and refactor `screens/install.go` to support a `Progress` state with step tracking and logs. Break `steps/install.go` into discrete structs that only execute if their component/skill is in the `ResolvedPlan`. Add all 21 missing skills and MCP templates.
   - **Pros**: Matches `gentle-ai`'s UX exactly; scalable for future components; leverages existing `planner` logic; allows fine-grained error tracking.
   - **Cons**: High refactoring effort, requires adapting the Pipeline runner to emit progress and log tick messages to the TUI.
   - **Effort**: High

2. **Monolithic Plan Execution (Static TUI)**
   - **Description**: Add `screens/review.go` before installation. Keep the current `Running` spinner in `install.go`. Update `StepInstallGlobalRules` and `StepInstallGlobalSkills` to read the `ResolvedPlan` and conditionally copy files.
   - **Pros**: Low effort, minimal changes to the `Pipeline` engine.
   - **Cons**: Fails the requirement of "install views are not interactive like gentle-ai's". No progress bar, no step statuses.
   - **Effort**: Low

### Recommendation
**Approach 1 (Modular Component Steps & Interactive TUI)**. 
Since the primary goal is to replicate the rich, interactive install views of `gentle-ai` (Strict TDD mode, Install Plan, Review and Confirm, Progress tracking) and ensure all supported components are injected, we must adopt the modular architecture. The `planner` package in SpecAI is already capable of producing the necessary payloads; it just needs to be wired into BubbleTea and the underlying pipeline engine needs to emit progress updates.

### Risks
- **Pipeline Engine Rewrite**: SpecAI's current pipeline execution might not support streaming logs and granular step progress back to the BubbleTea model. The engine may need an interface update to send `TickMsg` or `ProgressMsg`.
- **Template Completeness**: We must ensure we accurately backport the 18 missing skills and the MCP configs (`Context7`, `Engram`) from `gentle-ai` without hallucinating components that don't exist in SpecAI's catalog.

### Ready for Proposal
Yes. The orchestrator can proceed to the `sdd-propose` phase for Approach 1, as the structural differences and required changes are fully mapped out.
