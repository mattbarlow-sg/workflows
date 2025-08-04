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

## Stage 3: Test Generation & Implementation Planning

### Objective
Generate comprehensive test suite before implementation.

### Testing Types Generated
1. **Property-Based Tests**: From formal spec
2. **Deterministic Simulation Tests**: For workflows
3. **Fuzz Tests**: For boundaries and parsers
4. **Contract Tests**: For API boundaries

### Example Generated Tests
```typescript
// 1. Property-based tests (from Stage 2 spec)
describe('Order Properties', () => {
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

## Stage 4: Implementation

### Objective
Implement system to pass all pre-written tests.

### Testing During Implementation
- **Test-Driven Development**: Tests already exist from Stage 3
- **Runtime Invariant Checking**: Embed invariants in code
- **Continuous Property Discovery**: Add new properties as discovered

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

## Stage 5: Integration & System Testing

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

## Stage 6: Runtime Verification & Monitoring

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

## Summary

This workflow ensures:
1. **Early Testing**: Tests exist before implementation
2. **Comprehensive Coverage**: Property + Deterministic + Fuzz testing
3. **Continuous Validation**: From design to production
4. **Feedback Loops**: Discoveries improve specifications
5. **Runtime Safety**: Invariants enforced in production

The key insight: **Testing isn't a phase, it's woven throughout the entire process**.
