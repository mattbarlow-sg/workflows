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
- Export `CURRENT_IMPLEMENTATION_ID` as an environment variabl3.
- Create a directory called `ai/<CURRENT_IMPLEMENTATION_ID>`.
- In this directory, create four files: `plan.yaml`, `log.md`, `architecture.md`, and `summary.md`.
## Design Of The Plan
- Examine the context provided by the user.
- It the provided context enough to create the plan?
- If so, generate the plan but don't save anything. Store it in your local context.
- **Important** Look at the codebase. Use the tree-stter MCP to assst. Does this plan require an overall BPNM specification?
### When BPMN is Required at Global MPC Plan Level
**BPMN Required for:**
- **Cross-functional processes** - Multiple teams/systems with handoffs
- **Regulated workflows** - Compliance requirements needing audit trails
- **Complex orchestration** - Multiple services with conditional routing
- **Event-driven architectures** - Async event flows between components
- **Integration patterns** - External system integrations with retries/fallbacks
- **Business process automation** - Multi-step approval flows or state machines

**BPMN NOT Required for:**
- Simple CRUD applications
- Single-service implementations
- Linear data transformations
- Stateless request/response APIs
- Pure computational tasks
- UI-only features

**Indicators to Use BPMN:**
- 3+ decision points in the workflow
- Multiple actors/systems involved
- Async operations with callbacks
- Error handling/compensation logic
- Time-based constraints or SLAs
- Parallel processing branches
## BPMN Decision
- If you decide BPMN is required, read the file `.claude/commands/ai-bpmn-create.md` and follow the interactive instructions.
- If you decide BPMN is not required, read the file `.claude/commands/ai-plan-create.md` and follow the interactive instructions.
## Mising Context
- If context is missing to satisfy the requirements for @schemas/mpc.json, you are a
  workshop coordinator. As the user for data necessary to formulate the plan. Ask
  questions a few at a tiome until you have enough context to satisfy the scheme.
- Invoke the @plan-generator subagent to complete implementaton of the plan.
- STOP after the subagent has completed.
