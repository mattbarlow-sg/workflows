---
name: plan-generator
description: Begin ai assisted workflow for a work session
model: opus
---

# Context
- `CURRENT_WORK_SESSION` is !`echo $CURRENT_WORK_SESSION` as set in the `.envrc` file.
- If `CURRENT_WORK_SESSION` is not set as an environment variable, export it according to the value in the `.envrc`.
- Past work sessions are !`ls ./ai`.
- Work Session Directory is `ai/<CURRENT_WORK_SESSION>`.


# Instructions
- In this session you will only edit `plan.yaml` and `architecture.md`. Do not edit other files.
## Discovery
- Analyze the request and list the related components.
- Document component properties and relationships: Imports, method calls, interfaces.
- Scan for semantically similar components: Files with similar patterns.
- Search for related documentation online.
- If after the investigation you need additional context, stop and ask the user.
- Document the results of your discovery, including upstream documentation links, to `ai/<work-session-id>/architecture.md`.
## Plan
- Create the plan according to the MPC specification `schemas/mpc.json`
- Do not add artifacts to any nodes. These will be generated in the future.
- Set materialization scores for initial planning:
  - All nodes start at 0.1 (initial planning stage - identified but flexible)
  - Materialization will increase as artifacts are added:
    - BPMN diagrams: 0.2-0.3
    - Specifications: 0.4-0.5
    - Tests: 0.6-0.7
    - Implementation: 0.8-0.9
    - Production: 1.0
- Write the plan to the `ai/<work-session-id>/plan.yaml`.
- Validate the plan with `./workflows mpc validate ai/<work-session-id>/plan.yaml`
- Ensure the structure of the plan looks correct with `./workflows mpc discover ai/<work-session-id>/plan.yaml`
- Stop when the plan is generated. Do not start implementation.

# Tools
- ast-grep
- madge (for typescript projects)
- pycallgraph (for python projects)
