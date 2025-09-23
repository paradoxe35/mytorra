# Monetization Engine Specification (Subscription & Transaction Fees)

## Overview
The Monetization Engine is a comprehensive, centralized revenue management system that handles both subscription plans and transaction-based fees across all MyTorra platform services. This hybrid system serves as the single source of truth for all revenue streams, providing dynamic plan creation, commission management, subscription lifecycle handling, billing automation, and usage tracking. The engine supports multi-tenant architecture where each business entity (restaurant, boutique, hotel, driver) is treated as a tenant with their own financial management capabilities.

## Core Responsibilities

### Primary Functions
1. **Dynamic Plan Management** - Admin-created and managed subscription plans without code changes
2. **Commission & Fee Management** - Configure transaction fees, commissions, and listing fees per service
3. **Multi-Tenant Revenue Management** - Isolated financial management for each business entity
4. **Subscription Lifecycle** - Handle subscription creation, changes, and cancellation
5. **Split Payment Processing** - Automatic distribution of payments between sellers and platform
6. **Billing Automation** - Automated recurring billing and invoice generation
7. **Usage Tracking** - Monitor and track feature usage and limits
8. **Proration Handling** - Calculate prorated charges for plan changes
9. **Grace Period Management** - Handle payment failures and account suspension

## Features & Functionality

### 1. Dynamic Monetization Management

#### Revenue Models
- **Subscription-Based Revenue**
  - Admin-defined recurring payment plans
  - Customizable billing cycles (daily, weekly, monthly, yearly)
  - Feature-based access control
  - Service-specific plan templates
  - Dynamic pricing adjustments

- **Transaction-Based Revenue**
  - Commission rates per service/category
  - Fixed transaction fees
  - Listing fees for marketplaces
  - Volume-based commission tiers
  - Promotional period adjustments

- **Hybrid Revenue Models**
  - Subscription + reduced commission rates
  - Base plans with transaction overages
  - Freemium with transaction fees
  - Credit-based systems with top-ups

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

### 2. Admin-Managed Plan Configuration

#### Dynamic Plan Builder
- **Plan Creation Interface**
  - Visual plan designer with drag-drop features
  - Feature catalog with searchable components
  - Pricing calculator with revenue projections
  - A/B testing configuration for plans
  - Plan template library for quick setup

- **Feature Registry System**
  - Centralized catalog of all available features
  - Feature dependencies and relationships
  - Feature categories (Core, Premium, Add-on)
  - Custom feature creation for specific services
  - Feature usage metrics and analytics

- **Commission Configuration**
  - Service-specific commission rates
  - Category-based fee structures
  - Volume discount tiers
  - Promotional rate scheduling
  - Geographic region-based rates

#### Service Plan Templates (Admin Customizable)

**Restaurant Service (MyTorra Food)**
- Admin can create unlimited plan variations
- Mix of subscription fees and commission rates
- Example configurations:
  - Free listing with 15% commission
  - Monthly fee with 8% commission
  - Premium tier with 5% commission + features

**Hotel Service (MyTorra Booking)**
- Flexible room listing limits
- Dynamic commission structures
- Seasonal rate adjustments
- Channel management features

**Marketplace Service**
- Listing fee options
- Transaction percentage configuration
- Category-specific rates
- Seller tier systems

**Delivery & Cab Services**
- Pure commission models
- Distance-based fee calculations
- Peak hour rate adjustments
- Driver incentive programs

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

### 4. Dual-Mode Payment Processing

#### Collection Phase (Both Modes)
- **Payment Collection Methods**
  - Mobile money (M-Pesa, MTN, Airtel, Orange)
  - Credit/debit cards (Visa, Mastercard)
  - Bank transfers
  - Digital wallets
  - USSD payments
  - QR code payments

- **Transaction Recording**
  - Full amount collection to platform
  - Commission calculation
  - Fee breakdown storage
  - Transaction reference tracking
  - Multi-currency handling

#### Disbursement Phase

##### Mode 1: Instant Split (Card Markets)
- **Real-time Processing**
  - Immediate vendor credit
  - Instant commission deduction
  - Same-day settlement
  - Automated reconciliation
  - Direct account deposits

##### Mode 2: Collect-and-Disburse (Mobile Money Markets)
- **Scheduled Disbursements**
  - Daily/weekly batch processing
  - Minimum threshold enforcement (e.g., $10)
  - Bulk transfer optimization
  - Provider-specific routing
  - Failed disbursement retry

- **Settlement Management**
  - Holding account reconciliation
  - Float balance management
  - Reserve requirements (10-20%)
  - Regulatory compliance
  - Audit trail maintenance

#### Reconciliation & Reporting
- **Daily Reconciliation**
  - Collection vs disbursement matching
  - Fee verification
  - Discrepancy detection
  - Automatic adjustment entries
  - Exception handling

- **Financial Reporting**
  - Platform float tracking
  - Vendor balance sheets
  - Commission reports
  - Settlement status
  - Regulatory compliance reports

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

### 1. Dynamic Plan Management
- **Plan Builder Interface**
  - Drag-and-drop feature composer
  - Real-time pricing preview
  - Plan comparison matrix generator
  - Version control for plan changes
  - Sandbox testing environment
  - Bulk plan operations

- **Feature Management**
  - Feature catalog administration
  - Feature dependency mapping
  - Usage limit configuration
  - Feature flag management
  - Cross-service feature bundles

- **Commission & Fee Management**
  - Commission rate configurator
  - Transaction fee calculator
  - Listing fee management
  - Promotional rate scheduler
  - Revenue split configuration
  - Payment gateway fee settings

- **Plan Analytics**
  - Real-time revenue dashboards
  - Plan performance metrics
  - Conversion funnel analysis
  - Churn prediction models
  - Revenue optimization suggestions
  - A/B test results

### 2. Tenant & Subscription Management
- **Multi-Tenant Dashboard**
  - Tenant onboarding workflow
  - Business verification status
  - Tenant categorization (restaurant, boutique, driver)
  - Revenue per tenant analytics
  - Tenant health scoring

- **Subscription Management**
  - Real-time subscription monitoring
  - Bulk subscription operations
  - Migration tools for plan changes
  - Grandfathering rules configuration
  - Trial period management
  - Custom deal creation

- **Tenant Self-Service Portal**
  - Subscription upgrade/downgrade
  - Payment method management
  - Earnings and commission tracking
  - Transaction history
  - Payout schedule viewing
  - Financial reports download

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

### Admin Plan Management (Dynamic Creation/Modification)
```
# Plan Builder
POST /admin/monetization/plans/create        # Create new plan from scratch
POST /admin/monetization/plans/from-template # Create from template
GET /admin/monetization/plans/templates      # List available templates
PUT /admin/monetization/plans/{id}          # Modify existing plan
POST /admin/monetization/plans/{id}/version # Create new version of plan
POST /admin/monetization/plans/{id}/publish # Publish draft plan
POST /admin/monetization/plans/{id}/archive # Archive active plan
DELETE /admin/monetization/plans/{id}       # Delete draft plan only

# Feature Registry Management
GET /admin/monetization/features            # List all available features
POST /admin/monetization/features           # Add new feature to registry
PUT /admin/monetization/features/{id}       # Update feature definition
GET /admin/monetization/features/categories # Get feature categories

# Commission Rules Management
GET /admin/monetization/commission-rules    # List all commission rules
POST /admin/monetization/commission-rules   # Create new rule
PUT /admin/monetization/commission-rules/{id} # Update rule
DELETE /admin/monetization/commission-rules/{id} # Delete rule
POST /admin/monetization/commission-rules/simulate # Simulate rule impact

# Plan Migration Management
GET /admin/monetization/migrations/pending  # List pending migrations
POST /admin/monetization/migrations/schedule # Schedule bulk migration
POST /admin/monetization/migrations/{id}/execute # Execute migration
GET /admin/monetization/migrations/history  # Migration history
```

### Tenant-Facing Plan Operations
```
# Browse and Subscribe
GET /monetization/plans/available           # List plans for tenant type
GET /monetization/plans/{id}/details       # Get plan details with features
POST /monetization/plans/{id}/preview      # Preview plan change (pricing/features)
POST /monetization/subscribe               # Subscribe to a plan
POST /monetization/trial/start            # Start trial period

# Manage Subscription
GET /monetization/subscription/current     # Get current subscription
PUT /monetization/subscription/upgrade     # Upgrade to higher plan
PUT /monetization/subscription/downgrade   # Downgrade to lower plan
POST /monetization/subscription/cancel     # Cancel subscription
POST /monetization/subscription/reactivate # Reactivate cancelled subscription
```

### Transaction & Commission Management
```
# Transaction Fee Recording
POST /monetization/transactions/record     # Record transaction and calculate fees
GET /monetization/transactions/{id}        # Get transaction fee details
GET /monetization/transactions/history     # Transaction history for tenant

# Earnings & Payouts
GET /monetization/earnings/current-period  # Current period earnings
GET /monetization/earnings/history         # Historical earnings
GET /monetization/payouts/schedule         # Upcoming payouts
GET /monetization/payouts/history          # Payout history
PUT /monetization/payouts/account          # Update payout account details
```

### Billing & Payment Operations
```
# Invoices
GET /monetization/billing/invoices         # List all invoices
GET /monetization/billing/invoices/{id}    # Get specific invoice
GET /monetization/billing/next-charge      # Next charge details

# Payment Methods
GET /monetization/payment-methods          # List payment methods
POST /monetization/payment-methods         # Add payment method
PUT /monetization/payment-methods/{id}/default # Set default
DELETE /monetization/payment-methods/{id}  # Remove payment method

# Payment Processing
POST /monetization/payments/process        # Process payment
POST /monetization/payments/{id}/refund    # Process refund
POST /monetization/payments/{id}/dispute   # Handle dispute
```

### Analytics & Reporting
```
# Admin Analytics
GET /admin/monetization/analytics/revenue  # Platform revenue metrics
GET /admin/monetization/analytics/plans    # Plan performance metrics
GET /admin/monetization/analytics/tenants  # Tenant revenue analysis
GET /admin/monetization/analytics/forecasting # Revenue forecasting

# Tenant Analytics
GET /monetization/analytics/revenue        # Tenant's revenue metrics
GET /monetization/analytics/transactions   # Transaction analytics
GET /monetization/analytics/performance    # Performance metrics
```

## Database Schema

### Dynamic Plans Table
```sql
CREATE TABLE monetization_plans (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    service VARCHAR(50) NOT NULL,
    tenant_type ENUM('restaurant', 'hotel', 'marketplace_seller', 'driver', 'delivery') NOT NULL,
    plan_type ENUM('subscription', 'commission', 'hybrid') DEFAULT 'hybrid',
    billing_cycle ENUM('monthly', 'yearly', 'weekly', 'daily', 'one_time', 'transaction') DEFAULT 'monthly',

    -- Subscription pricing
    subscription_price DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    trial_days INTEGER DEFAULT 0,
    setup_fee DECIMAL(10,2) DEFAULT 0.00,

    -- Commission/transaction fees
    commission_rate DECIMAL(5,2) DEFAULT 0.00, -- Percentage
    fixed_transaction_fee DECIMAL(10,2) DEFAULT 0.00,
    listing_fee DECIMAL(10,2) DEFAULT 0.00,

    -- Dynamic configuration
    features JSONB NOT NULL, -- Array of feature IDs from feature_registry
    limits JSONB,
    commission_rules JSONB, -- Complex commission configurations

    -- Versioning
    version INTEGER DEFAULT 1,
    parent_plan_id UUID REFERENCES monetization_plans(id),

    -- Status
    status ENUM('draft', 'active', 'archived', 'testing') DEFAULT 'draft',
    published_at TIMESTAMP,
    archived_at TIMESTAMP,

    -- Admin tracking
    created_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),

    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_service (service),
    INDEX idx_tenant_type (tenant_type),
    INDEX idx_status (status),
    INDEX idx_plan_type (plan_type),
    INDEX idx_version (version)
);
```

### Feature Registry Table
```sql
CREATE TABLE feature_registry (
    id UUID PRIMARY KEY,
    feature_key VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    category ENUM('core', 'premium', 'addon', 'limit', 'permission') NOT NULL,
    service VARCHAR(50),
    data_type ENUM('boolean', 'integer', 'string', 'json') DEFAULT 'boolean',
    default_value TEXT,
    validation_rules JSONB,
    dependencies JSONB, -- Array of feature IDs this depends on
    metadata JSONB,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_service (service),
    INDEX idx_category (category),
    INDEX idx_active (active)
);
```

### Commission Rules Table
```sql
CREATE TABLE commission_rules (
    id UUID PRIMARY KEY,
    plan_id UUID REFERENCES monetization_plans(id),
    rule_type ENUM('category', 'volume', 'geographic', 'promotional', 'time_based') NOT NULL,
    priority INTEGER DEFAULT 0, -- Higher priority rules apply first

    -- Rule conditions
    conditions JSONB NOT NULL, -- Complex condition definitions

    -- Rule actions
    commission_rate DECIMAL(5,2),
    fixed_fee DECIMAL(10,2),

    -- Validity
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,

    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_plan_id (plan_id),
    INDEX idx_rule_type (rule_type),
    INDEX idx_priority (priority),
    INDEX idx_active (active)
);
```

### Tenant Subscriptions Table
```sql
CREATE TABLE tenant_subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL, -- Can be user_id or organization_id
    tenant_type ENUM('user', 'organization') NOT NULL,
    plan_id UUID REFERENCES monetization_plans(id),
    plan_version INTEGER NOT NULL, -- Track which version of plan they're on

    -- Subscription status
    status ENUM('trial', 'active', 'past_due', 'cancelled', 'expired', 'suspended', 'grandfathered') DEFAULT 'active',

    -- Billing periods
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,

    -- Customizations
    custom_commission_rate DECIMAL(5,2), -- Override plan's commission
    custom_features JSONB, -- Additional features granted
    negotiated_price DECIMAL(10,2), -- Custom pricing for this tenant

    -- Cancellation
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    ended_at TIMESTAMP,

    -- Grandfathering
    is_grandfathered BOOLEAN DEFAULT FALSE,
    grandfathered_from_plan_id UUID,

    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_tenant_id (tenant_id),
    INDEX idx_tenant_type (tenant_type),
    INDEX idx_plan_id (plan_id),
    INDEX idx_status (status),
    INDEX idx_current_period_end (current_period_end),
    INDEX idx_is_grandfathered (is_grandfathered),
    UNIQUE KEY active_subscription (tenant_id, plan_id)
);
```

### Transaction Fees Table
```sql
CREATE TABLE transaction_fees (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL, -- Reference to the actual transaction in service
    tenant_id UUID NOT NULL,
    service VARCHAR(50) NOT NULL,
    transaction_type VARCHAR(50), -- order, booking, ride, delivery

    -- Transaction details
    transaction_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',

    -- Fee calculation
    applied_plan_id UUID REFERENCES monetization_plans(id),
    commission_rate DECIMAL(5,2),
    commission_amount DECIMAL(10,2),
    fixed_fee_amount DECIMAL(10,2),
    total_fee_amount DECIMAL(10,2) NOT NULL,

    -- Split payment details
    tenant_payout DECIMAL(10,2) NOT NULL,
    platform_revenue DECIMAL(10,2) NOT NULL,

    -- Payment gateway fees (if applicable)
    gateway_fee DECIMAL(10,2) DEFAULT 0.00,
    net_platform_revenue DECIMAL(10,2),

    -- Status
    status ENUM('pending', 'processed', 'settled', 'refunded', 'disputed') DEFAULT 'pending',
    processed_at TIMESTAMP,
    settled_at TIMESTAMP,

    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_tenant_id (tenant_id),
    INDEX idx_service (service),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Tenant Payouts Table
```sql
CREATE TABLE tenant_payouts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    payout_batch_id VARCHAR(100),

    -- Payout amount
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',

    -- Period covered
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,

    -- Transaction summary
    transaction_count INTEGER NOT NULL,
    total_sales DECIMAL(10,2) NOT NULL,
    total_fees DECIMAL(10,2) NOT NULL,

    -- Payout details
    payout_method ENUM('bank_transfer', 'stripe', 'paypal', 'mobile_money') NOT NULL,
    payout_account_id VARCHAR(200),

    -- Status
    status ENUM('pending', 'processing', 'completed', 'failed', 'cancelled') DEFAULT 'pending',
    processed_at TIMESTAMP,
    completed_at TIMESTAMP,
    failure_reason TEXT,

    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_tenant_id (tenant_id),
    INDEX idx_status (status),
    INDEX idx_period_end (period_end),
    INDEX idx_payout_batch_id (payout_batch_id)
);
```

### Plan Migration History Table
```sql
CREATE TABLE plan_migrations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    from_plan_id UUID REFERENCES monetization_plans(id),
    to_plan_id UUID REFERENCES monetization_plans(id),
    migration_type ENUM('upgrade', 'downgrade', 'lateral', 'forced') NOT NULL,

    -- Migration details
    reason TEXT,
    initiated_by ENUM('tenant', 'admin', 'system') NOT NULL,
    admin_user_id UUID REFERENCES users(id),

    -- Grandfathering
    grandfathered BOOLEAN DEFAULT FALSE,
    retained_features JSONB,
    retained_pricing DECIMAL(10,2),

    -- Execution
    scheduled_for TIMESTAMP,
    executed_at TIMESTAMP,

    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_tenant_id (tenant_id),
    INDEX idx_from_plan_id (from_plan_id),
    INDEX idx_to_plan_id (to_plan_id),
    INDEX idx_executed_at (executed_at)
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

### With Payment Gateways (Marketplace-Specific)
- **Stripe Connect Integration**
  - Automated seller onboarding
  - Split payment processing
  - Managed account creation
  - Compliance handling
  - Instant payouts

- **PayStack/Flutterwave Integration**
  - Sub-account management
  - Commission splitting
  - Multi-currency support
  - Mobile money integration

### With Core Service
- User authentication and authorization
- Payment gateway abstraction
- Tenant profile management
- Notification handling
- Permission synchronization

### With Service-Specific Features
```go
// Restaurant service integration - Hybrid model
func ProcessRestaurantOrder(tenantID string, orderAmount decimal.Decimal) (TenantPayout, PlatformRevenue decimal.Decimal) {
    subscription := monetizationEngine.GetTenantSubscription(tenantID, "restaurant")
    commissionRate := subscription.GetEffectiveCommissionRate()

    // Calculate split
    platformRevenue := orderAmount.Mul(commissionRate.Div(decimal.NewFromInt(100)))
    tenantPayout := orderAmount.Sub(platformRevenue)

    // Record transaction fee
    monetizationEngine.RecordTransactionFee(tenantID, orderAmount, platformRevenue)

    return tenantPayout, platformRevenue
}

// Marketplace service - Pure commission model
func ProcessMarketplaceSale(sellerID string, saleAmount decimal.Decimal) error {
    // Check if seller has subscription for reduced fees
    subscription := monetizationEngine.GetTenantSubscription(sellerID, "marketplace")

    if subscription == nil {
        // Apply default commission for non-subscribers
        return monetizationEngine.ApplyDefaultCommission(sellerID, saleAmount, "marketplace")
    }

    // Apply subscription-based commission
    return monetizationEngine.ApplySubscriptionCommission(sellerID, saleAmount)
}

// Cab service - Pure transaction fee model
func ProcessCabRide(driverID string, fareAmount decimal.Decimal) error {
    // No subscription model - pure commission
    return monetizationEngine.ApplyServiceCommission(driverID, fareAmount, "cab", 0.15) // 15% commission
}
```

### Event Publishing
- `PlanCreated`: Admin creates new monetization plan
- `PlanModified`: Admin modifies existing plan
- `TenantSubscribed`: Tenant subscribes to a plan
- `SubscriptionUpdated`: Tenant changes subscription
- `SubscriptionCancelled`: Tenant cancels subscription
- `TransactionProcessed`: Transaction fee applied
- `PayoutScheduled`: Tenant payout scheduled
- `PayoutCompleted`: Tenant payout completed
- `CommissionRuleApplied`: Special commission rule triggered
- `PaymentSucceeded`: Payment processed successfully
- `PaymentFailed`: Payment processing failed
- `UsageLimitReached`: Feature usage limit exceeded
- `TrialExpired`: Trial period ended

### Event Consumption
- `TenantOnboarded`: Create default monetization profile
- `TransactionCompleted`: Calculate and apply fees
- `PaymentMethodUpdated`: Update billing information
- `ServiceUsageRecorded`: Update usage metrics
- `TenantVerified`: Enable full monetization features

## Business Rules & Logic

### Dynamic Plan Management
```go
// Admin creates new plan without code changes
func CreateMonetizationPlan(adminID string, planConfig PlanConfig) (*MonetizationPlan, error) {
    // 1. Validate admin permissions
    // 2. Validate feature combinations from registry
    // 3. Calculate revenue projections
    // 4. Create plan in draft status
    // 5. Submit for approval workflow
    // 6. Publish plan when approved
}

// Handle plan modifications for existing subscribers
func ModifyPlan(planID string, changes PlanChanges) error {
    // 1. Create new plan version
    // 2. Identify affected tenants
    // 3. Apply grandfathering rules
    // 4. Schedule migrations
    // 5. Send notifications
}
```

### Tenant Plan Changes
```go
// Tenant upgrades plan - immediate with proration
func UpgradeTenantPlan(tenantID string, newPlanID string) error {
    currentSub := GetTenantSubscription(tenantID)
    newPlan := GetPlan(newPlanID)

    // Check if downgrading commission (better deal for tenant)
    if newPlan.CommissionRate < currentSub.Plan.CommissionRate {
        // Apply immediately - tenant benefits
        ApplyPlanImmediately(tenantID, newPlanID)
    }

    // Calculate prorated subscription fee
    proratedAmount := CalculateProration(currentSub, newPlan)

    // Process payment and update
    ProcessPaymentAndUpgrade(tenantID, proratedAmount, newPlanID)
}

// Tenant downgrades - scheduled for period end
func DowngradeTenantPlan(tenantID string, newPlanID string) error {
    // Schedule for end of current period
    SchedulePlanChange(tenantID, newPlanID, GetPeriodEnd(tenantID))

    // If increasing commission rate, keep current rate until period end
    PreserveFavorableCommission(tenantID)
}
```

### Dual-Mode Payment Processing Logic
```go
// Main payment processing function that routes to appropriate mode
func ProcessPayment(order Order, customer Customer, tenant Tenant) (*PaymentResult, error) {
    // Determine payment mode based on region and payment method
    paymentMode := DeterminePaymentMode(customer.Region, customer.PaymentMethod)
    gateway := GetPaymentGateway(customer.Region)

    // Calculate commission
    commission := CalculateEffectiveCommission(tenant.ID, order.Service, order.Amount, nil)

    switch paymentMode {
    case INSTANT_SPLIT:
        return ProcessInstantSplitPayment(order, customer, tenant, commission, gateway)
    case COLLECT_DISBURSE:
        return ProcessCollectDisbursePayment(order, customer, tenant, commission, gateway)
    default:
        return nil, errors.New("unsupported payment mode")
    }
}

// Mode 1: Instant Split Payment (Stripe Connect style)
func ProcessInstantSplitPayment(order Order, customer Customer, tenant Tenant, commission Commission, gateway PaymentGateway) (*PaymentResult, error) {
    // Create split payment instruction
    splitInstruction := SplitInstruction{
        TotalAmount: order.Amount,
        Splits: []Split{
            {Recipient: tenant.StripeAccountID, Amount: order.Amount.Sub(commission.Amount)},
            {Recipient: PLATFORM_ACCOUNT, Amount: commission.Amount},
        },
    }

    // Process payment with immediate split
    tx, err := gateway.ProcessSplitPayment(customer, splitInstruction)
    if err != nil {
        return nil, err
    }

    // Record transaction (already split)
    RecordCollection(PaymentCollection{
        TransactionID: tx.ID,
        TenantID: tenant.ID,
        AmountCollected: order.Amount,
        PaymentMode: INSTANT_SPLIT,
        DisbursementStatus: "completed", // Instant disbursement
    })

    return &PaymentResult{TransactionID: tx.ID, Status: "completed"}, nil
}

// Mode 2: Collect and Disburse (Mobile Money)
func ProcessCollectDisbursePayment(order Order, customer Customer, tenant Tenant, commission Commission, gateway PaymentGateway) (*PaymentResult, error) {
    // Step 1: Collect full payment to platform holding account
    tx, err := gateway.CollectPayment(order.Amount, customer)
    if err != nil {
        return nil, err
    }

    // Step 2: Record collection with pending disbursement
    collection := RecordCollection(PaymentCollection{
        TransactionID: tx.ID,
        TenantID: tenant.ID,
        AmountCollected: order.Amount,
        TenantShare: order.Amount.Sub(commission.Amount),
        PlatformShare: commission.Amount,
        PaymentMode: COLLECT_DISBURSE,
        CollectionStatus: "successful",
        DisbursementStatus: "pending",
    })

    // Step 3: Schedule disbursement
    ScheduleDisbursement(tenant.ID, collection.TenantShare, collection.ID)

    return &PaymentResult{TransactionID: tx.ID, Status: "collected_pending_disbursement"}, nil
}

// Batch disbursement processor (runs on schedule)
func ProcessScheduledDisbursements() error {
    // Get pending disbursements grouped by provider
    batches := GetPendingDisbursementBatches()

    for provider, disbursements := range batches {
        gateway := GetGatewayForProvider(provider)

        // Check if provider supports bulk transfer
        if gateway.SupportsBulkTransfer() {
            ProcessBulkDisbursement(gateway, disbursements)
        } else {
            // Process individually
            for _, disbursement := range disbursements {
                ProcessSingleDisbursement(gateway, disbursement)
            }
        }
    }

    return nil
}

// Process bulk disbursement for efficiency
func ProcessBulkDisbursement(gateway PaymentGateway, disbursements []Disbursement) error {
    // Create batch
    batch := CreateDisbursementBatch(disbursements)

    // Validate float balance
    if !HasSufficientFloat(batch.TotalAmount) {
        return errors.New("insufficient float balance")
    }

    // Submit to gateway
    result, err := gateway.BatchDisburse(batch)
    if err != nil {
        HandleDisbursementFailure(batch, err)
        return err
    }

    // Update disbursement statuses
    for _, disbursement := range disbursements {
        if result.IsSuccessful(disbursement.ID) {
            MarkDisbursementComplete(disbursement.ID)
            NotifyTenantOfPayout(disbursement.TenantID, disbursement.Amount)
        } else {
            ScheduleRetry(disbursement.ID)
        }
    }

    // Update float balance
    UpdateFloatBalance(-batch.TotalAmount)

    return nil
}

// Daily reconciliation process
func RunDailyReconciliation() error {
    date := time.Now().AddDate(0, 0, -1) // Previous day

    // Get all collections and disbursements
    collections := GetCollections(date)
    disbursements := GetDisbursements(date)

    reconciliation := SettlementReconciliation{
        ReconciliationDate: date,
        TotalCollected: SumCollections(collections),
        TotalDisbursed: SumDisbursements(disbursements),
        TotalCommission: SumCommissions(collections),
    }

    // Calculate expected vs actual float
    expectedFloat := reconciliation.TotalCollected - reconciliation.TotalDisbursed
    actualFloat := GetFloatBalance()

    reconciliation.DiscrepancyAmount = actualFloat - expectedFloat

    if reconciliation.DiscrepancyAmount.Abs().GreaterThan(RECONCILIATION_THRESHOLD) {
        // Flag for manual investigation
        reconciliation.Status = "discrepancy"
        NotifyFinanceTeam(reconciliation)
    } else {
        reconciliation.Status = "balanced"
    }

    SaveReconciliation(reconciliation)

    return nil
}
```

## Pricing & Revenue Calculations

### Hybrid Revenue Calculation
```go
// Calculate total revenue from a tenant (subscription + transactions)
func CalculateTenantRevenue(tenantID string, period Period) TenantRevenue {
    subscription := GetTenantSubscription(tenantID)

    revenue := TenantRevenue{
        TenantID: tenantID,
        Period: period,
    }

    // Subscription revenue
    if subscription != nil && subscription.IsActive() {
        revenue.SubscriptionRevenue = subscription.GetPeriodPrice()
    }

    // Transaction revenue (commissions)
    transactions := GetTenantTransactions(tenantID, period)
    for _, tx := range transactions {
        revenue.TransactionRevenue = revenue.TransactionRevenue.Add(tx.CommissionAmount)
    }

    // Total platform revenue from this tenant
    revenue.TotalRevenue = revenue.SubscriptionRevenue.Add(revenue.TransactionRevenue)

    return revenue
}

// Calculate tenant's net earnings after platform fees
func CalculateTenantEarnings(tenantID string, period Period) TenantEarnings {
    transactions := GetTenantTransactions(tenantID, period)

    earnings := TenantEarnings{
        TenantID: tenantID,
        Period: period,
    }

    for _, tx := range transactions {
        earnings.GrossSales = earnings.GrossSales.Add(tx.TransactionAmount)
        earnings.TotalFees = earnings.TotalFees.Add(tx.TotalFeeAmount)
    }

    // Net earnings after all fees
    earnings.NetEarnings = earnings.GrossSales.Sub(earnings.TotalFees)

    // Deduct subscription fee if applicable
    subscription := GetTenantSubscription(tenantID)
    if subscription != nil {
        earnings.SubscriptionFee = subscription.GetPeriodPrice()
        earnings.NetEarnings = earnings.NetEarnings.Sub(earnings.SubscriptionFee)
    }

    return earnings
}
```

### Proration for Plan Changes
```go
func CalculateProration(currentSub Subscription, newPlan Plan, changeDate time.Time) decimal.Decimal {
    // Calculate days remaining in current period
    daysRemaining := currentSub.CurrentPeriodEnd.Sub(changeDate).Hours() / 24
    totalDays := currentSub.CurrentPeriodEnd.Sub(currentSub.CurrentPeriodStart).Hours() / 24

    // Calculate unused subscription fee
    unusedSubscriptionFee := currentSub.Plan.SubscriptionPrice.
        Mul(decimal.NewFromFloat(daysRemaining)).
        Div(decimal.NewFromFloat(totalDays))

    // Calculate new plan fee for remaining period
    newSubscriptionFee := newPlan.SubscriptionPrice.
        Mul(decimal.NewFromFloat(daysRemaining)).
        Div(decimal.NewFromFloat(totalDays))

    // Return amount to charge (positive) or credit (negative)
    return newSubscriptionFee.Sub(unusedSubscriptionFee)
}
```

## Performance Requirements

### Response Times
- **Plan Retrieval**: <30ms
- **Commission Calculation**: <50ms
- **Subscription Validation**: <50ms
- **Transaction Fee Recording**: <100ms
- **Split Payment Processing**: <200ms
- **Plan Changes**: <2 seconds
- **Billing Processing**: <5 seconds
- **Payout Processing**: <10 seconds

### Scalability
- **Concurrent Tenants**: Support 500,000+ active tenants
- **Plans Management**: Handle 1,000+ active plans across services
- **Transaction Processing**: 50,000+ transactions per second
- **Commission Calculations**: 100,000+ per second
- **Usage Events**: Process 10,000+ events per second
- **Billing Operations**: Handle 200,000+ invoices per month
- **Payout Operations**: Process 100,000+ payouts per month
- **API Throughput**: 5,000+ requests per second

### High Availability
- **Uptime SLA**: 99.99% availability
- **Failover Time**: <30 seconds
- **Data Replication**: Multi-region replication
- **Backup Frequency**: Real-time for transactions, hourly for configurations
