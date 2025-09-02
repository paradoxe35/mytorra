# Booking Service Specification

## Overview
The Booking Service manages TorraBooking (hotels, accommodations) and TorraFlight (flight bookings). It handles availability management, reservation lifecycle, and booking confirmations across various accommodation and travel providers.

## Service Responsibilities

### Primary Functions
1. **Accommodation Management** - Hotel, resort, and property management
2. **Room Inventory Management** - Real-time availability and pricing
3. **Reservation Processing** - Booking lifecycle management
4. **Flight Integration** - Flight search and booking capabilities
5. **Property Owner Dashboard** - Tools for accommodation providers
6. **Guest Management** - Guest profiles and preferences

## Features & Functionality

### 1. TorraBooking (Accommodation Platform)

#### Property Management
- **Property Registration**
  - Hotel/property verification
  - Business license validation
  - Property address confirmation
  - Tax registration details
  - Insurance verification

- **Property Profile Setup**
  - Property information and branding
  - Amenities and facilities
  - Photo gallery management
  - Location and accessibility info
  - Policies and rules

#### Room & Rate Management
- **Room Type Management**
  - Room categories and descriptions
  - Occupancy limits
  - Bed configurations
  - Room amenities
  - Photo galleries per room type

- **Pricing and Availability**
  - Seasonal pricing strategies
  - Dynamic pricing algorithms
  - Promotional rates
  - Last-minute deals
  - Blackout dates management

- **Inventory Control**
  - Real-time availability updates
  - Overbooking protection
  - Allocation management
  - Minimum stay requirements
  - Advance booking limits

### 2. Reservation Management

#### Booking Flow
- **Property Discovery**
  - Location-based search
  - Date availability filtering
  - Price range filtering
  - Amenity-based filtering
  - Guest rating filtering

- **Booking Customization**
  - Room selection
  - Guest information collection
  - Special requests handling
  - Add-on services (meals, spa, etc.)
  - Group booking capabilities

#### Reservation Lifecycle
```
inquiry -> quote -> booked -> confirmed -> checked_in -> checked_out -> completed
        \-> cancelled -> refunded
```

- **Booking Confirmations**
  - Instant confirmation for available rooms
  - Pending confirmation for request-based bookings
  - Payment processing integration
  - Booking modification handling

### 3. TorraFlight (Flight Booking)

#### Flight Search Integration
- **GDS Integration**
  - Amadeus API integration
  - Sabre API integration
  - Travelport integration
  - Direct airline APIs

- **Search Capabilities**
  - Multi-city search
  - Flexible date search
  - Cabin class filtering
  - Airline preference
  - Price comparison

#### Flight Booking Features
- **Passenger Management**
  - Passenger information collection
  - Frequent flyer integration
  - Special assistance requests
  - Group booking capabilities

- **Ancillary Services**
  - Seat selection
  - Baggage upgrades
  - Meal preferences
  - Travel insurance
  - Airport transfers

### 4. Guest Experience Management

#### Guest Profiles
- **Personal Information**
  - Contact details
  - Preference tracking
  - Loyalty program integration
  - Special requirements
  - Travel history

- **Booking Management**
  - Current reservations
  - Booking history
  - Cancellation management
  - Modification requests
  - Review submissions

#### Communication Features
- **Automated Messaging**
  - Booking confirmations
  - Check-in reminders
  - Stay instructions
  - Post-stay follow-ups

- **Direct Communication**
  - Property messaging
  - Special request handling
  - Concierge services
  - Emergency contacts

## Property Owner Dashboard Features

### 1. Reservation Management
- **Live Reservation Dashboard**
  - Incoming booking notifications
  - Reservation acceptance/modification
  - Guest communication
  - Special request handling

- **Reservation Calendar**
  - Visual availability calendar
  - Booking overlays
  - Rate adjustments
  - Maintenance blocking

### 2. Property Management
- **Room Management**
  - Room status updates
  - Maintenance scheduling
  - Inventory adjustments
  - Rate management

- **Guest Services**
  - Check-in/check-out management
  - Guest request handling
  - Service coordination
  - Issue resolution

### 3. Business Analytics
- **Performance Metrics**
  - Occupancy rates
  - Average daily rate (ADR)
  - Revenue per available room (RevPAR)
  - Booking lead times

- **Financial Management**
  - Revenue tracking
  - Commission breakdown
  - Payout schedules
  - Tax reporting

### 4. Marketing Tools
- **Promotional Management**
  - Discount creation
  - Package deals
  - Seasonal promotions
  - Loyalty programs

- **Content Management**
  - Photo updates
  - Description editing
  - Amenity management
  - Policy updates

## Admin Dashboard Features

### 1. Property Administration
- **Property Oversight**
  - Property approval process
  - Quality monitoring
  - Compliance verification
  - Performance tracking

- **Content Moderation**
  - Photo approval
  - Description review
  - Policy compliance
  - Rating monitoring

### 2. Booking Administration
- **Reservation Monitoring**
  - Booking analytics
  - Dispute resolution
  - Refund processing
  - Fraud detection

- **System Management**
  - Rate monitoring
  - Availability audits
  - Integration health
  - Performance optimization

### 3. Financial Oversight
- **Revenue Management**
  - Commission tracking
  - Payout processing
  - Financial reporting
  - Tax compliance

## API Endpoints

### Property Management
```
POST /properties/register
GET /properties/profile
PUT /properties/profile
GET /properties/rooms
POST /properties/rooms
PUT /properties/rooms/{roomId}
DELETE /properties/rooms/{roomId}
```

### Availability & Pricing
```
GET /properties/{id}/availability
PUT /properties/{id}/availability
GET /properties/{id}/rates
PUT /properties/{id}/rates
POST /properties/{id}/block-dates
```

### Booking Management
```
POST /bookings
GET /bookings
GET /bookings/{id}
PUT /bookings/{id}
POST /bookings/{id}/cancel
POST /bookings/{id}/modify
GET /bookings/{id}/confirmation
```

### Flight Integration
```
GET /flights/search
POST /flights/book
GET /flights/booking/{id}
POST /flights/cancel/{id}
GET /flights/seat-map/{flightId}
```

### Property Search
```
GET /search/properties
GET /search/availability
GET /properties/{id}/details
GET /properties/{id}/reviews
```

## Database Schema

### Properties Table
```sql
CREATE TABLE properties (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    property_type ENUM('hotel', 'resort', 'apartment', 'villa', 'hostel', 'guesthouse') NOT NULL,
    address JSONB NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(255),
    website VARCHAR(500),
    business_license VARCHAR(100),
    tax_id VARCHAR(50),
    check_in_time TIME DEFAULT '15:00',
    check_out_time TIME DEFAULT '11:00',
    policies JSONB,
    amenities JSONB,
    images JSONB,
    rating DECIMAL(3,2),
    total_reviews INTEGER DEFAULT 0,
    status ENUM('pending', 'active', 'suspended', 'closed') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_owner_id (owner_id),
    INDEX idx_property_type (property_type),
    INDEX idx_status (status),
    INDEX idx_rating (rating),
    SPATIAL INDEX idx_location ((CAST(address->>'$.coordinates' AS POINT))),
    FULLTEXT idx_search (name, description)
);
```

### Room Types Table
```sql
CREATE TABLE room_types (
    id UUID PRIMARY KEY,
    property_id UUID REFERENCES properties(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    max_occupancy INTEGER NOT NULL,
    bed_configuration VARCHAR(200),
    size_sqm DECIMAL(6,2),
    amenities JSONB,
    images JSONB,
    base_price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    total_rooms INTEGER NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_property_id (property_id),
    INDEX idx_max_occupancy (max_occupancy),
    INDEX idx_base_price (base_price),
    INDEX idx_active (active)
);
```

### Room Inventory Table
```sql
CREATE TABLE room_inventory (
    id UUID PRIMARY KEY,
    room_type_id UUID REFERENCES room_types(id),
    date DATE NOT NULL,
    available_rooms INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    minimum_stay INTEGER DEFAULT 1,
    closed_to_arrival BOOLEAN DEFAULT FALSE,
    closed_to_departure BOOLEAN DEFAULT FALSE,

    UNIQUE KEY unique_room_date (room_type_id, date),
    INDEX idx_room_type_id (room_type_id),
    INDEX idx_date (date),
    INDEX idx_price (price)
);
```

### Bookings Table
```sql
CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    guest_id UUID NOT NULL,
    property_id UUID REFERENCES properties(id),
    room_type_id UUID REFERENCES room_types(id),
    confirmation_number VARCHAR(20) UNIQUE NOT NULL,
    status ENUM('inquiry', 'quote', 'booked', 'confirmed', 'checked_in', 'checked_out', 'completed', 'cancelled', 'refunded') DEFAULT 'booked',
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    nights INTEGER NOT NULL,
    adults INTEGER NOT NULL,
    children INTEGER DEFAULT 0,
    rooms INTEGER DEFAULT 1,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_id VARCHAR(100),
    special_requests TEXT,
    guest_details JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_guest_id (guest_id),
    INDEX idx_property_id (property_id),
    INDEX idx_room_type_id (room_type_id),
    INDEX idx_status (status),
    INDEX idx_check_in_date (check_in_date),
    INDEX idx_confirmation_number (confirmation_number)
);
```

### Flight Bookings Table
```sql
CREATE TABLE flight_bookings (
    id UUID PRIMARY KEY,
    traveler_id UUID NOT NULL,
    pnr VARCHAR(10) UNIQUE NOT NULL,
    airline_pnr VARCHAR(20),
    status ENUM('booked', 'confirmed', 'ticketed', 'cancelled', 'refunded') DEFAULT 'booked',
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payment_id VARCHAR(100),
    booking_class ENUM('economy', 'premium_economy', 'business', 'first') DEFAULT 'economy',
    passenger_details JSONB NOT NULL,
    flight_segments JSONB NOT NULL,
    ancillary_services JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_traveler_id (traveler_id),
    INDEX idx_pnr (pnr),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Property Reviews Table
```sql
CREATE TABLE property_reviews (
    id UUID PRIMARY KEY,
    booking_id UUID REFERENCES bookings(id),
    property_id UUID REFERENCES properties(id),
    guest_id UUID NOT NULL,
    overall_rating INTEGER CHECK (overall_rating >= 1 AND overall_rating <= 5),
    cleanliness_rating INTEGER CHECK (cleanliness_rating >= 1 AND cleanliness_rating <= 5),
    service_rating INTEGER CHECK (service_rating >= 1 AND service_rating <= 5),
    location_rating INTEGER CHECK (location_rating >= 1 AND location_rating <= 5),
    value_rating INTEGER CHECK (value_rating >= 1 AND value_rating <= 5),
    comment TEXT,
    images JSONB,
    helpful_count INTEGER DEFAULT 0,
    verified BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_property_id (property_id),
    INDEX idx_guest_id (guest_id),
    INDEX idx_overall_rating (overall_rating),
    INDEX idx_verified (verified),
    UNIQUE KEY unique_booking_review (booking_id, guest_id)
);
```

## Integration Patterns

### With Core Service
- Property owner authentication
- Guest authentication and profiles
- Payment processing coordination
- Subscription validation for properties

### With Marketplace Service
- Travel package integration
- Experience bookings
- Local activity coordination

### With Delivery Service
- Airport transfer coordination
- Room service integration
- Local delivery services

### With Media Service
- Property photo management
- Virtual tour integration
- Review photo processing

### With Realtime Service
- Booking notifications
- Live availability updates
- Guest communication

### Event Publishing
- `BookingCreated`: New booking made
- `BookingConfirmed`: Booking confirmed by property
- `CheckInCompleted`: Guest checked in
- `CheckOutCompleted`: Guest checked out
- `BookingCancelled`: Booking cancelled
- `ReviewSubmitted`: Guest review submitted

### Event Consumption
- `PaymentProcessed`: Confirm booking payment
- `PaymentFailed`: Cancel pending bookings
- `UserSuspended`: Cancel user bookings
- `SubscriptionExpired`: Restrict property features
