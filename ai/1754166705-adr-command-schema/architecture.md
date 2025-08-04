# ADR Command Architecture

## Final Architecture

### System Overview

The ADR system provides a complete solution for creating, validating, and managing Architecture Decision Records through both CLI and AI-assisted workflows.

```
┌─────────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   AI Command        │     │    CLI Tool      │     │  JSON Schema    │
│ ai-adr-create.md    │────▶│  workflows adr   │────▶│   adr.json      │
└─────────────────────┘     └──────────────────┘     └─────────────────┘
         │                           │                         │
         │                           ▼                         ▼
         │                  ┌──────────────────┐     ┌─────────────────┐
         │                  │   Go Packages    │     │   Validation    │
         └─────────────────▶│  internal/adr/   │────▶│    Engine       │
                            └──────────────────┘     └─────────────────┘
                                     │
                                     ▼
                            ┌──────────────────┐
                            │  Output Files    │
                            │  .json & .md     │
                            └──────────────────┘
```

### Component Architecture

#### 1. JSON Schema Layer (`schemas/`)
- **adr.json**: Complete ADR schema with all fields and validation rules
- **Based on**: MADR 4.0.0 standard with AI extensions
- **Validation**: JSON Schema Draft-07 specification
- **Features**: Required fields, constraints, enumerations, patterns

#### 2. Go Implementation (`internal/adr/`)
- **template.go**: ADR struct definitions and generation logic
  - Mirrors JSON schema structure
  - ID and timestamp generation
  - JSON serialization/deserialization
- **markdown.go**: Markdown rendering engine
  - Converts ADR structs to formatted markdown
  - Supports tables, lists, and mermaid diagrams
  - Clean, readable output

#### 3. CLI Tool (`cmd/workflows/`)
- **main.go**: Entry point with command routing
- **adr.go**: ADR-specific commands
  - `new`: Generate ADRs with comprehensive flag support
  - `validate`: Validate against JSON schema
  - `render`: Convert JSON to Markdown
- **Features**:
  - Verbose help with examples
  - Complete non-interactive support
  - Clear error messages
  - Flexible output options

#### 4. AI Command (`.claude/commands/`)
- **ai-adr-create.md**: Interactive ADR creation workflow
- **Process**:
  1. Context gathering through conversation
  2. Option research via WebSearch
  3. Stakeholder identification
  4. Decision analysis and documentation
  5. CLI command generation and execution
- **Integration**: Uses CLI tool for all operations

### Data Flow

1. **Interactive Creation** (AI Command)
   ```
   User Input → AI Assistant → Build CLI Command → Execute → Validate → Output
   ```

2. **Direct Creation** (CLI)
   ```
   Command Flags → Parse & Validate → Generate ADR → Validate Schema → Output
   ```

3. **Rendering Pipeline**
   ```
   JSON File → Parse → Struct → Markdown Renderer → Formatted Output
   ```

### Key Design Decisions

#### 1. Separation of Concerns
- Schema defines structure and validation
- Go code handles generation and rendering
- CLI provides user interface
- AI command adds intelligence layer

#### 2. Schema-First Design
- JSON schema is source of truth
- Go structs mirror schema exactly
- Validation happens at multiple levels
- Ensures consistency across system

#### 3. CLI Completeness
- All fields accessible via flags
- No interactive prompts required
- Suitable for automation
- Verbose help documentation

#### 4. AI Enhancement
- Optional intelligent layer
- Guides users through process
- Researches best practices
- Generates complete CLI commands

### File Structure

```
workflows/
├── .claude/commands/
│   └── ai-adr-create.md        # AI-assisted workflow
├── cmd/workflows/
│   ├── main.go                 # CLI entry point
│   └── adr.go                  # ADR commands
├── internal/adr/
│   ├── template.go             # ADR generation
│   └── markdown.go             # Markdown rendering
├── schemas/
│   └── adr.json                # ADR JSON schema
├── docs/adr/
│   ├── README.md               # User documentation
│   └── SCHEMA.md               # Schema documentation
└── examples/
    ├── sample-adr.json         # Complete example
    └── sample-adr.md           # Rendered example
```

### Technical Stack

- **Language**: Go 1.24.5
- **Schema**: JSON Schema Draft-07
- **Validation**: github.com/xeipuuv/gojsonschema
- **Standards**: MADR 4.0.0
- **AI**: Claude with tool integration

### Extension Points

1. **Schema Extensions**: Add fields to adr.json
2. **Rendering Customization**: Modify markdown.go
3. **New Commands**: Add to adr.go
4. **AI Workflows**: Create new .claude/commands
5. **Output Formats**: Extend renderer for HTML, PDF, etc.

### Security Considerations

- No network calls in core implementation
- File I/O limited to specified paths
- Schema validation prevents injection
- CLI input sanitization
- No execution of user content

### Performance Characteristics

- Instant generation and validation
- No database dependencies
- Linear scaling with ADR count
- Minimal memory footprint
- Suitable for CI/CD integration
