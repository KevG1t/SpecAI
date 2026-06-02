# Specs: embed-alignment

## Behavior changes
- The application will now read all assets (personas, commands, skills) from a single centralized `internal/assets` package.
- `AssetInjector.InjectAgentFolder` will correctly copy agent-specific folders (`claude`, `gemini`, etc.) to the target installation directory, instead of gracefully skipping due to missing embedded files.
- `AssetInjector.InjectSharedSkills` will continue to correctly copy shared skills to the target installation directory.
- `internal/tui/screens/install.go` will retrieve persona contents by reading them from the centralized embedded FS rather than falling back to disk reads.
- `scripts/extract_assets.go` will output the fetched assets directly into `internal/assets/` to ensure the embedded FS is kept up-to-date upon build.

## State changes
- No persistent user data state changes.
- **Filesystem refactor**: `assets/` at the project root and `internal/steps/assets/` will be moved and merged into a new `internal/assets/` directory.

## In Scope
- Creating `internal/assets/assets.go` with the correct `//go:embed` directive mimicking `gentle-ai`.
- Moving all asset directories.
- Updating all consumers of assets (`inject_assets.go`, `install.go`).
- Updating the `extract_assets.go` script and any broken tests.

## Out of Scope
- Modifying the actual contents of the personas or skills.
- Altering the installation logic (the planner steps remain the same, just the data source changes).
