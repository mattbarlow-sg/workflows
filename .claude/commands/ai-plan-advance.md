---
description: Chooses the next step
allowed-tools: Bash(echo:*),Bash(ls:*),Bash(./workflows:*)
---
# Context
- `CURRENT_IMPLEMENTATION_ID` is !`echo $CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- If `echo ${CURRENT_IMPLEMENTATION_ID}` is not set, then set it.
- The cli can validate and discover information about the plan, but you must make changes to the `plan.yaml` file directly.
- Always validate the plan file after making changes to it.

# Node Requirements
1. BPMN (Stage 1): Required when node involves:
  - Multi-step processes with decision points
  - State transitions
  - User/system interactions
  - Workflow orchestration
2. Formal Properties (Stage 2): Required when node has:
  - Invariants to maintain
  - State machines
  - Business rules
  - Data transformations
3. Contracts/Schemas (Stage 3): Always beneficial for:
  - API boundaries
  - Data validation
  - Interface definitions

## Instructions
### Decide Next Step
- Run discovery on the plan: !`./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml` `./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml`
- The command will tell you the next required step and artifacts, and you must make a judgement call on required artifacts or work to be done:

## Choose Path
- If nodes are READY TO WORK ON NOW, STOP and inform the user to run @.claude/commands/ai-nodes-implement.md
- If you judge that BPMN is not required for this node type, create a bpmn artifact with a null value.
- If you judge that BPMN is required, STOP and inform the user to run @.claude/commands/ai-bpmn-create.md.
- If you judge that Formal Spec is not required for this node type, create a formal_spec artifact with a null value.
- If you judge that Formal Spec is required, STOP and inform the user to run @.claude/commands/ai-spec-create.md.
- If you judge that Schemas are not required for this node type, create a schemas artifact with a null value.
- If you judge that Schemas are required, STOP and inform the user to run @.claude/commands/ai-spec-create.md.
