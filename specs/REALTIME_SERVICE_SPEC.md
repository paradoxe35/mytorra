# Realtime Service Specification

## Overview
The Realtime Service is a specialized service running Centrifugo that manages persistent WebSocket connections and broadcasts real-time messages across the MyTorra platform. It serves as the central hub for all real-time communication needs.

## Service Responsibilities

### Primary Functions
1. **WebSocket Connection Management** - Persistent client connections
2. **Message Broadcasting** - Real-time message delivery
3. **Channel Management** - Topic-based message routing
4. **Authentication Integration** - Secure connection establishment
5. **Presence System** - Online/offline user status
6. **Message History** - Recent message storage and replay

## Features & Functionality

### 1. Connection Management

#### WebSocket Connections
- **Connection Establishment**
  - JWT-based authentication
  - Connection rate limiting
  - Automatic reconnection support
  - Connection health monitoring
  - Graceful connection handling

- **Connection Lifecycle**
  - Connect/disconnect events
  - Heartbeat/ping-pong mechanism
  - Connection timeout management
  - Resource cleanup on disconnect
  - Connection metadata tracking

#### Client Management
- **Client Identification**
  - Unique client ID assignment
  - User association tracking
  - Device/platform identification
  - Session management
  - Multi-device support

- **Connection Scaling**
  - Horizontal scaling support
  - Load balancing across instances
  - Connection distribution
  - Failover mechanisms
  - Performance monitoring

### 2. Message Broadcasting System

#### Channel-Based Messaging
- **Channel Types**
  - Public channels (open to all)
  - Private channels (invitation only)
  - Personal channels (user-specific)
  - Service channels (service-specific)
  - Broadcast channels (one-to-many)

- **Channel Naming Convention**
  ```
  public:general                    # Public discussion
  private:order:123                # Private order updates
  user:456                         # Personal notifications
  service:marketplace:updates      # Service-specific updates
  location:city:lagos              # Location-based channels
  ```

#### Message Types
- **System Messages**
  - Service status updates
  - Maintenance notifications
  - Emergency broadcasts
  - Platform announcements
  - Version updates

- **User Messages**
  - Chat messages
  - Notifications
  - Status updates
  - Activity feeds
  - Real-time reactions

- **Business Messages**
  - Order status updates
  - Delivery tracking
  - Booking confirmations
  - Payment notifications
  - Ride updates

### 3. Service-Specific Real-time Features

#### Marketplace Real-time
- **Product Updates**
  - New product listings
  - Price changes
  - Availability updates
  - Bidding updates
  - Sale notifications

- **Order Tracking**
  - Order status changes
  - Payment confirmations
  - Shipping updates
  - Delivery confirmations
  - Review notifications

#### Delivery Service Real-time
- **Order Management**
  - New order notifications
  - Order acceptance/rejection
  - Preparation status updates
  - Ready for pickup alerts
  - Customer notifications

- **Delivery Tracking**
  - Driver assignment
  - Live location updates
  - ETA calculations
  - Delivery confirmations
  - Customer feedback

#### MyTorraCab Real-time
- **Ride Management**
  - Ride requests
  - Driver assignment
  - Driver arrival updates
  - Trip status changes
  - Live location tracking

- **Driver Operations**
  - Job notifications
  - Passenger communications
  - Navigation updates
  - Earnings updates
  - Status changes

#### Booking Service Real-time
- **Reservation Updates**
  - Booking confirmations
  - Availability changes
  - Check-in reminders
  - Property communications
  - Special offers

### 4. Presence System

#### Online Status Management
- **User Presence**
  - Online/offline status
  - Last seen timestamps
  - Activity status (active/idle)
  - Device information
  - Location awareness

- **Business Presence**
  - Restaurant online status
  - Driver availability
  - Property manager status
  - Customer service availability
  - Service operational status

### 5. Message History & Replay

#### Message Storage
- **Recent Message History**
  - Last 100 messages per channel
  - Time-based message retention
  - Message pagination support
  - Selective message storage
  - Compressed message storage

- **Message Recovery**
  - Connection resume support
  - Missed message delivery
  - Message acknowledgment
  - Duplicate prevention
  - Message ordering guarantees

## API & WebSocket Interface

### WebSocket Connection
```javascript
// Connection establishment
const ws = new WebSocket('wss://realtime.mytorra.com/connection/websocket');

// Authentication
ws.send(JSON.stringify({
  id: 1,
  method: 'connect',
  params: {
    token: 'jwt_token_here',
    name: 'client_name'
  }
}));
```

### Channel Operations
```javascript
// Subscribe to channel
{
  id: 2,
  method: 'subscribe',
  params: {
    channel: 'user:123'
  }
}

// Unsubscribe from channel
{
  id: 3,
  method: 'unsubscribe',
  params: {
    channel: 'user:123'
  }
}

// Publish message
{
  id: 4,
  method: 'publish',
  params: {
    channel: 'private:order:456',
    data: {
      type: 'status_update',
      message: 'Order confirmed',
      timestamp: '2024-01-01T12:00:00Z'
    }
  }
}
```

### HTTP API Endpoints
```
POST /api/publish              # Publish message to channel
POST /api/broadcast            # Broadcast to multiple channels
GET /api/channels              # List active channels
GET /api/channels/{name}/info  # Get channel information
POST /api/presence             # Set user presence
GET /api/presence/{channel}    # Get channel presence
GET /api/history/{channel}     # Get channel message history
```

## Service Integration Patterns

### Message Publishing from Services

#### Core Service Integration
```go
// Publish user notification
err := realtimeClient.Publish("user:123", map[string]interface{}{
    "type": "notification",
    "title": "Payment Successful",
    "message": "Your payment has been processed successfully",
    "timestamp": time.Now(),
})
```

#### Marketplace Service Integration
```go
// Notify about order status
err := realtimeClient.Publish("private:order:456", map[string]interface{}{
    "type": "order_update",
    "order_id": "456",
    "status": "confirmed",
    "estimated_delivery": "2024-01-01T14:30:00Z",
})
```

#### MyTorraCab Service Integration
```go
// Broadcast driver location
err := realtimeClient.Publish("ride:789", map[string]interface{}{
    "type": "location_update",
    "driver_id": "driver123",
    "location": {
        "lat": 6.5244,
        "lng": 3.3792
    },
    "bearing": 45,
    "speed": 30,
    "timestamp": time.Now(),
})
```

## Configuration Management

### Centrifugo Configuration
```json
{
  "token_hmac_secret_key": "secret_key_here",
  "admin_password": "admin_password",
  "admin_secret": "admin_secret",
  "allowed_origins": ["https://mytorra.com"],
  "websocket_compression": true,
  "websocket_compression_min_size": 1024,
  "client_stale_close_delay": "10s",
  "client_expired_close_delay": "10s",
  "client_expired_sub_close_delay": "10s",
  "presence": true,
  "join_leave": true,
  "history_size": 100,
  "history_lifetime": 300,
  "recovery": true
}
```

### Channel Configuration
```json
{
  "namespaces": [
    {
      "name": "user",
      "private": true,
      "presence": true,
      "join_leave": true,
      "history_size": 50,
      "history_lifetime": 300
    },
    {
      "name": "public",
      "presence": true,
      "join_leave": true,
      "history_size": 100,
      "history_lifetime": 600
    },
    {
      "name": "service",
      "private": true,
      "history_size": 200,
      "history_lifetime": 3600
    }
  ]
}
```

## Authentication & Authorization

### JWT Token Structure
```json
{
  "sub": "user_id_123",
  "exp": 1640995200,
  "channels": [
    "user:123",
    "public:general"
  ],
  "info": {
    "name": "John Doe",
    "avatar": "https://cdn.mytorra.com/avatars/123.jpg"
  }
}
```

### Channel Permissions
- **Public Channels**: No authentication required
- **Private Channels**: JWT token with channel access required
- **User Channels**: Only the user can access their channel
- **Service Channels**: Service-specific authentication required

## Monitoring & Analytics

### Connection Metrics
- **Active Connections**
  - Total connections
  - Connections per user
  - Connections per service
  - Geographic distribution
  - Device breakdown

- **Performance Metrics**
  - Connection establishment time
  - Message delivery latency
  - Throughput (messages/second)
  - Error rates
  - Memory usage

### Channel Analytics
- **Channel Activity**
  - Active channels
  - Messages per channel
  - Subscriber counts
  - Peak usage times
  - Popular channels

- **Message Analytics**
  - Message types distribution
  - Message size distribution
  - Delivery success rates
  - Replay requests
  - Failed deliveries

## Admin Dashboard Features

### 1. Real-time Monitoring
- **Live Statistics**
  - Active connections count
  - Messages per second
  - Channel activity
  - Server performance
  - Error monitoring

- **Connection Management**
  - Active connections list
  - User session details
  - Connection history
  - Force disconnect capability
  - Connection troubleshooting

### 2. Channel Management
- **Channel Overview**
  - Active channels list
  - Subscriber counts
  - Message history
  - Channel configuration
  - Access control

- **Message Management**
  - Broadcast messages
  - Message history viewing
  - Message filtering
  - Emergency broadcasts
  - Scheduled messages

### 3. System Administration
- **Server Management**
  - Server status monitoring
  - Configuration updates
  - Log viewing
  - Performance tuning
  - Maintenance mode

- **Security Management**
  - Authentication logs
  - Access control
  - Rate limiting configuration
  - Blocked connections
  - Security alerts

## Performance Requirements

### Connection Handling
- **Concurrent Connections**: Support 100,000+ simultaneous connections
- **Connection Time**: <100ms connection establishment
- **Message Latency**: <50ms message delivery
- **Throughput**: 50,000+ messages per second

### Scalability
- **Horizontal Scaling**: Support multiple Centrifugo instances
- **Load Balancing**: Distribute connections across instances
- **Auto Scaling**: Dynamic scaling based on load
- **Geographic Distribution**: Multi-region deployment support

### Reliability
- **Uptime**: 99.99% availability
- **Message Delivery**: 99.9% delivery success rate
- **Connection Recovery**: Automatic reconnection support
- **Data Consistency**: Message ordering guarantees

## Security Considerations

### Connection Security
- **Authentication**: JWT-based authentication required
- **Authorization**: Channel-based access control
- **Rate Limiting**: Connection and message rate limits
- **DDoS Protection**: Connection flood protection

### Message Security
- **Message Validation**: Content validation and sanitization
- **Encryption**: TLS encryption for all connections
- **Message Integrity**: Message tampering prevention
- **Audit Trail**: Message delivery logging

### Privacy Protection
- **Data Minimization**: Only necessary data transmitted
- **User Consent**: Presence sharing consent
- **Data Retention**: Automatic message cleanup
- **Anonymization**: User data anonymization options

## Integration with Other Services

### Event-Driven Integration
```go
// Listen for events from message bus
func handleOrderUpdate(event OrderUpdateEvent) {
    message := map[string]interface{}{
        "type": "order_update",
        "order_id": event.OrderID,
        "status": event.Status,
        "timestamp": event.Timestamp,
    }

    // Publish to customer channel
    realtimeClient.Publish(fmt.Sprintf("user:%s", event.CustomerID), message)

    // Publish to seller channel
    realtimeClient.Publish(fmt.Sprintf("user:%s", event.SellerID), message)
}
```

### Service Health Integration
```go
// Broadcast service status updates
func broadcastServiceStatus(service string, status string) {
    message := map[string]interface{}{
        "type": "service_status",
        "service": service,
        "status": status,
        "timestamp": time.Now(),
    }

    realtimeClient.Broadcast("service:status", message)
}
```

## Disaster Recovery

### Backup & Recovery
- **Message History Backup**: Regular backup of message history
- **Configuration Backup**: Automated configuration backups
- **Connection State**: Connection state preservation
- **Failover Procedures**: Automated failover mechanisms

### High Availability
- **Multi-Instance Deployment**: Multiple Centrifugo instances
- **Load Balancer Integration**: Health check and failover
- **Database Replication**: Redis cluster for scaling
- **Geographic Redundancy**: Multi-region deployment
