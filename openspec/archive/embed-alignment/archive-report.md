# SDD Archive Report: embed-alignment

## Overview
This change successfully aligned the asset embedding architecture in `SpecAI` with the `gentle-ai` pattern. A centralized `internal/assets` package was created to house and serve all static files (agent personas, generic assets, skills), resolving issues where certain embedded files were missing during the asset injection phase and removing the need for disk fallbacks in the TUI installation screens.

## Accomplished
- Consolidated previously scattered asset directories:
  - Moved root `assets/` to `internal/assets/`.
  - Moved `internal/steps/assets/` to `internal/assets/`.
- Centralized `embed.FS` creation:
  - Created `internal/assets/assets.go` exposing the embedded filesystem as `assets.FS`, along with `Read` and `MustRead` helper functions.
- Updated consumers to use the centralized `assets.FS`:
  - Refactored `internal/steps/inject_assets.go` to use the global `assets.FS` instead of managing its own local embed directive.
  - Updated `internal/tui/screens/install.go` to read personas directly from `assets.Read()`, eliminating previous hardcoded paths and disk read fallbacks.
- Updated maintenance scripts:
  - Directed `scripts/extract_assets.go` to output the bundled assets to the new `internal/assets/` destination.
- Updated related unit tests (`install_test.go`, `inject_assets_test.go`) to accommodate the new structural changes, ensuring all tests passed.

## Architecture & Decisions
- **Mimic `gentle-ai` architecture**: By centralizing all assets into `internal/assets` and explicitly embedding all subdirectories via `//go:embed all:claude all:generic all:skills`, we establish a single source of truth for assets. This pattern ensures self-contained binaries, reduces code duplication, and enables cleaner access across the entire application.

## Relevant Files
- `internal/assets/assets.go` - The new centralized asset embed package
- `internal/steps/inject_assets.go` - Updated to use `assets.FS` for payload injection
- `internal/tui/screens/install.go` - Updated to use `assets.Read`
- `scripts/extract_assets.go` - Updated output destination

## Review
All defined tasks were completed, and tests run successfully. The change was developed as a single PR due to the small scope and lack of architectural risk, effectively resolving the missing embedded assets bug.
