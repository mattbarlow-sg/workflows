# TypeScript Schema Specification Examples

## Core Schema with Zod

```typescript
import { z } from 'zod';

// Branded types for type safety
export const OrderIdSchema = z.string().uuid().brand('OrderId');
export const CustomerIdSchema = z.string().uuid().brand('CustomerId');
export const ProductIdSchema = z.string().uuid().brand('ProductId');

// Domain value objects
export const MoneySchema = z.number()
  .positive()
  .multipleOf(0.01)
  .brand('Money');

export const QuantitySchema = z.number()
  .int()
  .positive()
  .brand('Quantity');

// State machine types with discriminated unions
export type OrderState = 
  | { status: 'draft'; paymentId?: never; trackingNumber?: never }
  | { status: 'pending'; paymentId?: never; trackingNumber?: never }
  | { status: 'paid'; paymentId: string; trackingNumber?: never }
  | { status: 'shipped'; paymentId: string; trackingNumber: string }
  | { status: 'delivered'; paymentId: string; trackingNumber: string }
  | { status: 'cancelled'; reason: string };

// Complex domain entity
export const OrderItemSchema = z.object({
  productId: ProductIdSchema,
  quantity: QuantitySchema,
  price: MoneySchema,
  subtotal: MoneySchema
}).refine(
  item => item.subtotal === item.price * item.quantity,
  "Subtotal must equal price * quantity"
);

export const OrderSchema = z.object({
  id: OrderIdSchema,
  customerId: CustomerIdSchema,
  items: z.array(OrderItemSchema).min(1),
  total: MoneySchema,
  status: z.enum(['draft', 'pending', 'paid', 'shipped', 'delivered', 'cancelled']),
  createdAt: z.date(),
  updatedAt: z.date()
}).refine(
  order => order.total === order.items.reduce((sum, item) => sum + item.subtotal, 0),
  "Order total must equal sum of item subtotals"
);
```

## Validation Pipeline

```typescript
import { z } from 'zod';

// Multi-stage validation pipeline
export const LoginValidationPipeline = {
  // Stage 1: Input validation
  validateInput: z.object({
    email: z.string().email().toLowerCase().trim(),
    password: z.string().min(8).max(128)
  }),
  
  // Stage 2: Business rule validations
  validateBusinessRules: {
    passwordComplexity: (password: string): boolean => {
      const hasUpperCase = /[A-Z]/.test(password);
      const hasLowerCase = /[a-z]/.test(password);
      const hasNumbers = /\d/.test(password);
      const hasSpecialChar = /[!@#$%^&*]/.test(password);
      return hasUpperCase && hasLowerCase && hasNumbers && hasSpecialChar;
    },
    
    emailDomainAllowed: (email: string): boolean => {
      const blockedDomains = ['temp-mail.com', 'guerrillamail.com'];
      return !blockedDomains.some(domain => email.endsWith(`@${domain}`));
    },
    
    rateLimitCheck: async (ip: string): Promise<boolean> => {
      // Check rate limit from cache/db
      const attempts = await getLoginAttempts(ip);
      return attempts < 5;
    }
  },
  
  // Stage 3: Output validation with invariants
  validateOutput: z.object({
    token: z.string().min(1),
    refreshToken: z.string().min(1),
    expiresIn: z.number().int().positive().max(3600),
    userId: CustomerIdSchema
  }).refine(
    data => data.expiresIn <= 3600,
    "Token expiry must not exceed 1 hour"
  )
};
```

## Transformation Contracts

```typescript
import { z } from 'zod';

// Type-safe transformation specification
export const ProcessOrderTransformation = {
  name: 'processOrder',
  
  // Input schema
  from: z.object({
    cart: z.object({
      items: z.array(z.object({
        productId: ProductIdSchema,
        quantity: QuantitySchema
      }))
    }),
    customer: z.object({
      id: CustomerIdSchema,
      email: z.string().email(),
      shippingAddress: AddressSchema
    }),
    paymentMethod: z.enum(['credit_card', 'paypal', 'stripe'])
  }),
  
  // Output schema
  to: OrderSchema,
  
  // Transformation invariants
  invariants: [
    'output.customerId === input.customer.id',
    'output.items.length === input.cart.items.length',
    'output.status === "pending"',
    'output.total > 0'
  ],
  
  // Error specifications
  errors: {
    INSUFFICIENT_INVENTORY: z.object({
      code: z.literal('INSUFFICIENT_INVENTORY'),
      productId: ProductIdSchema,
      requested: QuantitySchema,
      available: QuantitySchema
    }),
    
    INVALID_PAYMENT: z.object({
      code: z.literal('INVALID_PAYMENT'),
      method: z.string(),
      reason: z.string()
    }),
    
    CUSTOMER_BLOCKED: z.object({
      code: z.literal('CUSTOMER_BLOCKED'),
      customerId: CustomerIdSchema,
      reason: z.string()
    })
  },
  
  // Transformation pipeline stages
  pipeline: [
    'validateCart',
    'checkInventory',
    'calculatePricing',
    'validatePayment',
    'createOrder',
    'notifyCustomer'
  ]
};
```

## Schema Composition

```typescript
import { z } from 'zod';

// Base schemas for composition
const TimestampedSchema = z.object({
  createdAt: z.date(),
  updatedAt: z.date()
});

const IdentifiableSchema = z.object({
  id: z.string().uuid()
});

const SoftDeletableSchema = z.object({
  deletedAt: z.date().nullable().optional()
});

// Composed domain entities
export const UserSchema = z.intersection(
  IdentifiableSchema,
  z.intersection(
    TimestampedSchema,
    z.intersection(
      SoftDeletableSchema,
      z.object({
        email: z.string().email(),
        passwordHash: z.string(),
        verified: z.boolean(),
        roles: z.array(z.enum(['user', 'admin', 'moderator'])),
        profile: z.object({
          firstName: z.string(),
          lastName: z.string(),
          avatar: z.string().url().optional()
        })
      })
    )
  )
);

export const SessionSchema = z.intersection(
  IdentifiableSchema,
  z.object({
    userId: z.string().uuid(),
    token: z.string(),
    refreshToken: z.string(),
    expiresAt: z.date(),
    createdAt: z.date(),
    lastActivity: z.date()
  })
).refine(
  session => session.expiresAt > new Date(),
  "Session must not be expired"
).refine(
  session => session.lastActivity <= new Date(),
  "Last activity cannot be in the future"
);
```

## API Contract Specification

```typescript
import { z } from 'zod';

// REST API endpoint contract
export const CreateOrderEndpoint = {
  method: 'POST' as const,
  path: '/api/orders',
  
  request: {
    headers: z.object({
      'authorization': z.string().regex(/^Bearer .+/),
      'content-type': z.literal('application/json')
    }),
    
    body: z.object({
      items: z.array(z.object({
        productId: ProductIdSchema,
        quantity: QuantitySchema
      })).min(1),
      shippingAddress: AddressSchema,
      paymentMethodId: z.string()
    }),
    
    query: z.object({
      promocode: z.string().optional()
    }).optional()
  },
  
  responses: {
    201: {
      description: 'Order created successfully',
      body: OrderSchema
    },
    400: {
      description: 'Invalid request',
      body: z.object({
        error: z.string(),
        details: z.array(z.object({
          field: z.string(),
          message: z.string()
        }))
      })
    },
    401: {
      description: 'Unauthorized',
      body: z.object({
        error: z.literal('Unauthorized')
      })
    },
    409: {
      description: 'Inventory conflict',
      body: z.object({
        error: z.literal('INSUFFICIENT_INVENTORY'),
        items: z.array(z.object({
          productId: ProductIdSchema,
          available: QuantitySchema
        }))
      })
    }
  }
};
```

## Event Schema Specification

```typescript
import { z } from 'zod';

// Event-driven architecture schemas
export const OrderEventSchema = z.discriminatedUnion('type', [
  z.object({
    type: z.literal('OrderCreated'),
    timestamp: z.date(),
    payload: z.object({
      orderId: OrderIdSchema,
      customerId: CustomerIdSchema,
      total: MoneySchema,
      items: z.array(OrderItemSchema)
    })
  }),
  
  z.object({
    type: z.literal('OrderPaid'),
    timestamp: z.date(),
    payload: z.object({
      orderId: OrderIdSchema,
      paymentId: z.string(),
      amount: MoneySchema,
      method: z.enum(['credit_card', 'paypal', 'stripe'])
    })
  }),
  
  z.object({
    type: z.literal('OrderShipped'),
    timestamp: z.date(),
    payload: z.object({
      orderId: OrderIdSchema,
      trackingNumber: z.string(),
      carrier: z.enum(['fedex', 'ups', 'usps', 'dhl']),
      estimatedDelivery: z.date()
    })
  }),
  
  z.object({
    type: z.literal('OrderCancelled'),
    timestamp: z.date(),
    payload: z.object({
      orderId: OrderIdSchema,
      reason: z.string(),
      refundAmount: MoneySchema.optional()
    })
  })
]);

// Event handler contract
export const OrderEventHandler = {
  subscribeTo: 'orders.*',
  
  handle: async (event: z.infer<typeof OrderEventSchema>) => {
    switch (event.type) {
      case 'OrderCreated':
        // Handle order creation
        break;
      case 'OrderPaid':
        // Handle payment
        break;
      case 'OrderShipped':
        // Handle shipping
        break;
      case 'OrderCancelled':
        // Handle cancellation
        break;
    }
  },
  
  errors: {
    HANDLER_TIMEOUT: 'Event handler timed out after 30s',
    INVALID_EVENT: 'Event does not match schema',
    DUPLICATE_EVENT: 'Event already processed'
  }
};
```

## Testing Contract Specification

```typescript
import { z } from 'zod';

// Property-based testing specification
export const OrderPropertyTests = {
  generators: {
    validOrder: () => ({
      id: crypto.randomUUID(),
      customerId: crypto.randomUUID(),
      items: Array.from({ length: Math.floor(Math.random() * 5) + 1 }, () => ({
        productId: crypto.randomUUID(),
        quantity: Math.floor(Math.random() * 10) + 1,
        price: Math.random() * 100,
        subtotal: 0 // Will be calculated
      })),
      total: 0, // Will be calculated
      status: 'pending' as const,
      createdAt: new Date(),
      updatedAt: new Date()
    }),
    
    invalidOrder: () => ({
      id: 'not-a-uuid',
      customerId: 'not-a-uuid',
      items: [],
      total: -100,
      status: 'invalid-status',
      createdAt: 'not-a-date',
      updatedAt: 'not-a-date'
    })
  },
  
  properties: {
    totalConservation: (order: any) => {
      const calculatedTotal = order.items.reduce(
        (sum: number, item: any) => sum + (item.price * item.quantity), 
        0
      );
      return Math.abs(order.total - calculatedTotal) < 0.01;
    },
    
    positiveQuantities: (order: any) => {
      return order.items.every((item: any) => item.quantity > 0);
    },
    
    validTimestamps: (order: any) => {
      return order.createdAt <= order.updatedAt;
    }
  },
  
  invariants: [
    'All orders must have at least one item',
    'Order total must equal sum of item subtotals',
    'All quantities must be positive integers',
    'Created date must be before or equal to updated date'
  ]
};
```