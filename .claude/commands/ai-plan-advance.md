---
description: Chooses the next step
allowed-tools: Bash(echo:*),Bash(ls:*),Bash(./workflows:*)
---
# Context
- `CURRENT_IMPLEMENTATION_ID` is !`echo $CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- If `echo ${CURRENT_IMPLEMENTATION_ID}` is not set, then set it.

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
- Run discovery on the plan: !`./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml` `./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml`
- If the next step is BPMN generation, stop and inform the user to run the /ai-bpmn-create command with the node contents.
- If the next step is Spec generation, stop and inform the user to run the /ai-spec-create command with the node contents.
