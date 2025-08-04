# AI BPMN Process Creator

This command helps you create BPMN 2.0 workflow processes using the workflows CLI, ensuring proper validation and documentation.

## Important: CLI Integration

I will use the workflows CLI to:
1. Create BPMN processes using templates
2. Validate all generated processes
3. Analyze process complexity and potential issues
4. Generate documentation in markdown format

All BPMN files will be:
- Created in `definitions/bpmn/` directory
- Validated using `./workflows bpmn validate`
- Analyzed using `./workflows bpmn analyze`
- Documented in `docs/bpmn/` as markdown

## Core Concepts

### BPMN Elements
- **Events**: Start, End, Intermediate (timer, message, error)
- **Activities**: User Task, Service Task, Script Task, Manual Task, Business Rule Task
- **Gateways**: Exclusive (XOR), Parallel (AND), Inclusive (OR)
- **Flows**: Sequence flows connect elements; can have conditions
- **Artifacts**: Data objects, annotations, groups

### Agent Types
- **Human**: Manual tasks requiring human intervention
- **AI**: Tasks suitable for AI processing
- **System**: Automated system tasks
- **Hybrid**: Tasks requiring AI with human oversight

### Assignment Strategies
- **Static**: Fixed agent assignment
- **Dynamic**: Runtime assignment based on rules
- **Pool**: Select from agent pool
- **Load-balanced**: Distribute based on workload

### Review Patterns
- **AI→Human**: AI completes, human reviews
- **Human→AI**: Human creates, AI validates
- **Collaborative**: Both work together
- **Peer Review**: Similar agents review each other
- **Hierarchical**: Escalation through levels
- **Custom**: Define your own pattern

## Process Templates

While the workflows CLI doesn't provide a `new` command, I'll help you create processes based on common templates:

### Basic Template
Creates: Start → User Task → End
Structure: Simple sequential flow for basic processes

### Parallel Template  
Creates: Start → Split → (Task A || Task B) → Join → End
Structure: Parallel execution of independent tasks

### Decision Template
Creates: Start → Review → Decision → (Approved | Rejected) → End
Structure: Conditional branching based on decisions

## Process Creation Steps

### 1. Process Overview
First, I'll ask about:
- Process name and description
- Business context and goals
- Expected process complexity
- Performance requirements
- Which template best fits (if any)

### 2. Element Definition
For each process step:
- Element type (event, activity, gateway)
- Name and description
- Agent assignment (who does this?)
- Review requirements
- Data inputs/outputs
- Timing constraints

### 3. Flow Logic
Define connections:
- Sequential flows (A→B)
- Conditional branches (if/then)
- Parallel execution (AND split/join)
- Alternative paths (OR/XOR)
- Loop conditions
- Error handling

### 4. Agent Assignment
For each task:
- Agent type and capabilities
- Assignment strategy
- Backup/escalation rules
- Workload considerations
- Performance metrics

### 5. Review Workflows
Where needed:
- Review trigger conditions
- Reviewer selection
- Approval criteria
- Timeout handling
- Escalation paths

### 6. Validation
I'll validate:
- Process structure (start/end events)
- Element connectivity
- Gateway logic
- Agent assignments
- Review workflows
- Potential deadlocks
- Performance bottlenecks

### 7. Analysis
I'll provide:
- Process metrics
- Complexity analysis
- Path analysis
- Workload distribution
- Optimization suggestions

## Best Practices

### Process Design
1. **Start Simple**: Begin with main path, add exceptions later
2. **Clear Naming**: Use verb-noun format (e.g., "Review Document")
3. **Atomic Tasks**: Each task should have single responsibility
4. **Error Handling**: Plan for failures and timeouts
5. **Data Flow**: Define what data moves between tasks

### Agent Assignment
1. **Match Capabilities**: Assign agents based on task requirements
2. **Balance Workload**: Distribute tasks evenly
3. **Plan Fallbacks**: Have backup agents for critical tasks
4. **Consider Availability**: Account for agent schedules/capacity
5. **Track Performance**: Monitor agent effectiveness

### Review Workflows
1. **Risk-Based**: More reviews for high-risk tasks
2. **Confidence Thresholds**: Auto-approve high-confidence results
3. **Time Limits**: Set reasonable review timeouts
4. **Clear Criteria**: Define what reviewers check
5. **Audit Trail**: Track all review decisions

### Common Patterns

#### Approval Process
```
Start → Submit Request → Review (Human) → Decision Gateway
  → If approved: Process Request (System) → Notify → End
  → If rejected: Notify Rejection → End
```

#### Document Processing
```
Start → Receive Document → Parallel Gateway
  → Branch 1: Extract Data (AI) → Review (Human)
  → Branch 2: Validate Format (System)
→ Join Gateway → Store Document → End
```

#### Incident Management
```
Start → Report Incident → Assess Severity (AI)
  → High: Assign Senior Agent → Immediate Response
  → Medium: Assign Available Agent → Standard Response
  → Low: Queue for Batch Processing
→ Resolve → Document → End
```

## Example Prompts

### Basic Process
"Create a simple expense approval process where employees submit expenses, managers review, and finance processes approved expenses."

### Complex Process
"Design a content creation workflow where AI generates initial content, human editors review and refine, then AI does final formatting before publication. Include quality checks and revision loops."

### Dynamic Assignment
"Build an incident response process that assigns agents based on incident type, severity, and current agent workload. Include escalation if not resolved within SLA."

## Interactive Creation

I'll guide you through creating your process step by step:

1. **Describe your process** - What are you trying to automate?
2. **Identify key steps** - What needs to happen?
3. **Assign agents** - Who does what?
4. **Add logic** - Decisions, parallel work, loops?
5. **Define reviews** - What needs checking?
6. **Validate design** - Let me check for issues
7. **Generate process** - Create the BPMN file

## My Workflow

When creating a BPMN process, I will:

### 1. Initial Creation
```bash
# Create directory if needed
mkdir -p definitions/bpmn

# Create the BPMN process JSON file manually
# I'll help you structure it according to the schema
# Save to: definitions/bpmn/<process-name>.json
```

### 2. Validation
```bash
# Validate the process
./workflows bpmn validate definitions/bpmn/<process-name>.json

# If errors, fix and re-validate until clean
```

### 3. Analysis
```bash
# Analyze for complexity and issues
./workflows bpmn analyze definitions/bpmn/<process-name>.json

# Review metrics and optimize if needed
```

### 4. Documentation
```bash
# Create documentation directory
mkdir -p docs/bpmn

# Generate visual representation if needed
./workflows bpmn render -format dot definitions/bpmn/<process-name>.json > definitions/bpmn/<process-name>.dot
# Or use mermaid format for markdown embedding
./workflows bpmn render -format mermaid definitions/bpmn/<process-name>.json

# Create comprehensive markdown documentation in docs/bpmn/
```

### 5. Deliverables
- ✓ **Process file**: `definitions/bpmn/<process-name>.json` 
- ✓ **Validation**: Must pass `./workflows bpmn validate`
- ✓ **Analysis**: Review from `./workflows bpmn analyze`
- ✓ **Documentation**: `docs/bpmn/<process-name>.md`
- ✓ **README**: Updated `docs/bpmn/README.md` with process listing

## Tips

- Start with the happy path, add exceptions later
- Use meaningful IDs (e.g., "task_review_expense" not "task1")
- Consider both human and system perspectives
- Plan for scale - will this work with 10x volume?
- Think about monitoring and metrics needs
- Document assumptions and constraints

Ready to create your BPMN process? Describe what you want to build!