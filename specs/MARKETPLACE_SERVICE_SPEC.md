# Marketplace Service Specification

## Overview
The Marketplace Service manages all e-commerce activities including MyTorra (used items), MyTorraShop (new items), MyTorraBoutique (fashion), and the Universal Want/Request System. It serves as the heart of buying and selling on the platform.

## Service Responsibilities

### Primary Functions
1. **Product Management** - Catalog management for all marketplace types
2. **Order Processing** - Order lifecycle management
3. **Search Integration** - Product search and discovery
4. **Want/Request System** - Universal product request system
5. **Seller Management** - Vendor onboarding and management
6. **Review System** - Product and seller reviews

## Features & Functionality

### 1. Multi-Marketplace Management

#### MyTorra (Used Items P2P)
- **Individual Seller Support**
  - Personal account product listings
  - Photo-based product descriptions
  - Condition assessment tools
  - Price suggestion algorithms
  - Direct buyer-seller messaging

- **Product Categories**
  - Electronics & Gadgets
  - Vehicles & Automotive
  - Home & Garden
  - Fashion & Accessories
  - Books & Media
  - Sports & Recreation

#### MyTorraShop (New Items Business)
- **Business Account Support**
  - Bulk product management
  - Inventory tracking
  - Business verification process
  - Advanced analytics dashboard

- **Business Features**
  - Product variants management
  - Bulk pricing rules
  - Inventory alerts
  - Sales reporting

#### MyTorraBoutique (Fashion & Lifestyle)
- **Specialized Fashion Features**
  - Size guide integration
  - Color/pattern matching
  - Style recommendations
  - Fashion trend analytics
  - Virtual try-on integration

### 2. Product Management

#### Product Listing Features
- **Rich Media Support**
  - Multiple photo uploads
  - Video product demos
  - 360-degree product views
  - Augmented reality preview

- **Product Information**
  - Detailed descriptions
  - Specifications and attributes
  - Condition assessment (for used items)
  - Authenticity verification
  - Price comparison tools

#### Inventory Management
- **Stock Tracking**
  - Real-time inventory updates
  - Low stock alerts
  - Automatic delisting when sold
  - Batch inventory uploads

- **Pricing Management**
  - Dynamic pricing suggestions
  - Promotional pricing
  - Bulk discount management
  - Currency conversion

### 3. Universal Want/Request System

#### Request Features
- **Product Requests**
  - Detailed want descriptions
  - Photo-based requests
  - Budget specification
  - Location preferences
  - Urgency levels

- **Matching Algorithm**
  - Automatic request-product matching
  - Seller notification system
  - Custom offer generation
  - Negotiation facilitation

#### Seller Responses
- **Custom Offers**
  - Direct response to requests
  - Custom pricing options
  - Availability confirmation
  - Delivery options

### 4. Order Management

#### Order Lifecycle
- **Order Creation**
  - Cart management
  - Multiple seller orders
  - Shipping calculation
  - Tax computation
  - Payment processing integration

- **Order Processing**
  - Seller notification
  - Inventory deduction
  - Shipping label generation
  - Tracking integration
  - Delivery confirmation

#### Order States
```
pending -> confirmed -> processing -> shipped -> delivered -> completed
                   \-> cancelled -> refunded
```

### 5. Search & Discovery

#### Advanced Search Features
- **Multi-criteria Search**
  - Text, image, and voice search
  - Filter by price, location, condition
  - Category and brand filtering
  - Availability filtering

- **Personalized Discovery**
  - Recommendation engine
  - Recently viewed items
  - Wish list management
  - Price drop alerts

## Seller Dashboard Features

### 1. Product Management
- **Listing Management**
  - Add/edit/delete products
  - Bulk product operations
  - Performance analytics per product
  - Inventory tracking

- **Order Management**
  - Order notifications
  - Order processing tools
  - Shipping management
  - Customer communication

### 2. Business Analytics
- **Sales Metrics**
  - Revenue tracking
  - Product performance
  - Customer demographics
  - Seasonal trends

- **Financial Management**
  - Earnings overview
  - Payout schedules
  - Transaction history
  - Tax reporting tools

### 3. Customer Relations
- **Review Management**
  - Review monitoring
  - Response management
  - Quality improvement insights

- **Customer Communication**
  - Direct messaging
  - Bulk customer notifications
  - Customer service tools

## Admin Dashboard Features

### 1. Marketplace Oversight
- **Product Moderation**
  - New listing approval
  - Content moderation
  - Quality control
  - Authenticity verification

- **Seller Management**
  - Seller verification
  - Performance monitoring
  - Suspension management
  - Commission management

### 2. Order Administration
- **Dispute Resolution**
  - Order dispute handling
  - Refund processing
  - Mediation tools
  - Resolution tracking

- **Platform Analytics**
  - Marketplace performance
  - Revenue analytics
  - User behavior insights
  - Market trends

## API Endpoints

### Product Management
```
GET /products
POST /products
GET /products/{id}
PUT /products/{id}
DELETE /products/{id}
POST /products/{id}/images
GET /products/search
GET /products/categories
```

### Order Management
```
POST /orders
GET /orders
GET /orders/{id}
PUT /orders/{id}/status
POST /orders/{id}/cancel
GET /orders/{id}/tracking
POST /orders/{id}/review
```

### Want/Request System
```
POST /wants
GET /wants
GET /wants/{id}
POST /wants/{id}/offers
GET /wants/{id}/offers
PUT /wants/{id}/offers/{offerId}/accept
```

### Seller Dashboard
```
GET /seller/products
GET /seller/orders
GET /seller/analytics
GET /seller/earnings
PUT /seller/profile
```

### Admin Endpoints
```
GET /admin/products/pending
POST /admin/products/{id}/approve
GET /admin/sellers
GET /admin/orders/disputes
POST /admin/orders/{id}/resolve
```

## Database Schema

### Products Table
```sql
CREATE TABLE products (
    id UUID PRIMARY KEY,
    seller_id UUID NOT NULL,
    marketplace_type ENUM('mytorra', 'mytorrashop', 'mytorraboutique') NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    brand VARCHAR(100),
    condition ENUM('new', 'like_new', 'good', 'fair', 'poor') DEFAULT 'new',
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    quantity INTEGER DEFAULT 1,
    status ENUM('draft', 'active', 'sold', 'suspended') DEFAULT 'draft',
    location JSONB,
    attributes JSONB,
    images JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_seller_id (seller_id),
    INDEX idx_category (category_id),
    INDEX idx_status (status),
    INDEX idx_price (price),
    FULLTEXT idx_search (title, description)
);
```

### Orders Table
```sql
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    buyer_id UUID NOT NULL,
    seller_id UUID NOT NULL,
    product_id UUID REFERENCES products(id),
    quantity INTEGER DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status ENUM('pending', 'confirmed', 'processing', 'shipped', 'delivered', 'completed', 'cancelled', 'refunded') DEFAULT 'pending',
    shipping_address JSONB,
    tracking_number VARCHAR(100),
    payment_id VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_buyer_id (buyer_id),
    INDEX idx_seller_id (seller_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Want/Request System Tables
```sql
CREATE TABLE wants (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    max_price DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    location JSONB,
    urgency ENUM('low', 'medium', 'high') DEFAULT 'medium',
    images JSONB,
    status ENUM('active', 'fulfilled', 'expired') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_category (category_id),
    INDEX idx_status (status),
    FULLTEXT idx_search (title, description)
);

CREATE TABLE want_offers (
    id UUID PRIMARY KEY,
    want_id UUID REFERENCES wants(id),
    seller_id UUID NOT NULL,
    product_id UUID REFERENCES products(id),
    offer_price DECIMAL(10,2) NOT NULL,
    message TEXT,
    status ENUM('pending', 'accepted', 'rejected', 'expired') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_want_id (want_id),
    INDEX idx_seller_id (seller_id),
    INDEX idx_status (status)
);
```

### Reviews Table
```sql
CREATE TABLE reviews (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id),
    reviewer_id UUID NOT NULL,
    reviewee_id UUID NOT NULL,
    product_id UUID REFERENCES products(id),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    images JSONB,
    helpful_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_reviewee_id (reviewee_id),
    INDEX idx_product_id (product_id),
    INDEX idx_rating (rating),
    UNIQUE KEY unique_order_review (order_id, reviewer_id)
);
```

## Integration Patterns

### With Core Service
- User authentication and authorization
- Subscription validation for sellers
- Payment processing coordination
- Notification triggers

### With Media Service
- Product image/video uploads
- Image processing and optimization
- Content moderation integration

### With Delivery Service
- Shipping option integration
- Delivery tracking coordination
- Location-based delivery

### Event Publishing
- `ProductListed`: New product created
- `OrderPlaced`: Order created
- `OrderCompleted`: Order fulfilled
- `ReviewSubmitted`: Product/seller review

### Event Consumption
- `UserSuspended`: Hide seller products
- `PaymentFailed`: Cancel pending orders
- `SubscriptionExpired`: Suspend seller privileges
