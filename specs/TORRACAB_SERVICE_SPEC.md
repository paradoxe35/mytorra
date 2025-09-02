# TorraCab Service Specification

## Overview
TorraCab is a standalone ride-hailing service optimized for high-frequency, real-time operations. It manages drivers, vehicles, ride requests, fare calculations, and provides live location tracking for seamless transportation services.

## Service Responsibilities

### Primary Functions
1. **Driver Management** - Driver onboarding, verification, and performance tracking
2. **Vehicle Management** - Vehicle registration and maintenance tracking
3. **Ride Request Processing** - Real-time ride matching and dispatch
4. **Fare Calculation** - Dynamic pricing and fare computation
5. **Live Tracking** - Real-time GPS tracking for rides
6. **Route Optimization** - Efficient routing and navigation

## Features & Functionality

### 1. Driver Management System

#### Driver Onboarding
- **Driver Registration**
  - Personal information collection
  - Driver license verification
  - Background check integration
  - Profile photo capture
  - Emergency contact details

- **Vehicle Documentation**
  - Vehicle registration verification
  - Insurance certificate validation
  - Safety inspection reports
  - Vehicle photos (interior/exterior)
  - Commercial permit validation

- **Financial Setup**
  - Bank account verification
  - Tax information collection
  - Payment method setup
  - Earnings configuration
  - Direct deposit setup

#### Driver Verification Process
- **Document Verification**
  - ID document scanning
  - License authenticity check
  - Criminal background screening
  - Driving record review
  - Reference verification

- **Training & Certification**
  - Platform training completion
  - Safety protocol training
  - Customer service guidelines
  - City regulation compliance
  - Emergency procedure training

### 2. Vehicle Management

#### Vehicle Registration
- **Vehicle Information**
  - Make, model, year
  - License plate number
  - VIN verification
  - Color and description
  - Passenger capacity

- **Safety & Compliance**
  - Safety inspection certificates
  - Insurance coverage verification
  - Emission test compliance
  - Commercial license validation
  - Vehicle condition assessment

#### Vehicle Categories
- **Standard Vehicles**
  - Economy cars (4 passengers)
  - Sedan cars (4 passengers)
  - Comfort vehicles (4 passengers)
  - Premium vehicles (4 passengers)

- **Specialized Vehicles**
  - Large capacity (6-8 passengers)
  - Wheelchair accessible
  - Luxury vehicles
  - Electric vehicles

### 3. Ride Request & Matching System

#### Ride Booking Flow
- **Passenger Request**
  - Current location detection
  - Destination selection
  - Vehicle type preference
  - Ride scheduling options
  - Special requirements

- **Smart Matching Algorithm**
  - Proximity-based matching
  - Driver availability status
  - Vehicle type compatibility
  - Driver rating consideration
  - Estimated arrival time

#### Real-time Operations
- **Driver Dispatch**
  - Automatic driver assignment
  - Manual dispatch override
  - Multiple driver notifications
  - Acceptance timeout handling
  - Backup driver assignment

- **Ride Tracking**
  - Real-time GPS tracking
  - ETA calculations
  - Route optimization
  - Traffic condition integration
  - Live sharing with passengers

### 4. Fare & Pricing System

#### Dynamic Pricing Engine
- **Base Fare Structure**
  - Distance-based pricing
  - Time-based pricing
  - Minimum fare guarantees
  - Surge pricing algorithms
  - Peak hour multipliers

- **Pricing Factors**
  - Supply and demand analysis
  - Traffic conditions
  - Weather conditions
  - Special events impact
  - Geographic zone pricing

#### Fare Components
- **Standard Charges**
  - Base fare
  - Per kilometer rate
  - Per minute rate
  - Booking fee
  - Service fee

- **Additional Charges**
  - Airport pickup fee
  - Tolls and parking
  - Cancellation fee
  - Wait time charges
  - Cleaning fee

### 5. Real-time Location & Navigation

#### GPS Tracking System
- **Live Location Updates**
  - Driver location broadcasting
  - Passenger location tracking
  - Route progress monitoring
  - Geofence integration
  - Location accuracy validation

- **Navigation Integration**
  - Turn-by-turn directions
  - Traffic-aware routing
  - Alternative route suggestions
  - Landmark-based navigation
  - Voice guidance support

#### Safety Features
- **Emergency Support**
  - Panic button integration
  - Emergency contact alerts
  - Location sharing with contacts
  - 24/7 support hotline
  - Incident reporting system

- **Safety Monitoring**
  - Route deviation alerts
  - Unexpected stop detection
  - Speed monitoring
  - Driver fatigue detection
  - Passenger safety checks

## Driver Dashboard Features

### 1. Earnings & Performance
- **Earnings Overview**
  - Daily/weekly/monthly earnings
  - Ride completion statistics
  - Bonus and incentive tracking
  - Commission breakdown
  - Payment history

- **Performance Metrics**
  - Driver rating display
  - Acceptance rate tracking
  - Completion rate monitoring
  - Customer feedback review
  - Performance improvement tips

### 2. Ride Management
- **Active Rides**
  - Current ride details
  - Passenger contact information
  - Navigation integration
  - Ride status updates
  - Trip documentation

- **Ride History**
  - Completed ride records
  - Earnings per ride
  - Route information
  - Passenger ratings
  - Trip receipts

### 3. Vehicle & Account Management
- **Vehicle Status**
  - Vehicle information updates
  - Maintenance reminders
  - Document expiry alerts
  - Insurance status
  - Inspection schedules

- **Account Settings**
  - Profile information
  - Payment preferences
  - Notification settings
  - Privacy controls
  - Support access

## Passenger App Features

### 1. Ride Booking
- **Location Services**
  - Current location detection
  - Address search and selection
  - Favorite location management
  - Recent destination history
  - Map-based selection

- **Ride Options**
  - Vehicle type selection
  - Ride scheduling
  - Fare estimation
  - Payment method selection
  - Special requirements

### 2. Ride Experience
- **Live Tracking**
  - Real-time driver location
  - ETA updates
  - Route visualization
  - Driver contact options
  - Trip sharing

- **Communication**
  - In-app messaging
  - Voice calling
  - Special instructions
  - Emergency contacts
  - Support access

### 3. Payment & History
- **Payment Management**
  - Multiple payment methods
  - Automatic payment processing
  - Receipt generation
  - Expense tracking
  - Corporate billing

- **Ride History**
  - Trip records
  - Fare breakdowns
  - Driver information
  - Route maps
  - Rating submissions

## Admin Dashboard Features

### 1. Driver Administration
- **Driver Management**
  - Driver approval process
  - Performance monitoring
  - Document verification
  - Suspension management
  - Training coordination

- **Fleet Analytics**
  - Active driver count
  - Utilization rates
  - Performance statistics
  - Earnings distribution
  - Geographic coverage

### 2. Operations Management
- **Real-time Monitoring**
  - Live ride tracking
  - Demand heatmaps
  - Supply distribution
  - System performance
  - Issue alerts

- **Pricing Management**
  - Surge pricing control
  - Rate adjustments
  - Promotional campaigns
  - Market analysis
  - Revenue optimization

### 3. Safety & Compliance
- **Incident Management**
  - Incident reporting
  - Investigation tools
  - Resolution tracking
  - Safety analytics
  - Compliance monitoring

- **Regulatory Compliance**
  - License management
  - Insurance tracking
  - Safety standards
  - Local regulation adherence
  - Audit trail maintenance

## API Endpoints

### Driver Management
```
POST /drivers/register
GET /drivers/profile
PUT /drivers/profile
PUT /drivers/status
POST /drivers/documents
GET /drivers/earnings
```

### Vehicle Management
```
POST /drivers/vehicles
GET /drivers/vehicles
PUT /drivers/vehicles/{id}
POST /drivers/vehicles/{id}/documents
GET /drivers/vehicles/{id}/maintenance
```

### Ride Operations
```
POST /rides/request
GET /rides/available
POST /rides/accept/{rideId}
PUT /rides/{id}/status
GET /rides/{id}/tracking
POST /rides/{id}/complete
```

### Fare & Pricing
```
POST /fares/estimate
GET /fares/structure
POST /fares/calculate
GET /rides/{id}/fare
POST /rides/{id}/payment
```

### Location Services
```
PUT /drivers/location
GET /rides/{id}/location
POST /location/geocode
GET /location/places
POST /location/route
```

## Database Schema

### Drivers Table
```sql
CREATE TABLE drivers (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    license_number VARCHAR(50) UNIQUE NOT NULL,
    license_expiry DATE NOT NULL,
    background_check_status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
    background_check_date DATE,
    status ENUM('pending', 'active', 'inactive', 'suspended', 'banned') DEFAULT 'pending',
    rating DECIMAL(3,2) DEFAULT 5.00,
    total_trips INTEGER DEFAULT 0,
    acceptance_rate DECIMAL(5,2) DEFAULT 100.00,
    completion_rate DECIMAL(5,2) DEFAULT 100.00,
    last_online TIMESTAMP,
    earnings_total DECIMAL(12,2) DEFAULT 0.00,
    emergency_contact JSONB,
    documents JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_user_id (user_id),
    INDEX idx_license_number (license_number),
    INDEX idx_status (status),
    INDEX idx_rating (rating),
    INDEX idx_background_check_status (background_check_status)
);
```

### Vehicles Table
```sql
CREATE TABLE vehicles (
    id UUID PRIMARY KEY,
    driver_id UUID REFERENCES drivers(id),
    make VARCHAR(50) NOT NULL,
    model VARCHAR(50) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(30) NOT NULL,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    vin VARCHAR(17) UNIQUE NOT NULL,
    vehicle_type ENUM('economy', 'standard', 'comfort', 'premium', 'xl', 'luxury', 'accessible') DEFAULT 'standard',
    passenger_capacity INTEGER DEFAULT 4,
    insurance_policy VARCHAR(100),
    insurance_expiry DATE,
    registration_expiry DATE,
    safety_inspection_expiry DATE,
    status ENUM('active', 'inactive', 'maintenance', 'suspended') DEFAULT 'active',
    images JSONB,
    documents JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_driver_id (driver_id),
    INDEX idx_license_plate (license_plate),
    INDEX idx_vehicle_type (vehicle_type),
    INDEX idx_status (status)
);
```

### Rides Table
```sql
CREATE TABLE rides (
    id UUID PRIMARY KEY,
    passenger_id UUID NOT NULL,
    driver_id UUID REFERENCES drivers(id),
    vehicle_id UUID REFERENCES vehicles(id),
    status ENUM('requested', 'accepted', 'driver_arrived', 'in_progress', 'completed', 'cancelled') DEFAULT 'requested',
    pickup_location JSONB NOT NULL,
    pickup_address TEXT,
    destination_location JSONB NOT NULL,
    destination_address TEXT,
    requested_at TIMESTAMP DEFAULT NOW(),
    accepted_at TIMESTAMP,
    pickup_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    cancellation_reason TEXT,
    distance_km DECIMAL(8,3),
    duration_minutes INTEGER,
    base_fare DECIMAL(8,2),
    distance_fare DECIMAL(8,2),
    time_fare DECIMAL(8,2),
    surge_multiplier DECIMAL(4,2) DEFAULT 1.00,
    total_fare DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    payment_id VARCHAR(100),
    payment_status ENUM('pending', 'completed', 'failed', 'refunded') DEFAULT 'pending',
    passenger_rating INTEGER CHECK (passenger_rating >= 1 AND passenger_rating <= 5),
    driver_rating INTEGER CHECK (driver_rating >= 1 AND driver_rating <= 5),
    route_data JSONB,
    notes TEXT,

    INDEX idx_passenger_id (passenger_id),
    INDEX idx_driver_id (driver_id),
    INDEX idx_status (status),
    INDEX idx_requested_at (requested_at),
    INDEX idx_pickup_location ((CAST(pickup_location AS POINT))),
    INDEX idx_destination_location ((CAST(destination_location AS POINT)))
);
```

### Driver Locations Table
```sql
CREATE TABLE driver_locations (
    id UUID PRIMARY KEY,
    driver_id UUID REFERENCES drivers(id),
    location POINT NOT NULL,
    heading INTEGER, -- 0-359 degrees
    speed_kmh DECIMAL(5,2),
    accuracy_meters INTEGER,
    is_online BOOLEAN DEFAULT TRUE,
    is_available BOOLEAN DEFAULT TRUE,
    current_ride_id UUID,
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_driver_id (driver_id),
    SPATIAL INDEX idx_location (location),
    INDEX idx_is_online (is_online),
    INDEX idx_is_available (is_available),
    INDEX idx_updated_at (updated_at)
);
```

### Fare Structure Table
```sql
CREATE TABLE fare_structures (
    id UUID PRIMARY KEY,
    city VARCHAR(100) NOT NULL,
    vehicle_type ENUM('economy', 'standard', 'comfort', 'premium', 'xl', 'luxury', 'accessible') NOT NULL,
    base_fare DECIMAL(8,2) NOT NULL,
    per_km_rate DECIMAL(8,2) NOT NULL,
    per_minute_rate DECIMAL(8,2) NOT NULL,
    minimum_fare DECIMAL(8,2) NOT NULL,
    booking_fee DECIMAL(8,2) DEFAULT 0.00,
    cancellation_fee DECIMAL(8,2) DEFAULT 0.00,
    airport_fee DECIMAL(8,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    surge_zones JSONB,
    active BOOLEAN DEFAULT TRUE,
    effective_from TIMESTAMP DEFAULT NOW(),
    effective_until TIMESTAMP,

    INDEX idx_city (city),
    INDEX idx_vehicle_type (vehicle_type),
    INDEX idx_active (active),
    INDEX idx_effective_from (effective_from),
    UNIQUE KEY unique_city_vehicle_type (city, vehicle_type, effective_from)
);
```

### Driver Earnings Table
```sql
CREATE TABLE driver_earnings (
    id UUID PRIMARY KEY,
    driver_id UUID REFERENCES drivers(id),
    ride_id UUID REFERENCES rides(id),
    earning_date DATE NOT NULL,
    ride_fare DECIMAL(8,2) NOT NULL,
    commission_rate DECIMAL(5,2) NOT NULL,
    commission_amount DECIMAL(8,2) NOT NULL,
    driver_earning DECIMAL(8,2) NOT NULL,
    bonus_amount DECIMAL(8,2) DEFAULT 0.00,
    tip_amount DECIMAL(8,2) DEFAULT 0.00,
    total_earning DECIMAL(8,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    payout_id UUID,
    payout_status ENUM('pending', 'processing', 'paid', 'failed') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_driver_id (driver_id),
    INDEX idx_ride_id (ride_id),
    INDEX idx_earning_date (earning_date),
    INDEX idx_payout_status (payout_status)
);
```

## Integration Patterns

### With Core Service
- Driver and passenger authentication
- Payment processing coordination
- User profile management
- Background check integration

### With Delivery Service
- Shared driver pool coordination
- Cross-service availability
- Driver utilization optimization

### With Realtime Service
- Live location broadcasting
- Ride status updates
- Real-time notifications
- Emergency alerts

### With Media Service
- Driver photo verification
- Vehicle photo management
- Document upload processing

### Event Publishing
- `RideRequested`: New ride request created
- `RideAccepted`: Driver accepts ride
- `RideStarted`: Ride in progress
- `RideCompleted`: Ride finished
- `RideCancelled`: Ride cancelled
- `DriverLocationUpdated`: Driver location changes
- `EmergencyTriggered`: Emergency button pressed

### Event Consumption
- `PaymentProcessed`: Complete ride payment
- `PaymentFailed`: Handle payment failures
- `UserSuspended`: Suspend driver/passenger
- `BackgroundCheckCompleted`: Update driver status

## Performance Requirements

### Real-time Operations
- **Location Updates**: Sub-second GPS tracking
- **Matching Algorithm**: <3 seconds for driver assignment
- **ETA Calculations**: <1 second response time
- **Route Optimization**: Dynamic rerouting within 5 seconds

### Scalability Targets
- **Concurrent Rides**: Support 10,000+ simultaneous rides
- **Driver Pool**: Manage 50,000+ active drivers
- **API Throughput**: Handle 100,000+ requests per minute
- **Database Performance**: <100ms query response times
