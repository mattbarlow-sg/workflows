---
description: Begin ai assisted workflow for a work session
allowed-tools: Bash(echo:*),Bash(ls:*)
---

# Context
- `CURRENT_IMPLEMENTATION_ID` is !`echo $CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- If `echo ${CURRENT_IMPLEMENTATION_ID}` is not set, then set it.
- The cli can validate and discover information about the plan, but you must make changes to the `plan.yaml` file directly.
- Always validate the plan file after making changes to it.

# Instructions
## Call The Subagent
- View the next nodes ready to work in !`./workflows mpc discover --next-only ai/${CURRENT_IMPLEMENTATION_ID}/plan.yaml`
- Call the node-implementer subagent with the provided and generated context.

## When The Subagent Returns
- If necessary, update the plan based on information you learned, and based on the input from human validation.
- As you learn reference information, such as related schemas and documentation URLs, update `architecture.md`.
- Update `plan.yaml` with status updates.
- Summarize the work that was performed by appending to `log.md`.
- When finished with the entire plan, add manual test instructions to `summary.md`.
