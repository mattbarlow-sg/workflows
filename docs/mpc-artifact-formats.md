# MPC Artifact Formats

## MPC (Model Predictive Control) Philosophy

In robotics, Model Predictive Control (MPC) plans a full trajectory, executes a single control input, then replans from the new state—a rolling-horizon strategy that keeps the robot agile in uncertain environments. This approach brings the same principle to software planning: every node carries a 0-to-1 materialization score expressing how "locked-in" it is.

### Materialization Score Progression

The materialization score quantifies how concrete and committed a plan node is, increasing as artifacts are added:

- **0.1**: Initial planning - node identified but highly flexible
- **0.2-0.3**: BPMN diagrams added - workflow structure defined
- **0.4-0.5**: Specifications written - requirements locked in (API specs, models, schemas)
- **0.6-0.7**: Tests created - behavior contracts established (property, fuzz, contract tests)
- **0.8-0.9**: Implementation complete - code written and tested
- **1.0**: Fully materialized - deployed and validated in production

As you add artifacts (BPMN, specs, tests, code), the materialization score increases, reflecting decreased flexibility but increased confidence. The entry node typically has the highest score (ready to execute), while downstream nodes decay toward lower scores, signaling openness to change.

This quantifies the insight from "Why Greatness Cannot Be Planned": rigid end-to-end blueprints invite failure, whereas flexible, step-wise commitments let emergent opportunities guide the path to success. The system stays agile by keeping distant future nodes at low materialization, allowing pivots based on lessons learned during execution.

## Artifact Formats

The MPC workflow system supports two artifact formats:

## 1. Simple Format (Default)

The simple format uses string paths for artifacts and is backward compatible:

```yaml
artifacts:
  bpmn: "docs/bpmn/auth-flow.json"
  spec: "docs/specs/auth-endpoints.yaml" 
  tests: "tests/auth/*"
  properties: "docs/properties/auth-invariants.json"
```

This format is validated by `schemas/mpc.json`.

## 2. Enriched Format

The enriched format provides detailed categorization of specs and tests, following the Spec Driven Design workflow stages:

```yaml
artifacts:
  # Stage 1: BPMN Process Design
  bpmn: "ai/payment-processing-system/bpmn/payment-flow.json"
  
  # Stage 2: Formal Properties and Invariants
  properties:
    invariants: "ai/payment-processing-system/properties/payment-invariants.json"
    state_properties: "ai/payment-processing-system/properties/payment-state-machine.json"
    generators: "ai/payment-processing-system/properties/payment-generators.ts"
  
  # Stage 2-3: Specifications
  specs:
    api: "ai/payment-processing-system/specs/payment-api.yaml"      # OpenAPI/Swagger
    models: "ai/payment-processing-system/specs/payment-state.tla"  # TLA+/Alloy
    schemas: "ai/payment-processing-system/specs/payment-schemas.json"  # JSON Schema
  
  # Stage 3: Generated Tests
  tests:
    property: "ai/payment-processing-system/tests/property/payment-flow/*"
    deterministic: "ai/payment-processing-system/tests/simulation/payment-flow/*"
    fuzz: "ai/payment-processing-system/tests/fuzz/payment-validation/*"
    contract: "ai/payment-processing-system/tests/contract/payment-api/*"
    unit: "ai/payment-processing-system/tests/unit/payment-service/*"
    integration: "ai/payment-processing-system/tests/integration/payment-service/*"
    e2e: "ai/payment-processing-system/tests/e2e/payment-scenarios/*"
```

This format is validated by `schemas/mpc-enriched.json`.

## Namespacing Convention

In the enriched format, all artifacts should be namespaced under `ai/<plan_name>/` to:
- Avoid conflicts between different plans
- Clearly organize artifacts by project
- Follow a consistent structure

## Migration Path

To migrate from simple to enriched format:

1. Move artifacts under `ai/<plan_name>/` directory structure
2. Categorize specs and tests by type
3. Update the YAML to use nested structure
4. Validate against `mpc-enriched.json` schema

## Which Format to Use?

- **Simple Format**: Use for existing projects or when you don't need detailed categorization
- **Enriched Format**: Use for new projects following Spec Driven Design methodology, when you need:
  - Clear distinction between property-based, deterministic, and fuzz tests
  - Multiple specification formats (API specs, formal models, schemas)
  - Better organization of test artifacts by type