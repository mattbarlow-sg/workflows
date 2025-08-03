# ADR CLI Reference

Complete reference for the `workflows adr` command.

## Commands

### `workflows adr new`

Create a new Architecture Decision Record.

#### Synopsis

```bash
workflows adr new [flags]
```

#### Required Flags

| Flag | Type | Description |
|------|------|-------------|
| `-title` | string | ADR title describing the decision |
| `-problem` | string | Problem statement (minimum 10 characters) |
| `-background` | string | Background context explaining current state |
| `-chosen` | string | Name of the chosen option |
| `-rationale` | string | Why this option was chosen |

**Note**: At least one of `-positive` or `-negative` consequences is required.

#### Consequence Flags

| Flag | Type | Description |
|------|------|-------------|
| `-positive` | string | Positive consequences (comma-separated) |
| `-negative` | string | Negative consequences (comma-separated) |
| `-neutral` | string | Neutral consequences (comma-separated) |

#### Optional Context Flags

| Flag | Type | Description |
|------|------|-------------|
| `-constraints` | string | Technical/business constraints (comma-separated) |
| `-assumptions` | string | Assumptions made (comma-separated) |

#### Stakeholder Flags

| Flag | Type | Description |
|------|------|-------------|
| `-deciders` | string | People making the decision (comma-separated) |
| `-consulted` | string | People consulted (comma-separated) |
| `-informed` | string | People to be informed (comma-separated) |

#### Decision Driver Flags

| Flag | Type | Description |
|------|------|-------------|
| `-drivers` | string | Decision criteria (comma-separated) |
| `-driver-weights` | string | Weights 1-5 for each driver (comma-separated) |

#### Option Documentation Flags

| Flag | Type | Description |
|------|------|-------------|
| `-options` | string | Other options considered (comma-separated) |
| `-option-descs` | string | Descriptions for each option (comma-separated) |

#### Technical Story Flags

| Flag | Type | Description |
|------|------|-------------|
| `-story-id` | string | Issue/ticket ID (e.g., JIRA-123) |
| `-story-title` | string | Story title |
| `-story-desc` | string | Story description |

#### Metadata Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-status` | string | draft | Status: draft, proposed, accepted, deprecated, superseded, rejected |
| `-tags` | string | | Semantic tags (comma-separated) |
| `-keywords` | string | | Search keywords (comma-separated) |

#### Output Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-format` | string | json | Output format: json or markdown |
| `-output` | string | | Output file path (prints to stdout if not specified) |

#### Examples

##### Minimal ADR
```bash
workflows adr new \
  -title "Use PostgreSQL for main database" \
  -problem "Need reliable relational database for user data" \
  -background "Currently using SQLite which won't scale" \
  -chosen "PostgreSQL" \
  -rationale "Proven reliability and scalability" \
  -positive "ACID compliance,JSON support,Extensions"
```

##### Complete ADR
```bash
workflows adr new \
  -title "Adopt microservices architecture" \
  -problem "Monolith is becoming difficult to maintain and scale" \
  -background "Current monolith serves 1M users but development velocity is decreasing" \
  -chosen "Microservices" \
  -rationale "Enables independent scaling and deployment of services" \
  -positive "Independent scaling,Technology diversity,Fault isolation" \
  -negative "Increased complexity,Network latency,Distributed transactions" \
  -neutral "Requires DevOps culture shift" \
  -status "proposed" \
  -deciders "cto,lead-architect,engineering-manager" \
  -consulted "backend-team,devops-team,security-team" \
  -informed "all-engineers,product-team" \
  -constraints "Must maintain 99.9% uptime,Budget limited to 50k/month" \
  -assumptions "Team can learn Kubernetes,Cloud costs will decrease over time" \
  -drivers "Scalability,Development velocity,Operational complexity,Cost" \
  -driver-weights "5,5,3,4" \
  -options "Microservices,Modular monolith,Serverless" \
  -option-descs "Full microservices with k8s,Monolith with clear module boundaries,FaaS approach" \
  -story-id "ARCH-2024-001" \
  -story-title "Migrate to scalable architecture" \
  -tags "architecture,microservices,scalability" \
  -keywords "microservice,kubernetes,scaling,distributed" \
  -output "adr-microservices.json"
```

### `workflows adr validate`

Validate an ADR file against the JSON schema.

#### Synopsis

```bash
workflows adr validate <adr-file.json>
```

#### Arguments

| Argument | Description |
|----------|-------------|
| `<adr-file.json>` | Path to the ADR JSON file to validate |

#### Examples

```bash
# Validate a single file
workflows adr validate my-adr.json

# Validate multiple files
workflows adr validate adr-001.json adr-002.json
```

#### Output

Success:
```
✓ ADR file 'my-adr.json' is valid
  ID: ADR-0001
  Title: Use PostgreSQL for main database
  Status: accepted
```

Failure:
```
✗ ADR file 'my-adr.json' is invalid

Validation errors:
  1. context.problem: String length must be greater than or equal to 10
  2. decision.rationale: is required
```

### `workflows adr render`

Convert an ADR from JSON to Markdown format.

#### Synopsis

```bash
workflows adr render [flags] <adr-file.json>
```

#### Arguments

| Argument | Description |
|----------|-------------|
| `<adr-file.json>` | Path to the ADR JSON file to render |

#### Flags

| Flag | Type | Description |
|------|------|-------------|
| `-output` | string | Output file path (prints to stdout if not specified) |

#### Examples

```bash
# Render to stdout
workflows adr render my-adr.json

# Render to file
workflows adr render my-adr.json -output my-adr.md

# Render and pipe
workflows adr render my-adr.json | less
```

## Common Patterns

### Batch Processing

```bash
# Validate all ADRs
for f in docs/adr/*.json; do
  ./workflows adr validate "$f"
done

# Render all ADRs to markdown
for f in docs/adr/*.json; do
  ./workflows adr render "$f" -output "${f%.json}.md"
done
```

### Pipeline Integration

```bash
# Generate and validate in one command
./workflows adr new -title "..." -problem "..." ... | \
  tee new-adr.json | \
  ./workflows adr validate /dev/stdin

# Generate JSON and immediately convert to markdown
./workflows adr new -title "..." ... -format json | \
  ./workflows adr render /dev/stdin
```

### Template Usage

```bash
# Create a template function
create_adr() {
  ./workflows adr new \
    -deciders "tech-lead,architect" \
    -status "proposed" \
    -tags "architecture" \
    "$@"
}

# Use the template
create_adr \
  -title "New decision" \
  -problem "Problem to solve" \
  -background "Current situation" \
  -chosen "Solution" \
  -rationale "Why this solution" \
  -positive "Benefits"
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (invalid arguments, missing files, etc.) |
| 2 | Validation failure |

## Environment Variables

Currently, no environment variables affect ADR command behavior. All configuration is via command-line flags.

## See Also

- [ADR Schema Documentation](./SCHEMA.md)
- [ADR Best Practices](./README.md#best-practices)
- [AI-Assisted ADR Creation](../../.claude/commands/ai-adr-create.md)