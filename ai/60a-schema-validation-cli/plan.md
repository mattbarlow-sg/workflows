# Schema Validation CLI Implementation Plan

## Overview
Implement a CLI that supports JSON schemas with automatic discovery
from the filesystem. The CLI should list available schemas and
validate definitions against them without requiring code changes when
new schemas are added.

## Technical Approach
- Use Go's standard library for JSON schema validation
- Store schemas in a dedicated directory (e.g., `schemas/`)
- Implement filesystem-based schema discovery
- Create a simple CLI with subcommands for listing and validation

## Implementation Steps

### 1. Project Structure Setup
**Assignee: AI**
**Status: Completed**
- Create `schemas/` directory for storing JSON schema files
- Update project structure for CLI commands
- Add necessary dependencies for JSON schema validation

### 2. Core Schema Discovery Implementation
**Assignee: AI**
**Status: Completed**
- Implement filesystem scanner to discover `.json` schema files
- Create schema loader that reads and parses JSON schemas
- Build schema registry to manage available schemas

### 3. CLI Command Structure
**Assignee: AI**
**Status: Completed**
- Implement main CLI entry point with subcommands
- Create `list` command to show available schemas
- Create `validate` command with schema and file parameters

### 4. Schema Listing Feature
**Assignee: AI**
**Status: Completed**
- Implement logic to scan schemas directory
- Display schema names and descriptions
- Format output in user-friendly table or list

### 5. Validation Feature
**Assignee: AI**
**Status: Completed**
- Implement JSON schema validation logic
- Accept schema name and target file as parameters
- Provide clear validation results and error messages

### 6. Example Schemas
**Assignee: AI**
**Status: Completed**
- Create sample schemas for testing
- Include common use cases (config files, API responses, etc.)
- Document schema format and conventions

### 7. Testing and Documentation
**Assignee: AI**
**Status: Completed**
- Test various schema and input combinations
- Create README with usage examples
- Add inline code documentation

### 8. User Review and Feedback
**Assignee: Human**
**Status: Pending**
- Review implementation
- Test with custom schemas
- Provide feedback on improvements
