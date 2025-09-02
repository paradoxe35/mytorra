package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/restayway/gogis"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/types"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// AreaModel represents the GORM model for the areas table
type AreaModel struct {
	database.BaseModel
	ParentID *uuid.UUID `gorm:"type:uuid"`
	Parent   *AreaModel `gorm:"foreignKey:ParentID"`

	// Hierarchy
	Type string `gorm:"type:varchar(20);not null"`
	Name string `gorm:"type:varchar(255);not null"`
	Code string `gorm:"type:varchar(50);uniqueIndex"`

	// Geographic data using gogis types for PostGIS
	Boundary          gogis.Polygon `gorm:"type:geometry(Polygon,4326)"`
	CenterPoint       gogis.Point   `gorm:"type:geometry(Point,4326)"`
	OperationalRadius int           `gorm:"type:int"`

	// Legacy columns for backward compatibility
	BoundaryPolygon types.CoordinatesSlice `gorm:"type:jsonb"`
	CenterLat       float64                `gorm:"type:decimal(10,8)"`
	CenterLng       float64                `gorm:"type:decimal(11,8)"`

	// Configuration
	Status              string            `gorm:"type:varchar(20);not null;default:'active'"`
	ServiceAvailability types.JSONBoolMap `gorm:"type:jsonb"`
	Settings            types.JSONMap     `gorm:"type:jsonb"`

	// Area Behavior Settings
	SwitchingMode   string                    `gorm:"type:varchar(20);not null;default:'manual'"`
	AllowCrossTrade bool                      `gorm:"type:boolean;default:false"`
	NearbyAreaIDs   pq.StringArray            `gorm:"type:uuid[]"`
	TransitionZones types.TransitionZoneSlice `gorm:"type:jsonb"`

	// Operational data
	Timezone  string         `gorm:"type:varchar(50)"`
	Languages pq.StringArray `gorm:"type:text[]"`
	Currency  string         `gorm:"type:varchar(3);default:'USD'"`
	TaxRate   float64        `gorm:"type:decimal(5,4)"`

	// Service limits
	MinOrderValue        float64            `gorm:"type:decimal(10,2)"`
	DeliveryFeeStructure types.JSONMap      `gorm:"type:jsonb"`
	CrossAreaFees        types.JSONFloatMap `gorm:"type:jsonb"`
	OperationalHours     types.JSONMap      `gorm:"type:jsonb"`

	// Metadata
	ActivatedAt   *time.Time `gorm:"type:timestamp"`
	DeactivatedAt *time.Time `gorm:"type:timestamp"`
}

// TableName specifies the table name for the Area model
func (AreaModel) TableName() string {
	return "areas"
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

	// Create PostGIS point and polygon from coordinates
	if area.CenterLat != 0 && area.CenterLng != 0 {
		model.CenterPoint = gogis.Point{
			Lng: area.CenterLng,
			Lat: area.CenterLat,
		}
	}

	// Convert BoundaryPolygon to gogis.Polygon if needed
	// This requires mapping from valueobjects.Coordinates to gogis format

	return model
}
