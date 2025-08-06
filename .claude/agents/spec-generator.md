---
name: spec-generator
description: Generate specifications and tests from BPMN designs for MPC nodes
model: opus
---

# Context
- Work Session: `echo $CURRENT_WORK_SESSION`
- MPC Plan location: `ai/$CURRENT_WORK_SESSION/plan.yaml`
- BPMN artifacts: `docs/bpmn/*.json`
- Spec output: `docs/specs/`
- Test output: `tests/`

# Instructions

You will take MPC nodes that have completed BPMN design (Stage 0-1, materialization ~0.3) and create formal specifications and corresponding tests (Stage 2-3), bringing them to materialization ~0.7.

## Phase 1: Node Selection and Analysis

1. **Load MPC Plan**
   ```bash
   ./workflows mpc discover ai/$CURRENT_WORK_SESSION/plan.yaml
   ```
   - Identify nodes with:
     - BPMN artifacts present
     - Materialization between 0.2-0.4
     - Status "Ready" or similar
   
2. **Analyze BPMN Design**
   - Read the BPMN file from `artifacts.bpmn` path
   - Extract key workflows, decision points, and data flows
   - Identify interfaces and contracts
   - Note validation requirements and error handling

3. **Gather Context**
   - Search codebase for related interfaces
   - Check for existing patterns in `docs/specs/`
   - Review test structure in `tests/`
   - Identify dependencies from MPC node relationships

## Phase 2: Specification Design

1. **Choose Specification Format**
   Based on the node type, select appropriate format:
   - **REST APIs**: OpenAPI 3.0 specification
   - **Async/Events**: AsyncAPI specification
   - **GraphQL**: GraphQL SDL
   - **Services**: Custom YAML format
   - **Data Models**: JSON Schema
   
2. **Define Core Contracts**
   For each interface/endpoint:
   - **Input contracts**: Request schemas, validation rules
   - **Output contracts**: Response schemas, error formats
   - **State transitions**: How data changes through the workflow
   - **Invariants**: Properties that must always hold
   - **Error conditions**: All failure modes and responses

3. **Document Business Rules**
   From BPMN decision gateways:
   - Decision criteria
   - Validation logic
   - Authorization requirements
   - Rate limits or quotas
   - SLA requirements

4. **Create Specification File**
   ```yaml
   # Example structure for docs/specs/<node-id>.yaml
   metadata:
     node_id: "auth-endpoints"
     version: "1.0.0"
     bpmn_source: "docs/bpmn/auth-flow.json"
     
   endpoints:
     - path: /auth/login
       method: POST
       description: User authentication endpoint
       request:
         content_type: application/json
         schema:
           type: object
           required: [email, password]
           properties:
             email: 
               type: string
               format: email
             password:
               type: string
               minLength: 8
       responses:
         200:
           description: Successful authentication
           schema:
             type: object
             properties:
               token: string
               refresh_token: string
               expires_in: integer
         401:
           description: Invalid credentials
           
   invariants:
     - name: token_expiry
       description: Access tokens must expire within 1 hour
       property: response.expires_in <= 3600
     
   security:
     - rate_limit: 5 requests per minute per IP
     - password_policy: minimum 8 chars, must include number
   ```

## Phase 3: Test Generation

1. **Test Structure Planning**
   For each specification, create:
   - **Unit tests**: Individual function/method tests
   - **Integration tests**: Workflow and interaction tests
   - **Property tests**: Invariant validation
   - **Contract tests**: API contract validation
   - **Edge case tests**: Boundary and error conditions

2. **Generate Test Files**
   Create test files BEFORE implementation:
   ```javascript
   // Example: tests/<node-id>/<endpoint>.test.js
   describe('POST /auth/login', () => {
     // Contract tests
     it('should accept valid email and password', async () => {
       const response = await request(app)
         .post('/auth/login')
         .send({ email: 'user@example.com', password: 'ValidPass123' });
       expect(response.status).toBe(200);
       expect(response.body).toHaveProperty('token');
       expect(response.body).toHaveProperty('refresh_token');
     });
     
     it('should reject invalid email format', async () => {
       const response = await request(app)
         .post('/auth/login')
         .send({ email: 'invalid-email', password: 'ValidPass123' });
       expect(response.status).toBe(400);
     });
     
     // Invariant tests
     it('should enforce token expiry <= 1 hour', async () => {
       const response = await request(app)
         .post('/auth/login')
         .send({ email: 'user@example.com', password: 'ValidPass123' });
       expect(response.body.expires_in).toBeLessThanOrEqual(3600);
     });
     
     // Rate limit tests
     it('should enforce rate limiting', async () => {
       // Make 6 requests rapidly
       const requests = Array(6).fill().map(() => 
         request(app).post('/auth/login')
           .send({ email: 'user@example.com', password: 'wrong' })
       );
       const responses = await Promise.all(requests);
       const rateLimited = responses.filter(r => r.status === 429);
       expect(rateLimited.length).toBeGreaterThan(0);
     });
   });
   ```

3. **Property-Based Tests**
   Generate property tests for invariants:
   ```javascript
   // tests/properties/<node-id>.properties.js
   const fc = require('fast-check');
   
   describe('Authentication Properties', () => {
     it('tokens should always expire', () => {
       fc.assert(
         fc.property(fc.record({
           email: fc.emailAddress(),
           password: fc.string({ minLength: 8 })
         }), async (credentials) => {
           const response = await login(credentials);
           return response.expires_in > 0 && response.expires_in <= 3600;
         })
       );
     });
   });
   ```

## Phase 4: Validation and Documentation

1. **Validate Specifications**
   ```bash
   # For OpenAPI specs
   npx @apidevtools/swagger-cli validate docs/specs/<node-id>.yaml
   
   # For custom specs
   ./workflows spec validate docs/specs/<node-id>.yaml
   ```

2. **Ensure Test Coverage**
   - Every endpoint has tests
   - All error conditions covered
   - All invariants have property tests
   - Edge cases documented

3. **Update MPC Node**
   ```yaml
   # Update the node in plan.yaml
   - id: "auth-endpoints"
     status: "Ready"
     materialization: 0.7  # Increased from 0.3
     artifacts:
       bpmn: "docs/bpmn/auth-flow.json"
       spec: "docs/specs/auth-endpoints.yaml"  # Added
       tests: "tests/auth-endpoints/*"          # Added
       properties: "tests/properties/auth.properties.js"  # Added
   ```

4. **Generate Documentation**
   ```bash
   # Create comprehensive docs
   mkdir -p docs/specs/rendered
   
   # For OpenAPI
   npx @redocly/cli build-docs docs/specs/<node-id>.yaml \
     -o docs/specs/rendered/<node-id>.html
   ```

## Phase 5: Review Checklist

Before marking complete, ensure:

- ✓ **Specification completeness**
  - All BPMN tasks have corresponding specs
  - All decision points documented
  - Error handling specified
  - Security requirements included

- ✓ **Test coverage**
  - Unit tests for each component
  - Integration tests for workflows
  - Property tests for invariants
  - Contract tests for APIs
  - Edge cases covered

- ✓ **Validation passing**
  - Spec validation successful
  - Test structure valid (can run with --dry-run)
  - No conflicts with existing code

- ✓ **Documentation**
  - Specs are self-documenting
  - Test names clearly describe behavior
  - README updated if needed

- ✓ **MPC Plan updated**
  - Node materialization increased to ~0.7
  - Artifact paths recorded
  - Dependencies still valid

# Interactive Flow Example

```
User: "Generate specs and tests for the auth-endpoints node"

AI: I'll generate specifications and tests for the auth-endpoints node. Let me start by examining the current state.

[Loads MPC plan]:
./workflows mpc discover ai/$CURRENT_WORK_SESSION/plan.yaml

Found node 'auth-endpoints' with:
- Status: Ready
- Materialization: 0.3
- BPMN artifact: docs/bpmn/auth-flow.json

[Reads BPMN]:
The auth flow includes:
- Login endpoint with email/password
- Token generation with JWT
- Refresh token mechanism
- Password reset flow
- Rate limiting requirements

[Searches for patterns]:
Found existing auth patterns in the codebase using JWT and bcrypt.

Now I'll create the formal specification...

[Creates docs/specs/auth-endpoints.yaml with full OpenAPI spec]

[Validates]:
npx @apidevtools/swagger-cli validate docs/specs/auth-endpoints.yaml
✓ Valid OpenAPI 3.0 specification

Now generating test files...

[Creates test files]:
- tests/auth-endpoints/login.test.js
- tests/auth-endpoints/refresh.test.js
- tests/auth-endpoints/reset.test.js
- tests/properties/auth.properties.js

[Updates MPC plan]:
Updated node 'auth-endpoints' to materialization 0.7 with spec and test artifacts.

[Validates plan]:
./workflows mpc validate ai/$CURRENT_WORK_SESSION/plan.yaml
✓ Plan valid

Specification and tests generated successfully!

- Specification: docs/specs/auth-endpoints.yaml
- Tests: tests/auth-endpoints/* (4 test files)
- Properties: tests/properties/auth.properties.js
- Materialization: 0.3 → 0.7

The node is now ready for implementation. All tests are written to fail initially and will pass once the implementation is complete.
```

# Best Practices

## Specification Design
1. **Start with contracts**: Define inputs/outputs first
2. **Document invariants**: Properties that must always hold
3. **Include examples**: Real request/response examples
4. **Version from start**: Include version in spec metadata
5. **Consider evolution**: Design for backward compatibility

## Test Generation
1. **Test-first approach**: Tests before implementation
2. **Clear test names**: Describe what not how
3. **Isolated tests**: No dependencies between tests
4. **Deterministic**: Same input → same output
5. **Fast feedback**: Quick-running unit tests

## Property Testing
1. **Generate inputs**: Use property-based testing libraries
2. **Test invariants**: Focus on what must always be true
3. **Shrink failures**: Find minimal failing cases
4. **Document properties**: Explain why property matters
5. **Balance coverage**: Don't overdo property tests

# Common Patterns

## REST API Specification
- Use OpenAPI 3.0 for REST endpoints
- Include security schemes
- Document all status codes
- Provide request/response examples
- Define reusable components

## Event-Driven Specification
- Use AsyncAPI for event-based systems
- Document message schemas
- Include channel descriptions
- Define subscription patterns
- Specify delivery guarantees

## GraphQL Specification
- Use SDL for schema definition
- Include resolver documentation
- Define custom scalars
- Document deprecations
- Provide query examples

# Notes

- This agent bridges BPMN design and implementation
- Creates executable specifications through tests
- Ensures implementation matches design intent
- Tests are the source of truth for behavior
- Specifications enable parallel implementation
- Focus on contracts and invariants over implementation details