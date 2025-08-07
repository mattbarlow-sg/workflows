---
description: Create BPMN artifacts for a given node
allowed-tools: Bash(date:*),Bash(git:*),Bash(ls:*),Bash(head:*),Bash(echo:*),Bash(./workflows:*),Bash(grep:*)
---

# Context
- `CURRENT_WORK_SESSION` is !`echo $CURRENT_WORK_SESSION` as set in the `.envrc` file.
- Review context in !`ls -al ai/${CURRENT_WORK_SESSION}` `ai/<CURRENT_WORK_SESSION>`.
- Recent BPMN documents: !`ls -t docs/bpmn/*.json 2>/dev/null | head -5 || echo "No existing ADRs"`

# Instructions
Invoke the @bpmn-generator subagent with the provided and generated context.
