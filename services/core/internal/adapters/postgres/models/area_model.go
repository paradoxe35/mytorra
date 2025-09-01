package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/types"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// AreaModel represents the database model for the areas table
type AreaModel struct {
	ID       uuid.UUID  `db:"id"`
	ParentID *uuid.UUID `db:"parent_id"`

	// Hierarchy
	Type string `db:"type"`
	Name string `db:"name"`
	Code string `db:"code"`

	// Geographic data
	BoundaryPolygon   types.CoordinatesSlice `db:"boundary_polygon"`
	CenterLat         float64                `db:"center_lat"`
	CenterLng         float64                `db:"center_lng"`
	OperationalRadius int                    `db:"operational_radius"`

	// Configuration
	Status              string            `db:"status"`
	ServiceAvailability types.JSONBoolMap `db:"service_availability"`
	Settings            types.JSONMap     `db:"settings"`

	// Area Behavior Settings
	SwitchingMode   string                    `db:"switching_mode"`
	AllowCrossTrade bool                      `db:"allow_cross_trade"`
	NearbyAreaIDs   pq.StringArray            `db:"nearby_area_ids"`
	TransitionZones types.TransitionZoneSlice `db:"transition_zones"`

	// Operational data
	Timezone  string         `db:"timezone"`
	Languages pq.StringArray `db:"languages"`
	Currency  string         `db:"currency"`
	TaxRate   float64        `db:"tax_rate"`

	// Service limits
	MinOrderValue        float64            `db:"min_order_value"`
	DeliveryFeeStructure types.JSONMap      `db:"delivery_fee_structure"`
	CrossAreaFees        types.JSONFloatMap `db:"cross_area_fees"`
	OperationalHours     types.JSONMap      `db:"operational_hours"`

	// Metadata
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	ActivatedAt   *time.Time `db:"activated_at"`
	DeactivatedAt *time.Time `db:"deactivated_at"`
}

// ToEntity converts AreaModel to entities.Area
func (m *AreaModel) ToEntity() (*entities.Area, error) {
	area := &entities.Area{
		ID:                   m.ID,
		ParentID:             m.ParentID,
		Type:                 valueobjects.AreaType(m.Type),
		Name:                 m.Name,
		Code:                 m.Code,
		BoundaryPolygon:      []valueobjects.Coordinates(m.BoundaryPolygon),
		CenterLat:            m.CenterLat,
		CenterLng:            m.CenterLng,
		OperationalRadius:    m.OperationalRadius,
		Status:               valueobjects.AreaStatus(m.Status),
		ServiceAvailability:  map[string]bool(m.ServiceAvailability),
		Settings:             map[string]interface{}(m.Settings),
		SwitchingMode:        entities.AreaSwitchingMode(m.SwitchingMode),
		AllowCrossTrade:      m.AllowCrossTrade,
		Timezone:             m.Timezone,
		Languages:            []string(m.Languages),
		Currency:             m.Currency,
		TaxRate:              m.TaxRate,
		MinOrderValue:        m.MinOrderValue,
		DeliveryFeeStructure: map[string]interface{}(m.DeliveryFeeStructure),
		CrossAreaFees:        map[string]float64(m.CrossAreaFees),
		OperationalHours:     map[string]interface{}(m.OperationalHours),
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
		ActivatedAt:          m.ActivatedAt,
		DeactivatedAt:        m.DeactivatedAt,
	}

	// Convert NearbyAreaIDs
	if len(m.NearbyAreaIDs) > 0 {
		area.NearbyAreaIDs = make([]uuid.UUID, len(m.NearbyAreaIDs))
		for i, idStr := range m.NearbyAreaIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			area.NearbyAreaIDs[i] = id
		}
	}

	// Convert TransitionZones
	if m.TransitionZones != nil {
		area.TransitionZones = make([]entities.TransitionZone, len(m.TransitionZones))
		for i, tz := range m.TransitionZones {
			areaID, err := uuid.Parse(tz.AreaID)
			if err != nil {
				continue
			}
			area.TransitionZones[i] = entities.TransitionZone{
				AreaID:           areaID,
				OverlapPolygon:   tz.OverlapPolygon,
				TransitionRadius: tz.TransitionRadius,
			}
		}
	}

	return area, nil
}

// FromEntity converts entities.Area to AreaModel
func FromEntity(area *entities.Area) *AreaModel {
	model := &AreaModel{
		ID:                   area.ID,
		ParentID:             area.ParentID,
		Type:                 string(area.Type),
		Name:                 area.Name,
		Code:                 area.Code,
		BoundaryPolygon:      types.CoordinatesSlice(area.BoundaryPolygon),
		CenterLat:            area.CenterLat,
		CenterLng:            area.CenterLng,
		OperationalRadius:    area.OperationalRadius,
		Status:               string(area.Status),
		ServiceAvailability:  types.JSONBoolMap(area.ServiceAvailability),
		Settings:             types.JSONMap(area.Settings),
		SwitchingMode:        string(area.SwitchingMode),
		AllowCrossTrade:      area.AllowCrossTrade,
		Timezone:             area.Timezone,
		Languages:            pq.StringArray(area.Languages),
		Currency:             area.Currency,
		TaxRate:              area.TaxRate,
		MinOrderValue:        area.MinOrderValue,
		DeliveryFeeStructure: types.JSONMap(area.DeliveryFeeStructure),
		CrossAreaFees:        types.JSONFloatMap(area.CrossAreaFees),
		OperationalHours:     types.JSONMap(area.OperationalHours),
		CreatedAt:            area.CreatedAt,
		UpdatedAt:            area.UpdatedAt,
		ActivatedAt:          area.ActivatedAt,
		DeactivatedAt:        area.DeactivatedAt,
	}

	// Convert NearbyAreaIDs
	if len(area.NearbyAreaIDs) > 0 {
		model.NearbyAreaIDs = make(pq.StringArray, len(area.NearbyAreaIDs))
		for i, id := range area.NearbyAreaIDs {
			model.NearbyAreaIDs[i] = id.String()
		}
	}

	// Convert TransitionZones
	if len(area.TransitionZones) > 0 {
		model.TransitionZones = make(types.TransitionZoneSlice, len(area.TransitionZones))
		for i, tz := range area.TransitionZones {
			model.TransitionZones[i] = struct {
				AreaID           string                     `json:"area_id"`
				OverlapPolygon   []valueobjects.Coordinates `json:"overlap_polygon"`
				TransitionRadius float64                    `json:"transition_radius"`
			}{
				AreaID:           tz.AreaID.String(),
				OverlapPolygon:   tz.OverlapPolygon,
				TransitionRadius: tz.TransitionRadius,
			}
		}
	}

	return model
}
