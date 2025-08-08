---
description: Guide the user through an interactive workshop for creating BPMN context.
allowed-tools: Bash(date:*),Bash(git:*),Bash(ls:*),Bash(head:*),Bash(echo:*),Bash(./workflows:*),Bash(grep:*)
---

# Context
## Current Implementation
- An `IMPLEMENTATION` represents a major feature, capability, or implementation that is broken down into nodes.
- In this current work session we are working with a single node which is a member of an `IMPLEMENTATION`.
- `CURRENT_IMPLEMENTATION_ID` is `$CURRENT_IMPLEMENTATION_ID` as set in the `.envrc` file.
- If `CURRENT_IMPLEMENTATION_ID` is not set as an environment variable, export it according to the value in the `.envrc`.
- Work Session Directory is `ai/<CURRENT_IMPLEMENTATION_ID>`.

## Current State
- Run `./workflows mpc summary ./ai/<CURRENT_IMPLEMENTATION_ID>/plan.yaml` to gain context about the overall work session.
- To view details about a particular node, run ` ./workflows mpc node ./ai/<CURRENT_IMPLEMENTATION_ID>/plan.yaml <NODE_ID>`.

# Instructions
- You will not create or edit files directly.
- This work session is NOT about completing work in the plan. This is an interactive workshop.
**IMPORTANT**: This is an interactive process. You MUST ask questions and get your input before generating any BPMN.

Together we'll take a node that has completed Stage 1 (BPMN design) and guide it through:
- **Stage 2**: Formal Properties Extraction - Mathematical properties and invariants
- **Stage 3**: Schema-Driven Specifications - Validation contracts and transformations

This interactive process will create a complete specification that's ready for implementation.

## Phase 1: Understanding Your Node
1. **Review Node Context**
   - I'll review the BPMN design that was completed in Stage 1
   - Together we'll identify what needs to be formally specified
   
2. **Interactive Discovery**
   - I'll ask you about:
     - What invariants must always hold in this process?
     - What are the critical data validations?
     - Are there any business rules we need to enforce?
     - What state transitions are allowed or forbidden?
   
3. **Context Gathering**
   - I'll search the codebase for related patterns and interfaces
   - We'll review any existing specifications for consistency
   - I'll help identify dependencies from other nodes

## Phase 2: Formal Properties Workshop (Stage 2)
1. **Invariant Discovery Session**
   Together we'll identify properties that must always hold:
   - I'll analyze your BPMN to suggest potential invariants
   - You'll help me understand which are critical for your domain
   - We'll formalize these as mathematical properties
   - I'll save these in context as `invariants`
   
   Example questions I'll ask:
   - "I see a payment flow - should payment always precede shipping?"
   - "Are there any values that must stay within certain bounds?"
   - "What relationships between data fields must be preserved?"

2. **State Machine Exploration**
   We'll map out the allowed states and transitions:
   - I'll extract states from your BPMN gateways
   - You'll validate the transition rules
   - We'll identify forbidden transitions
   - I'll document liveness and safety properties
   - I'll save these in context as `states`
   
   Interactive prompts:
   - "Can an order go directly from 'draft' to 'shipped'?"
   - "What happens if payment fails - what states are possible?"
   - "Should cancelled orders be permanent?"

3. **Business Rules Workshop**
   We'll capture all business constraints:
   - I'll identify decision points from BPMN
   - You'll explain the business logic
   - We'll formalize validation rules
   - I'll document temporal and contextual constraints
   - I'll save these in context as `rules`
   
   Discovery questions:
   - "Are there minimum/maximum values for any fields?"
   - "Any time-based restrictions (business hours, deadlines)?"
   - "Rate limiting or throttling requirements?"

4. **Test Specification Planning**
   Together we'll define what tests should verify:
   - Property-based test scenarios
   - Deterministic test cases
   - Edge cases and error conditions
   - I'll save these in context as `test-specs`

## Phase 3: Schema Design Workshop (Stage 3)

1. **Language and Framework Selection**
   Let's determine the right schema approach for your project:
   - I'll check your project's language and frameworks
   - You'll confirm the validation library preference
   - We'll select the appropriate schema patterns
   
   Questions I'll ask:
   - "I see you're using TypeScript - shall we use Zod for schemas?"
   - "Do you prefer runtime validation or compile-time only?"
   - "Any existing schema patterns I should follow?"

2. **Interactive Schema Creation**
   We'll build schemas together based on your BPMN:
   - I'll propose initial schema structures
   - You'll help refine field types and constraints
   - We'll encode invariants as schema refinements
   - I'll implement branded types for safety
   
   Collaborative process:
   - "For the user ID, should this be a UUID or custom format?"
   - "What's the maximum length for this description field?"
   - "Should we allow null values or make everything required?"

3. **Transformation Contract Design**
   Together we'll map data flows:
   - I'll identify transformations from BPMN
   - You'll validate the input/output contracts
   - We'll define error schemas
   - I'll create type-safe pipelines
   
   Workshop questions:
   - "When login succeeds, what exact fields should the response contain?"
   - "How should we handle partial failures in this flow?"
   - "What validation should happen at each transformation?"

4. **API Specification Review**
   We'll document the complete interface:
   - I'll generate OpenAPI/AsyncAPI specs
   - You'll review endpoints and schemas
   - We'll add examples and descriptions
   - I'll save these in context as `api specs`

## Phase 4: Validation and Integration Workshop
1. **Contract Validation Session**
   We'll ensure all contracts are complete:
   - I'll create validation pipelines for your language
   - You'll review the validation logic
   - We'll test with example data
   - I'll refine based on your feedback
   
   Interactive validation:
   - "Here's a sample invalid input - is this the right error message?"
   - "Should validation fail fast or collect all errors?"
   - "Any custom validation rules beyond type checking?"

2. **Transformation Pipeline Review**
   Together we'll verify data flows:
   - I'll show you each transformation step
   - You'll confirm the business logic
   - We'll handle edge cases
   - I'll ensure invariants are preserved
   
   Collaborative refinement:
   - "This transformation assumes X - is that correct?"
   - "How should we handle missing optional fields?"
   - "What's the fallback for transformation failures?"

3. **Schema Composition Workshop**
   We'll build complex schemas from simple ones:
   - I'll propose composition patterns
   - You'll guide the relationships
   - We'll ensure consistency
   - I'll document the schema hierarchy

## Phase 5: Final Validation and Documentation
1. **Validation Checkpoint**
   Together we'll validate everything:
   - You'll confirm the specifications match intent
   - I'll make final adjustments
   
   I'll guide you through:
   - "All TypeScript schemas compile successfully!"
   - "Found one validation gap - should we add this constraint?"
   - "Everything passes - ready to proceed?"

2. **Coverage Review Session**
   Interactive checklist:
   - "✓ All invariants from BPMN captured?"
   - "✓ State transitions complete?"
   - "✓ Business rules documented?"
   - "✓ Test specifications comprehensive?"

Invoke the @spec-generator subagent with the provided and generated context.

# Interactive Flow Example
User: "I need to create specifications for the auth-endpoints node"
AI: I'll help you create formal properties and schemas for auth-endpoints. Let me examine your context.

[Examines plan and BPMN]
Found auth-endpoints node with login, JWT tokens, and rate limiting.

## Phase 1: Requirements Discovery
I need to understand your constraints. What should be the token expiry times?
User: "1 hour for access, 7 days for refresh"
AI: Good. What about password requirements and rate limiting?
User: "8+ chars with complexity rules. Lock after 5 failures for 15 minutes."

## Phase 2: Formal Properties
Creating invariants based on your requirements:
- Token expiry limits (access ≤ 1hr, refresh ≤ 7d)
- Password complexity rules
- Rate limit constraints
- Single-use refresh tokens

## Phase 3: Schema Design
You're using TypeScript - shall we use Zod for validation?
User: "Yes, with email, password, remember_me fields"
AI: Creating Zod schemas with refinements:
```typescript
const LoginRequestSchema = z.object({
  email: z.string().email(),
  password: PasswordSchema, // with complexity rules
  remember_me: z.boolean().optional()
});
```

## Phase 4: Validation
✓ All schemas compile
✓ Properties validated
✓ Contracts complete

# Common Questions
## For Invariants
- "What must always be true about this data?"
- "Are there any relationships between fields that must hold?"
- "What are the absolute limits or boundaries?"
- "Can this ever be null or empty?"

## For State Machines
- "What triggers this state transition?"
- "Can we go backwards from this state?"
- "What happens on timeout or error?"
- "Are there any forbidden state combinations?"

## For Business Rules
- "Is this a hard requirement or a preference?"
- "Does this apply always or only in certain contexts?"
- "What's the business impact if this rule is violated?"
- "How should we handle edge cases?"

## For Schemas
- "What format should this ID follow?"
- "What's the maximum reasonable size for this field?"
- "Should we accept partial data or require everything?"
- "How strict should validation be?"

# Notes
- This agent facilitates an interactive workshop for specification creation
- Guides you through Stage 2 (Formal Properties) and Stage 3 (Schemas)
- Updates node materialization to 1.0 (full confidence in execution)
- Focuses on collaboration and validation at each step
- Ensures specifications match your exact requirements
- Creates implementation-ready contracts and documentation
- Nothing is assumed - everything is confirmed with you
