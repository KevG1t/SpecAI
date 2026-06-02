# tui-pipeline-integration Specification

## Purpose
Wiring the interactive TUI to report actual pipeline execution progress instead of using a static 3-state flow.

## Requirements

### Requirement: Dynamic Pipeline Views
The system MUST replace the static 3-state TUI loop with dynamic view components driven by the `planner` pipeline.

#### Scenario: Displaying real-time progress
- GIVEN the install pipeline is running
- WHEN a pipeline step (e.g., Strict TDD, Install Plan) begins or completes
- THEN the TUI MUST dynamically update to reflect the current step and its progress
- AND the TUI MUST NOT block the synchronous pipeline work

### Requirement: Planner Progress Reporting
The `planner` MUST expose progress hooks (e.g., via Go channels or callbacks) to feed execution progress directly to Bubble Tea views.

#### Scenario: Emitting pipeline events
- GIVEN the `planner` is executing a task
- WHEN the task state changes
- THEN the `planner` MUST emit an event with the progress details
- AND the TUI MUST consume this event to update the view without UI lag
