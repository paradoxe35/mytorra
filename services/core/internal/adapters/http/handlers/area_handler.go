package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	httpPkg "github.com/paradoxe35/mytorra/packages/http"
	"github.com/paradoxe35/mytorra/packages/i18n"
	"github.com/paradoxe35/mytorra/packages/logger"
	"github.com/paradoxe35/mytorra/services/core/internal/application/services"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
)

type AreaHandler struct {
	areaService     *services.AreaService
	locationService *services.LocationService
	logger          logger.Interface
	translator      *i18n.Translator
}

func NewAreaHandler(
	areaService *services.AreaService,
	locationService *services.LocationService,
	logger logger.Interface,
	translator *i18n.Translator,
) *AreaHandler {
	return &AreaHandler{
		areaService:     areaService,
		locationService: locationService,
		logger:          logger,
		translator:      translator,
	}
}

// DetectAreaRequest represents the request for area detection
type DetectAreaRequest struct {
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,latitude"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,longitude"`
}

// AreaDetectionResult represents the result of area detection
type AreaDetectionResult struct {
	Area      *entities.Area `json:"area"`
	Latitude  float64        `json:"latitude"`
	Longitude float64        `json:"longitude"`
	Method    string         `json:"method"`
}

// DetectAreaResponse represents the response for area detection
type DetectAreaResponse struct {
	Area      *AreaResponse `json:"area"`
	Method    string        `json:"method"` // "gps", "ip", or "default"
	Latitude  *float64      `json:"latitude,omitempty"`
	Longitude *float64      `json:"longitude,omitempty"`
}

// AreaResponse represents an area in API responses
type AreaResponse struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Code              string                 `json:"code"`
	Type              string                 `json:"type"`
	CenterLat         float64                `json:"center_lat"`
	CenterLng         float64                `json:"center_lng"`
	OperationalRadius int                    `json:"operational_radius"`
	Status            string                 `json:"status"`
	Timezone          string                 `json:"timezone"`
	Languages         []string               `json:"languages"`
	Currency          string                 `json:"currency"`
	Settings          map[string]interface{} `json:"settings,omitempty"`
}

// DetectArea handles POST /api/v1/areas/detect
// This endpoint detects the area based on GPS coordinates or IP address
func (h *AreaHandler) DetectArea(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body (optional)
	var req DetectAreaRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Ignore JSON decode errors - coordinates might be in query params
			h.logger.WithError(err).Debug("Failed to decode JSON body, checking query params")
		}
	}

	// Check query parameters if not in body
	if req.Latitude == nil && req.Longitude == nil {
		if latStr := r.URL.Query().Get("lat"); latStr != "" {
			if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
				req.Latitude = &lat
			}
		}
		if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
			if lng, err := strconv.ParseFloat(lngStr, 64); err == nil {
				req.Longitude = &lng
			}
		}
	}

	// Use location service for area detection
	result, err := h.locationService.DetectAreaFromRequest(ctx, r, req.Latitude, req.Longitude)
	if err != nil {
		h.logger.WithContext(ctx).Error("Failed to detect area", "error", err)

		if strings.Contains(err.Error(), "no location data available") {
			httpPkg.RespondError(w, http.StatusBadRequest, "NO_LOCATION_DATA",
				h.translator.TWithContext(ctx, "area.no_location_data"))
		} else {
			httpPkg.RespondError(w, http.StatusNotFound, "AREA_NOT_FOUND",
				h.translator.TWithContext(ctx, "area.detection_failed"))
		}
		return
	}

	response := &DetectAreaResponse{
		Area:      h.mapAreaToResponse(result.Area),
		Method:    result.Method,
		Latitude:  &result.Latitude,
		Longitude: &result.Longitude,
	}
	httpPkg.RespondJSON(w, http.StatusOK, response)
}

// GetActiveAreas handles GET /api/v1/areas
// This endpoint returns all active areas
func (h *AreaHandler) GetActiveAreas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	areas, err := h.areaService.GetActiveAreas(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get active areas")
		httpPkg.RespondError(w, http.StatusInternalServerError, "FETCH_FAILED", h.translator.TWithContext(ctx, "area.fetch_failed"))
		return
	}

	// Map to response format
	response := make([]*AreaResponse, len(areas))
	for i, area := range areas {
		response[i] = h.mapAreaToResponse(area)
	}

	httpPkg.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"areas": response,
		"count": len(response),
	})
}

// Helper methods

// mapAreaToResponse converts an area entity to API response format
func (h *AreaHandler) mapAreaToResponse(area *entities.Area) *AreaResponse {
	return &AreaResponse{
		ID:                area.ID.String(),
		Name:              area.Name,
		Code:              area.Code,
		Type:              string(area.Type),
		CenterLat:         area.CenterLat,
		CenterLng:         area.CenterLng,
		OperationalRadius: area.OperationalRadius,
		Status:            string(area.Status),
		Timezone:          area.Timezone,
		Languages:         area.Languages,
		Currency:          area.Currency,
		Settings:          area.Settings,
	}
}
