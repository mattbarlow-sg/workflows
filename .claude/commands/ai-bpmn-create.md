---
description: Guide the user through an interactive workshop for creating BPMN context.
allowed-tools: Bash(date:*),Bash(git:*),Bash(ls:*),Bash(head:*),Bash(echo:*),Bash(./workflows:*),Bash(grep:*)
---

# Context
- View available schemas: `./workflows list | grep bpmn`
- Schemas are located in `ls ./schemas/ |grep bpmn`
- bpmn related commands are `./workflows bpmn -h`
- `CURRENT_IMPLEMENTATION_ID` is `echo $CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- If `echo ${CURRENT_IMPLEMENTATION_ID}` is not set, then set it.
- Read the `docs/bpmn-concepts.md` document.
- Run discovery on the plan: !`./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml` `./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml`

# Instructions
- You will not create or edit files directly.
- This work session is NOT about completing work in the plan. This is an interactive workshop to provide the data necessary for BPMN generation.
- Help the user design BPMN 2.0 workflows through an interactive process for a given node.
**IMPORTANT**: This is an interactive process. You MUST ask questions and get your input before generating any BPMN.

## Phase 1: Process Discovery

1. **Understand the Process Context**
   - **ALWAYS ASK QUESTIONS FIRST** before generating BPMN, even if context seems clear:
     - What is the business process you want to automate?
     - What triggers the process to start?
     - What is the expected outcome when the process completes?
     - Are there any compliance or regulatory requirements?
     - Should validation steps run in parallel or sequential?
     - What are the failure handling requirements?
     - What are appropriate timeout values?
   
2. **Identify Process Participants**
   - Ask about the actors involved:
     - Who initiates the process?
     - Who performs each task?
     - Who approves or reviews work?
     - What systems are involved?
   
3. **Confirm Understanding Before Generation**
   - **MANDATORY**: Summarize your understanding and get confirmation:
     - "Based on what you've shared, I understand the process as..."
     - "Before I generate the BPMN, please confirm these assumptions..."
     - Wait for user confirmation before proceeding
   
4. **Template Selection** (optional)
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

Invoke the @bpmn-generator subagent with the provided and generated context.

# Notes
- Interactive process guides users through all required fields
- Validates structure and logic at each step
- Can suggest optimizations based on analysis
- Flexible enough to handle simple to complex processes
