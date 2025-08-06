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
│   ├── cli/              # CLI commands
│   └── worker/           # Temporal worker
├── pkg/
│   ├── workflows/        # Workflow definitions
│   │   ├── base/        # Base workflow interfaces
│   │   ├── library/     # Reusable workflows
│   │   └── generated/   # Generated workflows
│   ├── activities/       # Activity implementations
│   ├── signals/         # Signal handlers
│   ├── queries/         # Query handlers
│   ├── generator/       # Workflow generator
│   └── validator/       # Workflow validator
├── internal/
│   ├── temporal/        # Temporal client setup
│   └── config/          # Configuration
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

### 7. Generator Templates

```go
// Base workflow template
type {{ .WorkflowName }}Workflow struct {
    // Workflow state
}

func (w *{{ .WorkflowName }}Workflow) Execute(ctx workflow.Context, input {{ .InputType }}) ({{ .OutputType }}, error) {
    // Generated workflow logic
}
```

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
├── internal/temporal/validator.go (new)
│   ├── go.temporal.io/sdk/workflow
│   ├── go.temporal.io/sdk/contrib/tools/workflowcheck
│   └── internal/errors
├── internal/temporal/generator.go (new)
│   ├── internal/bpmn/types.go
│   └── text/template
├── internal/temporal/client.go (new)
│   ├── go.temporal.io/sdk/client
│   └── internal/config
└── internal/temporal/worker.go (new)
    ├── go.temporal.io/sdk/worker
    └── internal/temporal/workflows/*
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