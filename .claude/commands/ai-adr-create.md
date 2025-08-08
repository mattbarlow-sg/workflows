---
description: Create an Architecture Decision Record with AI assistance
allowed-tools: Bash(date:*),Bash(git:*),Bash(ls:*),Bash(head:*),Bash(echo:*),Bash(./workflows:*),Bash(grep:*)
---
# Context
- `CURRENT_IMPLEMENTATION_ID` is !`echo $CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- Review context in !`ls -al ai/${CURRENT_IMPLEMENTATION_ID}` `ai/<CURRENT_IMPLEMENTATION_ID>`.
- Recent ADRs: !`ls -t docs/adr/*.json 2>/dev/null | head -5 || echo "No existing ADRs"`

# Instructions
Invoke the @adr-advisor subagent with the provided and generated context.
