# Rust Schema Specification Examples

## Core Types with Serde

```rust
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};
use rust_decimal::Decimal;

// Newtype pattern for type safety
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct OrderId(Uuid);

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct CustomerId(Uuid);

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct ProductId(Uuid);

// Domain value types
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct Money(Decimal);

impl Money {
    pub fn new(amount: Decimal) -> Result<Self, ValidationError> {
        if amount.is_sign_negative() {
            return Err(ValidationError::NegativeAmount);
        }
        Ok(Money(amount))
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct Quantity(u32);

impl Quantity {
    pub fn new(value: u32) -> Result<Self, ValidationError> {
        if value == 0 {
            return Err(ValidationError::ZeroQuantity);
        }
        Ok(Quantity(value))
    }
}

// State machine with enum
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(tag = "status")]
pub enum OrderState {
    Draft,
    Pending,
    Paid {
        payment_id: String,
    },
    Shipped {
        payment_id: String,
        tracking_number: String,
    },
    Delivered {
        payment_id: String,
        tracking_number: String,
        delivered_at: DateTime<Utc>,
    },
    Cancelled {
        reason: String,
        cancelled_at: DateTime<Utc>,
    },
}

// Domain entity with validation
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItem {
    pub product_id: ProductId,
    pub quantity: Quantity,
    pub price: Money,
    pub subtotal: Money,
}

impl OrderItem {
    pub fn new(
        product_id: ProductId,
        quantity: Quantity,
        price: Money,
    ) -> Result<Self, ValidationError> {
        let subtotal = Money::new(price.0 * Decimal::from(quantity.0))?;
        
        Ok(OrderItem {
            product_id,
            quantity,
            price,
            subtotal,
        })
    }
    
    pub fn validate(&self) -> Result<(), ValidationError> {
        let calculated = self.price.0 * Decimal::from(self.quantity.0);
        if self.subtotal.0 != calculated {
            return Err(ValidationError::InvalidSubtotal);
        }
        Ok(())
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: OrderId,
    pub customer_id: CustomerId,
    pub items: Vec<OrderItem>,
    pub total: Money,
    pub state: OrderState,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl Order {
    pub fn validate(&self) -> Result<(), ValidationError> {
        // Must have items
        if self.items.is_empty() {
            return Err(ValidationError::NoItems);
        }
        
        // Validate each item
        for item in &self.items {
            item.validate()?;
        }
        
        // Total must equal sum of subtotals
        let calculated_total = self.items
            .iter()
            .map(|item| item.subtotal.0)
            .sum::<Decimal>();
            
        if self.total.0 != calculated_total {
            return Err(ValidationError::InvalidTotal);
        }
        
        Ok(())
    }
}
```

## Validation Pipeline with Validator Trait

```rust
use async_trait::async_trait;
use std::net::IpAddr;

// Validation trait for composable validators
#[async_trait]
pub trait Validator<T> {
    type Error;
    async fn validate(&self, input: &T) -> Result<(), Self::Error>;
}

// Login request validation
#[derive(Debug, Deserialize)]
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

pub struct LoginValidator {
    pub min_password_length: usize,
    pub max_attempts: u32,
}

#[async_trait]
impl Validator<LoginRequest> for LoginValidator {
    type Error = ValidationError;
    
    async fn validate(&self, input: &LoginRequest) -> Result<(), Self::Error> {
        // Email validation
        if !input.email.contains('@') {
            return Err(ValidationError::InvalidEmail);
        }
        
        // Password length
        if input.password.len() < self.min_password_length {
            return Err(ValidationError::PasswordTooShort);
        }
        
        // Password complexity
        let has_upper = input.password.chars().any(|c| c.is_uppercase());
        let has_lower = input.password.chars().any(|c| c.is_lowercase());
        let has_digit = input.password.chars().any(|c| c.is_numeric());
        let has_special = input.password.chars().any(|c| "!@#$%^&*".contains(c));
        
        if !(has_upper && has_lower && has_digit && has_special) {
            return Err(ValidationError::WeakPassword);
        }
        
        // Check blocked domains
        let blocked_domains = ["temp-mail.com", "guerrillamail.com"];
        for domain in blocked_domains {
            if input.email.ends_with(&format!("@{}", domain)) {
                return Err(ValidationError::BlockedDomain);
            }
        }
        
        Ok(())
    }
}

// Rate limit validator
pub struct RateLimitValidator {
    cache: Arc<Mutex<HashMap<IpAddr, Vec<Instant>>>>,
    max_attempts: usize,
    window: Duration,
}

#[async_trait]
impl Validator<IpAddr> for RateLimitValidator {
    type Error = ValidationError;
    
    async fn validate(&self, ip: &IpAddr) -> Result<(), Self::Error> {
        let mut cache = self.cache.lock().await;
        let now = Instant::now();
        
        let attempts = cache.entry(*ip).or_insert_with(Vec::new);
        attempts.retain(|&t| now.duration_since(t) < self.window);
        
        if attempts.len() >= self.max_attempts {
            return Err(ValidationError::RateLimitExceeded);
        }
        
        attempts.push(now);
        Ok(())
    }
}

// Composable validation pipeline
pub struct ValidationPipeline<T> {
    validators: Vec<Box<dyn Validator<T, Error = ValidationError>>>,
}

impl<T> ValidationPipeline<T> {
    pub async fn validate(&self, input: &T) -> Result<(), ValidationError> {
        for validator in &self.validators {
            validator.validate(input).await?;
        }
        Ok(())
    }
}
```

## Transformation Specifications

```rust
use std::marker::PhantomData;

// Type-safe transformation trait
pub trait Transform<From, To> {
    type Error;
    fn transform(&self, input: From) -> Result<To, Self::Error>;
}

// Cart to Order transformation
pub struct CartToOrderTransform;

#[derive(Debug, Deserialize)]
pub struct Cart {
    pub items: Vec<CartItem>,
}

#[derive(Debug, Deserialize)]
pub struct CartItem {
    pub product_id: ProductId,
    pub quantity: u32,
}

impl Transform<(Cart, CustomerId), Order> for CartToOrderTransform {
    type Error = TransformError;
    
    fn transform(&self, (cart, customer_id): (Cart, CustomerId)) -> Result<Order, Self::Error> {
        // Validate cart not empty
        if cart.items.is_empty() {
            return Err(TransformError::EmptyCart);
        }
        
        // Transform cart items to order items
        let mut order_items = Vec::new();
        let mut total = Decimal::ZERO;
        
        for cart_item in cart.items {
            // Look up product price (in real app, from database)
            let price = self.get_product_price(&cart_item.product_id)?;
            
            let quantity = Quantity::new(cart_item.quantity)
                .map_err(|_| TransformError::InvalidQuantity)?;
            
            let order_item = OrderItem::new(cart_item.product_id, quantity, price)?;
            total += order_item.subtotal.0;
            order_items.push(order_item);
        }
        
        Ok(Order {
            id: OrderId(Uuid::new_v4()),
            customer_id,
            items: order_items,
            total: Money(total),
            state: OrderState::Pending,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        })
    }
}

// Transformation pipeline with invariant checking
pub struct TransformationPipeline<F, T> {
    transform: Box<dyn Transform<F, T, Error = TransformError>>,
    invariants: Vec<Box<dyn Fn(&T) -> bool>>,
}

impl<F, T> TransformationPipeline<F, T> {
    pub fn execute(&self, input: F) -> Result<T, TransformError> {
        let output = self.transform.transform(input)?;
        
        // Check all invariants
        for (i, invariant) in self.invariants.iter().enumerate() {
            if !invariant(&output) {
                return Err(TransformError::InvariantViolation(i));
            }
        }
        
        Ok(output)
    }
}
```

## Schema Derivation with Macros

```rust
use schemars::JsonSchema;
use validator::Validate;

// Automatic schema generation
#[derive(Debug, Serialize, Deserialize, JsonSchema, Validate)]
pub struct CreateOrderRequest {
    #[validate(length(min = 1))]
    pub items: Vec<OrderItemRequest>,
    
    #[validate]
    pub shipping_address: Address,
    
    #[validate(length(min = 1, max = 100))]
    pub payment_method_id: String,
    
    pub promo_code: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, JsonSchema, Validate)]
pub struct OrderItemRequest {
    pub product_id: Uuid,
    
    #[validate(range(min = 1, max = 100))]
    pub quantity: u32,
}

#[derive(Debug, Serialize, Deserialize, JsonSchema, Validate)]
pub struct Address {
    #[validate(length(min = 1, max = 100))]
    pub street: String,
    
    #[validate(length(min = 1, max = 50))]
    pub city: String,
    
    #[validate(length(min = 2, max = 2))]
    pub state: String,
    
    #[validate(regex = "ZIPCODE_REGEX")]
    pub zip: String,
    
    #[validate(length(min = 2, max = 2))]
    pub country: String,
}

// Custom validation attributes
#[derive(Debug)]
pub struct OrderValidator;

impl OrderValidator {
    pub fn validate_request(req: &CreateOrderRequest) -> Result<(), ValidationError> {
        // Use built-in validation
        req.validate().map_err(|e| ValidationError::InvalidRequest(e.to_string()))?;
        
        // Additional business rules
        if req.items.len() > 50 {
            return Err(ValidationError::TooManyItems);
        }
        
        // Check for duplicate products
        let mut product_ids = HashSet::new();
        for item in &req.items {
            if !product_ids.insert(item.product_id) {
                return Err(ValidationError::DuplicateProduct);
            }
        }
        
        Ok(())
    }
}
```

## Event-Driven Specifications

```rust
use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

// Event trait for type safety
pub trait Event {
    fn event_type(&self) -> &'static str;
    fn timestamp(&self) -> DateTime<Utc>;
}

// Order events using enum
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
pub enum OrderEvent {
    Created {
        order_id: OrderId,
        customer_id: CustomerId,
        total: Money,
        items: Vec<OrderItem>,
        timestamp: DateTime<Utc>,
    },
    Paid {
        order_id: OrderId,
        payment_id: String,
        amount: Money,
        method: PaymentMethod,
        timestamp: DateTime<Utc>,
    },
    Shipped {
        order_id: OrderId,
        tracking_number: String,
        carrier: Carrier,
        estimated_delivery: DateTime<Utc>,
        timestamp: DateTime<Utc>,
    },
    Delivered {
        order_id: OrderId,
        signature: Option<String>,
        timestamp: DateTime<Utc>,
    },
    Cancelled {
        order_id: OrderId,
        reason: String,
        refund_amount: Option<Money>,
        timestamp: DateTime<Utc>,
    },
}

impl Event for OrderEvent {
    fn event_type(&self) -> &'static str {
        match self {
            OrderEvent::Created { .. } => "order.created",
            OrderEvent::Paid { .. } => "order.paid",
            OrderEvent::Shipped { .. } => "order.shipped",
            OrderEvent::Delivered { .. } => "order.delivered",
            OrderEvent::Cancelled { .. } => "order.cancelled",
        }
    }
    
    fn timestamp(&self) -> DateTime<Utc> {
        match self {
            OrderEvent::Created { timestamp, .. } |
            OrderEvent::Paid { timestamp, .. } |
            OrderEvent::Shipped { timestamp, .. } |
            OrderEvent::Delivered { timestamp, .. } |
            OrderEvent::Cancelled { timestamp, .. } => *timestamp,
        }
    }
}

// Event handler trait
#[async_trait]
pub trait EventHandler<E: Event> {
    async fn handle(&self, event: E) -> Result<(), HandleError>;
}

// Type-safe event bus
pub struct EventBus {
    handlers: HashMap<String, Vec<Box<dyn EventHandler<OrderEvent>>>>,
}

impl EventBus {
    pub async fn publish(&self, event: OrderEvent) -> Result<(), PublishError> {
        let event_type = event.event_type();
        
        if let Some(handlers) = self.handlers.get(event_type) {
            for handler in handlers {
                handler.handle(event.clone()).await?;
            }
        }
        
        Ok(())
    }
}
```

## Property-Based Testing Specifications

```rust
use proptest::prelude::*;
use quickcheck::{Arbitrary, Gen};

// Property test generators
impl Arbitrary for OrderId {
    fn arbitrary(g: &mut Gen) -> Self {
        OrderId(Uuid::new_v4())
    }
}

impl Arbitrary for Money {
    fn arbitrary(g: &mut Gen) -> Self {
        let amount = (u32::arbitrary(g) % 100000) as f64 / 100.0;
        Money(Decimal::from_f64(amount).unwrap())
    }
}

// Property strategies
fn order_strategy() -> impl Strategy<Value = Order> {
    (
        any::<OrderId>(),
        any::<CustomerId>(),
        prop::collection::vec(order_item_strategy(), 1..10),
        any::<OrderState>(),
    ).prop_map(|(id, customer_id, items, state)| {
        let total = items.iter().map(|i| i.subtotal.0).sum();
        Order {
            id,
            customer_id,
            items,
            total: Money(total),
            state,
            created_at: Utc::now(),
            updated_at: Utc::now(),
        }
    })
}

// Property tests
#[cfg(test)]
mod property_tests {
    use super::*;
    use proptest::prelude::*;
    
    proptest! {
        #[test]
        fn total_conservation(order in order_strategy()) {
            let calculated_total = order.items
                .iter()
                .map(|item| item.subtotal.0)
                .sum::<Decimal>();
            
            prop_assert_eq!(order.total.0, calculated_total);
        }
        
        #[test]
        fn positive_quantities(order in order_strategy()) {
            for item in &order.items {
                prop_assert!(item.quantity.0 > 0);
            }
        }
        
        #[test]
        fn valid_timestamps(order in order_strategy()) {
            prop_assert!(order.created_at <= order.updated_at);
        }
        
        #[test]
        fn state_transitions_valid(
            initial_state in any::<OrderState>(),
            action in any::<OrderAction>()
        ) {
            let result = apply_transition(initial_state.clone(), action);
            
            match result {
                Ok(new_state) => {
                    prop_assert!(is_valid_transition(&initial_state, &new_state));
                }
                Err(_) => {
                    prop_assert!(!can_transition(&initial_state, &action));
                }
            }
        }
    }
}
```

## API Contract with Actix-Web

```rust
use actix_web::{web, HttpResponse, Result};
use garde::Validate;

// Request/Response types with validation
#[derive(Debug, Deserialize, Validate)]
pub struct CreateOrderRequest {
    #[garde(length(min = 1))]
    pub items: Vec<OrderItemRequest>,
    
    #[garde(dive)]
    pub shipping_address: AddressRequest,
    
    #[garde(length(min = 1, max = 100))]
    pub payment_method_id: String,
}

// API handler with contract enforcement
pub async fn create_order(
    req: web::Json<CreateOrderRequest>,
    auth: BearerAuth,
    db: web::Data<Database>,
) -> Result<HttpResponse> {
    // Validate request
    req.validate(&())
        .map_err(|e| ErrorBadRequest(e.to_string()))?;
    
    // Check authentication
    let user_id = auth.validate().await
        .map_err(|_| ErrorUnauthorized("Invalid token"))?;
    
    // Transform request to domain model
    let order = CartToOrderTransform
        .transform((req.into_inner(), user_id))?;
    
    // Validate domain invariants
    order.validate()
        .map_err(|e| ErrorBadRequest(e.to_string()))?;
    
    // Persist order
    let saved_order = db.save_order(order).await
        .map_err(|e| ErrorInternalServerError(e.to_string()))?;
    
    Ok(HttpResponse::Created().json(saved_order))
}

// Route configuration with OpenAPI schema
pub fn configure_routes(cfg: &mut web::ServiceConfig) {
    cfg.service(
        web::resource("/api/orders")
            .route(web::post().to(create_order))
            .wrap(RateLimitMiddleware::new(5, Duration::from_secs(60)))
    );
}
```