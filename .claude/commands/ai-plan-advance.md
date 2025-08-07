---
description: Chooses the next step
allowed-tools: Bash(echo:*),Bash(ls:*),Bash(./workflows:*)
---
- `CURRENT_WORK_SESSION` is !`echo $CURRENT_WORK_SESSION` as set in the `.envrc` file.
- If `echo ${CURRENT_WORK_SESSION}` is not set, then set it.

## Instructions
- Run discovery on the plan: !`./workflows mpc discover --next-only ai/${CURRENT_WORK_SESSION}/plan.yaml` `./workflows mpc discover --next-only ai/${CURRENT_WORK_SESSION}/plan.yaml`
- If the next step is BPMN generation, stop and inform the user to run the /ai-bpmn-create command with the node contents.
