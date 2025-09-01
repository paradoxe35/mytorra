package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
)

// LocationService handles location detection and area matching
type LocationService struct {
	areaService *AreaService
	logger      logger.Interface
}

// NewLocationService creates a new location service
func NewLocationService(areaService *AreaService, logger logger.Interface) *LocationService {
	return &LocationService{
		areaService: areaService,
		logger:      logger,
	}
}

// LocationDetectionResult represents the result of location detection
type LocationDetectionResult struct {
	Area      *entities.Area `json:"area"`
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Method    string         `json:"method"` // "gps", "ip", or "default"
}

// DetectAreaFromCoordinates detects area using GPS coordinates
func (s *LocationService) DetectAreaFromCoordinates(ctx context.Context, lat, lng float64) (*LocationDetectionResult, error) {
	s.logger.WithContext(ctx).Info("Detecting area using GPS coordinates",
		"lat", lat, "lng", lng)

	area, err := s.areaService.FindAreaByCoordinate(ctx, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("failed to find area by coordinates: %w", err)
	}

	if !area.IsActive() {
		return nil, fmt.Errorf("detected area is not active")
	}

	s.logger.WithContext(ctx).Info("Area detected using GPS",
		"area_id", area.ID, "area_name", area.Name)

	return &LocationDetectionResult{
		Area:      area,
		Latitude:  lat,
		Longitude: lng,
		Method:    "gps",
	}, nil
}

// DetectAreaFromIP detects area using IP address
func (s *LocationService) DetectAreaFromIP(ctx context.Context, ip string) (*LocationDetectionResult, error) {
	s.logger.WithContext(ctx).Info("Attempting IP-based location detection", "ip", ip)

	lat, lng, err := s.getLocationFromIP(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to get location from IP: %w", err)
	}

	s.logger.WithContext(ctx).Info("Location detected from IP",
		"ip", ip, "lat", lat, "lng", lng)

	area, err := s.areaService.FindAreaByCoordinate(ctx, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("failed to find area using IP location: %w", err)
	}

	if !area.IsActive() {
		return nil, fmt.Errorf("detected area is not active")
	}

	s.logger.WithContext(ctx).Info("Area detected using IP location",
		"area_id", area.ID, "area_name", area.Name)

	return &LocationDetectionResult{
		Area:      area,
		Latitude:  lat,
		Longitude: lng,
		Method:    "ip",
	}, nil
}

// DetectAreaFromRequest detects area from HTTP request (coordinates or IP)
func (s *LocationService) DetectAreaFromRequest(ctx context.Context, r *http.Request, lat, lng *float64) (*LocationDetectionResult, error) {
	// Try GPS coordinates first
	if lat != nil && lng != nil {
		result, err := s.DetectAreaFromCoordinates(ctx, *lat, *lng)
		if err == nil {
			return result, nil
		}
		s.logger.WithContext(ctx).Warn("Failed to find area using GPS coordinates", "error", err)
	}

	// Fallback to IP-based detection
	clientIP := s.ExtractClientIP(r)
	if clientIP == "" {
		return nil, fmt.Errorf("no location data available for area detection")
	}

	return s.DetectAreaFromIP(ctx, clientIP)
}

// DetectAreaFromContext detects area from context (for user service)
func (s *LocationService) DetectAreaFromContext(ctx context.Context, lat, lng *float64) (*LocationDetectionResult, error) {
	// Try GPS coordinates first
	if lat != nil && lng != nil {
		result, err := s.DetectAreaFromCoordinates(ctx, *lat, *lng)
		if err == nil {
			return result, nil
		}
		s.logger.WithContext(ctx).Warn("Failed to find area using GPS coordinates", "error", err)
	}

	// Fallback to IP-based detection
	clientIP := s.getClientIPFromContext(ctx)
	if clientIP == "" {
		return nil, fmt.Errorf("no location data available for area detection")
	}

	return s.DetectAreaFromIP(ctx, clientIP)
}

// ExtractClientIP extracts the client IP from HTTP request headers
func (s *LocationService) ExtractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header (common with nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfip := r.Header.Get("CF-Connecting-IP"); cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx] // Remove port
	}
	return ip
}

// getClientIPFromContext extracts client IP from request context (for user service)
func (s *LocationService) getClientIPFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(ClientIPKey).(string); ok {
		return ip
	}
	return ""
}

// getLocationFromIP uses IP geolocation service to get coordinates
func (s *LocationService) getLocationFromIP(ctx context.Context, ip string) (float64, float64, error) {
	// For development/testing, skip localhost IPs
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		// Return default coordinates for Bukavu, DRC as fallback
		return -2.5085, 28.8473, nil
	}

	// Use ip-api.com (free service, 1000 requests/month)
	// In production, consider using a paid service for better reliability
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,lat,lon", ip)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query IP geolocation service: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status string  `json:"status"`
		Lat    float64 `json:"lat"`
		Lon    float64 `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, fmt.Errorf("failed to decode IP geolocation response: %w", err)
	}

	if result.Status != "success" {
		return 0, 0, fmt.Errorf("IP geolocation service returned error status")
	}

	return result.Lat, result.Lon, nil
}
