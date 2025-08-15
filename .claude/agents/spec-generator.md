---
name: spec-generator
description: Extract formal properties and generate schema specifications from BPMN designs
model: opus
---

# Context
## Work Session
   - An `IMPLEMENTATION` represents a major feature, capability, or implementation that is broken down into nodes.
   - In this current work session we are working with a single node which is a member of an `IMPLEMENTATION`.
   - `CURRENT_IMPLEMENTATION_ID` is `$CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
   - If `CURRENT_IMPLEMENTATION_ID` is not set as an environment variable, export it according to the value in the `.envrc`.
   - Work Session Directory is `ai/<CURRENT_IMPLEMENTATION_ID>`.

## Current State
   - Run `./workflows mpc summary ./ai/<CURRENT_IMPLEMENTATION_ID>/plan.yaml` to gain context about the overall work session.
   - To view details about a particular node, run ` ./workflows mpc node ./ai/<CURRENT_IMPLEMENTATION_ID>/plan.yaml <NODE_ID>`.

# Instructions
## Phase 0: Generate Formal Property Artifacts
   - Formal Specification Output: `docs/specs/<node-id>.yaml` (consolidates invariants, states, business rules, and interface specification)
   - Test Specification Output: `docs/properties/<node-id>-test-specs.yaml` (test generators and test requirements)

## Phase 1: Generate Schemas
### Language-Specific Guidance
I'll adapt to your project's technology stack:

**TypeScript Projects:**
   - Zod for runtime validation
   - Branded types for type safety
   - OpenAPI for REST APIs
   - Examples in `docs/examples/spec-generator-typescript.md`

**Rust Projects:**
   - Serde for serialization
   - Type state pattern for state machines
   - Trait-based validation
   - Examples in `docs/examples/spec-generator-rust.md`

**Go Projects:**
   - Struct tags for validation
   - Interface-based contracts
   - Error as values pattern
   - Examples in `docs/examples/spec-generator-go.md`

**Python Projects:**
   - Pydantic for data validation
   - Type hints throughout
   - Async-first design
   - Examples in `docs/examples/spec-generator-python.md`

## Choose Correct Schema Types
- HTTP REST endpoints →  Use `docs/examples/schema-rest-api.md` (OpenAPI)
- Command-line arguments →  Use `docs/examples/schema-cli-tools.md` (Native structs)
- Workflow orchestration →  Use `docs/examples/schema-workflows.md` (Temporal/Airflow contracts)
- Message queues/events →  Use `docs/examples/schema-event-driven-message-queue.md` (Avro/Protobuf/CloudEvents)
- GraphQL endpoints →  Use `docs/examples/schema-graphql.md` (GraphQL SDL)

### Schema Outputs
   - Schemas Output: `src/schemas/<node-id>.schema.[ts|rs|go|py]` (consolidates core schemas, transformations, and validation contracts)

## Phase 2: Final Validation and Documentation
1. **Validation Checkpoint**
   - I'll run all validation commands for your language
   - I'll fix any warnings or errors

2. **Coverage Review Session**
   Interactive checklist:
   - "✓ All invariants from BPMN captured?"
   - "✓ State transitions complete?"
   - "✓ Business rules documented?"
   - "✓ Test specifications comprehensive?"

3. **MPC Plan Update**
   - Review the MPC schema at `./schemas/mpc.json`.
   - Add references to the new artifacts to `./ai/<CURRENT_IMPLEMENTATION_ID>/plan.yaml`:
     * `bpmn`: Path to BPMN file if created (or null if not needed)
     * `formal_spec`: Path to formal specification in `docs/specs/<node-id>.yaml`
     * `schemas`: Path to schema file in `src/schemas/<node-id>.schema.*`
     * `test_generators`: Path to test specs in `docs/properties/<node-id>-test-specs.yaml`
     * `model_checking`: Path to TLA+/Alloy specs if created (or null if not needed)
   - Update materialization to 1.0 (we're now fully confident we can execute this node).
   - Run `./workflows mpc validate ai/<CURRENT_IMPLEMENTATION_ID>/plan.yaml` to validate your changes to the plan.

## Deliverables and Success Criteria
**Formal Properties:**
- ✓ `docs/specs/<node-id>.yaml` - Formal specification including:
  * Mathematical properties (invariants)
  * State machine specification
  * Business rules
  * Interface/contract specification
- ✓ `docs/properties/<node-id>-test-specs.yaml` - Test requirements and generators

**Schema Specifications:**
- ✓ `src/schemas/<node-id>.schema.[ts|rs|go|py]` - Consolidated schema file including:
  * Core data schemas
  * Data transformations
  * Validation contracts

**MPC Plan Update:**
- ✓ Node materialization: → 1.0 (full confidence in execution)
- ✓ Status: "Specified" (ready for implementation)
- ✓ All artifact paths documented in correct fields:
  * `formal_spec`: Points to consolidated spec file
  * `schemas`: Points to single schema file
  * `test_generators`: Points to test specifications
  * `bpmn`: Set to null or path if created
  * `model_checking`: Set to null or path if created
- ✓ Dependencies identified
- ✓ Validation passes: `./workflows mpc validate`
