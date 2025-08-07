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
