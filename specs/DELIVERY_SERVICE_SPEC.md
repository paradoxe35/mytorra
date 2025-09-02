# Delivery Service Specification

## Overview
The Delivery Service manages TorraFood (restaurant ordering and delivery) and TorraDelivery (package/parcel delivery). It coordinates a shared pool of delivery agents, handles route optimization, and manages the complete order fulfillment lifecycle.

## Service Responsibilities

### Primary Functions
1. **Restaurant Organization Management** - Multi-location restaurant chain management
2. **Menu Management** - Digital menu creation and management
3. **Order Processing** - Food order lifecycle management
4. **Delivery Agent Management** - Shared pool of delivery personnel
5. **Route Optimization** - Efficient delivery routing and scheduling
6. **Package Delivery** - Parcel and package delivery services

## Features & Functionality

### 1. MyTorra Food (Restaurant Platform)

#### Restaurant Organization Types
- **Independent Restaurants**
  - Single location operations
  - Owner-operated establishments
  - Individual subscription plans
  - Simple management structure

- **Restaurant Chains/Organizations**
  - Multi-location operations
  - Centralized management
  - Hierarchical staff structure
  - Organization-level subscriptions
  - Shared branding and menus

#### Restaurant Organization Management
- **Chain Registration**
  - Corporate entity verification
  - Multi-location license validation
  - Brand management setup
  - Operational area definition
  - Regional management structure

- **Location Management**
  - Individual restaurant registration
  - Location-specific settings
  - Staff assignment per location
  - Local menu customization
  - Performance tracking per location

#### Restaurant Organization Roles
- **Restaurant Chain Owner**
  - Complete chain oversight
  - Strategic planning and expansion
  - Brand management
  - Financial oversight
  - Franchise management

- **Regional Manager**
  - Multi-location management
  - Regional performance optimization
  - Staff coordination across locations
  - Regional marketing campaigns
  - Compliance oversight

- **Restaurant Manager**
  - Single location operations
  - Local staff management
  - Daily operations oversight
  - Customer service management
  - Local performance metrics

- **Kitchen Manager**
  - Kitchen operations oversight
  - Staff scheduling and training
  - Quality control
  - Inventory management
  - Food safety compliance

- **Staff Members**
  - Order processing
  - Kitchen operations
  - Customer service
  - Shift-based access
  - Location-specific permissions

#### Restaurant Onboarding
- **Business Registration**
  - Restaurant license verification
  - Health permit validation
  - Business address confirmation
  - Bank account verification
  - Tax registration details

- **Restaurant Profile Setup**
  - Restaurant information and branding
  - Operating hours and days
  - Delivery zones definition
  - Service fees configuration
  - Photo gallery management

#### Menu Management System
- **Digital Menu Creation**
  - Category-based organization
  - Item descriptions and photos
  - Nutritional information
  - Allergen warnings
  - Availability scheduling

- **Pricing and Variants**
  - Multiple size options
  - Customization choices (toppings, sides)
  - Combo meal creation
  - Promotional pricing
  - Happy hour specials

- **Inventory Integration**
  - Real-time availability updates
  - Ingredient tracking
  - Auto-disable out-of-stock items
  - Estimated preparation times

### 2. Food Order Processing

#### Customer Order Flow
- **Restaurant Discovery**
  - Location-based restaurant search
  - Cuisine type filtering
  - Rating and review filtering
  - Delivery time estimates
  - Minimum order requirements

- **Order Customization**
  - Item customization options
  - Special instructions
  - Dietary preference indicators
  - Group ordering functionality
  - Scheduled ordering

#### Order Lifecycle Management
```
placed -> accepted -> preparing -> ready -> picked_up -> in_transit -> delivered -> completed
      \-> rejected -> cancelled -> refunded
```

- **Real-time Updates**
  - Order status notifications
  - Preparation time estimates
  - Delivery tracking
  - ETA updates

### 3. TorraDelivery (Package Delivery)

#### Package Delivery Services
- **Same-Day Delivery**
  - Urgent package delivery
  - Document courier services
  - Small parcel delivery
  - Fragile item handling

- **Scheduled Delivery**
  - Next-day delivery
  - Multi-day scheduling
  - Bulk delivery services
  - Recurring delivery setups

#### Package Management
- **Package Information**
  - Size and weight specifications
  - Fragile/special handling flags
  - Insurance options
  - Signature requirements
  - Photo confirmation

### 4. Delivery Agent Management

#### Agent Onboarding
- **Driver Verification**
  - Background checks
  - Vehicle registration
  - Insurance verification
  - License validation
  - Equipment requirements

- **Agent Profile**
  - Personal information
  - Vehicle details
  - Service area definition
  - Availability schedule
  - Performance metrics

#### Agent Assignment & Routing
- **Smart Assignment Algorithm**
  - Distance optimization
  - Agent availability
  - Vehicle type matching
  - Load capacity consideration
  - Historical performance

- **Route Optimization**
  - Multi-drop routing
  - Traffic condition integration
  - Real-time route adjustments
  - Fuel-efficient paths
  - Time window optimization

### 5. Delivery Tracking & Analytics

#### Real-time Tracking
- **GPS Integration**
  - Live location updates
  - Delivery progress tracking
  - Estimated arrival times
  - Route deviation alerts

- **Customer Communication**
  - SMS/push notifications
  - Real-time chat with driver
  - Delivery photos
  - Customer rating system

## Restaurant Dashboard Features

### 1. Order Management
- **Live Order Dashboard**
  - Incoming order notifications
  - Order acceptance/rejection
  - Preparation time management
  - Special request handling

- **Order History**
  - Historical order data
  - Customer information
  - Revenue tracking
  - Popular items analysis

### 2. Menu Management
- **Menu Editing Tools**
  - Real-time menu updates
  - Item availability toggle
  - Price adjustments
  - Photo management

- **Analytics Dashboard**
  - Sales performance metrics
  - Popular items tracking
  - Peak hours analysis
  - Customer feedback review

### 3. Financial Management
- **Revenue Tracking**
  - Daily/weekly/monthly revenue
  - Commission breakdown
  - Payout schedules
  - Tax reporting

- **Promotional Tools**
  - Discount creation
  - Loyalty program integration
  - Marketing campaign management

## Delivery Agent Dashboard Features

### 1. Job Management
- **Available Jobs**
  - Job notifications
  - Job acceptance/rejection
  - Route preview
  - Earnings estimation

- **Active Deliveries**
  - Current delivery status
  - Customer contact information
  - Navigation integration
  - Delivery confirmation tools

### 2. Performance Tracking
- **Earnings Dashboard**
  - Daily/weekly earnings
  - Commission structure
  - Bonus opportunities
  - Payout schedules

- **Performance Metrics**
  - Delivery completion rate
  - Customer ratings
  - On-time delivery rate
  - Distance traveled

## Admin Dashboard Features

### 1. Restaurant Administration
- **Restaurant Management**
  - Restaurant approval process
  - Performance monitoring
  - Subscription management
  - Quality control

- **Menu Moderation**
  - Menu item approval
  - Content moderation
  - Pricing oversight
  - Compliance monitoring

### 2. Delivery Operations
- **Agent Management**
  - Agent verification
  - Performance monitoring
  - Dispute resolution
  - Territory management

- **Delivery Analytics**
  - Delivery performance metrics
  - Route efficiency analysis
  - Customer satisfaction tracking
  - Operational cost analysis

### 3. Order Administration
- **Order Monitoring**
  - Real-time order tracking
  - Dispute resolution
  - Refund processing
  - Quality assurance

## API Endpoints

### Restaurant Management
```
POST /restaurants/register
GET /restaurants/profile
PUT /restaurants/profile
GET /restaurants/analytics
PUT /restaurants/status
```

### Menu Management
```
GET /restaurants/{id}/menu
POST /restaurants/{id}/menu/items
PUT /restaurants/{id}/menu/items/{itemId}
DELETE /restaurants/{id}/menu/items/{itemId}
PUT /restaurants/{id}/menu/items/{itemId}/availability
```

### Order Management
```
POST /orders
GET /orders
GET /orders/{id}
PUT /orders/{id}/status
POST /orders/{id}/accept
POST /orders/{id}/reject
GET /orders/{id}/tracking
```

### Delivery Agent
```
POST /agents/register
GET /agents/profile
PUT /agents/status
GET /agents/available-jobs
POST /agents/jobs/{id}/accept
PUT /agents/jobs/{id}/status
```

### Package Delivery
```
POST /packages
GET /packages/{id}
PUT /packages/{id}/status
GET /packages/{id}/tracking
POST /packages/{id}/confirm-delivery
```

## Database Schema

### Restaurants Table
```sql
CREATE TABLE restaurants (
    id UUID PRIMARY KEY,
    organization_id UUID, -- References Core Service organizations (NULL for independent)
    owner_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    cuisine_type VARCHAR(100),
    address JSONB NOT NULL,
    location geometry(POINT, 4326), -- PostGIS geometry using gogis
    phone VARCHAR(20),
    email VARCHAR(255),
    business_license VARCHAR(100),
    tax_id VARCHAR(50),
    operating_hours JSONB,
    delivery_zones JSONB,
    min_order_amount DECIMAL(8,2),
    delivery_fee DECIMAL(8,2),
    commission_rate DECIMAL(5,2),
    status ENUM('pending', 'active', 'suspended', 'closed') DEFAULT 'pending',
    rating DECIMAL(3,2),
    total_orders INTEGER DEFAULT 0,
    images JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_organization_id (organization_id),
    INDEX idx_owner_id (owner_id),
    INDEX idx_status (status),
    INDEX idx_cuisine_type (cuisine_type),
    SPATIAL INDEX idx_location (location)
);
```

### Menu Items Table
```sql
CREATE TABLE menu_items (
    id UUID PRIMARY KEY,
    restaurant_id UUID REFERENCES restaurants(id),
    category_id UUID REFERENCES menu_categories(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(8,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    preparation_time INTEGER, -- minutes
    calories INTEGER,
    ingredients TEXT,
    allergens JSONB,
    images JSONB,
    variants JSONB, -- sizes, customizations
    available BOOLEAN DEFAULT TRUE,
    popular BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_restaurant_id (restaurant_id),
    INDEX idx_category_id (category_id),
    INDEX idx_available (available),
    FULLTEXT idx_search (name, description)
);
```

### Food Orders Table
```sql
CREATE TABLE food_orders (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    restaurant_id UUID REFERENCES restaurants(id),
    agent_id UUID NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    delivery_fee DECIMAL(8,2),
    tax_amount DECIMAL(8,2),
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('placed', 'accepted', 'preparing', 'ready', 'picked_up', 'in_transit', 'delivered', 'completed', 'rejected', 'cancelled', 'refunded') DEFAULT 'placed',
    delivery_address JSONB NOT NULL,
    special_instructions TEXT,
    estimated_delivery_time TIMESTAMP,
    actual_delivery_time TIMESTAMP,
    payment_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_customer_id (customer_id),
    INDEX idx_restaurant_id (restaurant_id),
    INDEX idx_agent_id (agent_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Order Items Table
```sql
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES food_orders(id),
    menu_item_id UUID REFERENCES menu_items(id),
    quantity INTEGER NOT NULL,
    unit_price DECIMAL(8,2) NOT NULL,
    customizations JSONB,
    special_instructions TEXT,

    INDEX idx_order_id (order_id),
    INDEX idx_menu_item_id (menu_item_id)
);
```

### Delivery Agents Table
```sql
CREATE TABLE delivery_agents (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(255),
    vehicle_type ENUM('bicycle', 'motorcycle', 'car', 'scooter') NOT NULL,
    vehicle_registration VARCHAR(50),
    license_number VARCHAR(50),
    insurance_details JSONB,
    service_areas JSONB,
    status ENUM('offline', 'online', 'busy', 'suspended') DEFAULT 'offline',
    current_location geometry(POINT, 4326), -- PostGIS geometry using gogis
    rating DECIMAL(3,2),
    total_deliveries INTEGER DEFAULT 0,
    earnings DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_vehicle_type (vehicle_type),
    SPATIAL INDEX idx_current_location (current_location)
);
```

### Package Deliveries Table
```sql
CREATE TABLE package_deliveries (
    id UUID PRIMARY KEY,
    sender_id UUID NOT NULL,
    recipient_name VARCHAR(200) NOT NULL,
    recipient_phone VARCHAR(20) NOT NULL,
    pickup_address JSONB NOT NULL,
    delivery_address JSONB NOT NULL,
    package_details JSONB NOT NULL,
    agent_id UUID NULL,
    delivery_type ENUM('same_day', 'next_day', 'scheduled') DEFAULT 'same_day',
    status ENUM('pending', 'assigned', 'picked_up', 'in_transit', 'delivered', 'completed', 'cancelled') DEFAULT 'pending',
    scheduled_pickup TIMESTAMP,
    scheduled_delivery TIMESTAMP,
    actual_pickup TIMESTAMP,
    actual_delivery TIMESTAMP,
    total_cost DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_id VARCHAR(100),
    tracking_number VARCHAR(50) UNIQUE,
    delivery_instructions TEXT,
    proof_of_delivery JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_sender_id (sender_id),
    INDEX idx_agent_id (agent_id),
    INDEX idx_status (status),
    INDEX idx_tracking_number (tracking_number),
    INDEX idx_scheduled_pickup (scheduled_pickup)
);
```

## Integration Patterns

### With Core Service
- Restaurant owner authentication
- Delivery agent authentication
- Payment processing coordination
- Subscription validation for restaurants

### With Marketplace Service
- Product delivery integration
- Seller shipping options
- Order fulfillment coordination

### With TorraCab Service
- Shared delivery agent pool
- Cross-service agent availability
- Unified location tracking

### With Realtime Service
- Live delivery tracking
- Order status updates
- Agent location broadcasting

### Event Publishing
- `OrderPlaced`: New food order created
- `OrderStatusUpdated`: Order status changes
- `DeliveryCompleted`: Delivery finished
- `AgentLocationUpdated`: Agent location changes

### Event Consumption
- `PaymentProcessed`: Activate confirmed orders
- `UserSuspended`: Suspend restaurant/agent privileges
- `SubscriptionExpired`: Restrict restaurant features
