# Generic Subscription Engine Specification

## Overview
The Generic Subscription Engine is a comprehensive, service-agnostic subscription management system that handles plan creation, subscription lifecycle management, billing automation, and usage tracking across all MyTorra platform services.

## Core Responsibilities

### Primary Functions
1. **Plan Management** - Create and manage subscription plans for all services
2. **Subscription Lifecycle** - Handle subscription creation, changes, and cancellation
3. **Billing Automation** - Automated recurring billing and invoice generation
4. **Usage Tracking** - Monitor and track feature usage and limits
5. **Proration Handling** - Calculate prorated charges for plan changes
6. **Grace Period Management** - Handle payment failures and account suspension

## Features & Functionality

### 1. Subscription Plan Management

#### Plan Types
- **Feature-Based Plans**
  - Boolean feature flags (enabled/disabled)
  - Feature limits (max products, max orders, etc.)
  - Service access permissions
  - API rate limits
  - Storage quotas

- **Usage-Based Plans**
  - Pay-per-use billing
  - Tiered usage pricing
  - Overage charges
  - Usage thresholds
  - Volume discounts

- **Hybrid Plans**
  - Base subscription + usage charges
  - Feature bundles + overages
  - Freemium with premium features
  - Credit-based systems

#### Plan Configuration
- **Billing Cycles**
  - Monthly subscriptions
  - Annual subscriptions
  - Weekly subscriptions
  - Custom billing cycles
  - One-time payments

- **Pricing Models**
  - Fixed monthly/annual pricing
  - Per-user pricing
  - Per-location pricing (restaurants, hotels)
  - Volume-based pricing
  - Custom pricing tiers

### 2. Service-Specific Plans

#### Restaurant Plans (MyTorra Food)
- **Starter Plan**
  - Up to 50 menu items
  - Basic analytics
  - Standard commission rate
  - Email support

- **Professional Plan**
  - Unlimited menu items
  - Advanced analytics
  - Reduced commission rate
  - Priority support
  - Marketing tools

- **Enterprise Plan**
  - All Professional features
  - Custom branding
  - API access
  - Dedicated account manager
  - White-label options

#### Hotel Plans (MyTorra Booking)
- **Basic Plan**
  - Up to 10 room types
  - Basic availability management
  - Standard commission rate
  - Email support

- **Premium Plan**
  - Unlimited room types
  - Dynamic pricing tools
  - Channel management
  - Reduced commission rate
  - Phone support

- **Enterprise Plan**
  - All Premium features
  - Revenue management tools
  - Custom integrations
  - Dedicated support
  - Analytics dashboard

#### Marketplace Seller Plans
- **Individual Plan**
  - Up to 20 listings
  - Basic seller tools
  - Standard transaction fees
  - Community support

- **Business Plan**
  - Unlimited listings
  - Advanced seller tools
  - Reduced transaction fees
  - Priority listing placement
  - Business analytics

- **Enterprise Plan**
  - All Business features
  - API access
  - Bulk listing tools
  - Custom branding
  - Dedicated support

### 3. Subscription Lifecycle Management

#### Subscription Creation
- **Plan Selection**
  - Available plans display
  - Feature comparison
  - Pricing calculation
  - Trial period handling
  - Upgrade/downgrade paths

- **Subscription Activation**
  - Payment method collection
  - Initial payment processing
  - Account provisioning
  - Feature enablement
  - Welcome notifications

#### Subscription Changes
- **Plan Upgrades**
  - Immediate feature access
  - Prorated billing calculation
  - Payment processing
  - Usage limit adjustments
  - Notification handling

- **Plan Downgrades**
  - End-of-period changes
  - Feature restriction scheduling
  - Refund calculations
  - Data retention policies
  - Customer notifications

#### Subscription Cancellation
- **Cancellation Types**
  - Immediate cancellation
  - End-of-period cancellation
  - Scheduled cancellation
  - Administrative cancellation
  - Automatic cancellation

- **Cancellation Process**
  - Reason collection
  - Retention offers
  - Data export options
  - Final billing
  - Account deactivation

### 4. Billing & Payment Management

#### Automated Billing
- **Billing Cycles**
  - Recurring charge scheduling
  - Billing date management
  - Timezone handling
  - Holiday scheduling
  - Grace period configuration

- **Invoice Generation**
  - Automated invoice creation
  - Tax calculation
  - Discount application
  - Proration handling
  - Multi-currency support

#### Payment Processing
- **Payment Collection**
  - Multiple payment methods
  - Automatic retry logic
  - Failed payment handling
  - Payment method updates
  - PCI compliance

- **Dunning Management**
  - Payment failure notifications
  - Retry schedules
  - Account suspension rules
  - Collection procedures
  - Recovery workflows

### 5. Usage Tracking & Limits

#### Feature Usage Monitoring
- **Real-time Tracking**
  - API call counting
  - Feature usage logging
  - Resource consumption
  - Bandwidth monitoring
  - Storage tracking

- **Limit Enforcement**
  - Soft limit warnings
  - Hard limit enforcement
  - Overage calculations
  - Temporary limit increases
  - Usage notifications

#### Usage Analytics
- **Usage Reports**
  - Daily/weekly/monthly usage
  - Feature adoption metrics
  - Peak usage patterns
  - Cost per feature analysis
  - Trend identification

- **Business Intelligence**
  - Revenue per user
  - Churn prediction
  - Upgrade opportunities
  - Usage optimization
  - Plan effectiveness

## Admin Dashboard Features

### 1. Plan Management
- **Plan Creation**
  - Visual plan builder
  - Feature configuration
  - Pricing setup
  - Billing cycle definition
  - A/B testing support

- **Plan Analytics**
  - Plan popularity metrics
  - Revenue per plan
  - Conversion rates
  - Churn rates
  - Upgrade/downgrade patterns

### 2. Subscription Management
- **Subscription Overview**
  - Active subscriptions list
  - Revenue dashboard
  - Subscription status monitoring
  - Billing schedule view
  - Usage analytics

- **Customer Management**
  - Customer subscription history
  - Payment history
  - Support ticket integration
  - Account status management
  - Custom plan assignments

### 3. Billing Administration
- **Payment Monitoring**
  - Failed payment tracking
  - Retry status monitoring
  - Refund processing
  - Chargeback handling
  - Revenue reconciliation

- **Financial Reporting**
  - Monthly recurring revenue (MRR)
  - Annual recurring revenue (ARR)
  - Customer lifetime value (CLV)
  - Churn rate analysis
  - Revenue forecasting

### 4. Usage Management
- **Usage Monitoring**
  - Real-time usage dashboard
  - Limit breach alerts
  - Usage trend analysis
  - Cost optimization
  - Capacity planning

## API Endpoints

### Plan Management
```
GET /subscription/plans
POST /subscription/plans
GET /subscription/plans/{id}
PUT /subscription/plans/{id}
DELETE /subscription/plans/{id}
GET /subscription/plans/service/{service}
```

### Subscription Operations
```
POST /subscription/subscribe
GET /subscription/current
PUT /subscription/change-plan
POST /subscription/cancel
POST /subscription/reactivate
GET /subscription/billing-history
```

### Usage Management
```
POST /subscription/usage/record
GET /subscription/usage/current
GET /subscription/usage/limits
GET /subscription/usage/history
POST /subscription/usage/reset
```

### Billing Operations
```
GET /subscription/billing/invoices
GET /subscription/billing/next-bill
POST /subscription/billing/charge
POST /subscription/billing/refund
PUT /subscription/billing/payment-method
```

### Admin Operations
```
GET /admin/subscriptions
GET /admin/subscriptions/{id}
PUT /admin/subscriptions/{id}/status
GET /admin/billing/failed-payments
POST /admin/billing/retry-payment
GET /admin/analytics/mrr
GET /admin/analytics/churn
```

## Database Schema

### Subscription Plans Table
```sql
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    service VARCHAR(50) NOT NULL,
    plan_type ENUM('feature_based', 'usage_based', 'hybrid') DEFAULT 'feature_based',
    billing_cycle ENUM('monthly', 'yearly', 'weekly', 'daily', 'one_time') DEFAULT 'monthly',
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    trial_days INTEGER DEFAULT 0,
    setup_fee DECIMAL(10,2) DEFAULT 0.00,
    features JSONB NOT NULL,
    limits JSONB,
    metadata JSONB,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_service (service),
    INDEX idx_active (active),
    INDEX idx_billing_cycle (billing_cycle),
    INDEX idx_price (price)
);
```

### User Subscriptions Table
```sql
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    plan_id UUID REFERENCES subscription_plans(id),
    status ENUM('trial', 'active', 'past_due', 'cancelled', 'expired', 'suspended') DEFAULT 'active',
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMP,
    ended_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_user_id (user_id),
    INDEX idx_plan_id (plan_id),
    INDEX idx_status (status),
    INDEX idx_current_period_end (current_period_end),
    UNIQUE KEY active_subscription (user_id, plan_id) -- Ensure one active subscription per plan per user
);
```

### Subscription Usage Table
```sql
CREATE TABLE subscription_usage (
    id UUID PRIMARY KEY,
    subscription_id UUID REFERENCES user_subscriptions(id),
    feature_key VARCHAR(100) NOT NULL,
    usage_count INTEGER DEFAULT 0,
    usage_limit INTEGER,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    reset_count INTEGER DEFAULT 0,
    overage_count INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_subscription_id (subscription_id),
    INDEX idx_feature_key (feature_key),
    INDEX idx_period_start (period_start),
    INDEX idx_period_end (period_end),
    UNIQUE KEY unique_subscription_feature_period (subscription_id, feature_key, period_start)
);
```

### Billing Invoices Table
```sql
CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY,
    subscription_id UUID REFERENCES user_subscriptions(id),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    status ENUM('draft', 'open', 'paid', 'void', 'uncollectible') DEFAULT 'open',
    subtotal DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0.00,
    discount_amount DECIMAL(10,2) DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_period_start TIMESTAMP NOT NULL,
    billing_period_end TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    payment_intent_id VARCHAR(100),
    line_items JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_subscription_id (subscription_id),
    INDEX idx_invoice_number (invoice_number),
    INDEX idx_status (status),
    INDEX idx_due_date (due_date),
    INDEX idx_created_at (created_at)
);
```

### Payment Attempts Table
```sql
CREATE TABLE payment_attempts (
    id UUID PRIMARY KEY,
    invoice_id UUID REFERENCES billing_invoices(id),
    payment_method_id VARCHAR(100),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('pending', 'succeeded', 'failed', 'cancelled') DEFAULT 'pending',
    failure_reason TEXT,
    payment_intent_id VARCHAR(100),
    charge_id VARCHAR(100),
    attempted_at TIMESTAMP DEFAULT NOW(),
    succeeded_at TIMESTAMP,
    failed_at TIMESTAMP,

    INDEX idx_invoice_id (invoice_id),
    INDEX idx_status (status),
    INDEX idx_attempted_at (attempted_at)
);
```

### Subscription Events Table
```sql
CREATE TABLE subscription_events (
    id UUID PRIMARY KEY,
    subscription_id UUID REFERENCES user_subscriptions(id),
    event_type ENUM('created', 'updated', 'cancelled', 'reactivated', 'trial_started', 'trial_ended', 'payment_succeeded', 'payment_failed') NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_subscription_id (subscription_id),
    INDEX idx_event_type (event_type),
    INDEX idx_created_at (created_at)
);
```

## Integration Patterns

### With Core Service
- User authentication and authorization
- Payment processing coordination
- User profile management
- Notification handling

### With Service-Specific Features
```go
// Restaurant service integration
func ValidateRestaurantSubscription(userID string, feature string) bool {
    subscription := subscriptionEngine.GetActiveSubscription(userID, "restaurant")
    return subscription.HasFeature(feature) && !subscription.IsLimitExceeded(feature)
}

// Marketplace service integration
func TrackProductListing(userID string) error {
    return subscriptionEngine.RecordUsage(userID, "marketplace", "product_listings", 1)
}
```

### Event Publishing
- `SubscriptionCreated`: New subscription created
- `SubscriptionUpdated`: Subscription plan changed
- `SubscriptionCancelled`: Subscription cancelled
- `PaymentSucceeded`: Subscription payment successful
- `PaymentFailed`: Subscription payment failed
- `UsageLimitReached`: Usage limit exceeded
- `TrialExpired`: Trial period ended

### Event Consumption
- `UserRegistered`: Create trial subscription
- `PaymentMethodUpdated`: Update billing information
- `ServiceUsageRecorded`: Update usage metrics

## Business Rules & Logic

### Plan Change Logic
```go
// Plan upgrade - immediate access, prorated billing
func UpgradePlan(subscriptionID string, newPlanID string) error {
    // 1. Calculate prorated amount
    // 2. Process immediate payment
    // 3. Update subscription
    // 4. Enable new features immediately
    // 5. Send confirmation notification
}

// Plan downgrade - end of period change
func DowngradePlan(subscriptionID string, newPlanID string) error {
    // 1. Schedule change for period end
    // 2. Disable features at period end
    // 3. Send confirmation notification
    // 4. Calculate refund if applicable
}
```

### Usage Limit Enforcement
```go
func EnforceUsageLimit(subscriptionID string, feature string) error {
    usage := GetCurrentUsage(subscriptionID, feature)
    limit := GetFeatureLimit(subscriptionID, feature)

    if usage >= limit {
        return errors.New("usage limit exceeded")
    }

    if usage >= (limit * 0.8) {
        SendUsageWarning(subscriptionID, feature, usage, limit)
    }

    return nil
}
```

## Pricing Calculations

### Proration Calculation
```go
func CalculateProration(oldPlan Plan, newPlan Plan, daysRemaining int, totalDays int) decimal.Decimal {
    unusedAmount := oldPlan.Price.Mul(decimal.NewFromInt(int64(daysRemaining))).Div(decimal.NewFromInt(int64(totalDays)))
    newAmount := newPlan.Price.Mul(decimal.NewFromInt(int64(daysRemaining))).Div(decimal.NewFromInt(int64(totalDays)))
    return newAmount.Sub(unusedAmount)
}
```

### Usage-Based Billing
```go
func CalculateUsageBilling(subscription Subscription, usageRecords []UsageRecord) decimal.Decimal {
    var total decimal.Decimal

    for _, record := range usageRecords {
        if record.Count > record.IncludedAmount {
            overage := record.Count - record.IncludedAmount
            total = total.Add(decimal.NewFromInt(int64(overage)).Mul(record.OverageRate))
        }
    }

    return total
}
```

## Performance Requirements

### Response Times
- **Subscription Validation**: <50ms
- **Usage Recording**: <100ms
- **Plan Changes**: <2 seconds
- **Billing Processing**: <5 seconds

### Scalability
- **Concurrent Users**: Support 100,000+ active subscriptions
- **Usage Events**: Process 10,000+ events per second
- **Billing Operations**: Handle 50,000+ invoices per month
- **API Throughput**: 1,000+ requests per second
