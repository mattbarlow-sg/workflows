---
description: Create an Architecture Decision Record with AI assistance
allowed-tools: Bash(./workflows:*),Grep(*),Glob(*),WebSearch(*),Read(*),Write(*),Edit(*),LS(*)
---

# Context
- ADR ID: Will be auto-generated
- Current date: !`date +%Y-%m-%d`
- Current branch: !`git branch --show-current`
- Recent ADRs: !`ls -t docs/adr/*.json 2>/dev/null | head -5 || echo "No existing ADRs"`
- Available schemas: !`./workflows list | grep adr || echo "ADR schema available"`

# Instructions

You will help the user create an Architecture Decision Record (ADR) through an interactive process. The ADR will be generated using the `workflows adr new` CLI command.

## Phase 1: Context Gathering

1. **Understand the Decision Context**
   - Ask the user to describe the problem or decision that needs to be made
   - If not clear, ask for clarification about:
     - What problem are we trying to solve? (minimum 10 characters required)
     - What's the current situation/background?
     - Are there any constraints or assumptions we should document?
   
2. **Identify Stakeholders**
   - Ask who is involved in this decision:
     - Who are the deciders? (required for accountability)
     - Who should be consulted?
     - Who needs to be informed?
   
3. **Technical Context** (if applicable)
   - Is there a related ticket/story ID?
   - Does this relate to or supersede any existing ADRs?
   - Search codebase for related context using Grep/Glob if needed

## Phase 2: Options Discovery

1. **Research Options**
   - Based on the problem, research potential solutions:
     - Use WebSearch for industry best practices
     - Look for similar decisions in the codebase
     - Identify at least 2-3 viable options
   
2. **Document Each Option**
   - For each option, gather:
     - Clear name and description
     - Pros (advantages)
     - Cons (disadvantages)
   
3. **Decision Drivers** (optional but recommended)
   - Ask if the user wants to define decision criteria
   - If yes, for each driver:
     - Name of the criterion
     - Weight (1-5, where 5 is most important)

## Phase 3: Decision Making

1. **Choose Option**
   - Present the options with their pros/cons
   - Ask which option the user wants to choose
   - Ask for the rationale behind this choice
   
2. **Identify Consequences**
   - Based on the chosen option, help identify:
     - Positive consequences (at least one required)
     - Negative consequences (important to document)
     - Neutral consequences (optional)

## Phase 4: Generate ADR

1. **Collect Metadata** (optional)
   - Status (default is "draft", can be: proposed, accepted, etc.)
   - Tags for categorization
   - Keywords for searchability
   
2. **Build CLI Command**
   - Construct the complete `workflows adr new` command with all collected information
   - Example structure:
   ```bash
   ./workflows adr new \
     -title "Your decision title" \
     -problem "The problem statement" \
     -background "Current situation and context" \
     -chosen "Selected option name" \
     -rationale "Why this option was chosen" \
     -positive "Positive consequence 1,Positive consequence 2" \
     -negative "Negative consequence 1,Negative consequence 2" \
     -deciders "person1,person2" \
     -output "docs/adr/ADR-XXXX.json"
   ```
   
3. **Execute Generation**
   - Create the docs/adr directory if it doesn't exist: `mkdir -p docs/adr`
   - Run the CLI command to generate the ADR
   - Validate the generated file: `./workflows adr validate <filename>`
   - Generate markdown version: `./workflows adr render <filename>.json -output <filename>.md`

## Phase 5: Review and Finalize

1. **Show Results**
   - Display the generated ADR (use Read to show the markdown version)
   - Ask if any modifications are needed
   
2. **Final Checklist**
   - ✓ Problem statement is clear (min 10 chars)
   - ✓ Background provides context
   - ✓ At least one option was documented
   - ✓ Decision and rationale are clear
   - ✓ Consequences are identified
   - ✓ File passes validation

# Interactive Flow Example

```
User: "I need to decide on a caching strategy"

AI: I'll help you create an ADR for your caching strategy decision. Let me gather some information.

First, can you describe the specific problem you're trying to solve with caching? What performance or scalability issues are you facing?

User: "Our API response times are slow (3-5 seconds) for frequently accessed data. We're hitting the database too often for data that rarely changes."

AI: That's a clear problem. Now, what's the current situation? Are you using any caching currently, and what's your system architecture like?

User: "No caching yet. We have a Node.js API with PostgreSQL. About 10k requests/hour, mostly read-heavy."

AI: Good context. Who should be involved in this decision?
- Who are the decision makers?
- Who should be consulted?
- Who needs to be informed?

User: "Deciders: tech-lead, backend-lead. Consulted: devops-team. Informed: frontend-team, product-manager"

AI: Let me research caching options for your use case... [Uses WebSearch to find best practices]

Based on my research and your requirements, here are three viable options:

1. **Redis** - In-memory data store
   - Pros: Very fast, mature, supports complex data types, good Node.js support
   - Cons: Requires separate infrastructure, memory cost, complexity

2. **Node.js In-Process Cache** - Using libraries like node-cache
   - Pros: Simple to implement, no infrastructure, fast
   - Cons: Limited to single process, lost on restart, memory constraints

3. **CDN Caching** - Using CloudFlare or similar
   - Pros: Global distribution, reduces server load, handles static content well
   - Cons: Limited for dynamic content, cost, less control

Which option would you prefer, or would you like to explore any specific aspect further?

User: "Let's go with Redis. It gives us room to grow and handles our dynamic data needs."

AI: Good choice. Can you explain your rationale for choosing Redis?

User: "Redis provides the performance we need with flexibility for future scaling. Our DevOps team has experience with it."

AI: Now let's identify the consequences of implementing Redis:

Positive consequences?
User: "Faster response times, reduced database load, scalable solution, supports our growth"

Negative consequences?
User: "Additional infrastructure to maintain, memory costs, cache invalidation complexity"

Any neutral consequences?
User: "Need to train some developers on Redis patterns"

AI: Would you like to set any decision drivers with weights, or shall we proceed with the ADR generation?

User: "Skip drivers for now"

AI: Great! I'll now generate your ADR. Creating the docs/adr directory and generating the files...

[Executes]:
mkdir -p docs/adr

./workflows adr new \
  -title "Use Redis for API response caching" \
  -problem "API response times are slow (3-5 seconds) for frequently accessed data due to repeated database queries" \
  -background "No caching currently implemented. Node.js API with PostgreSQL handling 10k requests/hour, mostly read-heavy" \
  -chosen "Redis" \
  -rationale "Redis provides the performance we need with flexibility for future scaling. Our DevOps team has experience with it" \
  -positive "Faster response times,Reduced database load,Scalable solution,Supports our growth" \
  -negative "Additional infrastructure to maintain,Memory costs,Cache invalidation complexity" \
  -neutral "Need to train some developers on Redis patterns" \
  -deciders "tech-lead,backend-lead" \
  -consulted "devops-team" \
  -informed "frontend-team,product-manager" \
  -options "Redis,Node.js In-Process Cache,CDN Caching" \
  -option-descs "In-memory data store,Using libraries like node-cache,Using CloudFlare or similar" \
  -status "proposed" \
  -tags "performance,caching,infrastructure" \
  -keywords "redis,cache,api,performance,scalability" \
  -output "docs/adr/ADR-0001-redis-caching.json"

[Validates]:
./workflows adr validate docs/adr/ADR-0001-redis-caching.json

[Generates Markdown]:
./workflows adr render docs/adr/ADR-0001-redis-caching.json -output docs/adr/ADR-0001-redis-caching.md

AI: ✅ ADR created successfully! 

- ID: ADR-0001
- Title: "Use Redis for API response caching"
- Status: proposed
- Files created:
  - docs/adr/ADR-0001-redis-caching.json
  - docs/adr/ADR-0001-redis-caching.md

Would you like me to show you the generated ADR or make any modifications?
```

# Notes

- The command uses the CLI tool for all ADR operations
- Interactive process guides users through all required fields
- Validates input as it's collected
- Generates both JSON and Markdown outputs
- Can research options using WebSearch
- Flexible enough to skip optional sections
- Provides a complete audit trail of the decision