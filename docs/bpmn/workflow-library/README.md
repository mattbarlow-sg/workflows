# Temporal Workflow Library - BPMN 2.0 Patterns

This directory contains comprehensive BPMN 2.0 designs for reusable Temporal workflow patterns. These patterns serve as foundational building blocks for creating complex, scalable workflow implementations.

## Overview

The workflow library provides five core patterns that can be composed and extended to handle virtually any workflow scenario. Each pattern is designed with durability, error handling, and composability in mind.

## Workflow Patterns

### 1. [Approval Workflow Pattern](./approval-workflow.md)
**File**: `definitions/bpmn/workflow-library/approval-workflow.json`

A comprehensive pattern for human approval of AI-generated work with automatic retry capabilities.

**Key Features**:
- Human review and approval process
- Automatic retry on rejection with AI-powered fixes
- Configurable retry limits
- Timeout handling for review tasks
- Complete audit trail

**Complexity Metrics**:
- Complexity Score: 21
- Path Depth: 6
- Contains 1 loop for retry logic
- 11 total elements (4 events, 5 activities, 2 gateways)

### 2. [Scheduled Task Workflow Pattern](./scheduled-task-workflow.md)
**File**: `definitions/bpmn/workflow-library/scheduled-task-workflow.json`

Executes tasks on schedule or event triggers with sophisticated retry and catch-up logic.

**Key Features**:
- Cron-based scheduling
- Event-driven triggers
- Missed schedule handling with catch-up strategies
- Configurable retry logic (5xx vs 4xx errors)
- Parallel execution support

**Complexity Metrics**:
- Complexity Score: 36
- Path Depth: 13
- Contains 2 loops for scheduling and retry
- 17 total elements (5 events, 6 activities, 6 gateways)

### 3. [Human Task with Escalation Pattern](./human-escalation-workflow.md)
**File**: `definitions/bpmn/workflow-library/human-escalation-workflow.json`

AI agents escalate to human operators when encountering unknowns or requiring human judgment.

**Key Features**:
- Confidence-based escalation
- Hierarchical escalation levels (L1, L2, L3)
- Comprehensive diagnostic information capture
- Learning record system for future improvements
- Timeout handling with further escalation

**Complexity Metrics**:
- Complexity Score: 32
- Path Depth: 12
- Contains 1 loop for validation retry
- 16 total elements (5 events, 7 activities, 4 gateways)

### 4. [Long-Running Educational Workflow Pattern](./educational-workflow.md)
**File**: `definitions/bpmn/workflow-library/educational-workflow.json`

Manages educational processes over weeks/months with lesson delivery, assessment, and progress tracking.

**Key Features**:
- Dynamic curriculum generation and adaptation
- Quiz generation and evaluation
- Progress tracking and checkpointing
- Pause/resume capabilities
- Durable state management for system restarts
- Performance-based curriculum adaptation

**Complexity Metrics**:
- Complexity Score: 50 (highest)
- Path Depth: 17 (deepest)
- Contains 3 loops for learning iterations
- 24 total elements (6 events, 12 activities, 6 gateways)

### 5. [Workflow Composition Patterns](./composition-patterns.md)
**File**: `definitions/bpmn/workflow-library/composition-patterns.json`

Patterns for composing workflows including parent-child, signaling, and chaining.

**Key Features**:
- Parent-child workflow spawning
- Inter-workflow signaling and messaging
- Workflow chains (sequential execution)
- Synchronous and asynchronous execution modes
- Parallel workflow coordination
- Result aggregation

**Complexity Metrics**:
- Complexity Score: 50
- Path Depth: 10
- Contains 1 loop for chain continuation
- 24 total elements (6 events, 11 activities, 7 gateways)

## Agent Types

All workflows support three primary agent types:

1. **AI Agents**: Automated task execution using models like Claude-3-Opus
2. **Human Agents**: Manual tasks requiring human intervention
3. **System Agents**: Infrastructure and orchestration tasks

## Common Features

### Error Handling
- All patterns include comprehensive error handling
- Retry logic with configurable limits
- Error escalation paths
- Fallback strategies

### Monitoring & Alerting
- Built-in metrics collection
- Configurable alerting rules
- Performance tracking for all agent types
- Execution time monitoring

### State Management
- Durable state persistence
- Checkpoint and restore capabilities
- Support for long-running processes
- System restart resilience

## Usage Guidelines

### When to Use Each Pattern

| Pattern | Use When |
|---------|----------|
| Approval | AI work needs human verification before proceeding |
| Scheduled Task | Tasks need to run on a schedule or be triggered by events |
| Human Escalation | AI needs help with uncertain or complex decisions |
| Educational | Managing long-term learning processes with progress tracking |
| Composition | Orchestrating multiple workflows or creating workflow hierarchies |

### Composition Examples

#### Example 1: AI Code Generation with Review
```
Composition Pattern (parent) 
  → AI Task (child workflow)
  → Approval Workflow (child workflow)
  → Deployment (child workflow)
```

#### Example 2: Scheduled Learning Sessions
```
Scheduled Task Pattern
  → Educational Workflow (triggered on schedule)
  → Progress Report Generation (end of week)
```

#### Example 3: Complex Task with Escalation
```
Human Escalation Pattern
  → If escalated: Approval Workflow
  → Record Learning
  → Update AI Model
```

## Implementation Notes

### Temporal Integration
All patterns are designed for Temporal workflow engine implementation with:
- Activity-based task execution
- Signal and query support
- Child workflow spawning
- Durable timers
- State persistence

### Extension Points
Each pattern provides extension points for:
- Custom agent assignment strategies
- Additional validation rules
- Custom metrics and monitoring
- Integration with external systems

## Performance Considerations

### Scalability
- Patterns support parallel execution where applicable
- Configurable pool sizes for agent assignment
- Efficient state checkpointing

### Resource Usage
- Minimal state storage through selective checkpointing
- Lazy loading of workflow context
- Efficient signal propagation

## Best Practices

1. **Start Simple**: Begin with a single pattern and compose as needed
2. **Monitor Early**: Enable monitoring from the start to understand behavior
3. **Test Escalation Paths**: Ensure escalation and error paths are well-tested
4. **Document Customizations**: Keep track of pattern modifications
5. **Version Control**: Version your workflow definitions for rollback capability

## Validation and Testing

All workflows can be validated using:
```bash
./workflows bpmn validate definitions/bpmn/workflow-library/<pattern>.json
```

Analyze complexity and paths:
```bash
./workflows bpmn analyze definitions/bpmn/workflow-library/<pattern>.json
```

## Future Enhancements

Planned improvements to the workflow library:
- Additional composition patterns (scatter-gather, saga)
- Enhanced learning and adaptation capabilities
- Multi-tenant support
- Advanced scheduling strategies
- Integration with external workflow systems

## Contributing

When adding new patterns:
1. Follow the existing BPMN 2.0 structure
2. Include comprehensive error handling
3. Add monitoring and alerting configurations
4. Document all agent requirements
5. Provide usage examples

## Support

For questions or issues with these workflow patterns, please refer to the individual pattern documentation or contact the workflow platform team.