## Proposal: embed-alignment

### The Problem
The asset injector currently reads from `internal/steps/assets/` via `//go:embed assets/*`, missing all per-agent assets stored at the project root `assets/`. Also, `install.go` falls back to reading from disk because it cannot find personas in the embedded file system.

### Proposed Architecture
Mimic `gentle-ai`'s `internal/assets` pattern.
1. **Consolidation**: Move the root `assets/` directory to `internal/assets/`. Move `internal/steps/assets/skills` into `internal/assets/skills`.
2. **Centralization**: Create `internal/assets/assets.go` exposing `var FS embed.FS` via `//go:embed all:claude all:generic all:skills ...`
3. **Consumption**:
   - Update `internal/steps/inject_assets.go` to use `assets.FS` instead of defining its own local `embed.FS`.
   - Update `internal/tui/screens/install.go` to use `assets.FS` to read personas instead of checking the disk directly.
4. **Maintenance**: Update `scripts/extract_assets.go` to output to `internal/assets/` instead of `assets/`.

### Migration Strategy
1. **Refactor Phase:** Create the `internal/assets` package, move folders from root and `internal/steps/`, and create `assets.go`.
2. **Update Injection Phase:** Change `inject_assets.go` and `install.go` to depend on `assets.FS`.
3. **Update Scripts & Tests:** Adjust `extract_assets.go` and tests (`install_test.go`, `inject_assets_test.go`) to point to the new paths or use the centralized FS.

### Rollback Plan
Since this is a pure structural refactor and Go package rename/move, a rollback consists of simply reverting the commit or Git tree. No external database or state is modified.

### Risks & Mitigations
- Breakage of the `extract_assets` script. Mitigation: Run the script locally after changes to ensure it writes to the correct location.
- Breakage in how `inject_assets.go` computes relative paths. Mitigation: Make sure `inject_assets.go` properly trims prefixes according to the new `embed.FS` root.
