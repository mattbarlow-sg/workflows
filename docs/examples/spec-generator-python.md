# Python Schema Specification Examples

## Core Types with Pydantic

```python
from pydantic import BaseModel, Field, validator, root_validator
from typing import Optional, List, Union, Literal
from decimal import Decimal
from datetime import datetime
from uuid import UUID, uuid4
from enum import Enum

# Custom types for type safety
class OrderId(UUID):
    """Branded OrderId type"""
    @classmethod
    def __get_validators__(cls):
        yield cls.validate
    
    @classmethod
    def validate(cls, v):
        if isinstance(v, UUID):
            return cls(str(v))
        return cls(v)

class CustomerId(UUID):
    """Branded CustomerId type"""
    pass

class ProductId(UUID):
    """Branded ProductId type"""
    pass

# Domain value types with validation
class Money(BaseModel):
    amount: Decimal = Field(..., ge=0, decimal_places=2)
    
    @validator('amount')
    def validate_positive(cls, v):
        if v < 0:
            raise ValueError('Money amount cannot be negative')
        return v.quantize(Decimal('0.01'))
    
    def __add__(self, other: 'Money') -> 'Money':
        return Money(amount=self.amount + other.amount)
    
    def __mul__(self, quantity: int) -> 'Money':
        return Money(amount=self.amount * quantity)

class Quantity(BaseModel):
    value: int = Field(..., gt=0, le=1000)
    
    @validator('value')
    def validate_positive(cls, v):
        if v <= 0:
            raise ValueError('Quantity must be greater than zero')
        return v

# State machine using discriminated unions
class DraftState(BaseModel):
    status: Literal['draft'] = 'draft'

class PendingState(BaseModel):
    status: Literal['pending'] = 'pending'

class PaidState(BaseModel):
    status: Literal['paid'] = 'paid'
    payment_id: str

class ShippedState(BaseModel):
    status: Literal['shipped'] = 'shipped'
    payment_id: str
    tracking_number: str

class DeliveredState(BaseModel):
    status: Literal['delivered'] = 'delivered'
    payment_id: str
    tracking_number: str
    delivered_at: datetime

class CancelledState(BaseModel):
    status: Literal['cancelled'] = 'cancelled'
    reason: str
    cancelled_at: datetime

OrderState = Union[
    DraftState,
    PendingState,
    PaidState,
    ShippedState,
    DeliveredState,
    CancelledState
]

# Domain entity with invariants
class OrderItem(BaseModel):
    product_id: ProductId
    quantity: Quantity
    price: Money
    subtotal: Money
    
    @root_validator
    def validate_subtotal(cls, values):
        quantity = values.get('quantity')
        price = values.get('price')
        subtotal = values.get('subtotal')
        
        if quantity and price and subtotal:
            expected = Money(amount=price.amount * quantity.value)
            if subtotal.amount != expected.amount:
                raise ValueError('Subtotal must equal price * quantity')
        
        return values
    
    class Config:
        json_encoders = {
            UUID: str,
            Decimal: float
        }

class Order(BaseModel):
    id: OrderId
    customer_id: CustomerId
    items: List[OrderItem] = Field(..., min_items=1)
    total: Money
    state: OrderState
    created_at: datetime
    updated_at: datetime
    
    @root_validator
    def validate_total(cls, values):
        items = values.get('items', [])
        total = values.get('total')
        
        if items and total:
            calculated_total = sum(
                (item.subtotal.amount for item in items),
                Decimal('0')
            )
            if total.amount != calculated_total:
                raise ValueError('Order total must equal sum of item subtotals')
        
        return values
    
    @validator('updated_at')
    def validate_timestamps(cls, v, values):
        created_at = values.get('created_at')
        if created_at and v < created_at:
            raise ValueError('Updated time cannot be before created time')
        return v
    
    class Config:
        json_encoders = {
            UUID: str,
            Decimal: float,
            datetime: lambda v: v.isoformat()
        }
```

## Validation Pipeline with Decorators

```python
from functools import wraps
from typing import Callable, Any, Dict, List
import re
import asyncio
from datetime import datetime, timedelta
from collections import defaultdict
import ipaddress

# Validation decorator pattern
def validate(*validators: Callable) -> Callable:
    """Decorator to apply multiple validators to a function"""
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs):
            # Run all validators
            for validator in validators:
                result = await validator(*args, **kwargs)
                if not result:
                    raise ValidationError(f"Validation failed: {validator.__name__}")
            # Call original function
            return await func(*args, **kwargs)
        return wrapper
    return decorator

# Login request validation
class LoginRequest(BaseModel):
    email: str = Field(..., regex=r'^[\w\.-]+@[\w\.-]+\.\w+$')
    password: str = Field(..., min_length=8, max_length=128)

class PasswordComplexity(BaseModel):
    require_uppercase: bool = True
    require_lowercase: bool = True
    require_digit: bool = True
    require_special: bool = True
    min_length: int = 8
    max_length: int = 128

class LoginValidator:
    def __init__(
        self,
        password_complexity: PasswordComplexity,
        blocked_domains: List[str]
    ):
        self.password_complexity = password_complexity
        self.blocked_domains = blocked_domains
    
    def validate_email(self, email: str) -> bool:
        """Validate email format and domain"""
        # Check blocked domains
        domain = email.split('@')[-1].lower()
        if domain in self.blocked_domains:
            raise ValidationError(f"Email domain {domain} is blocked")
        
        return True
    
    def validate_password(self, password: str) -> bool:
        """Validate password complexity"""
        if len(password) < self.password_complexity.min_length:
            raise ValidationError("Password too short")
        
        if len(password) > self.password_complexity.max_length:
            raise ValidationError("Password too long")
        
        if self.password_complexity.require_uppercase:
            if not re.search(r'[A-Z]', password):
                raise ValidationError("Password must contain uppercase letter")
        
        if self.password_complexity.require_lowercase:
            if not re.search(r'[a-z]', password):
                raise ValidationError("Password must contain lowercase letter")
        
        if self.password_complexity.require_digit:
            if not re.search(r'\d', password):
                raise ValidationError("Password must contain digit")
        
        if self.password_complexity.require_special:
            if not re.search(r'[!@#$%^&*(),.?":{}|<>]', password):
                raise ValidationError("Password must contain special character")
        
        return True
    
    async def validate_request(self, request: LoginRequest) -> bool:
        """Validate complete login request"""
        self.validate_email(request.email)
        self.validate_password(request.password)
        return True

# Rate limiter with async support
class RateLimiter:
    def __init__(self, max_attempts: int, window: timedelta):
        self.max_attempts = max_attempts
        self.window = window
        self.attempts: Dict[str, List[datetime]] = defaultdict(list)
        self._lock = asyncio.Lock()
    
    async def check_rate_limit(self, identifier: str) -> bool:
        """Check if identifier has exceeded rate limit"""
        async with self._lock:
            now = datetime.now()
            
            # Clean old attempts
            self.attempts[identifier] = [
                attempt for attempt in self.attempts[identifier]
                if now - attempt < self.window
            ]
            
            # Check limit
            if len(self.attempts[identifier]) >= self.max_attempts:
                raise ValidationError("Rate limit exceeded")
            
            # Record attempt
            self.attempts[identifier].append(now)
            return True

# Validation pipeline class
class ValidationPipeline:
    def __init__(self):
        self.validators: List[Callable] = []
    
    def add_validator(self, validator: Callable) -> 'ValidationPipeline':
        """Add a validator to the pipeline"""
        self.validators.append(validator)
        return self
    
    async def validate(self, data: Any) -> bool:
        """Run all validators in sequence"""
        for validator in self.validators:
            if asyncio.iscoroutinefunction(validator):
                result = await validator(data)
            else:
                result = validator(data)
            
            if not result:
                raise ValidationError(f"Validation failed at {validator.__name__}")
        
        return True
```

## Transformation Specifications

```python
from typing import Protocol, TypeVar, Generic, Optional
from abc import ABC, abstractmethod
import asyncio

# Generic transformation protocol
Input = TypeVar('Input')
Output = TypeVar('Output')

class Transform(Protocol[Input, Output]):
    """Protocol for type-safe transformations"""
    async def transform(self, input: Input) -> Output:
        ...

# Cart to Order transformation
class Cart(BaseModel):
    items: List['CartItem']
    
class CartItem(BaseModel):
    product_id: ProductId
    quantity: int = Field(..., gt=0)

class TransformInput(BaseModel):
    cart: Cart
    customer_id: CustomerId

class CartToOrderTransform:
    def __init__(self, product_service, pricing_service):
        self.product_service = product_service
        self.pricing_service = pricing_service
    
    async def transform(self, input: TransformInput) -> Order:
        """Transform cart to order with validation"""
        if not input.cart.items:
            raise TransformError("Cart is empty")
        
        order_items = []
        total = Decimal('0')
        
        for cart_item in input.cart.items:
            # Get product details
            product = await self.product_service.get_product(cart_item.product_id)
            
            # Check inventory
            if product.stock < cart_item.quantity:
                raise TransformError(
                    f"Insufficient inventory for product {cart_item.product_id}"
                )
            
            # Get pricing
            price = await self.pricing_service.get_price(cart_item.product_id)
            
            # Create order item
            quantity = Quantity(value=cart_item.quantity)
            order_item = OrderItem(
                product_id=cart_item.product_id,
                quantity=quantity,
                price=price,
                subtotal=Money(amount=price.amount * quantity.value)
            )
            
            order_items.append(order_item)
            total += order_item.subtotal.amount
        
        # Create order
        order = Order(
            id=OrderId(uuid4()),
            customer_id=input.customer_id,
            items=order_items,
            total=Money(amount=total),
            state=PendingState(),
            created_at=datetime.now(),
            updated_at=datetime.now()
        )
        
        return order

# Transformation pipeline with invariants
class TransformationPipeline(Generic[Input, Output]):
    def __init__(
        self,
        transform: Transform[Input, Output],
        invariants: Optional[List[Callable[[Output], bool]]] = None
    ):
        self.transform = transform
        self.invariants = invariants or []
    
    async def execute(self, input: Input) -> Output:
        """Execute transformation and check invariants"""
        # Apply transformation
        output = await self.transform.transform(input)
        
        # Check invariants
        for i, invariant in enumerate(self.invariants):
            if not invariant(output):
                raise TransformError(f"Invariant {i} violated")
        
        return output

# Composable transformations
class ComposedTransform(Generic[Input, Output]):
    def __init__(self):
        self.steps: List[Callable] = []
    
    def add_step(self, transform: Callable) -> 'ComposedTransform':
        """Add a transformation step"""
        self.steps.append(transform)
        return self
    
    async def transform(self, input: Input) -> Output:
        """Execute all transformation steps"""
        result = input
        for step in self.steps:
            if asyncio.iscoroutinefunction(step):
                result = await step(result)
            else:
                result = step(result)
        return result
```

## Schema Validation with Marshmallow

```python
from marshmallow import Schema, fields, validate, validates_schema, ValidationError
from marshmallow_dataclass import dataclass as marshmallow_dataclass
import marshmallow_dataclass

# Request/Response schemas
class AddressSchema(Schema):
    street = fields.Str(required=True, validate=validate.Length(min=1, max=100))
    city = fields.Str(required=True, validate=validate.Length(min=1, max=50))
    state = fields.Str(required=True, validate=validate.Length(equal=2))
    zip = fields.Str(required=True, validate=validate.Regexp(r'^\d{5}(-\d{4})?$'))
    country = fields.Str(required=True, validate=validate.Length(equal=2))

class OrderItemRequestSchema(Schema):
    product_id = fields.UUID(required=True)
    quantity = fields.Int(required=True, validate=validate.Range(min=1, max=100))

class CreateOrderRequestSchema(Schema):
    items = fields.List(
        fields.Nested(OrderItemRequestSchema),
        required=True,
        validate=validate.Length(min=1, max=50)
    )
    shipping_address = fields.Nested(AddressSchema, required=True)
    payment_method_id = fields.Str(
        required=True,
        validate=validate.Length(min=1, max=100)
    )
    promo_code = fields.Str(
        missing=None,
        validate=validate.Length(min=3, max=20)
    )
    
    @validates_schema
    def validate_no_duplicates(self, data, **kwargs):
        """Check for duplicate products"""
        product_ids = [item['product_id'] for item in data.get('items', [])]
        if len(product_ids) != len(set(product_ids)):
            raise ValidationError('Duplicate products in order')

# Using dataclasses with Marshmallow
@marshmallow_dataclass
class OrderResponse:
    id: UUID
    customer_id: UUID
    items: List[OrderItem]
    total: Decimal
    status: str
    created_at: datetime
    updated_at: datetime
    
    class Meta:
        ordered = True
```

## Event-Driven Specifications

```python
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Dict, List, Type
from datetime import datetime
import json
import asyncio

# Event base class
@dataclass
class Event(ABC):
    event_id: UUID = field(default_factory=uuid4)
    timestamp: datetime = field(default_factory=datetime.now)
    
    @property
    @abstractmethod
    def event_type(self) -> str:
        pass
    
    @property
    @abstractmethod
    def aggregate_id(self) -> str:
        pass

# Order events
@dataclass
class OrderCreated(Event):
    order_id: OrderId
    customer_id: CustomerId
    total: Money
    items: List[OrderItem]
    
    @property
    def event_type(self) -> str:
        return "order.created"
    
    @property
    def aggregate_id(self) -> str:
        return str(self.order_id)

@dataclass
class OrderPaid(Event):
    order_id: OrderId
    payment_id: str
    amount: Money
    method: str
    
    @property
    def event_type(self) -> str:
        return "order.paid"
    
    @property
    def aggregate_id(self) -> str:
        return str(self.order_id)

@dataclass
class OrderShipped(Event):
    order_id: OrderId
    tracking_number: str
    carrier: str
    estimated_delivery: datetime
    
    @property
    def event_type(self) -> str:
        return "order.shipped"
    
    @property
    def aggregate_id(self) -> str:
        return str(self.order_id)

# Event handler protocol
class EventHandler(Protocol):
    async def handle(self, event: Event) -> None:
        ...

# Order event handler
class OrderEventHandler:
    def __init__(self, order_repo, notification_service):
        self.order_repo = order_repo
        self.notification_service = notification_service
    
    async def handle(self, event: Event) -> None:
        """Handle order events"""
        if isinstance(event, OrderCreated):
            await self._handle_order_created(event)
        elif isinstance(event, OrderPaid):
            await self._handle_order_paid(event)
        elif isinstance(event, OrderShipped):
            await self._handle_order_shipped(event)
        else:
            raise ValueError(f"Unknown event type: {event.event_type}")
    
    async def _handle_order_created(self, event: OrderCreated) -> None:
        # Send confirmation email
        await self.notification_service.send_order_confirmation(
            event.customer_id,
            event.order_id
        )
    
    async def _handle_order_paid(self, event: OrderPaid) -> None:
        # Update order status
        order = await self.order_repo.get(event.order_id)
        order.state = PaidState(payment_id=event.payment_id)
        await self.order_repo.save(order)
    
    async def _handle_order_shipped(self, event: OrderShipped) -> None:
        # Send shipping notification
        await self.notification_service.send_shipping_notification(
            event.order_id,
            event.tracking_number
        )

# Event bus implementation
class EventBus:
    def __init__(self):
        self.handlers: Dict[str, List[EventHandler]] = defaultdict(list)
    
    def subscribe(self, event_type: str, handler: EventHandler) -> None:
        """Subscribe handler to event type"""
        self.handlers[event_type].append(handler)
    
    async def publish(self, event: Event) -> None:
        """Publish event to all subscribers"""
        handlers = self.handlers.get(event.event_type, [])
        
        # Run handlers concurrently
        tasks = [handler.handle(event) for handler in handlers]
        await asyncio.gather(*tasks)
```

## Property-Based Testing with Hypothesis

```python
import hypothesis.strategies as st
from hypothesis import given, assume, settings, example
from hypothesis.stateful import RuleBasedStateMachine, rule, invariant
import pytest

# Custom strategies for domain types
@st.composite
def money_strategy(draw):
    amount = draw(st.decimals(min_value=0, max_value=10000, places=2))
    return Money(amount=amount)

@st.composite
def quantity_strategy(draw):
    value = draw(st.integers(min_value=1, max_value=100))
    return Quantity(value=value)

@st.composite
def order_item_strategy(draw):
    product_id = ProductId(uuid4())
    quantity = draw(quantity_strategy())
    price = draw(money_strategy())
    subtotal = Money(amount=price.amount * quantity.value)
    
    return OrderItem(
        product_id=product_id,
        quantity=quantity,
        price=price,
        subtotal=subtotal
    )

@st.composite
def order_strategy(draw):
    items = draw(st.lists(order_item_strategy(), min_size=1, max_size=10))
    total = sum((item.subtotal.amount for item in items), Decimal('0'))
    
    return Order(
        id=OrderId(uuid4()),
        customer_id=CustomerId(uuid4()),
        items=items,
        total=Money(amount=total),
        state=PendingState(),
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

# Property tests
class TestOrderProperties:
    @given(order_strategy())
    def test_total_conservation(self, order: Order):
        """Total must equal sum of item subtotals"""
        calculated_total = sum(
            (item.subtotal.amount for item in order.items),
            Decimal('0')
        )
        assert order.total.amount == calculated_total
    
    @given(order_strategy())
    def test_positive_quantities(self, order: Order):
        """All quantities must be positive"""
        for item in order.items:
            assert item.quantity.value > 0
    
    @given(order_strategy())
    def test_valid_timestamps(self, order: Order):
        """Created time must be before or equal to updated time"""
        assert order.created_at <= order.updated_at
    
    @given(order_item_strategy())
    def test_item_subtotal_correct(self, item: OrderItem):
        """Item subtotal must equal price * quantity"""
        expected = item.price.amount * item.quantity.value
        assert item.subtotal.amount == expected

# Stateful testing for order state machine
class OrderStateMachine(RuleBasedStateMachine):
    def __init__(self):
        super().__init__()
        self.order = Order(
            id=OrderId(uuid4()),
            customer_id=CustomerId(uuid4()),
            items=[],
            total=Money(amount=Decimal('0')),
            state=DraftState(),
            created_at=datetime.now(),
            updated_at=datetime.now()
        )
    
    @rule()
    def submit_order(self):
        """Submit draft order"""
        if isinstance(self.order.state, DraftState):
            self.order.state = PendingState()
    
    @rule(payment_id=st.text(min_size=1))
    def pay_order(self, payment_id: str):
        """Pay pending order"""
        if isinstance(self.order.state, PendingState):
            self.order.state = PaidState(payment_id=payment_id)
    
    @rule(tracking_number=st.text(min_size=1))
    def ship_order(self, tracking_number: str):
        """Ship paid order"""
        if isinstance(self.order.state, PaidState):
            self.order.state = ShippedState(
                payment_id=self.order.state.payment_id,
                tracking_number=tracking_number
            )
    
    @rule(reason=st.text(min_size=1))
    def cancel_order(self, reason: str):
        """Cancel order if not shipped"""
        if not isinstance(self.order.state, (ShippedState, DeliveredState)):
            self.order.state = CancelledState(
                reason=reason,
                cancelled_at=datetime.now()
            )
    
    @invariant()
    def valid_state(self):
        """Order must always have a valid state"""
        assert self.order.state is not None
        assert hasattr(self.order.state, 'status')

# Run stateful test
TestOrderStateMachine = OrderStateMachine.TestCase
```

## API Contract with FastAPI

```python
from fastapi import FastAPI, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from fastapi.middleware.cors import CORSMiddleware
from typing import Annotated
import uvicorn

app = FastAPI(title="Order API", version="1.0.0")

# Security
security = HTTPBearer()

async def get_current_user(
    credentials: Annotated[HTTPAuthorizationCredentials, Depends(security)]
) -> User:
    """Validate JWT token and return user"""
    # Token validation logic here
    pass

# Rate limiting dependency
async def rate_limit_check(request: Request) -> None:
    """Check rate limit for IP"""
    client_ip = request.client.host
    # Rate limiting logic here
    pass

# API endpoints with automatic OpenAPI documentation
@app.post(
    "/api/orders",
    response_model=OrderResponse,
    status_code=status.HTTP_201_CREATED,
    summary="Create a new order",
    description="Create a new order from cart items",
    dependencies=[Depends(rate_limit_check)]
)
async def create_order(
    request: CreateOrderRequest,
    current_user: Annotated[User, Depends(get_current_user)],
    order_service: Annotated[OrderService, Depends()]
) -> OrderResponse:
    """
    Create a new order with the following validations:
    - User must be authenticated
    - Rate limit: 5 requests per minute
    - Cart must not be empty
    - All products must be in stock
    - Payment method must be valid
    """
    # Validate request
    try:
        request_dict = request.dict()
        validated = CreateOrderRequestSchema().load(request_dict)
    except ValidationError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )
    
    # Transform to domain model
    transform = CartToOrderTransform(
        product_service=order_service,
        pricing_service=order_service
    )
    
    try:
        order = await transform.transform(
            TransformInput(
                cart=Cart(items=request.items),
                customer_id=CustomerId(current_user.id)
            )
        )
    except TransformError as e:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=str(e)
        )
    
    # Save order
    saved_order = await order_service.create_order(order)
    
    return OrderResponse.from_domain(saved_order)

@app.get(
    "/api/orders/{order_id}",
    response_model=OrderResponse,
    summary="Get order by ID"
)
async def get_order(
    order_id: UUID,
    current_user: Annotated[User, Depends(get_current_user)],
    order_service: Annotated[OrderService, Depends()]
) -> OrderResponse:
    """Get order details by ID"""
    order = await order_service.get_order(OrderId(order_id))
    
    # Check ownership
    if order.customer_id != CustomerId(current_user.id):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Not authorized to view this order"
        )
    
    return OrderResponse.from_domain(order)

# Error handlers
@app.exception_handler(ValidationError)
async def validation_exception_handler(request: Request, exc: ValidationError):
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={"detail": exc.errors()}
    )

# Middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```