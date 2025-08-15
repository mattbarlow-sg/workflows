# Human Task Lifecycle Management System

## Overview
A comprehensive human-in-the-loop task management system for Temporal workflows that handles approvals, reviews, data entry, decisions, and tasks that AI cannot complete independently.

## Process Description
This BPMN process defines a complete task lifecycle management system with the following key characteristics:
- **Single Operator Model**: Currently supports one human operator role
- **LIFO Queue Processing**: Last-in-first-out task processing strategy
- **Long Timeout Periods**: Weeks before escalation (14 days default)
- **No Queue Jumping**: Escalated tasks maintain their position in the LIFO queue
- **CLI Interface**: Command-line interface for operator interaction (future UI support planned)
- **Free Text Proof**: Completion proof is captured as unstructured text

## Main Process Flow

### Task Creation and Queuing
1. **Start Event**: Task Creation Request triggered by:
   - Workflow requirements
   - Scheduled events
   - Error conditions requiring human intervention

2. **Create Task Record**: System generates:
   - Unique task identifier
   - Task metadata (type, priority, creation timestamp)
   - Task context from requester

3. **Add to LIFO Queue**: Task is added to the single, shared LIFO queue for all task types

### Operator Review and Decision
4. **Operator Reviews Task**: Human operator:
   - Accesses task via CLI
   - Views task details and context
   - Makes decision on task outcome

5. **Operator Decision Gateway**: Five possible paths:
   - **Complete**: Task is finished with proof
   - **Approve**: Task is approved
   - **Reject**: Task is rejected
   - **Return for Clarification**: Needs more information
   - **Cancel**: Task is cancelled

### Task Resolution Paths

#### Completion Path
- Capture proof of completion (free text)
- Signal task completion to workflow
- Update task status
- Send response to requesting workflow

#### Approval Path
- Signal approval to workflow
- Update task status
- Send response to requesting workflow

#### Rejection Path
- Signal rejection to workflow
- Update task status
- Send response to requesting workflow

#### Clarification Loop
- Send request back to original requester
- Wait for clarification response
- Task returns to LIFO queue when clarification received
- Operator reviews updated task

#### Cancellation Path
- Signal cancellation to workflow
- Update task status
- Send response to requesting workflow

## Escalation Sub-Process
- **Timer Boundary Event**: Attached to operator review task
- **Timeout Period**: 14 days (2 weeks)
- **Non-interrupting**: Task remains with operator
- **Escalation Actions**:
  - Mark task as escalated
  - Update task in queue (maintains LIFO position)
  - Continue with normal operator handling

## Query Sub-Process
Separate process flow for CLI queries:
1. **CLI Query Request**: Operator initiates queue status check
2. **Fetch Queue State**: Retrieve current queue contents
3. **Format Task List**: Prepare data for CLI display
4. **Return Results**: Display queue information to operator

## Data Objects

### Task Record
```json
{
  "id": "unique-task-id",
  "type": "approval|review|data-entry|decision",
  "status": "pending|in-progress|completed|rejected|cancelled",
  "context": {}, // Task-specific data
  "proof": "free text proof of completion",
  "creation_time": "ISO-8601 timestamp",
  "escalated": false
}
```

### Queue State
```json
{
  "tasks": [], // Array of tasks in LIFO order
  "operator_status": "available|busy|offline"
}
```

### Task Response
```json
{
  "task_id": "unique-task-id",
  "outcome": "completed|approved|rejected|cancelled",
  "proof": "optional proof text",
  "timestamp": "ISO-8601 timestamp"
}
```

## Agent Assignments

### Human Agent (Operator)
- **Role**: Single operator for all tasks
- **Capabilities**:
  - Review tasks
  - Make decisions
  - Complete tasks
  - Generate proof
- **Access**: CLI interface (future UI planned)
- **Availability**: Manual polling (no push notifications)

### System Agent (Temporal)
- **Role**: Automated task management
- **Capabilities**:
  - Create and store tasks
  - Manage LIFO queue
  - Process signals
  - Track status
  - Handle timeouts
- **Availability**: Always available

## Process Metrics
- **Complexity Score**: 46
- **Maximum Path Length**: 12 steps
- **Minimum Path Length**: 4 steps
- **Average Path Length**: 9.8 steps
- **Loop Detection**: 1 (clarification loop)
- **Total Elements**: 24
  - Events: 5
  - Activities: 16
  - Gateways: 3
  - Sequence Flows: 26

## Implementation Notes

### Queue Management
- Single LIFO queue for all task types
- No task prioritization beyond LIFO ordering
- Escalated tasks don't jump the queue
- Tasks maintain their position even after escalation

### Timeout Handling
- Long timeout periods (weeks) before escalation
- Non-interrupting boundary events
- Escalation is informational only
- Operator continues to handle escalated tasks normally

### Clarification Process
- Tasks can be returned to requester for more information
- Clarification requests loop back to the queue
- Original task context is preserved
- Updated information is added to task

### Future Enhancements
- Web UI for operator interface
- Multiple operator support
- Task prioritization mechanisms
- Push notifications for operators
- Partial task completion
- Multiple escalation levels
- Advanced queue management strategies

## Process Visualization
The process can be visualized using the Mermaid diagram in `human-task-system.mermaid` or by running:
```bash
./workflows bpmn render -format mermaid definitions/bpmn/human-task-system.json
```

## Validation and Analysis
To validate the process definition:
```bash
./workflows bpmn validate definitions/bpmn/human-task-system.json
```

To analyze process complexity:
```bash
./workflows bpmn analyze definitions/bpmn/human-task-system.json
```