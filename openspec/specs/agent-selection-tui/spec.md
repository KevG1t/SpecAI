# agent-selection-tui Specification

## Purpose
A dedicated TUI screen (`AgentSelectModel`) that lets the user pick which agents to configure before the install pipeline fires.

## Requirements

### Requirement: Full Agent Catalog Displayed
The system MUST populate the agent selection screen from `catalog.AllAgents()`, presenting all 15 known agents.

#### Scenario: All agents listed
- GIVEN the application has loaded the agent catalog
- WHEN the `AgentSelectModel` screen is rendered
- THEN all 15 agents returned by `catalog.AllAgents()` MUST appear as selectable rows

### Requirement: Pre-Selection of Detected Agents
The system MUST pre-select agents that were detected by `StepScanGlobalIDEs` when the selection screen is first shown.

#### Scenario: Detected agents pre-checked
- GIVEN `StepScanGlobalIDEs` detected `claude-code` and `gemini-cli` on the user's system
- WHEN the `AgentSelectModel` screen opens
- THEN the checkboxes for `claude-code` and `gemini-cli` MUST be checked by default
- AND all other agents MUST be unchecked

### Requirement: Multi-Select Confirmation
The system MUST emit `AgentsSelectedMsg{Agents: []model.AgentID}` when the user confirms their selection.

#### Scenario: User confirms with Enter
- GIVEN one or more agents are checked on the selection screen
- WHEN the user presses Enter
- THEN the system MUST emit `AgentsSelectedMsg` carrying exactly the checked agent IDs
- AND the application MUST advance to the pipeline launch

#### Scenario: User confirms with no agents selected
- GIVEN no agents are checked
- WHEN the user presses Enter
- THEN the system MUST NOT advance to the pipeline
- AND MUST display an error or prompt requiring at least one selection

### Requirement: Selection Wired to Resolved Plan
The system MUST build `Selection{Agents: selected}`, pass it to `planner.Resolve`, and use the resulting `ResolvedPlan.Agents` to populate `ctx.IDEs` for the rest of the pipeline.

#### Scenario: Selection drives IDEs in context
- GIVEN the user selected only `codex`
- WHEN `AgentsSelectedMsg` is handled by `model.go`
- THEN `ctx.IDEs` MUST contain only the Codex adapter
- AND the hardcoded dummy agent list in `model.go` MUST NOT be used

### Requirement: Screen Inserted Between Welcome and Pipeline
The `AgentSelectModel` screen MUST appear after the welcome screen and before the pipeline fires; no other screen order is valid.

#### Scenario: Correct screen sequence
- GIVEN the application starts
- WHEN the user dismisses the welcome screen
- THEN the agent selection screen MUST be shown next
- AND the install pipeline MUST NOT start until `AgentsSelectedMsg` is received
