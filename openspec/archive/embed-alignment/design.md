# Design: embed-alignment

## Architecture
We are adopting the pattern used in `gentle-ai` for asset management: a dedicated `internal/assets` package providing a single global `embed.FS` instance. This acts as the single source of truth for all bundled static files.

### `internal/assets` package
- Responsible for holding the raw asset directories (`claude`, `gemini`, `skills`, `generic`, etc.).
- Exposes `var FS embed.FS` via a `//go:embed` directive that explicitly lists all required top-level directories (e.g., `//go:embed all:claude all:generic all:skills ...`).
- Exposes helper functions `MustRead(path string)` and `Read(path string) (string, error)` for convenient string access, identical to `gentle-ai`.

### `internal/steps/inject_assets.go`
- The `assetInjector` implementation will be updated to remove its local `//go:embed` directive.
- It will reference `assets.FS` instead.
- Path resolution logic within `walkAndCopy` will be adjusted. Previously it expected a root `assets/` folder in the embedded FS. Now, depending on the `go:embed` directive, the folders (`claude`, `skills`) will be at the root of the embedded FS.

### `internal/tui/screens/install.go`
- Functions like `renderPersonaOptions` or anything that previews assets will be updated to read from `assets.FS` directly using the helper functions.
- The fallback logic trying to read from the local disk `assets/` directory will be removed, ensuring the binary is truly self-contained.

### `scripts/extract_assets.go`
- The target output path will be updated from `assets/` to `internal/assets/`.

## Data Structures
- No new complex data structures.

## Error Handling
- Errors reading from the `embed.FS` during injection will continue to be returned to the planner.
- If a specific agent folder doesn't exist in the embedded FS, `InjectAgentFolder` will gracefully skip, returning `nil` (this logic is already present but will now actually work for existing folders).
