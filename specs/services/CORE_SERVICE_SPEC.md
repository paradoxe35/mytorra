# Core Service Specification

## Overview
The Core Service is the foundational nervous system of the MyTorra platform, providing centralized user management, authentication, payments, notifications, and subscription management for all other services.

## Service Responsibilities

### Primary Functions
1. **User Account Management** - Central user registry and profile management
2. **Area Management System** - Dynamic geographic operational boundaries and user area detection
3. **Organization Management** - Multi-user organization structures with hierarchical roles
4. **Authentication & Authorization** - JWT, MFA, OAuth2, permission management
5. **Payment Processing** - Abstracted payment gateway integration with split payment support
6. **Notification System** - Generic notification sender (email, SMS, push)
7. **Monetization Engine** - Dynamic subscription and commission management for all services
8. **Admin Management** - Platform-wide administrative functions
9. **Automatic Permission Assignment** - Event-driven permission management based on subscriptions
10. **Tenant Financial Management** - Earnings tracking, payouts, and financial reporting

## Features & Functionality

### 1. User Management
- **User Registration & Profiles**
  - Email/phone verification
  - Multi-profile support (personal, business, restaurant, etc.)
  - Profile completion tracking
  - KYC integration for business accounts

- **Account Types**
  - Consumer accounts (buyers)
  - Individual seller accounts (marketplace vendors)
  - Organization member accounts (employees, drivers, staff)
  - Organization owner accounts (business owners, fleet managers)
  - Admin accounts (platform administrators)

### 2. Area Management System

#### Dynamic Geographic Framework
- **Hierarchical Area Structure**
  - Country → Province → City/Town → Sub-areas
  - Admin-defined operational boundaries
  - No hardcoded locations - all database-driven
  - Flexible boundary definitions using PostGIS polygons
  - Custom service zones per area

- **Area Detection & Switching**
  - **Automatic Mode**: Area changes based on GPS location
  - **Manual Mode**: User explicitly selects area
  - **Hybrid Mode**: Suggests area change with confirmation
  - **Locked Mode**: Area fixed during active operations
  - Background location monitoring for seamless transitions

#### Area-Based User Experience
- **User Area Assignment**
  - Current active area tracking
  - Area history maintenance
  - Area-specific preferences
  - Cross-area activity tracking
  - Visiting mode for temporary areas

- **Content Localization**
  - Area-specific product/service visibility
  - Localized pricing and currency
  - Regional language preferences
  - Local holiday and event calendars
  - Area-specific promotional campaigns

#### Area Administration
- **Area Configuration Dashboard**
  - Add/edit/delete geographic areas
  - Define operational boundaries via map interface
  - Set area relationships (nearby areas, metropolitan zones)
  - Configure service availability per area
  - Area-specific business rules

- **Cross-Area Operations**
  - Enable/disable cross-area trading
  - Define shipping zones between areas
  - Set area proximity for expanded search
  - Configure area migration rules
  - Manage transition zones

### 3. Organization Management

#### Organization Structure
- **Organization Creation**
  - Business registration and verification
  - Organization type configuration (restaurant, hotel, cab fleet, delivery service)
  - Legal entity documentation
  - Tax ID and business license validation
  - Owner verification and onboarding

- **Hierarchical Organization Roles**
  - **Organization Owner**: Full control over organization settings, billing, and member management
  - **Organization Admin**: User management, operational oversight, limited billing access
  - **Manager**: Department/location management, staff scheduling, performance monitoring
  - **Supervisor**: Team leadership, task assignment, quality control
  - **Employee**: Standard operational access, task execution, time tracking
  - **Driver/Agent**: Mobile-focused role for delivery drivers and cab drivers

#### Multi-User Organization Features
- **Member Management**
  - Invite users to join organization via email
  - Role-based access control within organization
  - Department/location-based user grouping
  - Employee onboarding workflows
  - Performance tracking and evaluation

- **Subscription Management**
  - Organization-level subscription billing
  - User seat management (add/remove users)
  - Feature allocation across organization members
  - Usage tracking per user and department
  - Cost center allocation and reporting

- **Organization Settings**
  - Branding customization (logo, colors, themes)
  - Operational configurations (work hours, locations)
  - Policy management (break times, overtime rules)
  - Integration settings (third-party tools)
  - Notification preferences by role

#### Service-Specific Organization Types

**Restaurant Organizations**
- **Owner**: Complete restaurant chain management
- **Restaurant Manager**: Single location management
- **Kitchen Manager**: Kitchen operations and staff
- **Server**: Order taking and customer service
- **Delivery Coordinator**: Delivery logistics

**Hotel Organizations**
- **Hotel Owner**: Multi-property management
- **Property Manager**: Single hotel operations
- **Front Desk Manager**: Guest services oversight
- **Front Desk Staff**: Check-in/check-out operations
- **Housekeeping Manager**: Cleaning operations oversight
- **Housekeeping Staff**: Room cleaning and maintenance

**Cab Fleet Organizations**
- **Fleet Owner**: Complete fleet management and strategy
- **Fleet Manager**: Operations management and driver oversight
- **Dispatch Manager**: Ride assignment and coordination
- **Driver**: Ride execution and customer service
- **Maintenance Coordinator**: Vehicle maintenance scheduling

**Delivery Service Organizations**
- **Service Owner**: Multi-region delivery service management
- **Regional Manager**: Geographic area management
- **Delivery Coordinator**: Route optimization and driver management
- **Driver**: Package delivery and customer interaction
- **Warehouse Manager**: Hub operations and inventory

### 3. Authentication & Security
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

- **Automatic Permission Assignment**
  - Subscription-based permission templates
  - Event-driven permission updates
  - Plan upgrade/downgrade permission adjustment
  - Organization member permission distribution
  - Trial and grace period permission management

### 4. Dual-Mode Payment Processing
- **Payment Gateway Abstraction**
  - Dual-mode support (Instant Split vs Collect-and-Disburse)
  - Regional gateway selection:
    - **Card Markets**: Stripe Connect (instant splits)
    - **Mobile Money Markets**: Flutterwave, Paystack, MFS Africa (collect-disburse)
  - Payment method support:
    - Mobile money (M-Pesa, MTN, Airtel, Orange)
    - Cards (Visa, Mastercard, local cards)
    - Bank transfers
    - USSD payments
    - Digital wallets

- **Collection Operations**
  - Full payment collection to platform
  - Commission calculation at collection time
  - Multi-currency support
  - Payment validation and fraud detection
  - Transaction reference tracking

- **Disbursement Operations**
  - **Instant Mode**: Real-time vendor credits
  - **Scheduled Mode**: Batch disbursements (daily/weekly)
  - Minimum threshold enforcement
  - Bulk transfer optimization
  - Failed disbursement retry logic
  - Multi-provider routing

- **Float & Reconciliation**
  - Platform float management
  - Daily settlement reconciliation
  - Reserve requirement maintenance (10-20%)
  - Discrepancy detection and resolution
  - Regulatory compliance reporting

### 5. Monetization Engine (Subscriptions & Commissions)
- **Dynamic Plan Management**
  - Admin-created plans without code changes
  - Visual plan builder interface
  - Feature registry system
  - Plan versioning and grandfathering
  - A/B testing for plans

- **Hybrid Revenue Models**
  - Subscription-based revenue streams
  - Transaction fees and commissions
  - Combined subscription + commission models
  - Listing fees for marketplaces
  - Volume-based pricing tiers

- **Multi-Tenant Architecture**
  - Tenant isolation (restaurants, hotels, drivers, sellers)
  - Tenant-specific dashboards
  - Financial tracking per tenant
  - Payout management
  - Earnings analytics

- **Split Payment Processing**
  - Stripe Connect integration
  - Automatic payment splitting
  - Commission calculation
  - Gateway fee handling
  - Multi-currency support

- **Permission Integration**
  - Automatic permission assignment on subscription
  - Dynamic permission templates
  - Commission-based access levels
  - Grace period permission retention
  - Organization seat-based permissions

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

### 3. Monetization Management
- **Dynamic Plan Builder**
  - Visual plan creation interface
  - Feature catalog management
  - Commission rule configuration
  - Plan template library
  - Version control for plans

- **Revenue Management**
  - Subscription revenue tracking
  - Transaction fee monitoring
  - Commission analytics
  - Payout scheduling
  - Revenue forecasting

- **Tenant Management**
  - Tenant onboarding workflows
  - Financial profile management
  - Earnings tracking
  - Payout configuration
  - Custom deal creation

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

### Monetization Endpoints
```
# Tenant Operations
GET /monetization/plans/available
GET /monetization/subscription/current
POST /monetization/subscribe
PUT /monetization/subscription/change
POST /monetization/subscription/cancel
GET /monetization/earnings
GET /monetization/payouts
GET /monetization/transactions

# Admin Operations
POST /admin/monetization/plans
PUT /admin/monetization/plans/{id}
POST /admin/monetization/commission-rules
GET /admin/monetization/analytics
POST /admin/monetization/payouts/process
```

### Area Management Endpoints
```
GET /areas
GET /areas/countries
GET /areas/countries/{countryId}/provinces
GET /areas/provinces/{provinceId}/cities
GET /areas/{areaId}/boundaries
POST /areas/detect-from-location
GET /users/current-area
PUT /users/current-area
GET /users/area-history
POST /admin/areas
PUT /admin/areas/{id}
DELETE /admin/areas/{id}
PUT /admin/areas/{id}/boundaries
```

### Organization Endpoints
```
GET /organizations
POST /organizations
GET /organizations/{id}
PUT /organizations/{id}
DELETE /organizations/{id}
GET /organizations/{id}/members
POST /organizations/{id}/members
PUT /organizations/{id}/members/{userId}
DELETE /organizations/{id}/members/{userId}
GET /organizations/{id}/roles
POST /organizations/{id}/roles
GET /organizations/{id}/subscription
PUT /organizations/{id}/subscription
```

### Permission Management Endpoints
```
GET /permissions/user/{userId}/effective
POST /permissions/sync-subscription/{userId}
GET /permissions/templates
POST /permissions/templates
PUT /permissions/templates/{id}
GET /permissions/plan/{planId}/template
POST /permissions/apply-template/{userId}/{templateId}
```

### Admin Endpoints
```
GET /admin/users
GET /admin/payments
GET /admin/subscriptions
GET /admin/organizations
GET /admin/analytics
POST /admin/notifications/broadcast
GET /admin/permissions/templates
POST /admin/permissions/bulk-sync
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
    current_area_id UUID REFERENCES areas(id),
    status ENUM('active', 'suspended', 'pending') DEFAULT 'pending',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_current_area_id (current_area_id)
);
```

### Areas Table
```sql
CREATE TABLE areas (
    id UUID PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    type ENUM('country', 'province', 'city', 'sub_area') NOT NULL,
    parent_area_id UUID REFERENCES areas(id),
    code VARCHAR(10), -- Country/province/city codes
    boundaries geometry(POLYGON, 4326), -- PostGIS polygon for area boundaries
    center_point geometry(POINT, 4326), -- Geographic center of area
    timezone VARCHAR(50),
    currency VARCHAR(3),
    languages JSONB, -- Supported languages in this area
    settings JSONB, -- Area-specific configurations
    status ENUM('active', 'inactive', 'testing') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_parent_area_id (parent_area_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_code (code),
    SPATIAL INDEX idx_boundaries (boundaries),
    SPATIAL INDEX idx_center_point (center_point),
    UNIQUE KEY unique_code_type (code, type)
);
```

### User Area History Table
```sql
CREATE TABLE user_area_history (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    area_id UUID REFERENCES areas(id),
    switched_at TIMESTAMP DEFAULT NOW(),
    switch_method ENUM('automatic', 'manual', 'admin') DEFAULT 'manual',
    location_at_switch geometry(POINT, 4326), -- User's GPS location when area switched
    duration_minutes INTEGER, -- Time spent in this area

    INDEX idx_user_id (user_id),
    INDEX idx_area_id (area_id),
    INDEX idx_switched_at (switched_at),
    INDEX idx_switch_method (switch_method),
    SPATIAL INDEX idx_location_at_switch (location_at_switch)
);
```

### Area Relationships Table
```sql
CREATE TABLE area_relationships (
    id UUID PRIMARY KEY,
    area_id UUID REFERENCES areas(id),
    related_area_id UUID REFERENCES areas(id),
    relationship_type ENUM('nearby', 'metropolitan', 'sister_city', 'delivery_zone') NOT NULL,
    distance_km DECIMAL(10,2),
    settings JSONB, -- Relationship-specific settings
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_area_id (area_id),
    INDEX idx_related_area_id (related_area_id),
    INDEX idx_relationship_type (relationship_type),
    UNIQUE KEY unique_area_relationship (area_id, related_area_id, relationship_type)
);
```

### Organizations Table
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    type ENUM('restaurant', 'hotel', 'cab_fleet', 'delivery_service', 'marketplace_business') NOT NULL,
    description TEXT,
    business_license VARCHAR(100),
    tax_id VARCHAR(50),
    address JSONB,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    website VARCHAR(500),
    logo_url VARCHAR(500),
    settings JSONB,
    status ENUM('pending', 'active', 'suspended', 'closed') DEFAULT 'pending',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created_by (created_by),
    INDEX idx_created_at (created_at)
);
```

### Organization Members Table
```sql
CREATE TABLE organization_members (
    id UUID PRIMARY KEY,
    organization_id UUID REFERENCES organizations(id),
    user_id UUID REFERENCES users(id),
    role ENUM('owner', 'admin', 'manager', 'supervisor', 'employee', 'driver') NOT NULL,
    department VARCHAR(100),
    location_id UUID,
    permissions JSONB,
    invited_by UUID REFERENCES users(id),
    invited_at TIMESTAMP,
    joined_at TIMESTAMP,
    status ENUM('invited', 'active', 'inactive', 'suspended') DEFAULT 'invited',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_organization_id (organization_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role (role),
    INDEX idx_status (status),
    UNIQUE KEY unique_org_user (organization_id, user_id)
);
```

### Organization Subscriptions Table
```sql
CREATE TABLE organization_subscriptions (
    id UUID PRIMARY KEY,
    organization_id UUID REFERENCES organizations(id),
    plan_id UUID REFERENCES subscription_plans(id),
    status ENUM('trial', 'active', 'past_due', 'cancelled', 'expired', 'suspended') DEFAULT 'active',
    user_seats INTEGER DEFAULT 1,
    used_seats INTEGER DEFAULT 0,
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_organization_id (organization_id),
    INDEX idx_plan_id (plan_id),
    INDEX idx_status (status),
    INDEX idx_current_period_end (current_period_end)
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

### Monetization Tables
```sql
CREATE TABLE monetization_plans (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    service VARCHAR(50) NOT NULL,
    tenant_type ENUM('restaurant', 'hotel', 'marketplace_seller', 'driver', 'delivery') NOT NULL,
    plan_type ENUM('subscription', 'commission', 'hybrid') DEFAULT 'hybrid',

    -- Pricing
    subscription_price DECIMAL(10,2) DEFAULT 0.00,
    commission_rate DECIMAL(5,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    billing_cycle ENUM('monthly', 'yearly', 'weekly', 'transaction') DEFAULT 'monthly',

    -- Configuration
    features JSONB,
    limits JSONB,
    commission_rules JSONB,

    -- Status
    status ENUM('draft', 'active', 'archived') DEFAULT 'draft',
    version INTEGER DEFAULT 1,

    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE tenant_subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    tenant_type ENUM('user', 'organization') NOT NULL,
    plan_id UUID REFERENCES monetization_plans(id),

    -- Status
    status ENUM('trial', 'active', 'past_due', 'cancelled', 'suspended') DEFAULT 'active',

    -- Custom terms
    custom_commission_rate DECIMAL(5,2),
    negotiated_price DECIMAL(10,2),

    -- Periods
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transaction_fees (
    id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    service VARCHAR(50) NOT NULL,

    -- Amounts
    transaction_amount DECIMAL(10,2) NOT NULL,
    commission_amount DECIMAL(10,2) NOT NULL,
    tenant_payout DECIMAL(10,2) NOT NULL,
    platform_revenue DECIMAL(10,2) NOT NULL,

    status ENUM('pending', 'processed', 'settled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Integration Patterns

### With Other Services
- **Marketplace Service**:
  - Seller authentication and onboarding
  - Transaction fee calculation on sales
  - Seller payout processing
  - Dynamic commission rates based on seller plans

- **Delivery Service**:
  - Restaurant authentication
  - Delivery commission calculation
  - Driver earnings tracking
  - Split payment between restaurant and delivery driver

- **Booking Service**:
  - Hotel authentication and plan management
  - Booking commission processing
  - Room rate commission calculations
  - Hotel payout scheduling

- **MyTorraCab Service**:
  - Driver authentication and onboarding
  - Ride commission processing (no subscription model)
  - Driver earnings and payout management
  - Real-time fare splitting

- **Media Service**:
  - User authentication for secure uploads
  - Storage quota enforcement based on plans

- **Realtime Service**:
  - User authentication for WebSocket connections
  - Real-time financial updates to tenant dashboards
  - Live commission calculations for ongoing transactions

### Internal Event Handling (Core Service)
The Core Service internally handles monetization events to manage permissions and dual-mode payment operations:

#### Dual-Mode Payment Event Handlers
```go
// Determine payment mode based on region and method
func (m *MonetizationEngine) DeterminePaymentMode(region Region, paymentMethod PaymentMethod) PaymentMode {
    // Check region configuration
    if region.SupportsInstantSplit && paymentMethod.Type == "card" {
        return INSTANT_SPLIT
    }

    // Default to collect-and-disburse for mobile money regions
    if paymentMethod.Type == "mobile_money" || region.PrimaryPayment == "mobile_money" {
        return COLLECT_DISBURSE
    }

    return region.DefaultPaymentMode
}

// Handle transaction based on payment mode
func (m *MonetizationEngine) HandleTransactionCompleted(tx Transaction) {
    paymentMode := DeterminePaymentMode(tx.Region, tx.PaymentMethod)

    switch paymentMode {
    case INSTANT_SPLIT:
        ProcessInstantSplitTransaction(tx)
    case COLLECT_DISBURSE:
        ProcessCollectDisburseTransaction(tx)
    }
}

// Process instant split (Stripe Connect style)
func (m *MonetizationEngine) ProcessInstantSplitTransaction(tx Transaction) {
    // Calculate commission
    commission := CalculateCommission(tx.TenantID, tx.Amount)

    // Process with instant split
    gateway := GetGateway(tx.Region)
    splitResult := gateway.ProcessInstantSplit(tx, commission)

    // Record completed transaction
    RecordCollection(PaymentCollection{
        TransactionID: tx.ID,
        PaymentMode: INSTANT_SPLIT,
        DisbursementStatus: "completed",
    })

    // Notify tenant of instant payout
    NotifyInstantPayout(tx.TenantID, splitResult.TenantAmount)
}

// Process collect-and-disburse (Mobile Money)
func (m *MonetizationEngine) ProcessCollectDisburseTransaction(tx Transaction) {
    // Step 1: Collect to platform account
    gateway := GetGateway(tx.Region)
    collection := gateway.CollectToHoldingAccount(tx)

    // Step 2: Calculate and record splits
    commission := CalculateCommission(tx.TenantID, tx.Amount)
    RecordCollection(PaymentCollection{
        TransactionID: tx.ID,
        TenantShare: tx.Amount.Sub(commission.Amount),
        PlatformShare: commission.Amount,
        PaymentMode: COLLECT_DISBURSE,
        CollectionStatus: "successful",
        DisbursementStatus: "pending",
    })

    // Step 3: Add to disbursement queue
    QueueForDisbursement(tx.TenantID, tx.Amount.Sub(commission.Amount))

    // Notify tenant of pending payout
    NotifyPendingPayout(tx.TenantID, tx.Amount.Sub(commission.Amount))
}

// Scheduled disbursement processing (for collect-and-disburse regions)
func (m *MonetizationEngine) ProcessScheduledDisbursements() {
    // Group by region and provider for efficiency
    disbursementBatches := GroupPendingDisbursements()

    for region, batches := range disbursementBatches {
        gateway := GetGateway(region)

        for provider, disbursements := range batches {
            // Check float balance
            if !HasSufficientFloat(provider, SumDisbursements(disbursements)) {
                AlertFinanceTeam("Insufficient float", provider)
                continue
            }

            // Process batch disbursement
            if gateway.SupportsBulkTransfer() {
                result := gateway.ProcessBulkDisbursement(disbursements)
                HandleBulkDisbursementResult(result)
            } else {
                // Process individually for providers without bulk support
                for _, disbursement := range disbursements {
                    ProcessSingleDisbursement(gateway, disbursement)
                }
            }
        }
    }

    // Run reconciliation after disbursements
    RunDisbursementReconciliation()
}

// Daily reconciliation for collect-and-disburse mode
func (m *MonetizationEngine) RunDailyReconciliation() {
    reconciliation := SettlementReconciliation{
        Date: time.Now().AddDate(0, 0, -1),
        Collections: GetDailyCollections(),
        Disbursements: GetDailyDisbursements(),
    }

    // Calculate platform float
    reconciliation.ExpectedFloat = reconciliation.Collections.Total - reconciliation.Disbursements.Total
    reconciliation.ActualFloat = GetActualFloatBalance()
    reconciliation.Discrepancy = reconciliation.ActualFloat - reconciliation.ExpectedFloat

    // Check for discrepancies
    if reconciliation.Discrepancy.Abs() > THRESHOLD {
        reconciliation.Status = "DISCREPANCY"
        CreateInvestigationTicket(reconciliation)
    } else {
        reconciliation.Status = "BALANCED"
    }

    SaveReconciliation(reconciliation)
    GenerateFinancialReports(reconciliation)
}

// Plan migration handling
func (m *MonetizationEngine) HandlePlanModified(planID string, changes PlanChanges) {
    affectedTenants := GetTenantsOnPlan(planID)

    for _, tenant := range affectedTenants {
        if ShouldGrandfather(tenant, changes) {
            // Keep tenant on old plan version
            MarkAsGrandfathered(tenant.SubscriptionID)
        } else {
            // Schedule migration to new plan version
            SchedulePlanMigration(tenant.ID, planID, changes.EffectiveDate)
        }
    }

    // Publish event
    eventBus.Publish("plan.modified", planID, affectedTenants)
}
```

#### Organization Event Handlers
```go
func (s *OrganizationService) HandleMemberAdded(orgID, userID, role string) {
    // Get organization's subscription
    sub := subscriptionService.GetOrganizationSubscription(orgID)
    // Apply role-based permissions from plan
    permissionService.ApplyOrganizationMemberPermissions(userID, sub.PlanID, role)
    // Publish event for other services
    eventBus.Publish("organization.member_added", orgID, userID)
}

func (s *OrganizationService) HandleMemberRoleChanged(orgID, userID, newRole string) {
    // Get organization's subscription
    sub := subscriptionService.GetOrganizationSubscription(orgID)
    // Update permissions based on new role
    permissionService.UpdateOrganizationMemberPermissions(userID, sub.PlanID, newRole)
    // Publish event for other services
    eventBus.Publish("organization.member_role_changed", orgID, userID, newRole)
}
```

### Event Publishing (To Other Services)
- `UserRegistered`: When new user registers
- `TenantOnboarded`: When business entity joins platform
- `TenantSubscribed`: When tenant subscribes to plan
- `PlanChanged`: When tenant changes subscription plan
- `TransactionFeeProcessed`: When commission is calculated and applied
- `TenantEarningsUpdated`: When tenant earnings change
- `PayoutProcessed`: When tenant payout is completed
- `PaymentProcessed`: When payment is completed
- `UserSuspended`: When user account is suspended
- `PermissionsUpdated`: When user permissions change
- `OrganizationMemberAdded`: When member joins organization
- `OrganizationMemberRemoved`: When member leaves organization
- `CommissionRateChanged`: When tenant's commission rate changes

### Event Consumption (From Other Services)
- `OrderCompleted`: Calculate commission and update earnings
- `BookingCompleted`: Apply booking commission
- `RideCompleted`: Apply cab service commission
- `DeliveryCompleted`: Apply delivery commission
- `ProductSold`: Apply marketplace commission
- `ServiceUsageRecorded`: Update usage metrics
- `BusinessVerified`: Enable full monetization features
- `DriverApproved`: Set driver commission rates
- `PaymentReceived`: Trigger split payment processing
