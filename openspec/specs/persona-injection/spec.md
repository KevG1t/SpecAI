# Spec: persona-injection

## Requirements

### Requirement: Per-Agent Persona Resolution in StepInstallGlobalRules
The system MUST resolve the persona file for each adapter using the following priority chain:
1. `assets/{adapter.AssetFolder()}/persona-gentleman.md` — agent-specific formal persona
2. `assets/generic/persona-gentleman.md` — shared formal persona fallback
3. `assets/generic/persona-neutral.md` — neutral fallback of last resort

The system MUST use the first file that exists in the embedded FS. If none of the three paths exist, the step MUST return an error.

#### Scenario: Agent has dedicated persona file
- GIVEN the `claude` asset folder contains `persona-gentleman.md`
- WHEN `StepInstallGlobalRules` runs for the `ClaudeCodeAdapter`
- THEN the file `assets/claude/persona-gentleman.md` MUST be used as the persona
- AND neither of the two fallback paths MUST be read

#### Scenario: Agent falls back to generic formal persona
- GIVEN the `gemini` asset folder does NOT contain `persona-gentleman.md`
- AND `assets/generic/persona-gentleman.md` exists
- WHEN `StepInstallGlobalRules` runs for the `GeminiCLIAdapter`
- THEN `assets/generic/persona-gentleman.md` MUST be used as the persona

#### Scenario: Agent falls back to neutral persona
- GIVEN neither `assets/{folder}/persona-gentleman.md` nor `assets/generic/persona-gentleman.md` exists
- AND `assets/generic/persona-neutral.md` exists
- WHEN `StepInstallGlobalRules` runs
- THEN `assets/generic/persona-neutral.md` MUST be used as the persona

#### Scenario: No persona file found
- GIVEN none of the three candidate paths exist in the embedded FS
- WHEN `StepInstallGlobalRules` runs
- THEN the step MUST return a descriptive error and NOT write a persona file to `GlobalSkillsDir()`

## Historical Notes

### Removed Requirement: Static Base Persona
The prior single `templates/base/persona.md` lookup has been replaced by the three-tier per-agent resolution chain above. (`StepSetupLocalRules` is out of scope for this change and tracked separately.)
