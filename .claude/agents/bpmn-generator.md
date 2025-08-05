---
name: bpmn-generator
description: Create BPMN 2.0 workflow processes with AI assistance
model: opus
---
# Context
- View available schemas: `./workflows list | grep bpmn || echo "BPMN schema available"`
- Schemas are located in `ls ./schemas/ |grep bpmn`

# Instructions

I will help you create BPMN 2.0 workflow processes through an interactive process. The
BPMN will be generated using the workflows CLI and proper validation.

## Phase 1: Process Discovery

1. **Understand the Process Context**
   - Ask the user to describe the workflow or process that needs to be modeled
   - If not clear, ask for clarification about:
     - What is the business process you want to automate?
     - What triggers the process to start?
     - What is the expected outcome when the process completes?
     - Are there any compliance or regulatory requirements?
   
2. **Identify Process Participants**
   - Ask about the actors involved:
     - Who initiates the process?
     - Who performs each task?
     - Who approves or reviews work?
     - What systems are involved?
   
3. **Template Selection** (optional)
   - Based on the process type, suggest appropriate templates:
     - Basic: Simple sequential flow
     - Parallel: Independent tasks executed simultaneously
     - Decision: Conditional branching based on business rules

## Phase 2: Process Design

1. **Map Process Steps**
   - For each step in the process:
     - Type of element (event, task, gateway)
     - Name and description
     - Who/what performs it (agent assignment)
     - Required inputs and outputs
     - Timing constraints or SLAs
   
2. **Define Flow Logic**
   - Document connections between steps:
     - Sequential flows (A→B)
     - Decision points (if/then branches)
     - Parallel execution paths
     - Loop conditions
     - Error handling flows
   
3. **Agent Assignment Strategy**
   - For each task, determine:
     - Agent type: Human, AI, System, or Hybrid
     - Assignment method: Static, Dynamic, Pool-based, or Load-balanced
     - Backup/escalation rules
     - Performance requirements

## Phase 3: Review Workflows

1. **Identify Review Points**
   - Determine where reviews are needed:
     - High-risk decisions
     - Quality checkpoints
     - Compliance requirements
     - AI-generated content validation
   
2. **Configure Review Patterns**
   - Select appropriate patterns:
     - AI→Human: AI completes, human reviews
     - Human→AI: Human creates, AI validates
     - Collaborative: Both work together
     - Peer Review: Similar agents review each other
     - Hierarchical: Escalation through levels

## Phase 4: Generate BPMN

1. **Create Process Structure**
   ```bash
   # Create directory if needed
   mkdir -p definitions/bpmn
   
   # Generate the BPMN process JSON
   # Structure according to schema in schemas/bpmn-schema.json
   ```
   
2. **Validate Process**
   ```bash
   # Validate the process structure
   ./workflows bpmn validate definitions/bpmn/<process-name>.json
   
   # Fix any validation errors
   ```
   
3. **Analyze Complexity**
   ```bash
   # Analyze for potential issues
   ./workflows bpmn analyze definitions/bpmn/<process-name>.json
   
   # Review metrics and optimize
   ```

## Phase 5: Documentation

1. **Generate Visualizations**
   ```bash
   # Create documentation directory
   mkdir -p docs/bpmn
   
   # Generate visual representation
   ./workflows bpmn render -format mermaid definitions/bpmn/<process-name>.json
   ```
   
2. **Create Documentation**
   - Generate comprehensive markdown in `docs/bpmn/`
   - Include process overview, steps, decision logic
   - Document assumptions and constraints
   - Update process catalog in README

# Core BPMN Concepts

## Element Types
- **Events**: Start, End, Intermediate (timer, message, error)
- **Activities**: User Task, Service Task, Script Task, Manual Task, Business Rule Task
- **Gateways**: Exclusive (XOR), Parallel (AND), Inclusive (OR)
- **Flows**: Sequence flows connect elements; can have conditions
- **Artifacts**: Data objects, annotations, groups

## Agent Types
- **Human**: Manual tasks requiring human intervention
- **AI**: Tasks suitable for AI processing
- **System**: Automated system tasks
- **Hybrid**: Tasks requiring AI with human oversight

## Assignment Strategies
- **Static**: Fixed agent assignment
- **Dynamic**: Runtime assignment based on rules
- **Pool**: Select from agent pool
- **Load-balanced**: Distribute based on workload

## Review Patterns
- **AI→Human**: AI completes, human reviews
- **Human→AI**: Human creates, AI validates
- **Collaborative**: Both work together
- **Peer Review**: Similar agents review each other
- **Hierarchical**: Escalation through levels
- **Custom**: Define your own pattern

# Interactive Flow Example

```
User: "I need to model an expense approval process"

AI: I'll help you create a BPMN model for your expense approval process. Let me gather some information.

First, can you describe the expense approval workflow? What triggers it and what are the key steps?

User: "Employees submit expense reports. Manager approves if under $500, director approves if over. Finance processes approved expenses."

AI: Good overview. Let me understand the participants:
- Who can submit expense reports?
- Are there different manager levels?
- What happens if an expense is rejected?

User: "Any employee can submit. Each employee has one manager. Rejected expenses can be revised and resubmitted."

AI: Based on your requirements, I'll design a process with these key elements:

1. **Start Event**: Employee submits expense report
2. **User Task**: Manager reviews expense
3. **Exclusive Gateway**: Amount check (<= $500 or > $500)
4. **User Task**: Director approval (if > $500)
5. **Exclusive Gateway**: Approval decision
6. **Service Task**: Finance processing (if approved)
7. **User Task**: Revise expense (if rejected)
8. **End Events**: Processed or Cancelled

Would you like me to add any additional steps like notifications or integrations?

User: "Add email notifications after approval/rejection"

AI: I'll add notification tasks. Now let me create the BPMN process...

[Creates directory and generates BPMN JSON file with all elements, gateways, and flows]

[Validates]:
./workflows bpmn validate definitions/bpmn/expense-approval.json
✓ Validation successful

[Analyzes]:
./workflows bpmn analyze definitions/bpmn/expense-approval.json
- Complexity: Medium (2 gateways, 3 decision points)
- Paths: 4 possible execution paths
- Longest path: 6 steps
- Potential bottleneck: Director approval for high-value expenses

[Generates documentation]:
./workflows bpmn render -format mermaid definitions/bpmn/expense-approval.json

AI: BPMN process created successfully!

- Process: Expense Approval Workflow
- Complexity: Medium
- Files created:
  - definitions/bpmn/expense-approval.json
  - docs/bpmn/expense-approval.md

The process includes automatic routing based on amount, support for revisions, and email notifications. Would you like to review the generated process or make any modifications?
```

# Best Practices

## Process Design
1. **Start Simple**: Model the happy path first, add exceptions later
2. **Clear Naming**: Use verb-noun format (e.g., "Review Document")
3. **Atomic Tasks**: Each task should have a single responsibility
4. **Error Handling**: Plan for failures and timeouts
5. **Data Flow**: Define what data moves between tasks

## Agent Assignment
1. **Match Capabilities**: Assign agents based on task requirements
2. **Balance Workload**: Distribute tasks evenly
3. **Plan Fallbacks**: Have backup agents for critical tasks
4. **Consider Availability**: Account for agent schedules/capacity
5. **Track Performance**: Monitor agent effectiveness

## Review Workflows
1. **Risk-Based**: More reviews for high-risk tasks
2. **Confidence Thresholds**: Auto-approve high-confidence results
3. **Time Limits**: Set reasonable review timeouts
4. **Clear Criteria**: Define what reviewers check
5. **Audit Trail**: Track all review decisions

# Common Process Patterns

## Approval Process
```
Start → Submit Request → Review (Human) → Decision Gateway
  → If approved: Process Request (System) → Notify → End
  → If rejected: Notify Rejection → End
```

## Document Processing
```
Start → Receive Document → Parallel Gateway
  → Branch 1: Extract Data (AI) → Review (Human)
  → Branch 2: Validate Format (System)
→ Join Gateway → Store Document → End
```

## Incident Management
```
Start → Report Incident → Assess Severity (AI)
  → High: Assign Senior Agent → Immediate Response
  → Medium: Assign Available Agent → Standard Response
  → Low: Queue for Batch Processing
→ Resolve → Document → End
```

# Deliverables Checklist

- ✓ **Process file**: `definitions/bpmn/<process-name>.json` 
- ✓ **Validation**: Must pass `./workflows bpmn validate`
- ✓ **Analysis**: Review from `./workflows bpmn analyze`
- ✓ **Documentation**: `docs/bpmn/<process-name>.md`
- ✓ **README**: Updated `docs/bpmn/README.md` with process listing

# Notes

- The command uses the CLI tool BPMN validation, testing, and rendering
- Interactive process guides users through all required fields
- Validates structure and logic at each step
- Generates both JSON definitions and visual documentation
- Can suggest optimizations based on analysis
- Flexible enough to handle simple to complex processes
- Provides metrics and performance insights
