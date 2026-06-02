## Exploration: embed-alignment

### Current State
SpecAI currently stores agent-specific assets (`claude`, `gemini`, etc.) at the project root `assets/`, while SDD skills are in `internal/steps/assets/`. The asset injector (`internal/steps/inject_assets.go`) uses `//go:embed assets/*` which only captures `internal/steps/assets/`. As a result, agent folders are silently skipped during injection. Also `internal/tui/screens/install.go` falls back to disk reads for personas because it can't find them in the `embed.FS`.

### Affected Areas
- `assets/` (root) — Currently holds agent-specific files. Needs to be moved.
- `internal/steps/assets/` — Currently holds skills. Needs to be merged/moved.
- `internal/assets/` — New package location mimicking `gentle-ai`.
- `internal/steps/inject_assets.go` — References the embedded FS and needs to point to the new package.
- `internal/tui/screens/install.go` — Hardcodes paths and FS logic for personas.

### Approaches
1. **Mimic gentle-ai Architecture** — Move everything into `internal/assets/` and create `assets.go` that exports `var FS embed.FS`. Use `//go:embed all:claude all:generic all:skills ...`
   - Pros: Aligns perfectly with `gentle-ai` (as requested). Single source of truth. Easy to consume from any package.
   - Cons: Moves a lot of files, touching `extract_assets.go` script possibly.
   - Effort: Medium

2. **Move root assets into internal/steps/assets** — Just push the root `assets/` into `internal/steps/assets/`.
   - Pros: Less code change, keeps the current embed directive.
   - Cons: Not aligned with `gentle-ai`'s cleaner dedicated `assets` package.
   - Effort: Low

3. **Change the embed path** — Keep `assets/` at root, add an embed file at root or package level.
   - Pros: No file moving.
   - Cons: Clutters root, diverges from reference architecture.
   - Effort: Low

### Recommendation
**Mimic gentle-ai Architecture**. The user explicitly requested to copy the architecture, folder structure, and logic from `gentle-ai`. This means creating an `internal/assets` package, exporting an `embed.FS`, and moving all assets (from both root `assets/` and `internal/steps/assets/`) into it.

### Risks
- Moving the root `assets/` directory might break `scripts/extract_assets.go` which might be hardcoded to output there. We'll need to update it.
- Breaking tests in `install_test.go` and `inject_assets_test.go` that rely on the old filesystem structure.

### Ready for Proposal
Yes.
