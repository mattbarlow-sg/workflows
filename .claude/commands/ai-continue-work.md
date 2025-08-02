---
description: Begin ai assisted workflow for a work session
allowed-tools: Bash(echo:*),Bash(ls:*)
---
`CURRENT_WORK_SESSION` is !`echo $CURRENT_WORK_SESSION` as set in the `.envrc` file.

- Review context in !`ls -al ai/${CURRENT_WORK_SESSION}` `ai/<CURRENT_WORK_SESSION>`.
- Continue working on the plan in !`echo "ai/${CURRENT_WORK_SESSION}/plan.md"` `ai/<CURRENT_WORK_SESSION>/plan.md`
- Work on any steps in parallel that you can, but STOP as soon as you have something the human can manually validate.
- As you learn reference information, such as related schemas and documentation URLs, update `architecture.md`.
- Update `plan.md` with status updates.
- Summarize the work you are performing by appending to `log.md`.
- When finished with the entire plan, add manual test instructions to `summary.md`.
