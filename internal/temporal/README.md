# Temporal Workflow Validation Framework

A comprehensive validation framework for Temporal workflows that ensures code quality, determinism, and best practices before deployment.

## Features

**Determinism Checking** - Detects non-deterministic patterns that break workflow replay
**Activity Signature Validation** - Ensures activities follow Temporal conventions
**Workflow Graph Analysis** - Detects cycles and connectivity issues
**Policy Enforcement** - Validates timeout and retry configurations
**Human Task Support** - Special validation for human-in-the-loop workflows
**Performance Caching** - Speeds up repeated validations
**Parallel Validation** - Concurrent checks for better performance

## Quick Start

### 1. Create Sample Workflows

```bash
go run cmd/validate-temporal/main.go -sample
```

This creates three sample workflows in `./sample-workflows/`:
- `valid_workflow.go` - A properly structured Temporal workflow
- `invalid_workflow.go` - A workflow with multiple validation issues
- `cyclic_workflow.go` - Workflows with circular dependencies

### 2. Validate a Workflow

```bash
# Validate a valid workflow
go run cmd/validate-temporal/main.go -path /absolute/path/to/workflow.go

# Validate with verbose output
go run cmd/validate-temporal/main.go -path /absolute/path/to/workflow.go -v

# Enable caching for faster repeated validations
go run cmd/validate-temporal/main.go -path /absolute/path/to/workflow.go -cache
```

## Validation Checks

### 1. Determinism Validation

Detects non-deterministic patterns that break Temporal's replay mechanism:

- `time.Now()` → Use `workflow.Now()`
- Native goroutines → Use `workflow.Go()`
- `math/rand` → Use `workflow.SideEffect()`
- Native channels → Use `workflow.NewChannel()`
- Native select → Use `workflow.NewSelector()`
- Environment variables → Pass as workflow inputs
- File I/O → Move to activities
- Network calls → Move to activities
- Map iteration → Sort keys before iteration

### 2. Activity Signature Validation

Ensures activities follow Temporal best practices:

- PascalCase naming (e.g., `ProcessOrderActivity`)
- `context.Context` as first parameter
- Returns `error` as last value
- Maximum 3 parameters (use structs for complex inputs)
- Maximum 2 return values (result, error)

### 3. Policy Validation

Validates timeout and retry configurations:

- Workflow timeout ≤ 15 minutes (for non-human tasks)
- Activity timeout properly configured
- Retry count ≥ 3 for retryable errors
- Human tasks have infinite timeout
- Proper backoff configuration

### 4. Human Task Validation

Special checks for human-in-the-loop workflows:

- Escalation policy defined
- Task assignment rules configured
- Timeout handling implemented
- Signal channels properly set up

## Programmatic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/mattbarlow-sg/workflows/internal/temporal"
    "github.com/mattbarlow-sg/workflows/src/schemas"
)

func main() {
    // Create validator
    validator := temporal.NewTemporalValidator()
    
    // Configure validation request
    request := schemas.ValidationRequest{
        WorkflowPath: "/path/to/workflow.go",
        WorkflowID:   "my-workflow",
        Options: schemas.ValidationOptions{
            ParallelChecks:  true,
            Timeout:         30 * time.Second,
            MaxInfoMessages: 10,
        },
    }
    
    // Perform validation
    ctx := context.Background()
    result, err := validator.Validate(ctx, request)
    if err != nil {
        panic(err)
    }
    
    // Check results
    if result.Success {
        fmt.Println("Validation passed!")
    } else {
        fmt.Printf("Found %d errors\n", len(result.Errors))
        for _, err := range result.Errors {
            fmt.Printf("- %s: %s\n", err.Code, err.Message)
        }
    }
}
```

## Error Categories

| Category | Description | Retryable |
|----------|-------------|-----------|
| `DETERMINISM` | Non-deterministic code patterns | Yes |
| `TYPE_MISMATCH` | Signature or type errors | No |
| `TIMEOUT` | Timeout policy violations | Yes |
| `RETRY` | Retry policy violations | Yes |
| `SYNTAX` | Go syntax errors | No |
| `PERMISSION` | File access issues | No |
| `NAMING` | Naming convention violations | Yes |

## Architecture

```
internal/temporal/
├── validator.go              # Main validation orchestrator
├── determinism_checker.go    # Determinism validation logic
├── activity_validator.go     # Activity signature validation
├── validation_rules.go       # Policy & graph validators
├── validator_test.go         # Comprehensive test suite
├── example_test.go          # Usage examples
└── README.md                # This file
```

## Testing

```bash
# Run all tests
go test ./internal/temporal/...

# Run with coverage
go test ./internal/temporal/... -cover

# Run benchmarks
go test ./internal/temporal/... -bench=.

# Run specific test
go test ./internal/temporal/... -run TestDeterminismChecker
```

## Common Issues & Solutions

### Issue: "workflow path must be absolute"
**Solution:** Use absolute paths starting with `/`

### Issue: False positives on workflow functions
**Solution:** The validator skips functions with "Workflow" in the name

### Issue: Cache not working
**Solution:** Enable with `-cache` flag; cache is in-memory per run

### Issue: Validation takes too long
**Solution:** Use `-parallel` flag (default) for concurrent checks

## Next Steps

1. **Integration with CI/CD** - Add validation to your build pipeline
2. **Custom Rules** - Extend validators for domain-specific checks
3. **IDE Integration** - Add as a linter to your development environment
4. **Metrics Collection** - Track validation metrics over time

## Contributing

To add new validation rules:

1. Add pattern to `DeterminismPatterns` in `validation_rules.go`
2. Implement check logic in appropriate validator
3. Add test cases to `validator_test.go`
4. Update documentation

## License

Part of the Temporal Workflow Implementation project.