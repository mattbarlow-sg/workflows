# Architecture Decision Records (ADR) System

This directory contains Architecture Decision Records for the project. ADRs document important architectural decisions made during development.

## Overview

The ADR system provides:
- **Structured decision documentation** following MADR 4.0.0 standard
- **JSON schema validation** ensuring consistency
- **CLI tools** for creation, validation, and rendering
- **AI-assisted workflow** for guided decision making
- **Rich metadata** for searchability and AI context

## Quick Start

### Create an ADR via CLI

```bash
# Minimal ADR with required fields only
./workflows adr new \
  -title "Your decision title" \
  -problem "The problem you're solving (min 10 chars)" \
  -background "Current situation and context" \
  -chosen "Selected solution" \
  -rationale "Why this solution was chosen" \
  -positive "Benefit 1,Benefit 2" \
  -output docs/adr/ADR-0001.json

# Validate the ADR
./workflows adr validate docs/adr/ADR-0001.json

# Render to Markdown
./workflows adr render docs/adr/ADR-0001.json -output docs/adr/ADR-0001.md
```

### Create an ADR with AI Assistance

Use the interactive AI command for guided ADR creation:

```bash
ai-adr-create
```

The AI will:
1. Help you articulate the problem
2. Research potential solutions
3. Guide you through stakeholder identification
4. Generate a complete, validated ADR

## ADR Structure

Each ADR contains:

### Required Fields
- **ID**: Unique identifier (e.g., ADR-0001)
- **Title**: Brief description of the decision
- **Status**: Current state (draft, proposed, accepted, deprecated, superseded, rejected)
- **Date**: When the decision was made
- **Context**: Problem statement and background
- **Decision**: Chosen option and rationale
- **Consequences**: Positive and negative impacts

### Optional Fields
- **Stakeholders**: Deciders, consulted, informed parties
- **Technical Story**: Related tickets/issues
- **Decision Drivers**: Weighted criteria
- **Options**: Alternative solutions considered
- **Validation**: Success criteria and metrics
- **Compliance**: Standards and regulations
- **AI Metadata**: Tags, keywords, impact scores

## File Organization

```
docs/adr/
├── README.md           # This file
├── ADR-0001-*.json    # ADR in JSON format
├── ADR-0001-*.md      # ADR in Markdown format
└── ...
```

## Tools and Commands

### ADR CLI Commands

```bash
# Show detailed help
./workflows adr new --help

# Create new ADR (see all options)
./workflows adr new [flags]

# Validate ADR against schema
./workflows adr validate <file.json>

# Convert JSON to Markdown
./workflows adr render <file.json> [-output <file.md>]
```

### Common Workflows

#### 1. Simple Decision
```bash
./workflows adr new \
  -title "Use PostgreSQL for database" \
  -problem "Need scalable database solution" \
  -background "SQLite no longer meets our needs" \
  -chosen "PostgreSQL" \
  -rationale "Proven scalability and features" \
  -positive "ACID compliance,Scalable,Extensions" \
  -negative "Operational complexity"
```

#### 2. Complex Decision with Stakeholders
```bash
./workflows adr new \
  -title "Migrate to microservices" \
  -problem "Monolith is hard to scale and maintain" \
  -background "Single application serving 1M users" \
  -chosen "Microservices architecture" \
  -rationale "Enables independent scaling and deployment" \
  -positive "Scalability,Team autonomy,Technology flexibility" \
  -negative "Complexity,Network latency,Operational overhead" \
  -deciders "cto,architect" \
  -consulted "dev-team,ops-team" \
  -informed "all-engineers" \
  -status "proposed"
```

#### 3. Decision with Analysis
```bash
./workflows adr new \
  -title "Frontend framework selection" \
  -problem "Need modern UI framework for new project" \
  -background "Building customer-facing web application" \
  -options "React,Vue,Angular" \
  -option-descs "Facebook library,Progressive framework,Google framework" \
  -drivers "Learning curve,Performance,Ecosystem,Community" \
  -driver-weights "4,3,5,5" \
  -chosen "React" \
  -rationale "Best ecosystem and community support" \
  -positive "Large ecosystem,Easy hiring,Great tooling" \
  -negative "Just a library not framework,JSX learning curve"
```

## Schema Documentation

See [ADR Schema Documentation](./SCHEMA.md) for detailed field descriptions and validation rules.

## Best Practices

1. **Write ADRs promptly** - Document decisions while context is fresh
2. **Be specific** - Include concrete details, not vague statements
3. **Document alternatives** - Show what options were considered
4. **Include consequences** - Both positive and negative
5. **Link related ADRs** - Use supersedes/superseded-by relationships
6. **Keep ADRs immutable** - Create new ADRs instead of editing accepted ones
7. **Use consistent naming** - Follow the ADR-XXXX pattern

## Status Workflow

```
draft → proposed → accepted
         ↓           ↓
      rejected   deprecated
                     ↓
                 superseded
```

## Examples

- [Sample ADR - Frontend Framework](../../examples/sample-adr.json) - Complete example with all fields
- See validated examples in this directory

## Further Reading

- [MADR - Markdown Architectural Decision Records](https://adr.github.io/madr/)
- [ADR Tools](https://github.com/npryce/adr-tools)
- [Architectural Decision Records](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)