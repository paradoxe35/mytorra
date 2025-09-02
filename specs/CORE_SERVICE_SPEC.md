# Core Service Specification

## Overview
The Core Service is the foundational nervous system of the MyTorra platform, providing centralized user management, authentication, payments, notifications, and subscription management for all other services.

## Service Responsibilities

### Primary Functions
1. **User Account Management** - Central user registry and profile management
2. **Authentication & Authorization** - JWT, MFA, OAuth2, permission management
3. **Payment Processing** - Abstracted payment gateway integration
4. **Notification System** - Generic notification sender (email, SMS, push)
5. **Subscription Engine** - Generic subscription management for all services
6. **Admin Management** - Platform-wide administrative functions

## Features & Functionality

### 1. User Management
- **User Registration & Profiles**
  - Email/phone verification
  - Multi-profile support (personal, business, restaurant, etc.)
  - Profile completion tracking
  - KYC integration for business accounts

- **Account Types**
  - Consumer accounts (buyers)
  - Seller accounts (marketplace vendors)
  - Business accounts (restaurants, hotels, shops)
  - Driver accounts (delivery, cab)
  - Admin accounts (platform administrators)

### 2. Authentication & Security
- **Multi-Factor Authentication (MFA)**
  - SMS/Email OTP
  - TOTP (Google Authenticator)
  - Biometric authentication support

- **OAuth2 Integration**
  - Google, Facebook, Apple Sign-In
  - Platform-specific OAuth for partner integrations

- **JWT Token Management**
  - Access tokens (short-lived)
  - Refresh tokens (long-lived)
  - Token blacklisting for logout

### 3. Permission & Role Management
- **Hierarchical Role System**
  - Super Admin (platform owner)
  - Service Admin (per service management)
  - Business Owner (restaurant, hotel, shop owner)
  - Employee (restaurant staff, hotel staff)
  - Customer (end users)

- **Permission Granularity**
  - Resource-based permissions (users, products, orders, etc.)
  - Action-based permissions (create, read, update, delete)
  - Service-specific permissions
  - Geographic/location-based permissions

### 4. Payment Processing
- **Payment Gateway Abstraction**
  - Multiple payment provider support (Stripe, PayStack, Flutterwave)
  - Automatic failover between providers
  - Payment method management (cards, mobile money, bank transfer)

- **Financial Operations**
  - Payment processing and validation
  - Refund management
  - Payout scheduling for vendors
  - Transaction history and reconciliation
  - Multi-currency support

### 5. Subscription Engine
- **Generic Subscription Management**
  - Plan creation and management
  - Feature-based subscriptions
  - Usage-based billing
  - Tiered pricing models

- **Service-Specific Plans**
  - Restaurant plans (MyTorra Food)
  - Hotel plans (MyTorra Booking)
  - Marketplace vendor plans
  - Premium user plans

- **Billing Management**
  - Automatic billing cycles
  - Proration for plan changes
  - Grace periods and suspension
  - Invoice generation

### 6. Notification System
- **Multi-Channel Notifications**
  - Email notifications
  - SMS notifications
  - Push notifications (mobile/web)
  - In-app notifications

- **Notification Types**
  - Transactional (order confirmations, payments)
  - Marketing (promotions, newsletters)
  - System alerts (security, maintenance)
  - User-triggered (messages, reminders)

## Admin Dashboard Features

### 1. User Management
- User search and filtering
- Account verification management
- User profile editing
- Account suspension/activation
- User activity monitoring

### 2. Payment Administration
- Transaction monitoring
- Refund processing
- Payout management
- Financial reporting
- Payment provider configuration

### 3. Subscription Management
- Plan creation and editing
- Subscription monitoring
- Billing issue resolution
- Revenue analytics
- Plan performance metrics

### 4. System Administration
- Service health monitoring
- Configuration management
- Audit log viewing
- Security incident management
- System maintenance tools

## API Endpoints

### Authentication Endpoints
```
POST /auth/register
POST /auth/login
POST /auth/logout
POST /auth/refresh
POST /auth/verify-email
POST /auth/verify-phone
POST /auth/forgot-password
POST /auth/reset-password
POST /auth/enable-mfa
POST /auth/disable-mfa
POST /auth/verify-mfa
```

### User Management Endpoints
```
GET /users/profile
PUT /users/profile
GET /users/{id}
PUT /users/{id}
DELETE /users/{id}
GET /users/{id}/permissions
PUT /users/{id}/permissions
```

### Payment Endpoints
```
POST /payments/process
GET /payments/{id}
GET /payments/history
POST /payments/refund
GET /payments/methods
POST /payments/methods
DELETE /payments/methods/{id}
```

### Subscription Endpoints
```
GET /subscriptions/plans
GET /subscriptions/current
POST /subscriptions/subscribe
PUT /subscriptions/change-plan
POST /subscriptions/cancel
GET /subscriptions/billing-history
```

### Admin Endpoints
```
GET /admin/users
GET /admin/payments
GET /admin/subscriptions
GET /admin/analytics
POST /admin/notifications/broadcast
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url VARCHAR(500),
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    status ENUM('active', 'suspended', 'pending') DEFAULT 'pending',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### Roles & Permissions Tables
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    service VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    granted_at TIMESTAMP DEFAULT NOW(),
    granted_by UUID REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);
```

### Subscriptions Tables
```sql
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    service VARCHAR(50) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_cycle ENUM('monthly', 'yearly', 'weekly') DEFAULT 'monthly',
    features JSONB,
    max_limits JSONB,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    plan_id UUID REFERENCES subscription_plans(id),
    status ENUM('active', 'cancelled', 'expired', 'suspended') DEFAULT 'active',
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Integration Patterns

### With Other Services
- **Marketplace Service**: User authentication, seller subscription validation
- **Delivery Service**: Restaurant owner authentication, delivery plan management
- **Booking Service**: Hotel owner authentication, booking plan validation
- **TorraCab Service**: Driver authentication, ride revenue processing
- **Media Service**: User authentication for secure uploads
- **Realtime Service**: User authentication for WebSocket connections

### Event Publishing
- `UserRegistered`: When new user registers
- `SubscriptionChanged`: When user subscription changes
- `PaymentProcessed`: When payment is completed
- `UserSuspended`: When user account is suspended

### Event Consumption
- `OrderCompleted`: Update user transaction history
- `ServiceUsageRecorded`: Update subscription usage metrics
- `SecurityIncident`: Log and notify administrators
