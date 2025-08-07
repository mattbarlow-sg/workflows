# Go Schema Specification Examples

## Core Types with Validation

```go
package domain

import (
    "errors"
    "time"
    "github.com/google/uuid"
    "github.com/shopspring/decimal"
)

// Custom types for type safety
type OrderID uuid.UUID
type CustomerID uuid.UUID
type ProductID uuid.UUID

// Domain value types with validation
type Money struct {
    Amount decimal.Decimal
}

func NewMoney(amount decimal.Decimal) (Money, error) {
    if amount.IsNegative() {
        return Money{}, errors.New("money amount cannot be negative")
    }
    return Money{Amount: amount}, nil
}

type Quantity uint32

func NewQuantity(value uint32) (Quantity, error) {
    if value == 0 {
        return 0, errors.New("quantity must be greater than zero")
    }
    return Quantity(value), nil
}

// State machine using interfaces
type OrderState interface {
    isOrderState()
    Status() string
}

type DraftState struct{}
func (DraftState) isOrderState() {}
func (DraftState) Status() string { return "draft" }

type PendingState struct{}
func (PendingState) isOrderState() {}
func (PendingState) Status() string { return "pending" }

type PaidState struct {
    PaymentID string
}
func (PaidState) isOrderState() {}
func (PaidState) Status() string { return "paid" }

type ShippedState struct {
    PaymentID      string
    TrackingNumber string
}
func (ShippedState) isOrderState() {}
func (ShippedState) Status() string { return "shipped" }

type DeliveredState struct {
    PaymentID      string
    TrackingNumber string
    DeliveredAt    time.Time
}
func (DeliveredState) isOrderState() {}
func (DeliveredState) Status() string { return "delivered" }

type CancelledState struct {
    Reason      string
    CancelledAt time.Time
}
func (CancelledState) isOrderState() {}
func (CancelledState) Status() string { return "cancelled" }

// Domain entity with invariants
type OrderItem struct {
    ProductID ProductID
    Quantity  Quantity
    Price     Money
    Subtotal  Money
}

func NewOrderItem(productID ProductID, quantity Quantity, price Money) (*OrderItem, error) {
    subtotalAmount := price.Amount.Mul(decimal.NewFromInt32(int32(quantity)))
    subtotal, err := NewMoney(subtotalAmount)
    if err != nil {
        return nil, err
    }
    
    item := &OrderItem{
        ProductID: productID,
        Quantity:  quantity,
        Price:     price,
        Subtotal:  subtotal,
    }
    
    if err := item.Validate(); err != nil {
        return nil, err
    }
    
    return item, nil
}

func (oi *OrderItem) Validate() error {
    expectedSubtotal := oi.Price.Amount.Mul(decimal.NewFromInt32(int32(oi.Quantity)))
    if !oi.Subtotal.Amount.Equal(expectedSubtotal) {
        return errors.New("subtotal does not match price * quantity")
    }
    return nil
}

type Order struct {
    ID         OrderID
    CustomerID CustomerID
    Items      []OrderItem
    Total      Money
    State      OrderState
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

func (o *Order) Validate() error {
    if len(o.Items) == 0 {
        return errors.New("order must have at least one item")
    }
    
    calculatedTotal := decimal.Zero
    for _, item := range o.Items {
        if err := item.Validate(); err != nil {
            return err
        }
        calculatedTotal = calculatedTotal.Add(item.Subtotal.Amount)
    }
    
    if !o.Total.Amount.Equal(calculatedTotal) {
        return errors.New("order total does not match sum of item subtotals")
    }
    
    if o.CreatedAt.After(o.UpdatedAt) {
        return errors.New("created time cannot be after updated time")
    }
    
    return nil
}
```

## Validation Pipeline with Interfaces

```go
package validation

import (
    "context"
    "errors"
    "net"
    "regexp"
    "strings"
    "sync"
    "time"
)

// Validator interface for composable validation
type Validator interface {
    Validate(ctx context.Context) error
}

// Login request validation
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginValidator struct {
    Request            *LoginRequest
    MinPasswordLength  int
    BlockedDomains     []string
    PasswordComplexity PasswordRules
}

type PasswordRules struct {
    RequireUppercase bool
    RequireLowercase bool
    RequireDigit     bool
    RequireSpecial   bool
    MinLength        int
    MaxLength        int
}

func (lv *LoginValidator) Validate(ctx context.Context) error {
    // Email validation
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(lv.Request.Email) {
        return errors.New("invalid email format")
    }
    
    // Check blocked domains
    email := strings.ToLower(lv.Request.Email)
    for _, domain := range lv.BlockedDomains {
        if strings.HasSuffix(email, "@"+domain) {
            return errors.New("email domain is blocked")
        }
    }
    
    // Password length
    if len(lv.Request.Password) < lv.PasswordComplexity.MinLength {
        return errors.New("password too short")
    }
    if len(lv.Request.Password) > lv.PasswordComplexity.MaxLength {
        return errors.New("password too long")
    }
    
    // Password complexity
    if lv.PasswordComplexity.RequireUppercase {
        if !regexp.MustCompile(`[A-Z]`).MatchString(lv.Request.Password) {
            return errors.New("password must contain uppercase letter")
        }
    }
    if lv.PasswordComplexity.RequireLowercase {
        if !regexp.MustCompile(`[a-z]`).MatchString(lv.Request.Password) {
            return errors.New("password must contain lowercase letter")
        }
    }
    if lv.PasswordComplexity.RequireDigit {
        if !regexp.MustCompile(`\d`).MatchString(lv.Request.Password) {
            return errors.New("password must contain digit")
        }
    }
    if lv.PasswordComplexity.RequireSpecial {
        if !regexp.MustCompile(`[!@#$%^&*]`).MatchString(lv.Request.Password) {
            return errors.New("password must contain special character")
        }
    }
    
    return nil
}

// Rate limiter with validation
type RateLimiter struct {
    mu           sync.Mutex
    attempts     map[string][]time.Time
    maxAttempts  int
    window       time.Duration
}

func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        attempts:    make(map[string][]time.Time),
        maxAttempts: maxAttempts,
        window:      window,
    }
}

func (rl *RateLimiter) ValidateIP(ip net.IP) error {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    key := ip.String()
    now := time.Now()
    
    // Clean old attempts
    if attempts, exists := rl.attempts[key]; exists {
        var valid []time.Time
        for _, t := range attempts {
            if now.Sub(t) < rl.window {
                valid = append(valid, t)
            }
        }
        rl.attempts[key] = valid
    }
    
    // Check rate limit
    if len(rl.attempts[key]) >= rl.maxAttempts {
        return errors.New("rate limit exceeded")
    }
    
    // Record attempt
    rl.attempts[key] = append(rl.attempts[key], now)
    return nil
}

// Validation pipeline
type ValidationPipeline struct {
    validators []Validator
}

func NewValidationPipeline(validators ...Validator) *ValidationPipeline {
    return &ValidationPipeline{validators: validators}
}

func (vp *ValidationPipeline) Validate(ctx context.Context) error {
    for _, v := range vp.validators {
        if err := v.Validate(ctx); err != nil {
            return err
        }
    }
    return nil
}
```

## Transformation Specifications

```go
package transform

import (
    "context"
    "errors"
    "github.com/google/uuid"
)

// Transform interface for type-safe transformations
type Transform[From any, To any] interface {
    Transform(ctx context.Context, input From) (To, error)
}

// Cart to Order transformation
type CartToOrderTransform struct {
    ProductService ProductService
    PricingService PricingService
}

type Cart struct {
    Items []CartItem `json:"items"`
}

type CartItem struct {
    ProductID ProductID `json:"product_id"`
    Quantity  uint32    `json:"quantity"`
}

type TransformInput struct {
    Cart       Cart
    CustomerID CustomerID
}

func (t *CartToOrderTransform) Transform(ctx context.Context, input TransformInput) (*Order, error) {
    if len(input.Cart.Items) == 0 {
        return nil, errors.New("cart is empty")
    }
    
    var orderItems []OrderItem
    totalAmount := decimal.Zero
    
    for _, cartItem := range input.Cart.Items {
        // Get product details
        product, err := t.ProductService.GetProduct(ctx, cartItem.ProductID)
        if err != nil {
            return nil, err
        }
        
        // Check inventory
        if product.Stock < cartItem.Quantity {
            return nil, errors.New("insufficient inventory")
        }
        
        // Create order item
        quantity, err := NewQuantity(cartItem.Quantity)
        if err != nil {
            return nil, err
        }
        
        price, err := t.PricingService.GetPrice(ctx, cartItem.ProductID)
        if err != nil {
            return nil, err
        }
        
        orderItem, err := NewOrderItem(cartItem.ProductID, quantity, price)
        if err != nil {
            return nil, err
        }
        
        orderItems = append(orderItems, *orderItem)
        totalAmount = totalAmount.Add(orderItem.Subtotal.Amount)
    }
    
    total, err := NewMoney(totalAmount)
    if err != nil {
        return nil, err
    }
    
    order := &Order{
        ID:         OrderID(uuid.New()),
        CustomerID: input.CustomerID,
        Items:      orderItems,
        Total:      total,
        State:      PendingState{},
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }
    
    // Validate invariants
    if err := order.Validate(); err != nil {
        return nil, err
    }
    
    return order, nil
}

// Transformation pipeline with invariant checking
type TransformationPipeline[F any, T any] struct {
    transform  Transform[F, T]
    invariants []func(T) error
}

func NewTransformationPipeline[F any, T any](
    transform Transform[F, T],
    invariants ...func(T) error,
) *TransformationPipeline[F, T] {
    return &TransformationPipeline[F, T]{
        transform:  transform,
        invariants: invariants,
    }
}

func (tp *TransformationPipeline[F, T]) Execute(ctx context.Context, input F) (T, error) {
    output, err := tp.transform.Transform(ctx, input)
    if err != nil {
        var zero T
        return zero, err
    }
    
    // Check invariants
    for _, invariant := range tp.invariants {
        if err := invariant(output); err != nil {
            var zero T
            return zero, err
        }
    }
    
    return output, nil
}
```

## Schema Validation with struct tags

```go
package schema

import (
    "github.com/go-playground/validator/v10"
    "github.com/google/uuid"
)

// Request/Response schemas with validation tags
type CreateOrderRequest struct {
    Items           []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
    ShippingAddress Address           `json:"shipping_address" validate:"required"`
    PaymentMethodID string           `json:"payment_method_id" validate:"required,min=1,max=100"`
    PromoCode       *string          `json:"promo_code,omitempty" validate:"omitempty,min=3,max=20"`
}

type OrderItemRequest struct {
    ProductID uuid.UUID `json:"product_id" validate:"required"`
    Quantity  uint32    `json:"quantity" validate:"required,min=1,max=100"`
}

type Address struct {
    Street  string `json:"street" validate:"required,min=1,max=100"`
    City    string `json:"city" validate:"required,min=1,max=50"`
    State   string `json:"state" validate:"required,len=2"`
    Zip     string `json:"zip" validate:"required,regexp=^[0-9]{5}(-[0-9]{4})?$"`
    Country string `json:"country" validate:"required,len=2"`
}

// Custom validator with business rules
type OrderValidator struct {
    validator *validator.Validate
}

func NewOrderValidator() *OrderValidator {
    v := validator.New()
    
    // Register custom validations
    v.RegisterValidation("zipcode", validateZipCode)
    v.RegisterValidation("promo", validatePromoCode)
    
    return &OrderValidator{validator: v}
}

func (ov *OrderValidator) ValidateRequest(req *CreateOrderRequest) error {
    // Struct validation
    if err := ov.validator.Struct(req); err != nil {
        return err
    }
    
    // Business rules
    if len(req.Items) > 50 {
        return errors.New("too many items in order")
    }
    
    // Check for duplicate products
    productMap := make(map[uuid.UUID]bool)
    for _, item := range req.Items {
        if productMap[item.ProductID] {
            return errors.New("duplicate product in order")
        }
        productMap[item.ProductID] = true
    }
    
    return nil
}

func validateZipCode(fl validator.FieldLevel) bool {
    zip := fl.Field().String()
    matched, _ := regexp.MatchString(`^[0-9]{5}(-[0-9]{4})?$`, zip)
    return matched
}

func validatePromoCode(fl validator.FieldLevel) bool {
    code := fl.Field().String()
    return len(code) >= 3 && len(code) <= 20
}
```

## Event-Driven Specifications

```go
package events

import (
    "context"
    "encoding/json"
    "time"
)

// Event interface
type Event interface {
    EventType() string
    Timestamp() time.Time
    AggregateID() string
}

// Order events
type OrderCreated struct {
    OrderID    OrderID     `json:"order_id"`
    CustomerID CustomerID  `json:"customer_id"`
    Total      Money       `json:"total"`
    Items      []OrderItem `json:"items"`
    CreatedAt  time.Time   `json:"timestamp"`
}

func (e OrderCreated) EventType() string { return "order.created" }
func (e OrderCreated) Timestamp() time.Time { return e.CreatedAt }
func (e OrderCreated) AggregateID() string { return e.OrderID.String() }

type OrderPaid struct {
    OrderID   OrderID       `json:"order_id"`
    PaymentID string        `json:"payment_id"`
    Amount    Money         `json:"amount"`
    Method    PaymentMethod `json:"method"`
    PaidAt    time.Time     `json:"timestamp"`
}

func (e OrderPaid) EventType() string { return "order.paid" }
func (e OrderPaid) Timestamp() time.Time { return e.PaidAt }
func (e OrderPaid) AggregateID() string { return e.OrderID.String() }

type OrderShipped struct {
    OrderID           OrderID   `json:"order_id"`
    TrackingNumber    string    `json:"tracking_number"`
    Carrier           Carrier   `json:"carrier"`
    EstimatedDelivery time.Time `json:"estimated_delivery"`
    ShippedAt         time.Time `json:"timestamp"`
}

func (e OrderShipped) EventType() string { return "order.shipped" }
func (e OrderShipped) Timestamp() time.Time { return e.ShippedAt }
func (e OrderShipped) AggregateID() string { return e.OrderID.String() }

// Event handler interface
type EventHandler interface {
    Handle(ctx context.Context, event Event) error
    CanHandle(eventType string) bool
}

// Order event handler
type OrderEventHandler struct {
    orderRepo OrderRepository
    notifier  NotificationService
}

func (h *OrderEventHandler) Handle(ctx context.Context, event Event) error {
    switch e := event.(type) {
    case OrderCreated:
        return h.handleOrderCreated(ctx, e)
    case OrderPaid:
        return h.handleOrderPaid(ctx, e)
    case OrderShipped:
        return h.handleOrderShipped(ctx, e)
    default:
        return errors.New("unknown event type")
    }
}

func (h *OrderEventHandler) CanHandle(eventType string) bool {
    switch eventType {
    case "order.created", "order.paid", "order.shipped":
        return true
    default:
        return false
    }
}

// Event bus
type EventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        handlers: make(map[string][]EventHandler),
    }
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(ctx context.Context, event Event) error {
    eb.mu.RLock()
    handlers := eb.handlers[event.EventType()]
    eb.mu.RUnlock()
    
    for _, handler := range handlers {
        if err := handler.Handle(ctx, event); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Property-Based Testing

```go
package domain_test

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

// Generators for property testing
func genOrderID() gopter.Gen {
    return gen.UUIDv4().Map(func(id uuid.UUID) OrderID {
        return OrderID(id)
    })
}

func genMoney() gopter.Gen {
    return gen.Float64Range(0.01, 10000.00).Map(func(amount float64) Money {
        return Money{Amount: decimal.NewFromFloat(amount)}
    })
}

func genQuantity() gopter.Gen {
    return gen.UInt32Range(1, 100).Map(func(q uint32) Quantity {
        return Quantity(q)
    })
}

func genOrderItem() gopter.Gen {
    return gopter.CombineGens(
        genProductID(),
        genQuantity(),
        genMoney(),
    ).Map(func(values []interface{}) OrderItem {
        productID := values[0].(ProductID)
        quantity := values[1].(Quantity)
        price := values[2].(Money)
        
        subtotal := Money{
            Amount: price.Amount.Mul(decimal.NewFromInt32(int32(quantity))),
        }
        
        return OrderItem{
            ProductID: productID,
            Quantity:  quantity,
            Price:     price,
            Subtotal:  subtotal,
        }
    })
}

func genOrder() gopter.Gen {
    return gopter.CombineGens(
        genOrderID(),
        genCustomerID(),
        gen.SliceOfN(5, genOrderItem()),
    ).Map(func(values []interface{}) *Order {
        orderID := values[0].(OrderID)
        customerID := values[1].(CustomerID)
        items := values[2].([]OrderItem)
        
        total := decimal.Zero
        for _, item := range items {
            total = total.Add(item.Subtotal.Amount)
        }
        
        return &Order{
            ID:         orderID,
            CustomerID: customerID,
            Items:      items,
            Total:      Money{Amount: total},
            State:      PendingState{},
            CreatedAt:  time.Now(),
            UpdatedAt:  time.Now(),
        }
    })
}

// Property tests
func TestOrderProperties(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    // Total conservation property
    properties.Property("total equals sum of items", prop.ForAll(
        func(order *Order) bool {
            calculated := decimal.Zero
            for _, item := range order.Items {
                calculated = calculated.Add(item.Subtotal.Amount)
            }
            return order.Total.Amount.Equal(calculated)
        },
        genOrder(),
    ))
    
    // Positive quantities property
    properties.Property("all quantities are positive", prop.ForAll(
        func(order *Order) bool {
            for _, item := range order.Items {
                if item.Quantity == 0 {
                    return false
                }
            }
            return true
        },
        genOrder(),
    ))
    
    // Valid timestamps property
    properties.Property("created before updated", prop.ForAll(
        func(order *Order) bool {
            return !order.CreatedAt.After(order.UpdatedAt)
        },
        genOrder(),
    ))
    
    // Item subtotal correctness
    properties.Property("item subtotals are correct", prop.ForAll(
        func(item OrderItem) bool {
            expected := item.Price.Amount.Mul(decimal.NewFromInt32(int32(item.Quantity)))
            return item.Subtotal.Amount.Equal(expected)
        },
        genOrderItem(),
    ))
    
    properties.TestingRun(t)
}

// State machine property test
func TestOrderStateMachine(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("valid state transitions", prop.ForAll(
        func(commands []StateCommand) bool {
            order := &Order{State: DraftState{}}
            
            for _, cmd := range commands {
                newState, err := applyCommand(order.State, cmd)
                if err != nil {
                    // Invalid transition should be rejected
                    if !isValidTransition(order.State, cmd) {
                        continue
                    }
                    return false
                }
                order.State = newState
            }
            
            return true
        },
        gen.SliceOf(genStateCommand()),
    ))
    
    properties.TestingRun(t)
}
```

## API Contract with Echo Framework

```go
package api

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

// API handler with contract enforcement
type OrderHandler struct {
    orderService OrderService
    validator    *OrderValidator
    rateLimiter  *RateLimiter
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
    // Parse request
    var req CreateOrderRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
    }
    
    // Validate request
    if err := h.validator.ValidateRequest(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    
    // Check rate limit
    ip := c.RealIP()
    if err := h.rateLimiter.ValidateIP(net.ParseIP(ip)); err != nil {
        return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
    }
    
    // Get user from context
    user := c.Get("user").(*User)
    
    // Transform to domain model
    transform := &CartToOrderTransform{
        ProductService: h.orderService,
        PricingService: h.orderService,
    }
    
    order, err := transform.Transform(c.Request().Context(), TransformInput{
        Cart:       Cart{Items: req.Items},
        CustomerID: CustomerID(user.ID),
    })
    if err != nil {
        return echo.NewHTTPError(http.StatusConflict, err.Error())
    }
    
    // Save order
    savedOrder, err := h.orderService.CreateOrder(c.Request().Context(), order)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create order")
    }
    
    return c.JSON(http.StatusCreated, savedOrder)
}

// Route configuration
func SetupRoutes(e *echo.Echo, handler *OrderHandler) {
    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    
    // API group with authentication
    api := e.Group("/api")
    api.Use(middleware.JWTWithConfig(middleware.JWTConfig{
        SigningKey: []byte("secret"),
    }))
    
    // Order endpoints
    api.POST("/orders", handler.CreateOrder)
    api.GET("/orders/:id", handler.GetOrder)
    api.PUT("/orders/:id", handler.UpdateOrder)
    api.DELETE("/orders/:id", handler.CancelOrder)
    
    // Add OpenAPI documentation
    api.GET("/swagger/*", echoSwagger.WrapHandler)
}

// OpenAPI annotations for automatic documentation
// @Summary Create a new order
// @Description Create a new order from cart items
// @Tags orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "Order creation request"
// @Success 201 {object} Order
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /api/orders [post]
// @Security Bearer
func (h *OrderHandler) CreateOrderSwagger(c echo.Context) error {
    return h.CreateOrder(c)
}
```