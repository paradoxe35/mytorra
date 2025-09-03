# User Permission Management Specification

## Overview
The User Permission Management system provides a comprehensive, hierarchical, and granular permission system that controls access to resources and actions across all MyTorra platform services. It supports role-based access control (RBAC) with dynamic permission assignment and inheritance.

## Core Responsibilities

### Primary Functions
1. **Role Management** - Define and manage user roles across services
2. **Permission Control** - Granular permission assignment and validation
3. **Access Control** - Resource-based access management
4. **Delegation** - Permission delegation and temporary access
5. **Audit & Compliance** - Permission usage tracking and audit trails
6. **Multi-Tenant Support** - Service-specific and cross-service permissions

## Features & Functionality

### 1. Hierarchical Role System

#### Platform-Level Roles
- **Super Admin**
  - Full platform access
  - All service administration
  - User management across services
  - System configuration access
  - Audit log access

- **Platform Moderator**
  - Content moderation across services
  - User support capabilities
  - Limited administrative functions
  - Incident response access
  - Reporting capabilities

#### Service-Level Roles

##### Core Service Roles
- **Service Admin**
  - User account management
  - Payment system oversight
  - Subscription management
  - System configuration
  - Analytics access

- **Support Agent**
  - Customer support tickets
  - Basic user account assistance
  - Transaction inquiry access
  - Limited refund processing
  - Knowledge base management

##### Marketplace Service Roles
- **Marketplace Admin**
  - Seller management and verification
  - Product catalog oversight
  - Dispute resolution
  - Commission management
  - Marketplace analytics

- **Content Moderator**
  - Product listing approval
  - Review moderation
  - Content policy enforcement
  - Spam detection and removal
  - Quality control

- **Seller (Business)**
  - Product management
  - Order processing
  - Customer communication
  - Analytics access
  - Promotional tools

- **Individual Seller**
  - Basic product listing
  - Order management
  - Customer messaging
  - Limited analytics
  - Profile management

##### Restaurant Service Roles
- **Restaurant Owner**
  - Complete restaurant management
  - Staff management
  - Menu and pricing control
  - Financial analytics
  - Marketing tools

- **Restaurant Manager**
  - Daily operations management
  - Staff scheduling
  - Menu management
  - Order processing
  - Customer service

- **Kitchen Staff**
  - Order preparation updates
  - Inventory status updates
  - Basic menu visibility
  - Kitchen display access
  - Preparation time management

- **Front of House Staff**
  - Customer service
  - Order taking
  - Payment processing
  - Table management
  - Customer communication

##### Hotel Service Roles
- **Hotel Owner**
  - Complete property management
  - Staff management
  - Pricing and availability
  - Financial oversight
  - Marketing control

- **Hotel Manager**
  - Daily operations
  - Booking management
  - Guest services
  - Staff coordination
  - Reporting access

- **Front Desk Staff**
  - Guest check-in/check-out
  - Booking modifications
  - Guest services
  - Room status updates
  - Payment processing

- **Housekeeping Staff**
  - Room status updates
  - Maintenance requests
  - Inventory management
  - Cleaning schedules
  - Quality control

##### Delivery Service Roles
- **Delivery Coordinator**
  - Driver management
  - Route optimization
  - Delivery tracking
  - Issue resolution
  - Performance monitoring

- **Delivery Driver**
  - Order acceptance
  - Location sharing
  - Delivery confirmation
  - Customer communication
  - Earnings tracking

##### MyTorraCab Service Roles
- **Fleet Manager**
  - Driver management
  - Vehicle oversight
  - Performance monitoring
  - Territory management
  - Dispatch control

- **Cab Driver**
  - Ride acceptance
  - Location sharing
  - Navigation access
  - Passenger communication
  - Earnings management

### 2. Granular Permission System

#### Permission Categories
- **Resource Permissions**
  - Create: Can create new resources
  - Read: Can view resources
  - Update: Can modify resources
  - Delete: Can remove resources
  - List: Can view resource collections

- **Action Permissions**
  - Process: Can process transactions/orders
  - Approve: Can approve requests/content
  - Reject: Can reject requests/content
  - Suspend: Can suspend users/accounts
  - Export: Can export data

- **Administrative Permissions**
  - Manage Users: User account management
  - Manage Roles: Role assignment and modification
  - View Analytics: Access to analytics and reports
  - System Config: System configuration access
  - Audit Logs: Access to audit trails

#### Resource-Based Access Control
- **User Resources**
  - Own profile: Full access
  - Subordinate profiles: Limited management
  - Customer profiles: Service-specific access
  - Admin profiles: Restricted access

- **Business Resources**
  - Own business: Full management
  - Partner businesses: Collaboration access
  - Competitor businesses: No access
  - Platform businesses: Admin-only access

- **Content Resources**
  - Own content: Full control
  - Assigned content: Moderation rights
  - Public content: Read-only
  - Restricted content: Permission-based

### 3. Dynamic Permission Assignment

#### Context-Aware Permissions
- **Time-Based Permissions**
  - Scheduled access (work hours only)
  - Temporary elevated access
  - Expiring permissions
  - Seasonal access modifications
  - Emergency access protocols

- **Location-Based Permissions**
  - Geographic restrictions
  - Venue-specific access
  - Regional management rights
  - IP-based restrictions
  - Mobile location awareness

- **Condition-Based Permissions**
  - Performance-based access
  - Subscription-dependent features
  - Trust level requirements
  - Verification status dependency
  - Activity-based permissions

#### Permission Inheritance
- **Hierarchical Inheritance**
  - Parent role permissions
  - Cascading permissions
  - Override capabilities
  - Inheritance blocking
  - Custom inheritance rules

- **Group-Based Inheritance**
  - Team permissions
  - Department access
  - Project-based rights
  - Collaborative permissions
  - Shared resource access

### 4. Multi-Service Permission Management

#### Cross-Service Access
- **Service Integration Permissions**
  - API access rights
  - Data sharing permissions
  - Cross-service operations
  - Unified user experience
  - Service-to-service authentication

- **Unified Permission Dashboard**
  - Single sign-on (SSO)
  - Cross-service role management
  - Unified permission audit
  - Global user management
  - Service-specific overrides

#### Service-Specific Customization
- **Custom Permission Sets**
  - Service-defined permissions
  - Business logic permissions
  - Workflow-specific rights
  - Industry-specific access
  - Regulatory compliance permissions

### 5. Automatic Permission Assignment

#### Subscription-Based Permissions
- **Plan-Based Auto Assignment**
  - Permissions automatically granted on subscription activation
  - Permissions revoked on subscription cancellation/expiration
  - Plan upgrade/downgrade permission adjustments
  - Grace period permission retention
  - Trial permission management

- **Service-Specific Auto Permissions**
  - Restaurant plan permissions (menu management, analytics access)
  - Hotel plan permissions (room management, booking tools)
  - Marketplace seller permissions (product listing, order management)
  - Cab service permissions (driver management, fleet tracking)
  - Delivery service permissions (route optimization, delivery tracking)

#### Permission Templates by Plan
- **Starter/Basic Plans**
  - Core feature permissions
  - Limited resource creation rights
  - Basic analytics viewing
  - Standard support ticket creation
  - Read-only access to advanced features

- **Professional/Premium Plans**
  - Extended feature permissions
  - Unlimited resource creation
  - Advanced analytics and reporting
  - Priority support access
  - API access permissions
  - Bulk operations permissions
  - Export/import data rights

- **Enterprise Plans**
  - All available permissions for service
  - Custom permission configurations
  - White-label permissions
  - Multi-location management
  - Advanced integration access
  - Custom reporting permissions
  - Audit log full access

#### Organization Subscription Permissions
- **Seat-Based Permissions**
  - Automatic role assignment for new members
  - Permission scaling with seat count
  - Department-specific permission allocation
  - Location-based permission distribution
  - Temporary contractor permissions

- **Feature Bundle Permissions**
  - Marketing tool access permissions
  - Financial management permissions
  - Inventory management rights
  - Customer communication permissions
  - Analytics dashboard access

#### Event-Driven Permission Updates
- **Subscription Events**
  - `SubscriptionCreated`: Grant initial plan permissions
  - `SubscriptionUpgraded`: Add new tier permissions
  - `SubscriptionDowngraded`: Revoke premium permissions
  - `SubscriptionExpired`: Remove all plan permissions
  - `SubscriptionRenewed`: Restore active permissions
  - `PaymentFailed`: Enter grace period permissions
  - `TrialStarted`: Grant trial permissions
  - `TrialEnded`: Adjust to paid plan permissions

- **Business Events**
  - `BusinessVerified`: Grant verified seller permissions
  - `RestaurantOnboarded`: Enable restaurant management permissions
  - `HotelRegistered`: Activate hotel management permissions
  - `DriverApproved`: Enable driver operation permissions
  - `OrganizationCreated`: Setup owner permissions

### 6. Permission Delegation & Temporary Access

#### Delegation Mechanisms
- **Manager Delegation**
  - Temporary authority transfer
  - Limited scope delegation
  - Time-bound access
  - Approval workflows
  - Delegation audit trails

- **Emergency Access**
  - Break-glass procedures
  - Emergency escalation
  - Temporary super-user access
  - Incident response permissions
  - Post-incident reviews

#### Guest and Contractor Access
- **Limited Access Accounts**
  - Contractor permissions
  - Vendor access
  - Consultant rights
  - Audit firm access
  - External partner permissions

- **API Access Management**
  - Third-party integrations
  - Partner API access
  - Developer sandbox rights
  - Rate-limited access
  - Scope-restricted permissions

## Admin Dashboard Features

### 1. Role Management
- **Role Creation and Editing**
  - Visual role builder
  - Permission assignment interface
  - Role hierarchy visualization
  - Permission inheritance mapping
  - Role template library

- **Role Analytics**
  - Role usage statistics
  - Permission utilization metrics
  - Role effectiveness analysis
  - Access pattern insights
  - Security compliance reporting

### 2. User Permission Management
- **User Permission Overview**
  - Individual user permissions
  - Role assignment history
  - Permission conflict resolution
  - Access request management
  - Permission audit trails

- **Bulk Permission Operations**
  - Mass role assignment
  - Bulk permission changes
  - Group permission updates
  - Template-based assignments
  - Automated permission workflows

### 3. Access Control Monitoring
- **Real-Time Monitoring**
  - Active user sessions
  - Permission violations
  - Unauthorized access attempts
  - Suspicious activity alerts
  - Real-time permission changes

- **Security Analytics**
  - Access pattern analysis
  - Permission abuse detection
  - Privilege escalation monitoring
  - Compliance violation tracking
  - Risk assessment reporting

### 4. Audit & Compliance
- **Audit Trail Management**
  - Permission change logs
  - Access attempt records
  - Administrative action history
  - System modification tracking
  - User activity monitoring

- **Compliance Reporting**
  - Regulatory compliance reports
  - Access certification workflows
  - Permission review processes
  - Violation documentation
  - Remediation tracking

## API Endpoints

### Automatic Permission Management
```
GET /permissions/subscription-templates/{planId}
POST /permissions/apply-subscription-permissions
PUT /permissions/update-subscription-permissions
DELETE /permissions/revoke-subscription-permissions
GET /permissions/plan-permissions/{planId}
POST /permissions/sync-subscription-permissions
```

### Permission Templates
```
GET /permissions/templates
POST /permissions/templates
GET /permissions/templates/{id}
PUT /permissions/templates/{id}
DELETE /permissions/templates/{id}
POST /permissions/templates/{id}/apply/{userId}
```

### Role Management
```
GET /permissions/roles
POST /permissions/roles
GET /permissions/roles/{id}
PUT /permissions/roles/{id}
DELETE /permissions/roles/{id}
GET /permissions/roles/{id}/permissions
PUT /permissions/roles/{id}/permissions
```

### User Permission Management
```
GET /permissions/users/{userId}
PUT /permissions/users/{userId}/roles
POST /permissions/users/{userId}/permissions
DELETE /permissions/users/{userId}/permissions/{permissionId}
GET /permissions/users/{userId}/audit
```

### Permission Validation
```
POST /permissions/validate
POST /permissions/check
GET /permissions/user/{userId}/can/{action}/{resource}
POST /permissions/bulk-check
GET /permissions/effective/{userId}
```

### Access Control
```
POST /permissions/request-access
GET /permissions/requests
POST /permissions/approve-request/{requestId}
POST /permissions/deny-request/{requestId}
GET /permissions/delegations
POST /permissions/delegate
```

## Database Schema

### Subscription Permission Templates Table
```sql
CREATE TABLE subscription_permission_templates (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL, -- References subscription_plans
    service VARCHAR(50) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL, -- Array of permission IDs and configurations
    roles JSONB, -- Roles to auto-assign
    conditions JSONB, -- Additional conditions for activation
    priority INTEGER DEFAULT 0, -- For conflict resolution
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_plan_id (plan_id),
    INDEX idx_service (service),
    INDEX idx_is_active (is_active),
    UNIQUE KEY unique_plan_service (plan_id, service)
);
```

### Subscription Permissions Table
```sql
CREATE TABLE subscription_permissions (
    id UUID PRIMARY KEY,
    subscription_id UUID NOT NULL, -- References user_subscriptions or organization_subscriptions
    user_id UUID NOT NULL,
    template_id UUID REFERENCES subscription_permission_templates(id),
    permissions JSONB NOT NULL, -- Actual permissions granted
    roles JSONB, -- Roles granted through subscription
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP, -- NULL for permanent, set for trials
    status ENUM('active', 'suspended', 'expired', 'grace_period') DEFAULT 'active',
    metadata JSONB, -- Additional tracking data

    INDEX idx_subscription_id (subscription_id),
    INDEX idx_user_id (user_id),
    INDEX idx_template_id (template_id),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at),
    UNIQUE KEY unique_subscription_user (subscription_id, user_id)
);
```

### Permission Auto-Assignment Log Table
```sql
CREATE TABLE permission_auto_assignment_log (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    subscription_id UUID,
    plan_id UUID,
    trigger_event VARCHAR(50) NOT NULL, -- subscription_created, plan_upgraded, etc.
    permissions_added JSONB,
    permissions_removed JSONB,
    roles_added JSONB,
    roles_removed JSONB,
    reason TEXT,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_user_id (user_id),
    INDEX idx_subscription_id (subscription_id),
    INDEX idx_trigger_event (trigger_event),
    INDEX idx_success (success),
    INDEX idx_created_at (created_at)
);
```

### Plan Permission Mappings Table
```sql
CREATE TABLE plan_permission_mappings (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL,
    permission_id UUID REFERENCES permissions(id),
    role_id UUID REFERENCES roles(id),
    feature_key VARCHAR(100), -- For feature-based permissions
    resource_limits JSONB, -- Max items, API calls, etc.
    is_included BOOLEAN DEFAULT TRUE,
    conditions JSONB, -- Additional activation conditions
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_plan_id (plan_id),
    INDEX idx_permission_id (permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_feature_key (feature_key),
    UNIQUE KEY unique_plan_permission (plan_id, permission_id),
    UNIQUE KEY unique_plan_role (plan_id, role_id)
);
```

### Roles Table
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    service VARCHAR(50), -- NULL for platform-wide roles
    parent_role_id UUID REFERENCES roles(id),
    level INTEGER NOT NULL, -- Hierarchy level
    is_system_role BOOLEAN DEFAULT FALSE,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_name (name),
    INDEX idx_service (service),
    INDEX idx_parent_role_id (parent_role_id),
    INDEX idx_level (level),
    INDEX idx_is_system_role (is_system_role)
);
```

### Permissions Table
```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    service VARCHAR(50), -- Service this permission belongs to
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    conditions JSONB, -- Additional conditions for permission
    is_system_permission BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_name (name),
    INDEX idx_service (service),
    INDEX idx_resource (resource),
    INDEX idx_action (action),
    INDEX idx_is_system_permission (is_system_permission),
    UNIQUE KEY unique_service_resource_action (service, resource, action)
);
```

### Role Permissions Table
```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY,
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    granted BOOLEAN DEFAULT TRUE, -- FALSE for explicit denial
    conditions JSONB, -- Role-specific conditions
    inherited BOOLEAN DEFAULT FALSE, -- TRUE if inherited from parent role
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id),
    INDEX idx_granted (granted),
    INDEX idx_inherited (inherited),
    UNIQUE KEY unique_role_permission (role_id, permission_id)
);
```

### User Roles Table
```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    role_id UUID REFERENCES roles(id),
    granted_by UUID, -- User who granted this role
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP, -- NULL for permanent assignment
    context JSONB, -- Additional context (location, conditions, etc.)
    is_active BOOLEAN DEFAULT TRUE,

    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id),
    INDEX idx_granted_by (granted_by),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active),
    UNIQUE KEY unique_active_user_role (user_id, role_id, is_active) -- Prevent duplicate active assignments
);
```

### User Permissions Table (Direct Assignments)
```sql
CREATE TABLE user_permissions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    permission_id UUID REFERENCES permissions(id),
    granted BOOLEAN DEFAULT TRUE,
    granted_by UUID,
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    context JSONB,
    is_active BOOLEAN DEFAULT TRUE,

    INDEX idx_user_id (user_id),
    INDEX idx_permission_id (permission_id),
    INDEX idx_granted (granted),
    INDEX idx_granted_by (granted_by),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
);
```

### Permission Audit Log Table
```sql
CREATE TABLE permission_audit_log (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    action_type ENUM('role_assigned', 'role_removed', 'permission_granted', 'permission_revoked', 'access_granted', 'access_denied') NOT NULL,
    target_user_id UUID, -- For administrative actions
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    resource VARCHAR(100),
    resource_id VARCHAR(100),
    success BOOLEAN NOT NULL,
    failure_reason TEXT,
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_user_id (user_id),
    INDEX idx_action_type (action_type),
    INDEX idx_target_user_id (target_user_id),
    INDEX idx_resource (resource),
    INDEX idx_success (success),
    INDEX idx_created_at (created_at)
);
```

### Access Requests Table
```sql
CREATE TABLE access_requests (
    id UUID PRIMARY KEY,
    requester_id UUID NOT NULL,
    requested_role_id UUID REFERENCES roles(id),
    requested_permission_id UUID REFERENCES permissions(id),
    justification TEXT NOT NULL,
    urgency ENUM('low', 'medium', 'high', 'emergency') DEFAULT 'medium',
    status ENUM('pending', 'approved', 'denied', 'expired') DEFAULT 'pending',
    approver_id UUID,
    approved_at TIMESTAMP,
    expires_at TIMESTAMP,
    denial_reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_requester_id (requester_id),
    INDEX idx_status (status),
    INDEX idx_urgency (urgency),
    INDEX idx_approver_id (approver_id),
    INDEX idx_created_at (created_at)
);
```

### Permission Delegations Table
```sql
CREATE TABLE permission_delegations (
    id UUID PRIMARY KEY,
    delegator_id UUID NOT NULL,
    delegatee_id UUID NOT NULL,
    role_id UUID REFERENCES roles(id),
    permission_id UUID REFERENCES permissions(id),
    scope_restrictions JSONB, -- Limit delegation scope
    time_restrictions JSONB, -- Time-based limitations
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_delegator_id (delegator_id),
    INDEX idx_delegatee_id (delegatee_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
);
```

## Permission Validation Logic

### Effective Permissions Calculation
```go
func GetEffectivePermissions(userID string, service string) ([]Permission, error) {
    // 1. Get user's direct role assignments
    userRoles := getUserRoles(userID, service)

    // 2. Get permissions from roles (including inherited)
    rolePermissions := getRolePermissions(userRoles)

    // 3. Get direct user permissions
    directPermissions := getUserPermissions(userID, service)

    // 4. Get subscription-based permissions
    subscriptionPermissions := getSubscriptionPermissions(userID, service)

    // 5. Apply delegation permissions
    delegatedPermissions := getDelegatedPermissions(userID, service)

    // 6. Merge and resolve conflicts (explicit deny wins)
    effectivePermissions := mergePermissions(
        rolePermissions,
        directPermissions,
        subscriptionPermissions,
        delegatedPermissions,
    )

    return effectivePermissions, nil
}

// Get permissions based on active subscriptions
func getSubscriptionPermissions(userID string, service string) ([]Permission, error) {
    // 1. Get user's active subscriptions
    subscriptions := getUserSubscriptions(userID, service)

    // 2. Get organization subscriptions if user is member
    orgSubscriptions := getOrganizationSubscriptions(userID, service)

    // 3. Load permission templates for each subscription
    var permissions []Permission
    for _, sub := range append(subscriptions, orgSubscriptions...) {
        template := getPermissionTemplate(sub.PlanID, service)
        if template != nil {
            // 4. Check subscription status and apply permissions
            if sub.Status == "active" || sub.Status == "grace_period" {
                permissions = append(permissions, template.Permissions...)
            }
        }
    }

    return permissions, nil
}
```

### Permission Check Implementation
```go
func CheckPermission(userID, service, resource, action string, context map[string]interface{}) (bool, error) {
    // 1. Get effective permissions for user
    permissions, err := GetEffectivePermissions(userID, service)
    if err != nil {
        return false, err
    }

    // 2. Find matching permission
    for _, perm := range permissions {
        if perm.Resource == resource && perm.Action == action {
            // 3. Check conditions
            if evaluateConditions(perm.Conditions, context) {
                return perm.Granted, nil
            }
        }
    }

    // 4. Default deny
    return false, nil
}
```

## Automatic Permission Assignment Implementation

### Subscription Permission Flow
```go
// When a subscription is created or updated
func HandleSubscriptionEvent(event SubscriptionEvent) error {
    switch event.Type {
    case "subscription.created":
        return applySubscriptionPermissions(event.UserID, event.PlanID)
    case "subscription.upgraded":
        return upgradePermissions(event.UserID, event.OldPlanID, event.NewPlanID)
    case "subscription.downgraded":
        return downgradePermissions(event.UserID, event.OldPlanID, event.NewPlanID)
    case "subscription.cancelled":
        return schedulePermissionRevocation(event.UserID, event.PlanID, event.EndDate)
    case "subscription.expired":
        return revokeSubscriptionPermissions(event.UserID, event.PlanID)
    }
    return nil
}

// Apply permissions based on subscription plan
func applySubscriptionPermissions(userID, planID string) error {
    // 1. Get permission template for plan
    template, err := getPermissionTemplate(planID)
    if err != nil {
        return err
    }

    // 2. Create subscription permission record
    subPermission := &SubscriptionPermission{
        UserID:       userID,
        TemplateID:   template.ID,
        Permissions:  template.Permissions,
        Roles:        template.Roles,
        Status:       "active",
    }

    // 3. Apply permissions to user
    for _, permID := range template.Permissions {
        grantPermission(userID, permID, "subscription")
    }

    // 4. Assign roles
    for _, roleID := range template.Roles {
        assignRole(userID, roleID, "subscription")
    }

    // 5. Log the assignment
    logAutoAssignment(userID, planID, "subscription.created", template)

    return nil
}
```

### Permission Template Examples
```json
// Restaurant Professional Plan Template
{
  "plan_id": "restaurant_professional",
  "service": "food_delivery",
  "permissions": [
    "restaurant.menu.create",
    "restaurant.menu.update",
    "restaurant.menu.delete",
    "restaurant.orders.view",
    "restaurant.orders.process",
    "restaurant.analytics.advanced",
    "restaurant.marketing.campaigns",
    "restaurant.staff.manage",
    "restaurant.locations.multiple"
  ],
  "roles": ["restaurant_manager"],
  "resource_limits": {
    "max_menu_items": -1,
    "max_staff_accounts": 10,
    "max_locations": 3,
    "api_calls_per_hour": 1000
  }
}

// Hotel Premium Plan Template
{
  "plan_id": "hotel_premium",
  "service": "booking",
  "permissions": [
    "hotel.rooms.manage",
    "hotel.bookings.manage",
    "hotel.pricing.dynamic",
    "hotel.channels.integrate",
    "hotel.analytics.revenue",
    "hotel.reports.export",
    "hotel.api.access"
  ],
  "roles": ["hotel_manager"],
  "resource_limits": {
    "max_room_types": -1,
    "max_channels": 10,
    "api_calls_per_hour": 5000
  }
}
```

### Organization Subscription Permission Distribution
```go
// Distribute permissions to organization members based on subscription
func distributeOrganizationPermissions(orgID string, planID string) error {
    // 1. Get organization members
    members, err := getOrganizationMembers(orgID)
    if err != nil {
        return err
    }

    // 2. Get plan template
    template, err := getPermissionTemplate(planID)
    if err != nil {
        return err
    }

    // 3. Apply permissions based on member role
    for _, member := range members {
        var permissions []string
        var roles []string

        switch member.Role {
        case "owner":
            // Owners get all plan permissions
            permissions = template.Permissions
            roles = append(template.Roles, "organization_owner")
        case "admin":
            // Admins get most permissions except billing
            permissions = filterPermissions(template.Permissions, "exclude_billing")
            roles = template.Roles
        case "manager":
            // Managers get operational permissions
            permissions = filterPermissions(template.Permissions, "operational_only")
            roles = []string{"organization_manager"}
        case "employee":
            // Employees get basic permissions
            permissions = filterPermissions(template.Permissions, "basic_only")
            roles = []string{"organization_employee"}
        }

        // Apply permissions to member
        applyUserPermissions(member.UserID, permissions, roles, "organization_subscription")
    }

    return nil
}
```

## Integration Patterns

### With Core Service
- User authentication integration
- Role-based dashboard rendering
- Permission-aware API endpoints
- Admin user management
- Subscription status monitoring
- Automatic permission synchronization

### With All Services
```go
// Middleware for permission checking
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := getUserIDFromToken(c)
        service := getServiceFromPath(c)
        resource := getResourceFromPath(c)
        action := getActionFromMethod(c)

        hasPermission, err := permissionEngine.CheckPermission(userID, service, resource, action, nil)
        if err != nil || !hasPermission {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// Service-specific permission validation
func ValidateRestaurantAccess(userID, restaurantID string) bool {
    context := map[string]interface{}{
        "restaurant_id": restaurantID,
    }

    return permissionEngine.CheckPermission(userID, "delivery", "restaurant", "manage", context)
}
```

### Event Publishing
- `RoleAssigned`: User assigned to role
- `PermissionGranted`: Direct permission granted
- `AccessDenied`: Access attempt denied
- `PermissionRevoked`: Permission removed
- `DelegationCreated`: Permission delegated
- `SubscriptionPermissionsApplied`: Permissions granted via subscription
- `SubscriptionPermissionsRevoked`: Permissions removed due to subscription change
- `PermissionTemplateUpdated`: Template changes requiring sync

### Event Consumption
- `UserRegistered`: Assign default roles
- `BusinessVerified`: Grant business permissions
- `SubscriptionCreated`: Apply plan permission template
- `SubscriptionUpgraded`: Add new tier permissions
- `SubscriptionDowngraded`: Adjust permissions to lower tier
- `SubscriptionCancelled`: Schedule permission removal
- `SubscriptionExpired`: Remove subscription permissions
- `SubscriptionRenewed`: Restore plan permissions
- `PaymentFailed`: Apply grace period permissions
- `TrialStarted`: Grant trial permissions with expiry
- `TrialConverted`: Convert to full plan permissions
- `OrganizationMemberAdded`: Apply org subscription permissions
- `OrganizationMemberRemoved`: Revoke org permissions
- `UserSuspended`: Revoke all permissions

## Security Considerations

### Permission Security
- **Principle of Least Privilege**: Users get minimum required permissions
- **Defense in Depth**: Multiple permission checks at different layers
- **Explicit Deny**: Denial permissions override grants
- **Time-Based Access**: Automatic permission expiration
- **Regular Audits**: Periodic permission reviews

### Data Protection
- **Permission Encryption**: Sensitive permission data encryption
- **Audit Trail Protection**: Tamper-proof audit logs
- **Access Logging**: Comprehensive access attempt logging
- **Privacy Controls**: User consent for data access permissions

## Performance Requirements

### Permission Validation
- **Check Latency**: <10ms for permission checks
- **Cache Hit Ratio**: >95% for frequently checked permissions
- **Concurrent Checks**: Support 100,000+ simultaneous permission checks
- **Database Performance**: <5ms for permission queries

### Scalability
- **User Scale**: Support 1M+ users with permissions
- **Role Scale**: Handle 10,000+ roles across services
- **Permission Scale**: Manage 100,000+ individual permissions
- **Audit Scale**: Process 1M+ audit events per day
