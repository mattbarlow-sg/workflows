# Workflows CLI

A comprehensive Go-based command-line tool for managing JSON schemas, Architecture Decision Records (ADRs), and BPMN 2.0 workflow processes.

## Features

### Schema Validation
- **Automatic Schema Discovery**: Automatically discovers all JSON schema files in the `schemas/` directory
- **List Schemas**: View all available schemas with their titles and descriptions
- **Validate Files**: Validate JSON files against any discovered schema
- **Clear Error Messages**: Get detailed validation error messages when files don't match schemas
- **Extensible**: Add new schemas by simply dropping `.json` files in the `schemas/` directory

### Architecture Decision Records (ADR)
- **Create ADRs**: Generate new ADRs with comprehensive metadata and decision tracking
- **Validate ADRs**: Ensure ADRs conform to the schema
- **Render to Markdown**: Convert ADR JSON files to readable Markdown format

### BPMN 2.0 Workflows
- **Validate Processes**: Check BPMN process definitions for correctness
- **Analyze Complexity**: Get metrics on process complexity, paths, and potential issues
- **Render Diagrams**: Export processes to DOT, Mermaid, or text format

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd workflows

# Install dependencies
go mod download

# Build the CLI
go build -o workflows cmd/workflows/main.go
```

## Usage

### General Help

```bash
./workflows help
./workflows help <command>
```

### Schema Commands

#### List Available Schemas

```bash
./workflows list
```

Output:
```
NAME          TITLE                      DESCRIPTION
----          -----                      -----------
config        Application Configuration  Schema for application configuration files
user          User Profile               Schema for user profile data
api-response  API Response               Standard API response format
adr           Architecture Decision Record  Schema for Architecture Decision Records
```

#### Validate a File

```bash
./workflows validate <schema-name> <file-path>
```

Examples:
```bash
# Validate a configuration file
./workflows validate config test-data/valid-config.json

# Validate a user profile
./workflows validate user test-data/valid-user.json
```

### ADR Commands

#### Create a New ADR

```bash
./workflows adr new [flags]
```

Required flags:
- `-title`: ADR title
- `-problem`: Problem statement (min 10 chars)
- `-background`: Background context
- `-chosen`: Chosen option name
- `-rationale`: Why this option was chosen
- `-positive` or `-negative`: At least one consequence

Example:
```bash
./workflows adr new \
  -title "Use PostgreSQL for data storage" \
  -problem "Need reliable data persistence solution" \
  -background "Building new application requiring ACID compliance" \
  -chosen "PostgreSQL" \
  -rationale "Mature, feature-rich, strong community support" \
  -positive "ACID compliance,JSON support,Extensions available" \
  -negative "Requires DB administration knowledge" \
  -output my-adr.json
```

#### Validate an ADR

```bash
./workflows adr validate <file>
```

#### Render ADR to Markdown

```bash
./workflows adr render <file> [-output <output-file>]
```

### BPMN Commands

#### Validate a BPMN Process

```bash
./workflows bpmn validate <file>
```

#### Analyze Process Complexity

```bash
./workflows bpmn analyze <file>
```

Output includes:
- Process metrics (complexity, depth, width)
- Element breakdown
- Path analysis
- Agent workload distribution
- Potential issues and deadlocks

#### Render Process Diagrams

```bash
./workflows bpmn render -format <format> <file> [-output <output-file>]
```

Formats:
- `dot`: GraphViz DOT format
- `mermaid`: Mermaid diagram format  
- `text`: Simple text representation

## Adding New Schemas

1. Create a JSON schema file following the JSON Schema specification
2. Place the file in the `schemas/` directory with a `.json` extension
3. The schema will be automatically discovered and available for use

Example schema structure:
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Your Schema Title",
  "description": "Description of what this schema validates",
  "type": "object",
  "properties": {
    // Define your schema properties here
  },
  "required": ["field1", "field2"]
}
```

## Project Structure

```
workflows/
├── cmd/workflows/          # CLI entry point
│   ├── main.go            # Main entry point using command framework
│   └── commands/          # Command implementations
│       ├── list.go        # List schemas command
│       ├── validate.go    # Validate files command
│       ├── adr*.go        # ADR commands
│       └── bpmn*.go       # BPMN commands
├── internal/
│   ├── cli/              # Command framework
│   │   ├── command.go    # Command interface and base
│   │   ├── manager.go    # Command manager
│   │   └── validation.go # Validation chain
│   ├── schema/           # Schema discovery and validation
│   │   ├── discovery.go  # Schema discovery and registry
│   │   └── validator.go  # JSON schema validation
│   ├── adr/              # ADR domain logic
│   │   ├── builder.go    # ADR builder pattern
│   │   ├── renderer.go   # ADR rendering
│   │   └── template.go   # ADR types and structures
│   ├── bpmn/             # BPMN domain logic
│   │   ├── analyzer.go   # Process analysis
│   │   ├── validator.go  # BPMN validation
│   │   ├── renderer.go   # Diagram rendering
│   │   └── types.go      # BPMN type definitions
│   ├── config/           # Configuration management
│   ├── errors/           # Custom error types
│   └── validation/       # Input validation
├── schemas/              # JSON schema files
│   ├── adr.json         # ADR schema
│   ├── bpmn-*.json      # BPMN schemas
│   ├── config.json      # Config schema
│   └── user.json        # User schema
├── test-data/            # Sample files for testing
├── docs/                 # Documentation
│   └── adr/             # ADR documentation
└── README.md            # This file
```

## Example Schemas

### Config Schema
Validates application configuration files with:
- Semantic version validation
- Port number ranges
- Database connection settings
- Feature flags

### User Schema
Validates user profile data with:
- UUID format for IDs
- Email validation
- Username pattern matching
- Role enumeration
- Age constraints

### API Response Schema
Validates API responses with:
- Success/error state validation
- Conditional field requirements
- Metadata fields

## Error Handling

When validation fails, the tool provides detailed error messages:

```bash
./workflows validate config test-data/invalid-config.json
```

Output:
```
✗ File 'test-data/invalid-config.json' is invalid according to schema 'config'

Validation errors:
  1. database: port is required
  2. database: name is required
  3. version: Does not match pattern '^\d+\.\d+\.\d+$'
  4. port: Must be less than or equal to 65535
```

## Dependencies

- Go 1.24.5+
- github.com/xeipuuv/gojsonschema - JSON Schema validation library