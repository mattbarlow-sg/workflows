---
description: Begin ai assisted workflow for a work session
allowed-tools: Bash(date:*),Bash(ls:*)
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
- In this session you will only edit `plan.md` and `architecture.md`. Do not edit the other ones.
## Discovery
- Analyze the request and list the related components.
- Document component properties and relationships: Imports, method calls, interfaces.
- Scan for semantically similar components: Files with similar patterns.
- Search for related documentation online.
- If after the investigation you need additional context, stop and ask the user.
- Document the results of your discovery to `ai/<work-session-id>/architecture.md`.
## Plan
- Separate the plan into discrete steps prioritizing parallelization.
- Write the plan to the `ai/<work-session-id>/plan.md`.
- Each step in `ai/<work-session-id>/plan.md` should be assigned either to an ai or to a human.
- Stop when the plan is generated. Do not start implementation.

# Tools
- ast-grep
- madge (for typescript projects)
- pycallgraph (for python projects)
