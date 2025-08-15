# Temporal Workflow Generator

The Temporal Workflow Generator is a comprehensive code generation system that produces validated Temporal workflow code from templates and specifications. It ensures all generated code is deterministic, follows Go best practices, and passes Temporal's validation requirements.

## Features

- **Multiple Templates**: Supports various workflow patterns including basic, approval, scheduled, human task, and long-running workflows
- **Activity Generation**: Automatically generates activity function stubs
- **Test Generation**: Creates comprehensive test suites using Temporal's test framework
- **Signal & Query Support**: Generates signal and query handlers with proper type safety
- **Human Task Management**: Built-in support for human-in-the-loop workflows with escalation
- **Child Workflow Support**: Generates code for parent-child workflow relationships
- **Validation Integration**: Validates generated code using the Temporal validation framework
- **Fluent Builder API**: Provides an intuitive builder pattern for workflow specification

## Quick Start

### Basic Usage

```go
import "github.com/mattbarlow-sg/workflows/internal/temporal"

// Create a generator
gen, err := temporal.NewGenerator(temporal.GeneratorConfig{
    OutputDir: "pkg/workflows/generated",
})

// Define a workflow specification
spec := temporal.WorkflowSpec{
    Name:        "MyWorkflow",
    Package:     "mypackage",
    Description: "A simple workflow",
    InputType:   "string",
    OutputType:  "string",
    Template:    "basic",
}

// Generate the workflow
ctx := context.Background()
code, err := gen.GenerateWorkflow(ctx, spec)

// Save the generated code
err = gen.SaveGeneratedCode(code)
```

### Using the Builder API

```go
builder := temporal.NewWorkflowBuilder("OrderProcessing", "orders").
    WithDescription("Process customer orders").
    WithTemplate("approval").
    AddActivity("ValidateOrder", "Validate order details").
        WithTimeout(5 * time.Minute).
        Done().
    AddHumanTask("ManagerApproval", "Manager approval required").
        AssignTo("manager@example.com").
        WithEscalation(2*time.Hour, "director@example.com").
        Done()

// Generate the workflow
spec, _ := builder.Build()
code, _ := gen.GenerateWorkflow(ctx, spec)
```

## Available Templates

### 1. Basic Workflow (`basic`)
Standard workflow template with activities, signals, and queries.

### 2. Approval Workflow (`approval`)
Human approval workflow with escalation support:
- Approval chain processing
- Timeout handling with automatic escalation
- Signal-based approval/rejection
- Query handlers for status checking

### 3. Scheduled Workflow (`scheduled`)
Cron-like scheduled workflow execution:
- Periodic task execution
- Max run limits
- Query handlers for execution status

### 4. Human Task Workflow (`human_task`)
Comprehensive human task management:
- Task assignment and reassignment
- Priority levels
- Deadline management
- Escalation on timeout
- Signal-based task completion

### 5. Long-Running Workflow (`long_running`)
Workflows that run for extended periods:
- Checkpoint management
- Pause/resume capabilities
- Progress tracking
- Phase-based execution
- Continue-as-new support

## Pre-configured Generators

### Approval Workflow
```go
builder := temporal.GenerateApprovalWorkflow(
    "ExpenseApproval",
    "expenses",
    []string{"manager@example.com", "director@example.com"},
)
```

### Scheduled Workflow
```go
builder := temporal.GenerateScheduledWorkflow(
    "DailyReport",
    "reporting",
    "0 9 * * *", // Run at 9 AM daily
)
```

### Long-Running Workflow
```go
builder := temporal.GenerateLongRunningWorkflow(
    "DataMigration",
    "migration",
    72*time.Hour, // 3 days max duration
)
```

### ETL Workflow
```go
builder := temporal.GenerateETLWorkflow(
    "CustomerDataETL",
    "etl",
)
```

## Workflow Specification

### Core Fields

- `Name`: Workflow name (required)
- `Package`: Go package name (required)
- `Description`: Workflow description
- `InputType`: Input struct fields
- `OutputType`: Output struct fields
- `Template`: Template to use (default: "basic")

### Activities

```go
Activities: []temporal.ActivitySpec{
    {
        Name:        "ProcessData",
        Description: "Process input data",
        InputType:   "DataInput",
        OutputType:  "DataOutput",
        Timeout:     10 * time.Minute,
        RetryPolicy: temporal.RetryPolicy{
            InitialInterval:    1 * time.Second,
            BackoffCoefficient: 2.0,
            MaximumInterval:    1 * time.Minute,
            MaximumAttempts:    3,
        },
        IsHumanTask: false,
    },
}
```

### Signals

```go
Signals: []temporal.SignalSpec{
    {
        Name:        "UpdateConfig",
        Description: "Update workflow configuration",
        PayloadType: "ConfigUpdate",
    },
}
```

### Queries

```go
Queries: []temporal.QuerySpec{
    {
        Name:         "GetStatus",
        Description:  "Get current workflow status",
        ResponseType: "WorkflowStatus",
    },
}
```

### Human Tasks

```go
HumanTasks: []temporal.HumanTaskSpec{
    {
        Name:           "ApprovalRequired",
        Description:    "Manager approval needed",
        AssignedTo:     "manager@example.com",
        EscalationTime: 4 * time.Hour,
        EscalationTo:   "director@example.com",
        Priority:       "high",
        Deadline:       8 * time.Hour,
    },
}
```

## Generated Files

The generator creates three files:

1. **Workflow File** (`<name>_workflow.go`)
   - Workflow implementation
   - Input/output types
   - Signal and query handlers
   - Registration function

2. **Activity File** (`activities.go`)
   - Activity function stubs
   - Ready for implementation

3. **Test File** (`<name>_workflow_test.go`)
   - Test suite setup
   - Activity mocks
   - Signal and query tests
   - Ready-to-run test cases

## Validation

All generated code is automatically validated against:
- Temporal determinism requirements
- Activity signature correctness
- Timeout and retry policy validation
- Go syntax and formatting

## Best Practices

1. **Use appropriate templates**: Choose the template that best matches your use case
2. **Define clear types**: Specify input/output types explicitly
3. **Set reasonable timeouts**: Configure appropriate timeouts for activities
4. **Implement retry policies**: Add retry policies for transient failures
5. **Add tests**: Extend generated tests with your specific test cases
6. **Validate before deployment**: Always run validation before deploying workflows

## Examples

See `generator_example_test.go` for complete working examples of:
- Basic workflow generation
- Using the builder API
- Generating approval workflows
- Creating ETL workflows
- Long-running workflow patterns

## Integration with Registry

Generated workflows can be automatically registered:

```go
// Generate workflow
code, _ := gen.GenerateWorkflow(ctx, spec)

// Register with worker
registry := temporal.NewRegistry()
registry.RegisterWorkflow(MyWorkflow{}.Execute)
```

## Troubleshooting

### Common Issues

1. **Validation Errors**: Ensure your workflow follows Temporal determinism rules
2. **Formatting Errors**: Check that generated code has valid Go syntax
3. **Template Not Found**: Verify template name matches available templates
4. **Type Mismatches**: Ensure input/output types are properly defined

### Debug Tips

- Check generated code in the output directory
- Review validation errors in `GeneratedCode.Errors`
- Use the test file to verify workflow behavior
- Enable verbose logging in the generator configuration