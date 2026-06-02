# asset-injection-engine Specification

## Purpose
The system responsible for exactly copying and injecting all 21+ `gentle-ai` assets, including skills, personas, and MCP configurations.

## Requirements

### Requirement: Asset Extraction
The system MUST extract and literal copy/paste ALL assets from the `gentle-ai` source (skills, persona configs, and MCP configs).

#### Scenario: Extracting gentle-ai assets
- GIVEN the `gentle-ai` source directory is accessible
- WHEN the install phase begins
- THEN the system MUST extract all 21+ skills and their corresponding configurations exactly as they are

### Requirement: MCP Configuration Adaptation
The system MUST adapt the extracted MCP configurations specifically to replace the 'engram' MCP configuration with 'sdd-memory' for injection into SpecAI.

#### Scenario: Replacing engram with sdd-memory
- GIVEN the extracted MCP configurations include 'engram'
- WHEN injecting the MCP configurations into SpecAI
- THEN the system MUST replace the 'engram' configuration with 'sdd-memory'
- AND all other MCP configurations MUST remain unchanged

### Requirement: Asset Injection
The system MUST inject the extracted and adapted assets precisely during the setup phase of the install pipeline.

#### Scenario: Injecting assets during setup
- GIVEN the assets have been extracted and adapted
- WHEN the setup phase of the pipeline is executing
- THEN the system MUST place the assets into their correct target directories in the SpecAI environment
