# Schema Driven Development Workflow
## Stage 0: Design Thinking & Problem Definition

### Objective
Define problem statements and success criteria.

### Testing at this Stage
- **Stakeholder Validation**: Example mapping sessions
- **Acceptance Criteria**: Define measurable outcomes
- **Risk Analysis**: Identify areas needing robustness testing

### Example Output
```markdown
Problem: Order fulfillment system with high reliability requirements
Success Criteria:
- 99.9% order completion rate
- Zero inventory discrepancies
- Payment processing within 5 seconds
Risk Areas:
- Concurrent inventory updates
- Payment gateway failures
- Network partitions
```

## Stage 1: BPMN Process Design with Embedded Schemas

### Objective
Create executable visual process diagrams with testable specifications.

### Testing at this Stage
- **Spec Validation**: Test the BPMN itself for deadlocks/livelocks
- **Schema Testing**: Validate example data against schemas
- **Process Simulation**: Run sample flows through BPMN

### Example Output
```json
{
  "$type": "bpmn:process",
  "process": {
    "id": "order-fulfillment",
    "elements": {
      "activities": [{
        "id": "validate-order",
        "name": "Validate Order",
        "extensions": {
          "schema": {
            "input": "RawOrderSchema",
            "output": "ValidatedOrderSchema"
          },
          "invariants": [
            "order.items.length > 0",
            "order.total === sum(order.items.map(i => i.price * i.quantity))"
          ],
          "examples": {
            "valid": {"items": [{"id": "1", "quantity": 2, "price": 10}], "total": 20},
            "invalid": {"items": [], "total": 0}
          }
        }
      }]
    }
  }
}
```

### Testing Tools
```typescript
// BPMN process validator
function validateBPMN(process: BPMNProcess) {
  // Check for unreachable nodes
  // Verify all paths lead to end events
  // Validate schema references exist
}

// Example-based schema testing
process.elements.activities.forEach(activity => {
  const examples = activity.extensions.examples;
  test(`${activity.name} schema validation`, () => {
    expect(validateSchema(examples.valid, activity.schema.input)).toBe(true);
    expect(validateSchema(examples.invalid, activity.schema.input)).toBe(false);
  });
});
```

## Stage 2: Formal Modeling & Property Extraction

### Objective
Extract mathematical properties and generate comprehensive test suites.

### Testing at this Stage
- **Model Checking**: Use TLA+ or Alloy to verify properties
- **Property Test Generation**: Create property-based tests
- **Invariant Testing**: Test that invariants can't be violated

### Example Output
```typescript
// Formal specification with embedded tests
export const OrderFulfillmentSpec = {
  invariants: {
    "total matches items": (order: Order) => 
      order.total === order.items.reduce((sum, item) => sum + item.price * item.quantity, 0),
    
    "inventory conservation": (before: Inventory, after: Inventory, order: Order) =>
      order.items.every(item => 
        after[item.id] === before[item.id] - item.quantity
      )
  },

  // Property-based test generators
  generators: {
    validOrder: fc.record({
      items: fc.array(
        fc.record({
          id: fc.string(),
          quantity: fc.integer({ min: 1, max: 100 }),
          price: fc.integer({ min: 1, max: 10000 })
        }),
        { minLength: 1 }
      )
    }).map(order => ({
      ...order,
      total: order.items.reduce((sum, item) => sum + item.price * item.quantity, 0)
    })),

    invalidOrder: fc.oneof(
      fc.record({ items: fc.constant([]), total: fc.integer() }), // Empty items
      fc.record({
        items: fc.array(orderItem, { minLength: 1 }),
        total: fc.integer({ min: -1000, max: -1 }) // Negative total
      })
    )
  },

  // State machine properties
  stateProperties: [
    // Liveness: All orders eventually complete or fail
    "eventually(order.status === 'Completed' || order.status === 'Failed')",
    
    // Safety: Payment before shipping
    "always(order.status === 'Shipped' => order.paymentStatus === 'Paid')"
  ]
};

// Generate tests from properties
const propertyTests = generatePropertyTests(OrderFulfillmentSpec);
```

### Model Checking Example (TLA+)
```tla
---- MODULE OrderFulfillment ----
INVARIANT InventoryNonNegative == \A item \in Items: inventory[item] >= 0
INVARIANT PaymentBeforeShipping == 
  \A order \in Orders: order.status = "Shipped" => order.paid = TRUE
====
```

## Stage 3: Schema-Driven Specification & Validation Contracts

### Objective
Create formal schemas and validation contracts that implementations must satisfy.

### Schema Types Created
1. **Data Schemas**: Zod/JSON schemas with validation rules
2. **Transformation Contracts**: Input→Output pipelines
3. **State Machines**: Type-safe state transitions
4. **Invariant Refinements**: Business rules as schema refinements

### Example Schema Specifications
```typescript
// 1. Data validation schemas with refinements
import { z } from 'zod';

// Branded types for type safety
export const OrderIdSchema = z.string().uuid().brand('OrderId');
export const MoneySchema = z.number()
  .positive()
  .multipleOf(0.01)
  .brand('Money');

// Order schema with invariants as refinements
export const OrderSchema = z.object({
  id: OrderIdSchema,
  items: z.array(OrderItemSchema).nonempty(),
  total: MoneySchema
}).refine(
  // Invariant: total must match sum of items
  order => {
    const calculated = order.items.reduce(
      (sum, item) => sum + (item.price * item.quantity), 0
    );
    return Math.abs(order.total - calculated) < 0.01;
  },
  { message: "Order total doesn't match items sum" }
);

// 2. Transformation contracts
export const ProcessOrderTransformation = {
  name: 'processOrder',
  from: z.object({
    cart: CartSchema,
    customer: CustomerSchema,
    paymentMethod: PaymentMethodSchema
  }),
  to: OrderSchema,
  invariants: [
    'output.customerId === input.customer.id',
    'output.status === "pending"'
  ]
};

// 3. State machine types (compile-time safety)
type OrderState = 
  | { status: 'draft'; paymentId?: never }
  | { status: 'pending'; paymentId?: never }
  | { status: 'paid'; paymentId: string }
  | { status: 'shipped'; paymentId: string; trackingNumber: string };

// 4. Validation pipeline specifications
export const OrderValidationPipeline = {
  validateInput: OrderInputSchema,
  validateBusinessRules: {
    inventoryAvailable: (order) => checkInventory(order.items),
    customerActive: (customer) => customer.status === 'active',
    paymentMethodValid: (method) => method.verified === true
  },
  validateOutput: OrderOutputSchema.refine(
    data => data.status === 'pending',
    "New orders must start in pending status"
  )
};
```

### Example Test Generation (After Implementation)
```typescript
// Tests are generated AFTER implementation to validate against schemas
describe('Order Processing Implementation', () => {
  test('total always matches sum of items', () => {
    fc.assert(
      fc.property(OrderFulfillmentSpec.generators.validOrder, (order) => {
        return OrderFulfillmentSpec.invariants["total matches items"](order);
      })
    );
  });

  test('inventory conservation through order process', () => {
    fc.assert(
      fc.property(
        OrderFulfillmentSpec.generators.validOrder,
        inventoryGenerator,
        (order, inventory) => {
          const result = simulateOrder(order, inventory);
          return OrderFulfillmentSpec.invariants["inventory conservation"](
            inventory, 
            result.inventory, 
            order
          );
        }
      )
    );
  });
});

// 2. Deterministic simulation tests
describe('Order Workflows', () => {
  let env: DeterministicEnvironment;
  
  beforeEach(() => {
    env = createDeterministicEnv({
      time: '2024-01-01T00:00:00Z',
      random: seedRandom(42),
      services: {
        payment: mockPaymentGateway(),
        inventory: inMemoryInventory()
      }
    });
  });

  test('happy path order completion', async () => {
    const order = createTestOrder({ items: [{ id: 'SKU001', quantity: 2 }] });
    const result = await processOrder(order, env);
    
    expect(result.status).toBe('Completed');
    expect(result.timeline).toMatchSnapshot();
  });

  test('payment failure handling', async () => {
    env.services.payment.failNext();
    const order = createTestOrder();
    const result = await processOrder(order, env);
    
    expect(result.status).toBe('PaymentFailed');
    expect(env.services.inventory.wasModified()).toBe(false);
  });
});

// 3. Fuzz testing setup
describe('Input Validation Fuzzing', () => {
  const fuzzer = createFuzzer({
    baseInputs: fc.sample(OrderFulfillmentSpec.generators.validOrder, 10),
    mutations: ['truncate', 'type-confusion', 'boundary-values']
  });

  test('order validation handles malformed input', () => {
    fuzzer.run(1000, (malformedOrder) => {
      expect(() => validateOrder(malformedOrder))
        .not.toThrow(UnhandledException);
    });
  });
});

// 4. Contract tests (future implementation)
describe('API Contracts', () => {
  test.todo('order endpoint matches OpenAPI spec');
  test.todo('event schemas match AsyncAPI spec');
});
```

## Stage 4: Implementation with Schema Validation

### Objective
Implement system that satisfies the schema specifications from Stage 3.

### Implementation Approach
- **Schema-Driven Development**: Implementations must satisfy schemas
- **Runtime Validation**: All inputs/outputs validated against schemas
- **Type Safety**: Leverage TypeScript types derived from schemas

### Example Implementation
```typescript
// Runtime invariant enforcement
class OrderService {
  private invariants = OrderFulfillmentSpec.invariants;

  async processOrder(order: Order): Promise<OrderResult> {
    // Validate against schema
    const validated = OrderSchema.parse(order);
    
    // Check preconditions
    this.checkInvariants(validated, 'pre');
    
    // Process with state machine
    const result = await this.stateMachine.process(validated);
    
    // Check postconditions
    this.checkInvariants(result, 'post');
    
    // Runtime property verification in dev/staging
    if (process.env.ENABLE_RUNTIME_CHECKS) {
      this.verifyProperties(validated, result);
    }
    
    return result;
  }

  private checkInvariants(data: any, phase: 'pre' | 'post') {
    for (const [name, check] of Object.entries(this.invariants)) {
      if (!check(data)) {
        throw new InvariantViolation(`${phase}: ${name}`, data);
      }
    }
  }
}

// Discovered properties during implementation
test('new property: orders preserve customer preferences', () => {
  fc.assert(fc.property(
    orderWithPreferencesGenerator,
    (order) => {
      const result = processOrder(order);
      return result.shippingMethod === order.customer.preferredShipping;
    }
  ));
});
```

## Stage 5: Test Generation & Validation

### Objective
Generate tests that validate the implementation against schema specifications.

### Testing Types Generated
1. **Schema Validation Tests**: Verify data conforms to schemas
2. **Transformation Tests**: Validate input→output contracts
3. **Property-Based Tests**: From formal invariants
4. **Contract Tests**: API boundary validation

### Example Generated Tests
```typescript
// Tests validate that implementation satisfies schemas
describe('Order Service Schema Compliance', () => {
  test('processOrder respects transformation contract', async () => {
    const input = {
      cart: validCart,
      customer: validCustomer,
      paymentMethod: validPaymentMethod
    };
    
    // Validate input against schema
    const validatedInput = ProcessOrderTransformation.from.parse(input);
    
    // Call implementation
    const result = await orderService.processOrder(validatedInput);
    
    // Validate output against schema
    const validatedOutput = ProcessOrderTransformation.to.parse(result);
    
    // Verify invariants
    expect(validatedOutput.customerId).toBe(input.customer.id);
    expect(validatedOutput.status).toBe('pending');
  });
  
  test('order totals satisfy invariant', () => {
    fc.assert(
      fc.property(orderGenerator, (order) => {
        // Implementation must produce valid schema
        const processed = orderService.calculateTotal(order);
        const validated = OrderSchema.safeParse(processed);
        
        // Schema refinement will catch invariant violations
        return validated.success;
      })
    );
  });
});
```

## Stage 6: Integration & System Testing

### Objective
Verify system-level properties and emergent behaviors.

### Testing at this Stage
- **End-to-End Testing**: Full workflow validation
- **Performance Testing**: Verify SLAs from Stage 0
- **Chaos Engineering**: Test failure modes
- **Load Testing**: Verify system under stress

### Example Tests
```typescript
// System-level property tests
describe('System Properties', () => {
  test('system maintains payment-inventory consistency', async () => {
    // Run 100 concurrent orders
    const orders = fc.sample(OrderFulfillmentSpec.generators.validOrder, 100);
    const results = await Promise.all(orders.map(processOrder));
    
    // Verify system-wide invariants
    const totalPayments = sumPayments(results);
    const inventoryDelta = calculateInventoryDelta(results);
    
    expect(totalPayments).toBe(expectedRevenue(orders));
    expect(inventoryDelta).toEqual(expectedInventoryChange(orders));
  });
});

// Chaos testing (future implementation)
describe('Failure Mode Testing', () => {
  test.todo('handles payment gateway timeout gracefully');
  test.todo('recovers from database connection loss');
  test.todo('maintains consistency during network partition');
});
```

## Stage 7: Runtime Verification & Monitoring

### Objective
Continuous validation in production.

### Testing in Production
- **Synthetic Monitoring**: Run test transactions
- **Invariant Monitoring**: Alert on violations
- **Property Verification**: Sample and verify properties

### Example Implementation
```typescript
// Production monitoring
class ProductionMonitor {
  async verifyInvariants() {
    const sample = await this.sampleRecentOrders(100);
    
    for (const order of sample) {
      for (const [name, check] of Object.entries(OrderFulfillmentSpec.invariants)) {
        if (!check(order)) {
          this.alert({
            severity: 'critical',
            invariant: name,
            order: order.id,
            details: order
          });
        }
      }
    }
  }
}

// Synthetic testing in production
const syntheticTests = {
  async runHappyPath() {
    const testOrder = createSyntheticOrder();
    const result = await processOrder(testOrder);
    
    assert(result.status === 'Completed', 'Synthetic order failed');
    assert(result.duration < 5000, 'SLA violation: >5s processing');
  }
};
```

## Continuous Refinement Loop

### Feedback Integration
1. **Property Discovery**: New invariants found → Update Stage 2 spec
2. **Bug Patterns**: Common failures → Add to fuzz corpus
3. **Performance Issues**: Bottlenecks → Update Stage 0 criteria
4. **Business Changes**: New requirements → Update BPMN (Stage 1)

### Example Refinement
```typescript
// Discovered in production: orders can be partially fulfilled
// Update Stage 2 spec:
OrderFulfillmentSpec.invariants["partial fulfillment"] = (order, result) => {
  if (result.partiallyFulfilled) {
    return result.fulfilledItems.length < order.items.length &&
           result.fulfilledItems.every(item => 
             order.items.some(oi => oi.id === item.id)
           );
  }
  return true;
};

// Generate new tests for partial fulfillment
// Update implementation to handle new case
```

## The Developer Experience

### Session 1: Planning
```
Developer: "I need to build a secure authentication service"

AI Coordinator: "Let's start with Stage 0 - defining the problem..."
[Creates problem statement]

AI Coordinator: "Now for Stage 0.5 - let's break this into manageable nodes..."
[Creates MPC plan with nodes]

AI Coordinator: "Great! You have 5 nodes. The first ready node is 'setup-project-structure'. 
                 When you're ready, we'll work through Stages 1-5 for this node."
```

### Session 2: Working Through First Node
```
Developer: "Let's work on the 'implement-auth-endpoints' node"

AI Coordinator: "Starting Stage 1 for 'implement-auth-endpoints'. 
                 Let's design the BPMN process flow. What happens when a user logs in?"

Developer: "They submit credentials, we validate them, check if account is locked, 
           generate JWT if valid..."

AI Coordinator: "I'll create the BPMN. Here's what I captured:
                 [Shows BPMN JSON with decision points, schemas, invariants]
                 Does this flow look correct? What about MFA?"

Developer: "Oh right, we need to handle MFA after password validation"

AI Coordinator: "Updated. Now Stage 2 - let's extract formal properties from this flow.
                 I see these invariants:
                 - JWT tokens expire within 1 hour
                 - Failed attempts increment lockout counter
                 - MFA codes are single-use
                 Should we add any others?"

Developer: "Add that refresh tokens must be revokable"

AI Coordinator: "Added. Moving to Stage 3 - I'll generate tests for these properties:
                 [Generates property-based tests, deterministic tests, fuzz tests]
                 Here are 23 tests that verify our specifications."

Developer: "Looks good, let's implement"

AI Coordinator: "Stage 4 - I'll implement the auth endpoints to pass all these tests...
                 [Implements code]
                 
                 23/23 tests passing
                 All invariants enforced
                 Schemas validated
                 
                 The 'implement-auth-endpoints' node is complete!"
```

### Key Benefits of This Flow
+ The AI won't let you skip to implementation without specs.
+ Each stage reveals requirements.
+ Test-First Implementation
+ By Stage 4, the implementation has clear success criteria.
+ Traceable Decisions
+ The AI acts as a **methodology enforcer** and **knowledge keeper**:

### Example Multi-Session Flow

**Session 1**: Plan creation
- Output: MPC plan with 5 nodes

**Session 2**: Node "setup-project-structure"
- Stages 1-5 for basic setup
- Output: Working project skeleton

**Session 3**: Node "create-user-model"  
- Stages 1-5 for data model
- Output: Database schema, entities, migrations

**Session 4**: Node "implement-auth-endpoints"
- Stages 1-5 for auth logic
- Output: Working auth endpoints

Each session builds on the previous one, but focuses on one logical unit of work. The AI
maintains context across sessions through the plan file and artifacts.

This approach ensures that by the time the AI implements code, it has:
1. A visual process model (BPMN)
2. Mathematical properties to maintain (Formal spec)
3. Comprehensive tests to pass
4. Clear success criteria

Goal: Make the tests pass while following the spec.

## Summary

This workflow ensures:
1. **Early Testing**: Tests exist before implementation
2. **Comprehensive Coverage**: Property + Deterministic + Fuzz testing
3. **Continuous Validation**: From design to production
4. **Feedback Loops**: Discoveries improve specifications
5. **Runtime Safety**: Invariants enforced in production

The key insight: **Testing isn't a phase, it's woven throughout the entire process**.
