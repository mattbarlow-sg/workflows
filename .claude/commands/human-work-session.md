---
description: Begin manual workflow for a work session
allowed-tools: Bash(date:*)
---

# Context
- Past work sessions: !`ls ./ai`
- `work-session-id` is !`date +%s`-`<three-word-description>`

# Instructions
## Setup
- Run `export CURRENT_WORK_SESSION=<work-session-id>`.
- Update the `CURRENT_WORK_SESSION` value in `.envrc`. Replace it if it is already set.
- Based on the provided context, create a directory called `ai/<work-session-id>`.
- In this directory, create four files: `plan.md`, `log.md`, `architecture.md`, and `summary.md`.
## Plan
- Ask the user for context about the plan if it was not provided.
- Separate the plan into discrete steps prioritizing parallelization.
- Each step in `ai/<work-session-id>/plan.md` should be assigned to a human.
- Suggest three `plan`s for the user and ask them to select.
- Write the selected plan to the `ai/<work-session-id>/plan.md`.
- Stop when the plan is generated. Do not start implementation.
