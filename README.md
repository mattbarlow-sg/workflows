# Schema Validation CLI

A Go-based command-line tool for validating JSON files against JSON schemas with automatic schema discovery.

## Features

- **Automatic Schema Discovery**: Automatically discovers all JSON schema files in the `schemas/` directory
- **List Schemas**: View all available schemas with their titles and descriptions
- **Validate Files**: Validate JSON files against any discovered schema
- **Clear Error Messages**: Get detailed validation error messages when files don't match schemas
- **Extensible**: Add new schemas by simply dropping `.json` files in the `schemas/` directory

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

### List Available Schemas

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
```

### Validate a File

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
├── cmd/workflows/       # CLI entry point
│   └── main.go
├── internal/schema/     # Schema discovery and validation logic
│   ├── discovery.go     # Schema discovery and registry
│   └── validator.go     # JSON schema validation
├── schemas/            # JSON schema files
│   ├── config.json
│   ├── user.json
│   └── api-response.json
├── test-data/          # Sample JSON files for testing
│   ├── valid-config.json
│   ├── invalid-config.json
│   └── valid-user.json
└── README.md           # This file
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