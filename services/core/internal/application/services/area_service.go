package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/torra/packages/logger"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type AreaService struct {
	areaRepo       ports.AreaRepository
	eventPublisher ports.EventPublisher
	cache          ports.Cache
	logger         logger.Interface
}

func NewAreaService(
	areaRepo ports.AreaRepository,
	eventPublisher ports.EventPublisher,
	cache ports.Cache,
	logger logger.Interface,
) *AreaService {
	return &AreaService{
		areaRepo:       areaRepo,
		eventPublisher: eventPublisher,
		cache:          cache,
		logger:         logger,
	}
}

func (s *AreaService) GetAreaByID(ctx context.Context, areaID uuid.UUID) (*entities.Area, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("area:%s", areaID)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var area entities.Area
		if err := json.Unmarshal(cached, &area); err == nil {
			return &area, nil
		}
	}

	// Get from repository
	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("failed to find area: %w", err)
	}

	// Update cache
	if data, err := json.Marshal(area); err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return area, nil
}

func (s *AreaService) GetAreaByCode(ctx context.Context, code string) (*entities.Area, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("area:code:%s", code)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var area entities.Area
		if err := json.Unmarshal(cached, &area); err == nil {
			return &area, nil
		}
	}

	// Get from repository
	area, err := s.areaRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to find area by code: %w", err)
	}

	// Update cache
	if data, err := json.Marshal(area); err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return area, nil
}

func (s *AreaService) GetActiveAreas(ctx context.Context) ([]*entities.Area, error) {
	// Try cache first
	cacheKey := "areas:active"
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var areas []*entities.Area
		if err := json.Unmarshal(cached, &areas); err == nil {
			return areas, nil
		}
	}

	// Get from repository
	areas, err := s.areaRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find active areas: %w", err)
	}

	// Update cache
	if data, err := json.Marshal(areas); err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return areas, nil
}

func (s *AreaService) GetAreasByType(ctx context.Context, areaType valueobjects.AreaType) ([]*entities.Area, error) {
	return s.areaRepo.FindByType(ctx, areaType)
}

func (s *AreaService) GetAreasByParent(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error) {
	return s.areaRepo.FindByParentID(ctx, parentID)
}

func (s *AreaService) GetAreaHierarchy(ctx context.Context, areaID uuid.UUID) ([]*entities.Area, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("area:hierarchy:%s", areaID)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var hierarchy []*entities.Area
		if err := json.Unmarshal(cached, &hierarchy); err == nil {
			return hierarchy, nil
		}
	}

	// Get from repository
	hierarchy, err := s.areaRepo.GetHierarchy(ctx, areaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get area hierarchy: %w", err)
	}

	// Update cache
	if data, err := json.Marshal(hierarchy); err == nil {
		_ = s.cache.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return hierarchy, nil
}

func (s *AreaService) FindAreaByCoordinate(ctx context.Context, lat, lng float64) (*entities.Area, error) {
	areas, err := s.areaRepo.FindByCoordinate(ctx, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("failed to find area by coordinate: %w", err)
	}

	// Find the most specific area (highest level)
	var bestMatch *entities.Area
	maxLevel := 0

	for _, area := range areas {
		if area.ContainsCoordinate(lat, lng) {
			level := area.Type.Level()
			if level > maxLevel {
				maxLevel = level
				bestMatch = area
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no area found for coordinates (%f, %f)", lat, lng)
	}

	return bestMatch, nil
}

func (s *AreaService) CreateArea(ctx context.Context, req CreateAreaRequest) (*entities.Area, error) {
	// Validate parent exists if provided
	if req.ParentID != nil {
		parent, err := s.areaRepo.FindByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent area not found: %w", err)
		}

		// Check if parent can have children
		if !parent.CanAddSubArea() {
			return nil, fmt.Errorf("parent area of type %s cannot have sub-areas", parent.Type)
		}

		// Validate type hierarchy
		expectedLevel := parent.Type.Level() + 1
		if req.Type.Level() != expectedLevel {
			return nil, fmt.Errorf("invalid area type for parent")
		}
	}

	// Check for duplicate code
	if existing, _ := s.areaRepo.FindByCode(ctx, req.Code); existing != nil {
		return nil, fmt.Errorf("area with code %s already exists", req.Code)
	}

	// Create area entity
	area := &entities.Area{
		ID:                   uuid.New(),
		ParentID:             req.ParentID,
		Type:                 req.Type,
		Name:                 req.Name,
		Code:                 req.Code,
		CenterLat:            req.CenterLat,
		CenterLng:            req.CenterLng,
		OperationalRadius:    req.OperationalRadius,
		Status:               valueobjects.AreaStatusInactive,
		ServiceAvailability:  req.ServiceAvailability,
		Settings:             req.Settings,
		SwitchingMode:        req.SwitchingMode,
		AllowCrossTrade:      req.AllowCrossTrade,
		Timezone:             req.Timezone,
		Languages:            req.Languages,
		Currency:             req.Currency,
		TaxRate:              req.TaxRate,
		MinOrderValue:        req.MinOrderValue,
		DeliveryFeeStructure: req.DeliveryFeeStructure,
		CrossAreaFees:        req.CrossAreaFees,
		OperationalHours:     req.OperationalHours,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save to repository
	if err := s.areaRepo.Save(ctx, area); err != nil {
		return nil, fmt.Errorf("failed to save area: %w", err)
	}

	// Invalidate caches
	s.invalidateAreaCaches(ctx)

	return area, nil
}

func (s *AreaService) UpdateArea(ctx context.Context, areaID uuid.UUID, req UpdateAreaRequest) (*entities.Area, error) {
	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		area.Name = *req.Name
	}
	if req.Status != nil {
		area.Status = *req.Status
	}
	if req.ServiceAvailability != nil {
		area.ServiceAvailability = req.ServiceAvailability
	}
	if req.Settings != nil {
		area.Settings = req.Settings
	}
	if req.OperationalRadius != nil {
		area.OperationalRadius = *req.OperationalRadius
	}
	if req.SwitchingMode != nil {
		area.SwitchingMode = *req.SwitchingMode
	}
	if req.AllowCrossTrade != nil {
		area.AllowCrossTrade = *req.AllowCrossTrade
	}
	if req.MinOrderValue != nil {
		area.MinOrderValue = *req.MinOrderValue
	}
	if req.TaxRate != nil {
		area.TaxRate = *req.TaxRate
	}

	area.UpdatedAt = time.Now()

	// Save changes
	if err := s.areaRepo.Update(ctx, area); err != nil {
		return nil, fmt.Errorf("failed to update area: %w", err)
	}

	// Invalidate caches
	s.invalidateAreaCaches(ctx)
	s.invalidateAreaCache(ctx, areaID)

	return area, nil
}

func (s *AreaService) ActivateArea(ctx context.Context, areaID uuid.UUID) error {
	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return err
	}

	area.Activate()

	if err := s.areaRepo.Update(ctx, area); err != nil {
		return fmt.Errorf("failed to activate area: %w", err)
	}

	// Invalidate caches
	s.invalidateAreaCaches(ctx)
	s.invalidateAreaCache(ctx, areaID)

	return nil
}

func (s *AreaService) DeactivateArea(ctx context.Context, areaID uuid.UUID) error {
	area, err := s.areaRepo.FindByID(ctx, areaID)
	if err != nil {
		return err
	}

	area.Deactivate()

	if err := s.areaRepo.Update(ctx, area); err != nil {
		return fmt.Errorf("failed to deactivate area: %w", err)
	}

	// Invalidate caches
	s.invalidateAreaCaches(ctx)
	s.invalidateAreaCache(ctx, areaID)

	return nil
}

// Helper methods
func (s *AreaService) invalidateAreaCache(ctx context.Context, areaID uuid.UUID) {
	cacheKey := fmt.Sprintf("area:%s", areaID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to invalidate area cache", "error", err, "area_id", areaID)
	}
}

func (s *AreaService) invalidateAreaCaches(ctx context.Context) {
	_ = s.cache.Delete(ctx, "areas:active")
}

// Request DTOs
type CreateAreaRequest struct {
	ParentID             *uuid.UUID                 `json:"parent_id,omitempty"`
	Type                 valueobjects.AreaType      `json:"type"`
	Name                 string                     `json:"name"`
	Code                 string                     `json:"code"`
	CenterLat            float64                    `json:"center_lat"`
	CenterLng            float64                    `json:"center_lng"`
	OperationalRadius    int                        `json:"operational_radius"`
	ServiceAvailability  map[string]bool            `json:"service_availability"`
	Settings             map[string]interface{}     `json:"settings"`
	SwitchingMode        entities.AreaSwitchingMode `json:"switching_mode"`
	AllowCrossTrade      bool                       `json:"allow_cross_trade"`
	Timezone             string                     `json:"timezone"`
	Languages            []string                   `json:"languages"`
	Currency             string                     `json:"currency"`
	TaxRate              float64                    `json:"tax_rate"`
	MinOrderValue        float64                    `json:"min_order_value"`
	DeliveryFeeStructure map[string]interface{}     `json:"delivery_fee_structure"`
	CrossAreaFees        map[string]float64         `json:"cross_area_fees"`
	OperationalHours     map[string]interface{}     `json:"operational_hours"`
}

type UpdateAreaRequest struct {
	Name                 *string                     `json:"name,omitempty"`
	Status               *valueobjects.AreaStatus    `json:"status,omitempty"`
	ServiceAvailability  map[string]bool             `json:"service_availability,omitempty"`
	Settings             map[string]interface{}      `json:"settings,omitempty"`
	OperationalRadius    *int                        `json:"operational_radius,omitempty"`
	SwitchingMode        *entities.AreaSwitchingMode `json:"switching_mode,omitempty"`
	AllowCrossTrade      *bool                       `json:"allow_cross_trade,omitempty"`
	MinOrderValue        *float64                    `json:"min_order_value,omitempty"`
	TaxRate              *float64                    `json:"tax_rate,omitempty"`
	DeliveryFeeStructure map[string]interface{}      `json:"delivery_fee_structure,omitempty"`
	CrossAreaFees        map[string]float64          `json:"cross_area_fees,omitempty"`
	OperationalHours     map[string]interface{}      `json:"operational_hours,omitempty"`
}
