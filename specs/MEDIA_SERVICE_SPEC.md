# Media Service Specification

## Overview
The Media Service is a dedicated utility service that handles the complete lifecycle of user-generated content including secure uploads, processing, moderation, and delivery via CDN. It serves all other services with media management capabilities.

## Service Responsibilities

### Primary Functions
1. **Secure Media Upload** - Multi-format file upload with security validation
2. **Media Processing** - Image/video optimization and transformation
3. **Content Moderation** - Automated and manual content review
4. **CDN Integration** - Global content delivery and optimization
5. **Media Storage** - Scalable and secure media storage management
6. **Analytics & Reporting** - Media usage and performance analytics

## Features & Functionality

### 1. Secure Upload System

#### Upload Security
- **File Validation**
  - File type verification
  - File size limits enforcement
  - Virus/malware scanning
  - Metadata sanitization
  - MIME type validation

- **Access Control**
  - User authentication required
  - Permission-based uploads
  - Service-specific upload policies
  - Rate limiting per user
  - Quota management

#### Upload Methods
- **Direct Upload**
  - Browser-based upload
  - Mobile app integration
  - Drag-and-drop interface
  - Progress tracking
  - Resumable uploads

- **Programmatic Upload**
  - API-based uploads
  - Batch upload processing
  - Third-party integrations
  - Automated media imports
  - Service-to-service uploads

### 2. Media Processing Pipeline

#### Image Processing
- **Format Optimization**
  - WebP conversion for modern browsers
  - JPEG optimization with quality control
  - PNG lossless compression
  - HEIC/HEIF format support
  - SVG sanitization

- **Image Transformations**
  - Dynamic resizing and cropping
  - Thumbnail generation (multiple sizes)
  - Watermark application
  - Format conversion
  - Quality adjustments

- **Advanced Features**
  - Smart cropping with AI
  - Background removal
  - Color palette extraction
  - EXIF data handling
  - Progressive JPEG generation

#### Video Processing
- **Video Optimization**
  - Multi-bitrate encoding
  - Format conversion (MP4, WebM)
  - Compression optimization
  - Resolution scaling
  - Frame rate optimization

- **Video Features**
  - Thumbnail extraction
  - Preview clip generation
  - Video duration detection
  - Subtitle support
  - Chapter markers

### 3. Content Moderation System

#### Automated Moderation
- **AI-Powered Analysis**
  - NSFW content detection
  - Violence/inappropriate content screening
  - Text recognition in images (OCR)
  - Object/scene recognition
  - Brand safety analysis

- **Content Classification**
  - Age-appropriate rating
  - Content category tagging
  - Sensitivity scoring
  - Quality assessment
  - Spam detection

#### Manual Moderation
- **Review Workflow**
  - Flagged content queue
  - Moderator dashboard
  - Approval/rejection workflow
  - Escalation procedures
  - Appeal process

- **Moderation Tools**
  - Bulk actions
  - Custom moderation rules
  - Whitelist/blacklist management
  - Moderator performance tracking
  - Audit trail maintenance

### 4. CDN & Delivery System

#### Global Content Delivery
- **CDN Integration**
  - Multi-region distribution
  - Edge caching optimization
  - Automatic cache invalidation
  - Bandwidth optimization
  - Geographic routing

- **Delivery Features**
  - Adaptive quality serving
  - Progressive loading
  - Lazy loading support
  - Mobile-optimized delivery
  - Offline caching support

#### Performance Optimization
- **Image Optimization**
  - On-the-fly resizing
  - Format selection based on browser
  - Responsive image serving
  - Lossless compression
  - Cache-friendly URLs

- **Video Delivery**
  - Adaptive bitrate streaming
  - Progressive download
  - Video preloading
  - Quality switching
  - Bandwidth detection

### 5. Media Management

#### Storage Management
- **Storage Tiers**
  - Hot storage (frequently accessed)
  - Warm storage (occasionally accessed)
  - Cold storage (archival)
  - Automatic tier migration
  - Cost optimization

- **Backup & Recovery**
  - Multi-region replication
  - Point-in-time recovery
  - Disaster recovery procedures
  - Data integrity verification
  - Automated backup scheduling

#### Media Organization
- **Metadata Management**
  - Custom metadata fields
  - Automatic tag generation
  - Search and indexing
  - Category organization
  - Version control

- **Asset Management**
  - Media library interface
  - Bulk operations
  - Media collections
  - Usage tracking
  - Duplicate detection

## Admin Dashboard Features

### 1. Media Overview
- **Storage Analytics**
  - Storage usage by service
  - Growth trends
  - Cost breakdown
  - Performance metrics
  - Capacity planning

- **Upload Monitoring**
  - Upload success rates
  - Error analysis
  - Performance bottlenecks
  - User activity patterns
  - Service usage statistics

### 2. Content Moderation
- **Moderation Queue**
  - Pending content review
  - Flagged content management
  - Bulk moderation actions
  - Moderator assignment
  - Review performance tracking

- **Policy Management**
  - Content policy configuration
  - Automated rule setup
  - Threshold adjustments
  - Exception handling
  - Compliance reporting

### 3. System Management
- **Processing Pipeline**
  - Queue monitoring
  - Processing performance
  - Error handling
  - Resource utilization
  - Scaling controls

- **CDN Management**
  - Cache performance
  - Bandwidth usage
  - Geographic distribution
  - Cache invalidation
  - Origin server health

## API Endpoints

### Upload Management
```
POST /uploads/request-url
POST /uploads/complete
GET /uploads/{id}/status
DELETE /uploads/{id}
GET /uploads/user/{userId}
```

### Media Processing
```
POST /media/{id}/process
GET /media/{id}/variants
POST /media/{id}/transform
GET /media/{id}/status
POST /media/batch-process
```

### Content Moderation
```
POST /moderation/submit
GET /moderation/queue
POST /moderation/{id}/approve
POST /moderation/{id}/reject
GET /moderation/policy
```

### Media Delivery
```
GET /media/{id}
GET /media/{id}/thumbnail
GET /media/{id}/preview
GET /media/{id}/download
GET /media/{id}/stream
```

### Analytics & Reporting
```
GET /analytics/usage
GET /analytics/performance
GET /analytics/costs
GET /analytics/moderation
GET /analytics/cdn
```

## Database Schema

### Media Files Table
```sql
CREATE TABLE media_files (
    id UUID PRIMARY KEY,
    uploader_id UUID NOT NULL,
    service VARCHAR(50) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash VARCHAR(128) UNIQUE NOT NULL,
    storage_path VARCHAR(1000) NOT NULL,
    cdn_url VARCHAR(1000),
    width INTEGER,
    height INTEGER,
    duration DECIMAL(10,3), -- for videos/audio
    metadata JSONB,
    status ENUM('uploading', 'processing', 'ready', 'failed', 'deleted') DEFAULT 'uploading',
    moderation_status ENUM('pending', 'approved', 'rejected', 'flagged') DEFAULT 'pending',
    moderation_score DECIMAL(5,4),
    tags JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_uploader_id (uploader_id),
    INDEX idx_service (service),
    INDEX idx_content_type (content_type),
    INDEX idx_file_hash (file_hash),
    INDEX idx_status (status),
    INDEX idx_moderation_status (moderation_status),
    INDEX idx_created_at (created_at),
    FULLTEXT idx_tags ((tags->>'$.tags'))
);
```

### Media Variants Table
```sql
CREATE TABLE media_variants (
    id UUID PRIMARY KEY,
    media_file_id UUID REFERENCES media_files(id),
    variant_type ENUM('thumbnail', 'preview', 'compressed', 'watermarked', 'resized') NOT NULL,
    variant_name VARCHAR(100) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path VARCHAR(1000) NOT NULL,
    cdn_url VARCHAR(1000),
    width INTEGER,
    height INTEGER,
    processing_params JSONB,
    status ENUM('processing', 'ready', 'failed') DEFAULT 'processing',
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_media_file_id (media_file_id),
    INDEX idx_variant_type (variant_type),
    INDEX idx_status (status),
    UNIQUE KEY unique_media_variant (media_file_id, variant_type, variant_name)
);
```

### Upload Sessions Table
```sql
CREATE TABLE upload_sessions (
    id UUID PRIMARY KEY,
    uploader_id UUID NOT NULL,
    service VARCHAR(50) NOT NULL,
    filename VARCHAR(500) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    upload_url VARCHAR(1000),
    presigned_fields JSONB,
    status ENUM('initiated', 'uploading', 'completed', 'failed', 'expired') DEFAULT 'initiated',
    expires_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    media_file_id UUID REFERENCES media_files(id),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_uploader_id (uploader_id),
    INDEX idx_status (status),
    INDEX idx_expires_at (expires_at),
    INDEX idx_created_at (created_at)
);
```

### Moderation Queue Table
```sql
CREATE TABLE moderation_queue (
    id UUID PRIMARY KEY,
    media_file_id UUID REFERENCES media_files(id),
    moderator_id UUID,
    priority ENUM('low', 'normal', 'high', 'urgent') DEFAULT 'normal',
    reason VARCHAR(200),
    ai_analysis JSONB,
    ai_confidence DECIMAL(5,4),
    status ENUM('pending', 'in_review', 'approved', 'rejected', 'escalated') DEFAULT 'pending',
    moderator_notes TEXT,
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_media_file_id (media_file_id),
    INDEX idx_moderator_id (moderator_id),
    INDEX idx_priority (priority),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);
```

### Media Usage Analytics Table
```sql
CREATE TABLE media_usage_analytics (
    id UUID PRIMARY KEY,
    media_file_id UUID REFERENCES media_files(id),
    service VARCHAR(50) NOT NULL,
    access_date DATE NOT NULL,
    view_count BIGINT DEFAULT 0,
    download_count BIGINT DEFAULT 0,
    bandwidth_bytes BIGINT DEFAULT 0,
    unique_viewers BIGINT DEFAULT 0,
    geographic_data JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    INDEX idx_media_file_id (media_file_id),
    INDEX idx_service (service),
    INDEX idx_access_date (access_date),
    UNIQUE KEY unique_media_date (media_file_id, access_date)
);
```

### Processing Jobs Table
```sql
CREATE TABLE processing_jobs (
    id UUID PRIMARY KEY,
    media_file_id UUID REFERENCES media_files(id),
    job_type ENUM('thumbnail', 'compress', 'watermark', 'convert', 'moderate') NOT NULL,
    job_params JSONB NOT NULL,
    status ENUM('queued', 'processing', 'completed', 'failed', 'cancelled') DEFAULT 'queued',
    worker_id VARCHAR(100),
    progress_percent INTEGER DEFAULT 0,
    error_message TEXT,
    processing_time_seconds INTEGER,
    queued_at TIMESTAMP DEFAULT NOW(),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,

    INDEX idx_media_file_id (media_file_id),
    INDEX idx_job_type (job_type),
    INDEX idx_status (status),
    INDEX idx_queued_at (queued_at)
);
```

## Integration Patterns

### With Core Service
- User authentication for uploads
- Permission validation
- Usage quota management
- Billing integration for storage

### With Marketplace Service
- Product image uploads
- Image optimization for listings
- Bulk media operations
- Media moderation for marketplace content

### With Delivery Service
- Restaurant media management
- Menu item photos
- Delivery proof photos
- Marketing content

### With Booking Service
- Property photo management
- Room image galleries
- Virtual tour content
- Guest-generated content

### With TorraCab Service
- Driver profile photos
- Vehicle images
- Trip documentation
- Identity verification photos

### Event Publishing
- `MediaUploaded`: New media file uploaded
- `MediaProcessed`: Media processing completed
- `MediaModerated`: Content moderation completed
- `MediaDeleted`: Media file deleted
- `ProcessingFailed`: Media processing failed

### Event Consumption
- `UserSuspended`: Block user media uploads
- `ServiceQuotaExceeded`: Restrict service uploads
- `StorageQuotaWarning`: Notify of quota limits

## Security Considerations

### Upload Security
- **File Validation**
  - Strict MIME type checking
  - File signature verification
  - Content scanning for malware
  - Maximum file size enforcement
  - Extension whitelist/blacklist

### Access Control
- **Authentication Requirements**
  - Valid JWT tokens required
  - Service-specific permissions
  - User quota enforcement
  - Rate limiting per user
  - Temporary upload URLs

### Data Protection
- **Encryption**
  - Encryption at rest
  - Encryption in transit
  - Secure key management
  - Regular key rotation
  - Audit trail maintenance

## Performance Requirements

### Upload Performance
- **Throughput**: Support 10,000+ concurrent uploads
- **File Size**: Handle files up to 5GB
- **Response Time**: <100ms for upload URL generation
- **Reliability**: 99.9% upload success rate

### Processing Performance
- **Image Processing**: <5 seconds for standard operations
- **Video Processing**: Real-time for short clips
- **Batch Processing**: 100+ files per minute
- **Queue Management**: <1 second job queue response

### Delivery Performance
- **CDN Response**: <200ms global average
- **Bandwidth**: Support 10Gbps+ aggregate
- **Cache Hit Ratio**: >90% for popular content
- **Availability**: 99.99% uptime SLA
