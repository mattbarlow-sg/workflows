# BPMN Process Documentation

This directory contains documentation for all BPMN 2.0 processes defined in the workflows system.

## Available Processes

### 1. Temporal Workflow Validation Framework
- **File**: [temporal-validation-framework.md](temporal-validation-framework.md)
- **Definition**: `/definitions/bpmn/temporal-validation-framework.json`
- **Purpose**: Validates Temporal workflows for determinism, activity signatures, and timeout/retry policies
- **Type**: Automated validation process
- **Complexity**: Medium (Score: 21)

## Process Categories

### Validation & Testing
- [Temporal Workflow Validation Framework](temporal-validation-framework.md) - Comprehensive Temporal workflow validation

### Development Workflows
_(To be added)_

### CI/CD Processes
_(To be added)_

### Operational Processes
_(To be added)_

## Using BPMN Processes

### Validate a Process
```bash
./workflows bpmn validate definitions/bpmn/<process-name>.json
```

### Analyze Process Complexity
```bash
./workflows bpmn analyze definitions/bpmn/<process-name>.json
```

### Generate Visualizations
```bash
# Generate Mermaid diagram
./workflows bpmn render -format mermaid definitions/bpmn/<process-name>.json

# Generate text representation
./workflows bpmn render -format text definitions/bpmn/<process-name>.json

# Generate markdown documentation
./workflows bpmn render -format markdown definitions/bpmn/<process-name>.json
```

## Process Design Guidelines

### Naming Conventions
- **Process ID**: Use kebab-case (e.g., `temporal-validation-framework`)
- **Process Name**: Use title case (e.g., `Temporal Workflow Validation Framework`)
- **Element IDs**: Use kebab-case descriptive names
- **Element Names**: Use title case or sentence case

### Best Practices
1. **Start Simple**: Begin with the happy path, add error handling later
2. **Clear Documentation**: Include comprehensive documentation for each element
3. **Atomic Tasks**: Each task should have a single, clear responsibility
4. **Error Handling**: Always include error paths and end events
5. **Agent Assignment**: Clearly specify agent types (human, AI, system, hybrid)

### Complexity Guidelines
- **Low Complexity** (Score < 10): Simple linear processes
- **Medium Complexity** (Score 10-30): Processes with parallel paths or multiple gateways
- **High Complexity** (Score > 30): Complex processes requiring careful review

## Process Metrics

### Key Metrics
- **Complexity Score**: Calculated based on elements and connections
- **Process Depth**: Maximum path length through the process
- **Process Width**: Maximum parallel paths
- **Connectivity**: Ratio of connections to elements

### Analysis Thresholds
- **Depth > 10**: Consider breaking into sub-processes
- **Width > 5**: Review parallel execution requirements
- **Complexity > 50**: Strongly consider simplification

## Contributing

When adding new BPMN processes:

1. Create the process definition in `/definitions/bpmn/`
2. Validate using `./workflows bpmn validate`
3. Analyze complexity using `./workflows bpmn analyze`
4. Generate documentation using `./workflows bpmn render -format markdown`
5. Create detailed documentation in `/docs/bpmn/`
6. Update this README with the new process

## Tools and Commands

### BPMN CLI Commands
- `validate` - Validate BPMN structure and semantics
- `analyze` - Analyze process complexity and metrics
- `render` - Generate visualizations and documentation

### Supported Formats
- **JSON**: Native BPMN definition format
- **Mermaid**: Web-friendly diagram format
- **DOT**: Graphviz format for detailed diagrams
- **Text**: Human-readable text representation
- **Markdown**: Documentation format with embedded diagrams

## References
- [BPMN 2.0 Specification](https://www.omg.org/spec/BPMN/2.0/)
- [BPMN Concepts Guide](../bpmn-concepts.md)
- [Workflow Patterns](http://www.workflowpatterns.com/)