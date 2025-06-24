# SDL vs Traditional Programming Languages

## Overview

This document explains the fundamental differences between SDL (System Design Language) and traditional programming languages. Understanding these differences is crucial for effectively using SDL to model distributed systems.

## Key Philosophical Differences

### SDL is Descriptive, Not Prescriptive

**Traditional Programming:**
```go
func getUser(userId string) (*User, error) {
    // Connect to database
    conn, err := db.Connect()
    if err != nil {
        return nil, err
    }
    
    // Execute query
    result, err := conn.Query("SELECT * FROM users WHERE id = ?", userId)
    if err != nil {
        return nil, err
    }
    
    // Parse result
    return parseUser(result)
}
```

**SDL Equivalent:**
```sdl
method GetUser() Bool {
    // Model the behavior, not the implementation
    delay(5ms)  // Average query time
    return sample dist {
        99 => true,   // Success rate
        1 => false    // Failure rate
    }
}
```

SDL doesn't care **how** you get the user. It models **what happens** when you try: how long it takes and whether it succeeds.

### No Runtime Parameters

**Traditional Programming:**
```python
def calculate_price(quantity, unit_price, discount):
    return quantity * unit_price * (1 - discount)
```

**SDL Approach:**
```sdl
component PricingService {
    // Configuration is through parameters, not method arguments
    param AverageOrderSize Int = 10
    param TypicalDiscount Float = 0.1
    
    method CalculatePrice() Float {
        // Models the behavior of price calculation
        delay(1ms)  // Computation time
        return sample dist {
            80 => 50.0,   // Typical order value
            15 => 150.0,  // Large order
            5 => 10.0     // Small order
        }
    }
}
```

### Probabilistic vs Deterministic

**Traditional Programming:**
```java
public boolean processPayment(Payment payment) {
    // Deterministic logic
    if (payment.getAmount() > accountBalance) {
        return false;
    }
    
    accountBalance -= payment.getAmount();
    return true;
}
```

**SDL Modeling:**
```sdl
method ProcessPayment() Bool {
    // Model real-world uncertainty
    return sample dist {
        98 => true,    // Successful payment
        1 => false,    // Insufficient funds
        0.5 => false,  // Bank rejection
        0.5 => false   // Network error
    }
}
```

## No Arithmetic Operators

### The Problem with Math in System Modeling

Traditional programming languages need arithmetic for calculations. SDL intentionally omits these because:

1. **System behavior isn't math**: Response times, failure rates, and capacity limits are distributions, not calculations
2. **Abstraction over precision**: SDL models approximate behavior, not exact computation
3. **Focus on interactions**: What matters is how components interact, not internal calculations

**What NOT to do in SDL:**
```sdl
// WRONG - SDL doesn't support arithmetic operators
method CalculateTotal() Int {
    let subtotal = 100
    let tax = subtotal * 0.08  // ERROR: No * operator
    return subtotal + tax       // ERROR: No + operator
}
```

**Correct SDL approach:**
```sdl
// Model the behavior, not the calculation
method CalculateTotal() Bool {
    delay(1ms)  // Processing time
    return true  // Calculation succeeds
}

// Or if you need different outcomes
method ProcessOrder() Outcomes[String] {
    return dist {
        90 => "approved",
        5 => "insufficient_funds",
        3 => "invalid_card",
        2 => "system_error"
    }
}
```

## Virtual Time vs Real Time

### Traditional Programming: Real Time
```javascript
async function pollService() {
    while (true) {
        const result = await checkService();
        await sleep(1000);  // Real 1 second delay
    }
}
```

### SDL: Virtual Time
```sdl
method PollService() Bool {
    for true {
        self.checkService()
        delay(1s)  // Virtual 1 second - happens instantly in simulation
    }
}
```

In SDL:
- `delay(1hour)` completes instantly
- Simulations are deterministic and reproducible
- You can simulate days of system behavior in milliseconds

## State Management

### Traditional: Mutable State
```csharp
public class ShoppingCart {
    private List<Item> items = new List<Item>();
    private decimal total = 0;
    
    public void AddItem(Item item) {
        items.Add(item);
        total += item.Price;
    }
}
```

### SDL: Behavioral Modeling
```sdl
component ShoppingCart {
    // Model behavior, not state
    param AverageCartSize Int = 5
    param AbandonmentRate Float = 0.3
    
    method AddItem() Bool {
        // Model whether add succeeds
        return sample dist {
            95 => true,   // Successfully added
            5 => false    // Out of stock
        }
    }
    
    method Checkout() Bool {
        // Model checkout behavior
        return sample dist {
            70 => true,   // Successful checkout
            30 => false   // Cart abandoned
        }
    }
}
```

## Error Handling

### Traditional: Explicit Error Handling
```rust
fn read_file(path: &str) -> Result<String, io::Error> {
    match fs::read_to_string(path) {
        Ok(contents) => Ok(contents),
        Err(e) => Err(e),
    }
}
```

### SDL: Probabilistic Outcomes
```sdl
method ReadFile() Bool {
    // Model file system behavior
    delay(sample dist {
        80 => 1ms,    // Cache hit
        15 => 10ms,   // Disk read
        5 => 100ms    // Slow disk
    })
    
    return sample dist {
        99.9 => true,  // Success
        0.1 => false   // File system error
    }
}
```

## Concurrency

### Traditional: Complex Thread Management
```go
func processItems(items []Item) {
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10)  // Limit concurrency
    
    for _, item := range items {
        wg.Add(1)
        semaphore <- struct{}{}
        
        go func(it Item) {
            defer wg.Done()
            defer func() { <-semaphore }()
            
            process(it)
        }(item)
    }
    
    wg.Wait()
}
```

### SDL: Simple Concurrency Modeling
```sdl
component ItemProcessor {
    uses pool ResourcePool(Size = 10)  // Models concurrency limit
    
    method ProcessItems() Bool {
        // SDL handles the complexity
        gobatch 10 {
            self.processItem()
        }
        return true
    }
}
```

## When to Use Each

### Use Traditional Programming For:
- Building actual systems
- Implementing business logic
- Data transformations
- User interfaces
- APIs and services

### Use SDL For:
- Capacity planning
- Performance analysis
- Architecture validation
- Bottleneck identification
- "What if" scenarios

## Common Misconceptions

### Misconception 1: "SDL is a simpler programming language"
**Reality**: SDL is not a programming language at all. It's a modeling language with fundamentally different goals.

### Misconception 2: "I can port my code to SDL"
**Reality**: You model the behavior of your code, not the code itself.

### Misconception 3: "SDL limitations are weaknesses"
**Reality**: The constraints are intentional and force focus on system behavior over implementation details.

## Example: Modeling vs Programming

### Scenario: E-commerce Checkout

**Traditional Programming Approach:**
```python
class CheckoutService:
    def process_checkout(self, cart, payment_info, shipping_info):
        # Validate cart
        if not self.validate_cart(cart):
            raise InvalidCartError()
        
        # Calculate total
        subtotal = sum(item.price * item.quantity for item in cart.items)
        tax = subtotal * self.tax_rate
        shipping = self.calculate_shipping(shipping_info)
        total = subtotal + tax + shipping
        
        # Process payment
        payment_result = self.payment_gateway.charge(payment_info, total)
        if not payment_result.success:
            raise PaymentFailedError()
        
        # Create order
        order = self.create_order(cart, payment_result, shipping_info)
        
        # Send confirmation
        self.email_service.send_confirmation(order)
        
        return order
```

**SDL Modeling Approach:**
```sdl
component CheckoutService {
    uses inventory InventoryService
    uses payment PaymentGateway
    uses email EmailService
    uses shipping ShippingService
    
    method ProcessCheckout() Bool {
        // Validate inventory
        let available = self.inventory.CheckAvailability()
        if !available {
            return false
        }
        
        // Process payment (models payment gateway behavior)
        let paymentResult = self.payment.Process()
        if !paymentResult {
            return false
        }
        
        // Calculate shipping (models API call)
        self.shipping.Calculate()
        
        // Send confirmation (fire and forget)
        go self.email.SendConfirmation()
        
        // Model overall success rate
        return sample dist {
            97 => true,   // Successful checkout
            2 => false,   // Inventory became unavailable
            1 => false    // System error
        }
    }
}

component PaymentGateway {
    param SuccessRate Float = 0.98
    param Latency = dist {
        70 => 200ms,   // Fast
        25 => 500ms,   // Normal
        5 => 2s        // Slow
    }
    
    method Process() Bool {
        delay(sample self.Latency)
        
        return sample dist {
            98 => true,   // Payment approved
            1 => false,   // Card declined
            0.5 => false, // Gateway error
            0.5 => false  // Network timeout
        }
    }
}
```

## Key Takeaways

1. **SDL models behavior, not implementation**
2. **Constraints are features, not limitations**
3. **Probabilistic modeling captures real-world uncertainty**
4. **Virtual time enables fast, deterministic simulations**
5. **Focus on interactions between components, not internal logic**

Remember: You're not building a system with SDL; you're modeling how a system behaves under various conditions. This fundamental difference drives every design decision in the language.