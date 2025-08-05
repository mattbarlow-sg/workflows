# Use Temporal for workflow orchestration

## Metadata

| Field | Value |
|-------|-------|
| ID | ADR-0001 |
| Status | proposed |
| Date | 2025-08-05 |
| Deciders | project-owner |
| Technical Story | ARCH-001 |
| Tags | architecture, workflow, temporal, golang, infrastructure |

## Context and Problem Statement

**Problem:** Need a robust workflow framework for a Golang project that supports
long-running workflows mixing AI and human tasks, with workflows spanning days to weeks
(e.g., learning workflows with daily lessons)

**Background:** Greenfield Golang project requiring a library of workflows accessible via
CLI. Workflows must support human-in-the-loop interactions and provide durability for
long-running processes. Need UI with three views: full view showing all workflows and step
statuses, network graph view of single workflow, and queue view showing tasks awaiting
human input

**Constraints:**
- Must integrate with Golang codebase
- Must support long-running workflows (days to weeks)
- Must handle human-in-the-loop tasks
- Must be accessible via CLI

**Assumptions:**
- Workflows will require dynamic routing capabilities
- System needs to handle task escalation
- Persistence and fault tolerance are critical requirements

## Considered Options

### Option 1: Temporal

Workflow orchestration framework with code-based definitions and extreme durability

### Option 2: Prefect

Python-based workflow automation platform primarily for ETL pipelines

### Option 3: State Machines

Traditional state machine implementation without code-based definitions

## Decision Outcome

**Chosen option:** Temporal

**Rationale:** Temporal provides extreme durability and fault tolerance for long-running workflows, uses code-based workflow definitions that fit well with Golang, supports dynamic routing for task escalation, and includes built-in persistence, retries, and versioning. It's specifically designed for human-in-the-loop workflows

## Consequences

### Positive

- Extreme durability and fault tolerance
- Code-based workflow definitions integrate naturally with Golang
- Dynamic routing enables flexible task escalation
- Built-in persistence and versioning
- Excellent support for human-in-the-loop patterns
- Handles retries and error recovery automatically
- Strong community and enterprise support

### Negative

- Learning curve for developers new to Temporal
- Additional infrastructure to deploy and maintain
- Potential complexity for simple workflows

### Neutral

- May need to integrate with other systems later
- Requires dedicated Temporal server deployment
