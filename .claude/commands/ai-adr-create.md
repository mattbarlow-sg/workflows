---
description: Create an Architecture Decision Record with AI assistance
allowed-tools: Bash(git:*),Bash(printf:*),Bash(date:*),Bash(grep:*),Bash(head:*),Bash(wc:*),Bash(echo:*)
---

# Context
- ADR ID: !`printf "ADR-%04d" $(ls -1 docs/adr 2>/dev/null | grep -E '^ADR-[0-9]{4}' | wc -l || echo 0)`
- Current date: !`date +%Y-%m-%d`
- Current branch: !`git branch --show-current`
- Recent ADRs: !`ls -t docs/adr/*.json 2>/dev/null | head -5 || echo "No existing ADRs"`

# Instructions

## Phase 1: Context Gathering
1. **Understand the Decision Context**
   - Ask the user to describe the problem or decision that needs to be made
   - Identify the key stakeholders (deciders, consulted, informed)
   - Determine if this supersedes any existing ADRs
   - Link to any relevant technical stories/tickets

2. **Analyze Current State**
   - Search the codebase for related components using Grep and Glob
   - Identify existing patterns and conventions
   - Document technical constraints and assumptions
   - Note any relevant compliance or regulatory requirements

3. **Research Options**
   - Use WebSearch to find industry best practices for similar decisions
   - Look for relevant documentation, articles, and case studies
   - Identify at least 2-3 viable options to consider
   - Research pros/cons for each option from authoritative sources

## Phase 2: Decision Drivers
1. **Identify Decision Criteria**
   - Work with the user to identify key decision drivers
   - Assign weights (1-5) to each driver based on importance
   - Document the rationale for each driver's weight

2. **Score Options**
   - Evaluate each option against the decision drivers
   - Calculate weighted scores for objective comparison
   - Consider effort and risk levels for each option

## Phase 3: Decision Making
1. **Analyze Trade-offs**
   - Compare options based on weighted scores
   - Consider long-term consequences (positive, negative, neutral)
   - Assess implementation complexity and cost estimates

2. **Make Recommendation**
   - Present the analysis to the user
   - Recommend the highest-scoring option with clear rationale
   - Allow the user to override with their chosen option

## Phase 4: Document Creation
1. **Generate ADR JSON**
   - Create the ADR in JSON format following the schema
   - Ensure all required fields are populated
   - Add AI metadata: tags, impact scores, keywords
   - Validate against schemas/adr.json

2. **Create Markdown Version**
   - Generate a human-readable Markdown version
   - Include mermaid diagrams for dependencies if applicable
   - Format tables for decision matrices
   - Add links to references

3. **Save Files**
   - Create directory `docs/adr` if it doesn't exist
   - Save as `docs/adr/<ADR-ID>.json`
   - Save as `docs/adr/<ADR-ID>.md`
   - Run validation: `./cmd/workflows/workflows validate adr docs/adr/<ADR-ID>.json`

## Phase 5: Validation Checklist
- [ ] Problem statement is clear and specific
- [ ] At least 2 options were considered
- [ ] Decision rationale is well-documented
- [ ] Consequences are identified (positive/negative/neutral)
- [ ] Success criteria and metrics are defined
- [ ] ADR passes schema validation
- [ ] Markdown is properly formatted

# Output Format

When complete, provide:
1. The ADR ID and title
2. Brief summary of the decision
3. Link to the generated files
4. Any follow-up actions needed

# Tools
The allowed tools enable comprehensive research and analysis while creating well-documented, schema-compliant ADRs.
