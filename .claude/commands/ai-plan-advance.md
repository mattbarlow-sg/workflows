---
description: Chooses the next step
allowed-tools: Bash(echo:*),Bash(ls:*),Bash(./workflows:*)
---
- `CURRENT_WORK_SESSION` is !`echo $CURRENT_WORK_SESSION` as set in the `.envrc` file.
- If `echo ${CURRENT_WORK_SESSION}` is not set, then set it.

## Instructions
- Run discovery on the plan: !`./workflows mpc discover --next-only ai/${CURRENT_WORK_SESSION}/plan.yaml` `./workflows mpc discover --next-only ai/${CURRENT_WORK_SESSION}/plan.yaml`
- If the next step is to generate BPMN for a node, call the @bpmn-generator subagent with the provided and generated context.
- If the next step is to generate specs for a node, call the @spec-generator subagent with the provided and generated context.
