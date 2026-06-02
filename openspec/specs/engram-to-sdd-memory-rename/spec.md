# engram-to-sdd-memory-rename Specification

## Purpose
Remove all references to "engram" / "Engram" from the `assets/` tree: rename the four files that carry an `engram-` prefix and replace all text occurrences across affected files.

## Requirements

### Requirement: Four Files Renamed
The system MUST rename the following files; no other files are renamed as part of this change:

| From | To |
|------|----|
| `assets/claude/engram-protocol.md` | `assets/claude/sdd-memory-protocol.md` |
| `assets/codex/engram-compact-prompt.md` | `assets/codex/sdd-memory-compact-prompt.md` |
| `assets/codex/engram-instructions.md` | `assets/codex/sdd-memory-instructions.md` |
| `assets/skills/_shared/engram-convention.md` | `assets/skills/_shared/sdd-memory-convention.md` |

#### Scenario: Renamed files accessible under new paths
- GIVEN the rename has been applied
- WHEN the application reads `assets/claude/sdd-memory-protocol.md`
- THEN the file MUST exist and contain the expected content
- AND `assets/claude/engram-protocol.md` MUST NOT exist

### Requirement: Text Replacement Across All Affected Files
The system MUST replace every occurrence of `"engram"` with `"sdd-memory"` and `"Engram"` with `"sdd-memory"` in the 25 files within `assets/` that contain those strings.

#### Scenario: No residual engram references in assets
- GIVEN all replacements have been applied
- WHEN `grep -r "engram" assets/` is executed
- THEN the command MUST return zero matches

### Requirement: Section Tags Updated
HTML comment tags of the form `<!-- gentle-ai:engram-* -->` MUST be updated to `<!-- gentle-ai:sdd-memory-* -->` in all affected asset files.

#### Scenario: Tag renamed in protocol file
- GIVEN the rename and text replacement has been applied
- WHEN the content of `assets/claude/sdd-memory-protocol.md` is inspected
- THEN the opening tag MUST be `<!-- gentle-ai:sdd-memory-protocol -->`
- AND the closing tag MUST be `<!-- /gentle-ai:sdd-memory-protocol -->`

### Requirement: Internal References Updated
Any call-site that reads a renamed file by path (e.g. `assets.MustRead("claude/engram-protocol.md")`) MUST be updated to the new path.

#### Scenario: MustRead call uses new path
- GIVEN the rename has been applied
- WHEN `assets.MustRead("claude/sdd-memory-protocol.md")` is called at runtime
- THEN the call MUST succeed and return the file contents
- AND any call with the old path `"claude/engram-protocol.md"` MUST NOT exist in source code
