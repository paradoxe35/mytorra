# Payment Engine Specification

## Overview
The Payment Engine is a comprehensive payment processing system that provides secure, reliable, and multi-provider payment capabilities across all MyTorra platform services. It abstracts payment complexity and ensures PCI compliance while supporting multiple payment methods and currencies.

## Core Responsibilities

### Primary Functions
1. **Payment Processing** - Secure transaction processing across multiple providers
2. **Payment Method Management** - Store and manage customer payment methods
3. **Multi-Provider Support** - Abstraction layer for multiple payment gateways
4. **Refund Management** - Automated and manual refund processing
5. **Payout Management** - Vendor/seller payout processing and scheduling
6. **Fraud Detection** - Real-time fraud prevention and risk assessment

## Features & Functionality

### 1. Multi-Provider Payment Gateway Support

#### Supported Payment Providers
- **Stripe**
  - Credit/debit cards
  - Digital wallets (Apple Pay, Google Pay)
  - Bank transfers
  - SEPA payments
  - Local payment methods

- **PayStack** (African Markets)
  - Local cards and bank accounts
  - Mobile money (MTN, Airtel, etc.)
  - Bank transfers
  - USSD payments
  - QR code payments

- **Flutterwave** (African Markets)
  - Card payments
  - Mobile money
  - Bank transfers
  - Cryptocurrency
  - International payments

- **Paypal**
  - PayPal account payments
  - Credit/debit cards via PayPal
  - PayPal Credit
  - International payments

#### Provider Abstraction Layer
- **Unified API Interface**
  - Consistent payment processing interface
  - Standardized response formats
  - Error handling abstraction
  - Provider-agnostic webhooks
  - Automatic failover support

- **Provider Selection Logic**
  - Automatic provider selection based on:
    - Customer location
    - Payment method
    - Transaction amount
    - Provider availability
    - Cost optimization

### 2. Payment Methods Management

#### Supported Payment Methods
- **Credit/Debit Cards**
  - Visa, Mastercard, American Express
  - Local card schemes
  - Tokenization for security
  - Card verification (CVV)
  - 3D Secure authentication

- **Digital Wallets**
  - Apple Pay
  - Google Pay
  - Samsung Pay
  - Local wallet services
  - Cryptocurrency wallets

- **Bank Transfers**
  - ACH transfers
  - Wire transfers
  - SEPA transfers
  - Local bank networks
  - Instant bank payments

- **Mobile Money**
  - MTN Mobile Money
  - Airtel Money
  - Vodafone Cash
  - Orange Money
  - Other regional providers

- **Alternative Payments**
  - Buy now, pay later (BNPL)
  - Installment payments
  - Gift cards/vouchers
  - Loyalty points
  - Store credit

### 3. Transaction Processing

#### Payment Flow Management
- **Payment Intent Creation**
  - Amount calculation
  - Currency handling
  - Tax computation
  - Fee calculation
  - Metadata collection

- **Authentication & Authorization**
  - Customer authentication
  - Payment method verification
  - Risk assessment
  - Fraud detection
  - 3D Secure handling

- **Settlement & Capture**
  - Authorization capture
  - Partial captures
  - Multi-party settlements
  - Split payments
  - Escrow handling

#### Transaction States
```
created -> pending -> processing -> succeeded -> settled
        \-> requires_action -> processing
        \-> cancelled
        \-> failed -> refunded (optional)
```

### 4. Multi-Currency Support

#### Currency Management
- **Supported Currencies**
  - Major currencies (USD, EUR, GBP, etc.)
  - Local currencies (NGN, KES, GHS, etc.)
  - Cryptocurrency (BTC, ETH, USDC)
  - Digital currencies
  - Regional currencies

- **Exchange Rate Management**
  - Real-time exchange rates
  - Rate caching and optimization
  - Custom rate sources
  - Rate change notifications
  - Historical rate tracking

#### Currency Conversion
- **Dynamic Currency Conversion**
  - Real-time rate application
  - Customer currency selection
  - Rate transparency
  - Conversion fees
  - Multi-currency checkouts

### 5. Refund & Dispute Management

#### Refund Processing
- **Refund Types**
  - Full refunds
  - Partial refunds
  - Automatic refunds
  - Manual refunds
  - Split refunds

- **Refund Workflow**
  - Refund request validation
  - Multi-party refund splitting
  - Refund status tracking
  - Customer notification
  - Accounting integration

#### Dispute Management
- **Chargeback Handling**
  - Dispute notification
  - Evidence collection
  - Response automation
  - Status tracking
  - Resolution monitoring

### 6. Payout Management

#### Vendor Payout System
- **Payout Scheduling**
  - Daily payouts
  - Weekly payouts
  - Monthly payouts
  - Custom schedules
  - Instant payouts (premium)

- **Payout Methods**
  - Bank transfers
  - Digital wallets
  - Mobile money
  - Cryptocurrency
  - Check payments

#### Split Payment Architecture
- **Multi-Party Transactions**
  - Platform commission
  - Vendor payments
  - Service provider fees
  - Tax withholdings
  - Marketplace splits

### 7. Fraud Detection & Security

#### Real-time Fraud Detection
- **Risk Assessment**
  - Machine learning models
  - Velocity checks
  - Device fingerprinting
  - IP geolocation
  - Behavioral analysis

- **Risk Actions**
  - Transaction blocking
  - Additional verification
  - Manual review queuing
  - Customer communication
  - Account flagging

#### Security Measures
- **PCI Compliance**
  - PCI DSS Level 1 compliance
  - Secure card data handling
  - Tokenization
  - Encryption at rest/transit
  - Regular security audits

- **Data Protection**
  - Personal data encryption
  - GDPR compliance
  - Data retention policies
  - Secure API access
  - Audit trail maintenance

## Service-Specific Payment Flows

### 1. Marketplace Payments
- **Order Payment Flow**
  - Customer payment capture
  - Escrow holding
  - Seller verification
  - Automatic release on delivery
  - Dispute handling

- **Commission Structure**
  - Platform commission deduction
  - Seller payout calculation
  - Tax handling
  - Fee transparency
  - Revenue sharing

### 2. Restaurant Payments
- **Food Order Payments**
  - Pre-authorization
  - Order confirmation capture
  - Tip handling
  - Delivery fee processing
  - Restaurant payout

### 3. Booking Payments
- **Hotel Reservation Payments**
  - Booking deposit
  - Cancellation policies
  - No-show charges
  - Final payment capture
  - Property payout

### 4. Ride Payments
- **TorraCab Payments**
  - Automatic fare calculation
  - Dynamic pricing handling
  - Tip processing
  - Driver payout
  - Expense splitting

## Admin Dashboard Features

### 1. Transaction Monitoring
- **Real-time Dashboard**
  - Transaction volume metrics
  - Success/failure rates
  - Revenue analytics
  - Provider performance
  - Geographic distribution

- **Transaction Management**
  - Transaction search and filtering
  - Status tracking
  - Error investigation
  - Manual interventions
  - Refund processing

### 2. Provider Management
- **Gateway Configuration**
  - Provider settings
  - API key management
  - Webhook configuration
  - Rate limit settings
  - Failover rules

- **Performance Analytics**
  - Provider success rates
  - Response times
  - Cost analysis
  - Geographic performance
  - Uptime monitoring

### 3. Financial Management
- **Revenue Tracking**
  - Daily/weekly/monthly revenue
  - Service breakdown
  - Currency analysis
  - Commission tracking
  - Payout management

- **Reconciliation**
  - Provider settlement reconciliation
  - Discrepancy identification
  - Manual adjustments
  - Financial reporting
  - Audit trail

### 4. Risk Management
- **Fraud Monitoring**
  - Fraud alerts
  - Risk score monitoring
  - Investigation tools
  - Rule management
  - Pattern analysis

- **Compliance Monitoring**
  - PCI compliance status
  - Regulatory compliance
  - Audit preparation
  - Security monitoring
  - Incident response

## API Endpoints

### Payment Processing
```
POST /payments/intents
GET /payments/intents/{id}
POST /payments/intents/{id}/confirm
POST /payments/intents/{id}/cancel
GET /payments/methods
POST /payments/methods
DELETE /payments/methods/{id}
```

### Transaction Management
```
GET /transactions
GET /transactions/{id}
POST /transactions/{id}/refund
GET /transactions/{id}/events
POST /transactions/{id}/capture
```

### Payout Management
```
GET /payouts
POST /payouts
GET /payouts/{id}
POST /payouts/{id}/cancel
GET /payouts/schedule
```

### Provider Operations
```
POST /webhooks/stripe
POST /webhooks/paystack
POST /webhooks/flutterwave
GET /providers/status
GET /providers/rates
```

## Database Schema

### Payment Intents Table
```sql
CREATE TABLE payment_intents (
    id UUID PRIMARY KEY,
    external_id VARCHAR(200), -- Provider's payment intent ID
    user_id UUID NOT NULL,
    service VARCHAR(50) NOT NULL,
    provider ENUM('stripe', 'paystack', 'flutterwave', 'paypal') NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('created', 'pending', 'processing', 'succeeded', 'failed', 'cancelled', 'requires_action') DEFAULT 'created',
    payment_method_id UUID REFERENCES payment_methods(id),
    description TEXT,
    metadata JSONB,
    client_secret VARCHAR(500),
    confirmation_method ENUM('automatic', 'manual') DEFAULT 'automatic',
    capture_method ENUM('automatic', 'manual') DEFAULT 'automatic',
    setup_future_usage BOOLEAN DEFAULT FALSE,
    last_payment_error JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_external_id (external_id),
    INDEX idx_user_id (user_id),
    INDEX idx_service (service),
    INDEX idx_provider (provider),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Payment Methods Table
```sql
CREATE TABLE payment_methods (
    id UUID PRIMARY KEY,
    external_id VARCHAR(200), -- Provider's payment method ID
    user_id UUID NOT NULL,
    provider ENUM('stripe', 'paystack', 'flutterwave', 'paypal') NOT NULL,
    type ENUM('card', 'bank_account', 'digital_wallet', 'mobile_money') NOT NULL,
    card_brand VARCHAR(20),
    card_last4 VARCHAR(4),
    card_exp_month INTEGER,
    card_exp_year INTEGER,
    bank_name VARCHAR(100),
    account_holder_type VARCHAR(20),
    is_default BOOLEAN DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_external_id (external_id),
    INDEX idx_user_id (user_id),
    INDEX idx_provider (provider),
    INDEX idx_type (type),
    INDEX idx_is_default (is_default)
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    payment_intent_id UUID REFERENCES payment_intents(id),
    external_id VARCHAR(200), -- Provider's transaction/charge ID
    type ENUM('payment', 'refund', 'payout', 'transfer') NOT NULL,
    status ENUM('pending', 'succeeded', 'failed', 'cancelled') DEFAULT 'pending',
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    fee_amount DECIMAL(10,2) DEFAULT 0.00,
    net_amount DECIMAL(12,2) NOT NULL,
    description TEXT,
    metadata JSONB,
    failure_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_payment_intent_id (payment_intent_id),
    INDEX idx_external_id (external_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Refunds Table
```sql
CREATE TABLE refunds (
    id UUID PRIMARY KEY,
    transaction_id UUID REFERENCES transactions(id),
    external_id VARCHAR(200), -- Provider's refund ID
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('pending', 'succeeded', 'failed', 'cancelled') DEFAULT 'pending',
    reason ENUM('requested_by_customer', 'duplicate', 'fraudulent', 'subscription_cancelled') DEFAULT 'requested_by_customer',
    description TEXT,
    metadata JSONB,
    failure_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_transaction_id (transaction_id),
    INDEX idx_external_id (external_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Payouts Table
```sql
CREATE TABLE payouts (
    id UUID PRIMARY KEY,
    recipient_id UUID NOT NULL, -- Seller/vendor/driver ID
    external_id VARCHAR(200), -- Provider's payout ID
    provider ENUM('stripe', 'paystack', 'flutterwave', 'bank_transfer') NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('pending', 'in_transit', 'paid', 'failed', 'cancelled') DEFAULT 'pending',
    method ENUM('bank_account', 'debit_card', 'digital_wallet', 'mobile_money') NOT NULL,
    destination_info JSONB NOT NULL,
    description TEXT,
    metadata JSONB,
    failure_reason TEXT,
    scheduled_at TIMESTAMP,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_recipient_id (recipient_id),
    INDEX idx_external_id (external_id),
    INDEX idx_provider (provider),
    INDEX idx_status (status),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_created_at (created_at)
);
```

### Split Payments Table
```sql
CREATE TABLE split_payments (
    id UUID PRIMARY KEY,
    parent_transaction_id UUID REFERENCES transactions(id),
    recipient_id UUID NOT NULL,
    recipient_type ENUM('platform', 'seller', 'driver', 'restaurant', 'hotel') NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    percentage DECIMAL(5,2),
    fee_type ENUM('commission', 'processing_fee', 'service_fee', 'tax') NOT NULL,
    status ENUM('pending', 'completed', 'failed') DEFAULT 'pending',
    payout_id UUID REFERENCES payouts(id),
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_parent_transaction_id (parent_transaction_id),
    INDEX idx_recipient_id (recipient_id),
    INDEX idx_recipient_type (recipient_type),
    INDEX idx_status (status)
);
```

### Fraud Detection Table
```sql
CREATE TABLE fraud_assessments (
    id UUID PRIMARY KEY,
    payment_intent_id UUID REFERENCES payment_intents(id),
    risk_score DECIMAL(5,4) NOT NULL, -- 0.0000 to 1.0000
    risk_level ENUM('low', 'medium', 'high', 'highest') NOT NULL,
    risk_factors JSONB NOT NULL,
    decision ENUM('allow', 'review', 'block') NOT NULL,
    decision_reason TEXT,
    reviewer_id UUID,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_payment_intent_id (payment_intent_id),
    INDEX idx_risk_level (risk_level),
    INDEX idx_decision (decision),
    INDEX idx_created_at (created_at)
);
```

## Integration Patterns

### With Core Service
- User authentication and authorization
- Subscription payment processing
- User payment method management
- Billing coordination

### With Service Integration
```go
// Marketplace order payment
func ProcessMarketplacePayment(orderID string, amount decimal.Decimal, currency string) (*PaymentIntent, error) {
    intent, err := paymentEngine.CreatePaymentIntent(CreatePaymentIntentParams{
        Amount:      amount,
        Currency:    currency,
        Service:     "marketplace",
        Description: fmt.Sprintf("Order payment for order %s", orderID),
        Metadata: map[string]string{
            "order_id": orderID,
            "service":  "marketplace",
        },
    })

    return intent, err
}

// Restaurant payout processing
func ProcessRestaurantPayout(restaurantID string, amount decimal.Decimal) error {
    return paymentEngine.CreatePayout(CreatePayoutParams{
        RecipientID: restaurantID,
        Amount:      amount,
        Currency:    "USD",
        Method:      "bank_account",
        Description: "Weekly restaurant earnings",
    })
}
```

### Event Publishing
- `PaymentIntentCreated`: New payment intent created
- `PaymentSucceeded`: Payment completed successfully
- `PaymentFailed`: Payment processing failed
- `RefundProcessed`: Refund completed
- `PayoutCompleted`: Payout sent to recipient
- `FraudDetected`: Suspicious transaction detected

### Event Consumption
- `OrderCreated`: Create payment intent for order
- `SubscriptionBilled`: Process subscription payment
- `RefundRequested`: Process refund request

## Security & Compliance

### PCI DSS Compliance
- **Level 1 PCI DSS certification**
- Secure card data handling
- No storage of sensitive card data
- Tokenization of payment methods
- Encrypted data transmission
- Regular security assessments

### Data Protection
- **GDPR compliance**
- Data minimization
- Right to erasure
- Data portability
- Consent management
- Privacy by design

### API Security
- **Authentication**
  - API key authentication
  - JWT token validation
  - Rate limiting
  - IP whitelisting
  - Request signing

## Performance Requirements

### Transaction Processing
- **Processing Time**: <2 seconds for standard transactions
- **Throughput**: Support 10,000+ transactions per minute
- **Availability**: 99.99% uptime SLA
- **Latency**: <100ms API response times

### Scalability
- **Concurrent Processing**: Handle 50,000+ concurrent payment intents
- **Provider Failover**: <5 seconds failover time
- **Database Performance**: <50ms query response times
- **Cache Performance**: 99%+ cache hit ratio
