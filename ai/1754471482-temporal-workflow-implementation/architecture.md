# Temporal Workflow Architecture

## Overview
This document outlines the technical architecture for implementing Temporal workflow orchestration in our Golang project, based on ADR-0001.

## Discovery Results

### Existing Codebase Analysis
- **Current CLI Framework**: Project uses a modular CLI system at `cmd/workflows/` with command registration pattern
- **Validation Infrastructure**: Existing MPC validator at `internal/mpc/validator.go` provides schema validation framework
- **BPMN Types**: Rich BPMN type system at `internal/bpmn/types.go` can be adapted for Temporal workflows
- **Configuration System**: Config management at `internal/config/config.go`
- **Error Handling**: Structured error system at `internal/errors/errors.go`

### Temporal Best Practices (2025)
Based on latest documentation:
- **Testing First**: Use `go.temporal.io/sdk/testsuite` for integration testing
- **Static Analysis**: Use `workflowcheck` tool for determinism validation
- **Signal Patterns**: Channel-based signal handling for human-in-the-loop
- **Query Handlers**: For workflow state inspection without modification

### Related Documentation
- **Temporal Go SDK**: https://docs.temporal.io/develop/go/
- **Testing Suite**: https://docs.temporal.io/develop/go/testing-suite
- **Message Passing**: https://docs.temporal.io/develop/go/message-passing
- **Samples Repository**: https://github.com/temporalio/samples-go
- **Workflowcheck Tool**: https://pkg.go.dev/go.temporal.io/sdk/contrib/tools/workflowcheck
- **Go Template Package**: https://pkg.go.dev/text/template
- **Go Format Package**: https://pkg.go.dev/go/format

### Testing Framework Components
- **Test Environment**: `internal/temporal/testing/test_env.go` - Wrapper around Temporal test suite
- **Mock Framework**: `internal/temporal/testing/mocks.go` - Activity and workflow mocking utilities
- **Test Helpers**: `internal/temporal/testing/helpers.go` - Common testing patterns and utilities
- **Integration Tests**: `internal/temporal/testing/integration_test.go` - Complete testing scenarios
- **Human Task Simulation**: Built-in support for simulating human task lifecycles with escalation

## Core Components

### 1. Temporal Server
- **Deployment**: Dedicated Temporal server instance
- **Persistence**: PostgreSQL/MySQL for workflow state
- **Visibility**: Elasticsearch for workflow search and analytics
- **Metrics**: Prometheus integration

### 2. Workflow Engine Layer

```
┌─────────────────────────────────────────────┐
│             CLI Interface                    │
├─────────────────────────────────────────────┤
│          Workflow Library                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │Generator │ │Validator │ │ Registry │   │
│  └──────────┘ └──────────┘ └──────────┘   │
├─────────────────────────────────────────────┤
│          Temporal Client                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │Workflows │ │Activities│ │  Signals │   │
│  └──────────┘ └──────────┘ └──────────┘   │
├─────────────────────────────────────────────┤
│          Temporal Server                     │
└─────────────────────────────────────────────┘
```

### 3. Workflow Patterns

#### Long-Running Workflows
- Use Temporal's durable execution model
- Implement checkpointing for multi-day workflows
- Handle version migration for workflow updates

#### Human-in-the-Loop
- Signal-based task assignment
- Query handlers for task status
- Activity timeouts with escalation

#### Dynamic Routing
- Child workflow spawning
- Conditional task execution
- Event-driven branching

### 4. Code Structure

```
/temporal-workflows/
├── cmd/
│   ├── workflows/commands/  # CLI commands (existing)
│   └── validate-temporal/   # Validation tool (implemented)
├── internal/
│   └── temporal/           # Temporal infrastructure
│       ├── client.go       # Client with reconnection (implemented)
│       ├── worker.go       # Worker pool management (implemented)
│       ├── registry.go     # Workflow/activity registry (implemented)
│       ├── config.go       # Configuration management (implemented)
│       ├── validator.go    # Workflow validation (implemented)
│       ├── bpmn_adapter.go # BPMN conversion (implemented)
│       ├── generator.go    # Code generation (implemented)
│       ├── templates.go    # Workflow templates (implemented)
│       ├── generator_builder.go # Builder API (implemented)
│       └── testing/        # Test framework (implemented)
├── pkg/
│   ├── workflows/        # Workflow definitions
│   │   ├── base/        # Base workflow interfaces
│   │   ├── library/     # Reusable workflows
│   │   └── generated/   # Generated workflows
│   ├── activities/       # Activity implementations
│   ├── signals/         # Signal handlers
│   └── queries/         # Query handlers
└── ui/
    ├── full-view/       # All workflows view
    ├── graph-view/      # Network visualization
    └── queue-view/      # Human task queue

```

### 5. BPMN to Temporal Mapping

| BPMN Element | Temporal Implementation |
|--------------|------------------------|
| Process | Workflow Definition |
| Task | Activity |
| User Task | Activity with Signal |
| Gateway | Workflow Logic |
| Event | Signal/Query |
| Timer | Sleep/Timer |
| Sub-Process | Child Workflow |

### 6. Validation Framework

#### Pre-deployment Validation
- Workflow compilation checks
- Activity signature validation
- Signal/Query handler verification

#### Runtime Validation
- Workflow state consistency
- Activity retry policies
- Timeout configuration

#### Test Validation
- Unit tests with Temporal test framework
- Integration tests with test server
- End-to-end workflow simulations

### 7. Workflow Generator System (Implemented)

The workflow generator system provides a comprehensive code generation framework for creating validated Temporal workflows from specifications and templates.

#### Core Components

**Generator Engine** (`internal/temporal/generator.go`)
- Template-based code generation using Go's text/template
- Automatic code formatting with go/format
- Integration with validation framework
- Support for multiple output files (workflow, activities, tests)

**Template Library** (`internal/temporal/templates.go`)
- Basic workflow template - Simple workflow with activities
- Approval workflow template - Human-in-the-loop approval patterns
- Scheduled workflow template - Cron-based recurring workflows
- Human task workflow template - Task assignment with escalation
- Long-running workflow template - Checkpointing and continue-as-new

**Builder API** (`internal/temporal/generator_builder.go`)
- Fluent interface for workflow specification
- Pre-configured generators for common patterns
- Activity, signal, and query builders
- Human task configuration with priority and deadlines

#### Template System Architecture

```go
// Workflow specification structure
type WorkflowSpec struct {
    Package      string
    Name         string
    Description  string
    InputType    string
    OutputType   string
    Activities   []ActivitySpec
    Signals      []SignalSpec
    Queries      []QuerySpec
    HumanTasks   []HumanTaskSpec
    Options      WorkflowOptions
}

// Generated workflow structure
type {{ .Name }}Workflow struct {
    State   {{ .Name }}State
    Logger  log.Logger
}

func (w *{{ .Name }}Workflow) Execute(ctx workflow.Context, input {{ .InputType }}) ({{ .OutputType }}, error) {
    // Generated deterministic workflow logic
    // Includes activity invocations, signal handling, human tasks
}
```

#### Key Features

- **Deterministic Code Generation**: All generated code is deterministic with no random values or timestamps
- **Validation Integration**: Generated code automatically validated before saving
- **Test Generation**: Complete test suites generated with mocks and test scenarios
- **Human Task Patterns**: Built-in support for task assignment, escalation, and completion
- **Composition Support**: Child workflows, continue-as-new, and workflow chaining

### 8. Human Task Interface

```go
type HumanTask struct {
    ID          string
    AssignedTo  string
    Description string
    Deadline    time.Time
    Priority    Priority
}

type TaskQueue interface {
    GetPendingTasks() []HumanTask
    CompleteTask(taskID string, result interface{}) error
    EscalateTask(taskID string) error
}
```

## Integration Points

### CLI Integration
- `workflow list` - List all workflows
- `workflow start <name>` - Start workflow
- `workflow status <id>` - Check workflow status
- `workflow generate <template>` - Generate new workflow
- `workflow validate <path>` - Validate workflow

### Monitoring
- Temporal Web UI for workflow visualization
- Custom dashboards for business metrics
- Alert rules for stuck workflows

## Security Considerations
- mTLS for Temporal communication
- Role-based access control
- Audit logging for all workflow operations
- Encrypted activity payloads for sensitive data

## Performance Considerations
- Worker pool sizing
- Activity concurrency limits
- Workflow history size management
- Archival strategy for completed workflows

## Component Dependencies and Relationships

### Import Graph
```
cmd/workflows/commands/temporal.go
├── internal/temporal/validator.go (implemented)
│   ├── go.temporal.io/sdk/workflow
│   ├── go.temporal.io/sdk/contrib/tools/workflowcheck
│   └── internal/errors
├── internal/temporal/generator.go (new)
│   ├── internal/bpmn/types.go
│   └── text/template
├── internal/temporal/client.go (implemented)
│   ├── go.temporal.io/sdk/client
│   └── internal/temporal/config.go
├── internal/temporal/worker.go (implemented)
│   ├── go.temporal.io/sdk/worker
│   └── internal/temporal/registry.go
├── internal/temporal/registry.go (implemented)
│   ├── go.temporal.io/sdk/workflow
│   └── go.temporal.io/sdk/activity
└── internal/temporal/config.go (implemented)
    ├── go.temporal.io/sdk/client
    └── os/env variables
```

### Semantic Patterns
- **Validation Pattern**: Similar to `internal/mpc/validator.go` and `internal/bpmn/validator.go`
- **Generation Pattern**: Similar to `internal/bpmn/builder.go` 
- **CLI Command Pattern**: Similar to `cmd/workflows/commands/mpc.go`
- **Testing Pattern**: Similar to `internal/bpmn/*_test.go` files

## Validation Requirements

### Static Validation (Pre-deployment)
1. **Determinism Check**: All workflows must pass `workflowcheck` static analysis
2. **Type Safety**: Activity and signal signatures must be validated
3. **Graph Validation**: Workflow graph must be acyclic and connected
4. **Timeout Validation**: All activities must have reasonable timeouts
5. **Retry Policy Validation**: Retry policies must be bounded

### Runtime Validation (Test Environment)
1. **Integration Tests**: Full workflow execution in test environment
2. **Signal Testing**: Verify all signal handlers work correctly
3. **Query Testing**: Ensure queries return expected state
4. **Error Path Testing**: Test all failure scenarios
5. **Timeout Testing**: Verify timeout behavior

### Human Task Validation
1. **Assignment Validation**: Ensure tasks can be assigned
2. **Escalation Testing**: Verify escalation paths work
3. **Completion Testing**: Validate task completion flow
4. **Queue Testing**: Test task queue operations

## Temporal Client Implementation Details

### Client Features (internal/temporal/client.go)
- **Automatic Reconnection**: Exponential backoff retry mechanism for connection failures
- **Health Monitoring**: Periodic health checks every 30 seconds with status tracking
- **Connection States**: Tracks connected, disconnected, reconnecting states
- **Metrics Collection**: Built-in hooks for workflow/activity metrics
- **TLS Support**: Full TLS configuration for secure connections
- **Service Wrappers**: High-level wrappers for workflow and activity services
- **Graceful Shutdown**: Proper cleanup with context cancellation

### Worker Management (internal/temporal/worker.go)
- **Worker Pool**: Manages multiple workers across different task queues
- **Auto-Restart**: Automatic restart of failed workers with error tracking
- **Dynamic Scaling**: Add/remove workers at runtime based on load
- **Health Monitoring**: Per-worker health status and error tracking
- **Concurrent Management**: Safe concurrent operations with mutex protection
- **Graceful Stop**: Controlled shutdown with configurable timeout

### Registry System (internal/temporal/registry.go)
- **Type-Safe Registration**: Automatic type extraction for workflows and activities
- **Task Queue Organization**: Components organized by task queues
- **Builder Pattern**: Fluent API for easy registration setup
- **Metadata Support**: Additional metadata for each registered component
- **Snapshot Support**: Export registry state for monitoring/debugging
- **Dynamic Updates**: Runtime registration of new workflows/activities

### Configuration Management (internal/temporal/config.go)
- **Comprehensive Settings**: Client, worker, global, and development configurations
- **Environment Variables**: Support for environment-based configuration
- **Validation**: Built-in validation for all configuration values
- **TLS Configuration**: Full TLS/mTLS support with certificate management
- **Development Mode**: Special settings for local development
- **Worker Tuning**: Configurable concurrency, rate limits, and polling