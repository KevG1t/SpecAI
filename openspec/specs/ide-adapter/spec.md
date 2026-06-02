# Spec: ide-adapter

## ADDED Requirements

### Requirement: AgentID Method on IDEAdapter
The `IDEAdapter` interface MUST expose an `AgentID() model.AgentID` method returning the canonical identifier for the agent this adapter represents.

#### Scenario: AgentID returns correct value
- GIVEN the `ClaudeCodeAdapter` is instantiated
- WHEN `AgentID()` is called
- THEN it MUST return `model.AgentID("claude-code")`

### Requirement: AssetFolder Method on IDEAdapter
The `IDEAdapter` interface MUST expose an `AssetFolder() string` method returning the subfolder name under `assets/` that contains this agent's files.

#### Scenario: AssetFolder returns correct folder
- GIVEN the `ClaudeCodeAdapter` is instantiated
- WHEN `AssetFolder()` is called
- THEN it MUST return `"claude"`

#### Scenario: Generic adapters return "generic"
- GIVEN any of the adapters for `openclaw`, `pi`, `trae`, or `vscode-copilot`
- WHEN `AssetFolder()` is called
- THEN it MUST return `"generic"`

### Requirement: Fifteen Adapters Registered
`GetAdapters()` MUST return all 15 agent adapters with their correct `AgentID()` and `AssetFolder()` values as defined in the asset folder mapping table.

#### Scenario: Full catalog returned
- GIVEN the current runtime environment
- WHEN `GetAdapters()` is called
- THEN the returned slice MUST contain exactly 15 adapters
- AND each adapter MUST satisfy the AgentID → AssetFolder mapping in the proposal

### Requirement: Existing Mocks Implement New Methods
All test mocks that implement `IDEAdapter` MUST add `AgentID()` and `AssetFolder()` methods; compilation MUST succeed with no stub panics.

#### Scenario: Mock satisfies updated interface
- GIVEN a test that uses a mock `IDEAdapter`
- WHEN the test is compiled and run
- THEN it MUST compile without error
- AND calls to `AgentID()` and `AssetFolder()` on the mock MUST return deterministic test values

## Requirements

### Requirement: IDEAdapter Interface Methods
The `IDEAdapter` interface MUST define exactly 8 methods:
`Name()`, `ConfigDir()`, `GlobalRulesDir()`, `GlobalSkillsDir()`, `LocalRulesFile()`, `LocalSkillsDir()`, `AgentID() model.AgentID`, `AssetFolder() string`.
Any type claiming to implement `IDEAdapter` MUST implement all 8 methods.
