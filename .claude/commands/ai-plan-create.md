---
description: Begin ai assisted workflow for a work session
allowed-tools: Bash(date:*),Bash(ls:*)
---

# Context
- `work-session-id` is !`date +%s`-`<three-word-description-of-provided-context>`

# Instructions
## Setup
- Run `export CURRENT_IMPLEMENTATION_ID=<work-session-id>`.
- Update the `CURRENT_IMPLEMENTATION_ID` value in `.envrc`. Replace it if it is already set.
- Create a directory called `ai/<work-session-id>`.
- In this directory, create four files: `plan.yaml`, `log.md`, `architecture.md`, and `summary.md`.
- Call the subagent @plan-generator with all the provided and generated context.
