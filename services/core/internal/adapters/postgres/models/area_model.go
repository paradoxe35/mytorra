package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/types"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// AreaModel represents the Bun model for the areas table
type AreaModel struct {
	database.BaseModel `bun:"table:areas,alias:a"`

	ParentID *uuid.UUID `bun:"parent_id,type:uuid"`
	Parent   *AreaModel `bun:"rel:belongs-to,join:parent_id=id"`

	// Hierarchy
	Type string `bun:"type,notnull"`
	Name string `bun:"name,notnull"`
	Code string `bun:"code,unique"`

	// Geographic data using PostGIS types
	Boundary          *database.Polygon `bun:"boundary,type:geometry(Polygon,4326)"`
	CenterPoint       *database.Point   `bun:"center_point,type:geometry(Point,4326)"`
	OperationalRadius int               `bun:"operational_radius"`

	// Legacy columns for backward compatibility
	BoundaryPolygon types.CoordinatesSlice `bun:"boundary_polygon,type:jsonb"`
	CenterLat       float64                `bun:"center_lat,type:decimal(10,8)"`
	CenterLng       float64                `bun:"center_lng,type:decimal(11,8)"`

	// Configuration
	Status              string            `bun:"status,notnull,default:'active'"`
	ServiceAvailability types.JSONBoolMap `bun:"service_availability,type:jsonb"`
	Settings            types.JSONMap     `bun:"settings,type:jsonb"`

	// Area Behavior Settings
	SwitchingMode   string                    `bun:"switching_mode,notnull,default:'manual'"`
	AllowCrossTrade bool                      `bun:"allow_cross_trade,default:false"`
	NearbyAreaIDs   pq.StringArray            `bun:"nearby_area_ids,type:uuid[],array"`
	TransitionZones types.TransitionZoneSlice `bun:"transition_zones,type:jsonb"`

	// Operational data
	Timezone  string         `bun:"timezone"`
	Languages pq.StringArray `bun:"languages,type:text[],array"`
	Currency  string         `bun:"currency,default:'USD'"`
	TaxRate   float64        `bun:"tax_rate,type:decimal(5,4)"`

	// Service limits
	MinOrderValue        float64            `bun:"min_order_value,type:decimal(10,2)"`
	DeliveryFeeStructure types.JSONMap      `bun:"delivery_fee_structure,type:jsonb"`
	CrossAreaFees        types.JSONFloatMap `bun:"cross_area_fees,type:jsonb"`
	OperationalHours     types.JSONMap      `bun:"operational_hours,type:jsonb"`

	// Metadata
	ActivatedAt   *time.Time `bun:"activated_at"`
	DeactivatedAt *time.Time `bun:"deactivated_at"`
}

// BeforeAppendModel is called before inserting or updating the model
func (a *AreaModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery, *bun.UpdateQuery:
		// Sync PostGIS point from lat/lng if provided
		if a.CenterLat != 0 && a.CenterLng != 0 && a.CenterPoint == nil {
			a.CenterPoint = &database.Point{
				Lng: a.CenterLng,
				Lat: a.CenterLat,
			}
		}
	}
	return nil
}

// ToEntity converts the GORM model to domain entity
func (a *AreaModel) ToEntity() (*entities.Area, error) {
	area := &entities.Area{
		ID:                   a.ID,
		ParentID:             a.ParentID,
		Type:                 valueobjects.AreaType(a.Type),
		Name:                 a.Name,
		Code:                 a.Code,
		BoundaryPolygon:      []valueobjects.Coordinates(a.BoundaryPolygon),
		CenterLat:            a.CenterLat,
		CenterLng:            a.CenterLng,
		OperationalRadius:    a.OperationalRadius,
		Status:               valueobjects.AreaStatus(a.Status),
		ServiceAvailability:  map[string]bool(a.ServiceAvailability),
		Settings:             map[string]interface{}(a.Settings),
		SwitchingMode:        entities.AreaSwitchingMode(a.SwitchingMode),
		AllowCrossTrade:      a.AllowCrossTrade,
		Timezone:             a.Timezone,
		Languages:            []string(a.Languages),
		Currency:             a.Currency,
		TaxRate:              a.TaxRate,
		MinOrderValue:        a.MinOrderValue,
		DeliveryFeeStructure: map[string]interface{}(a.DeliveryFeeStructure),
		CrossAreaFees:        map[string]float64(a.CrossAreaFees),
		OperationalHours:     map[string]interface{}(a.OperationalHours),
		CreatedAt:            a.CreatedAt,
		UpdatedAt:            a.UpdatedAt,
		ActivatedAt:          a.ActivatedAt,
		DeactivatedAt:        a.DeactivatedAt,
	}

	// Convert NearbyAreaIDs
	if len(a.NearbyAreaIDs) > 0 {
		area.NearbyAreaIDs = make([]uuid.UUID, len(a.NearbyAreaIDs))
		for i, idStr := range a.NearbyAreaIDs {
			id, err := uuid.Parse(idStr)
			if err != nil {
				continue
			}
			area.NearbyAreaIDs[i] = id
		}
	}

	// Convert TransitionZones
	if a.TransitionZones != nil {
		area.TransitionZones = make([]entities.TransitionZone, len(a.TransitionZones))
		for i, tz := range a.TransitionZones {
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

// FromEntity creates a GORM model from domain entity
func AreaFromEntity(area *entities.Area) *AreaModel {
	model := &AreaModel{
		BaseModel: database.BaseModel{
			ID:        area.ID,
			CreatedAt: area.CreatedAt,
			UpdatedAt: area.UpdatedAt,
		},
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

	// Create PostGIS point from coordinates
	if area.CenterLat != 0 && area.CenterLng != 0 {
		model.CenterPoint = &database.Point{
			Lng: area.CenterLng,
			Lat: area.CenterLat,
		}
	}

	// Convert BoundaryPolygon to PostGIS Polygon if needed
	// This requires mapping from valueobjects.Coordinates to database.Polygon format

	return model
}
