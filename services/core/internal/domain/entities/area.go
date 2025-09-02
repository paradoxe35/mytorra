package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

type Area struct {
	ID       uuid.UUID  `db:"id" json:"id"`
	ParentID *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`

	// Hierarchy
	Type valueobjects.AreaType `db:"type" json:"type"`
	Name string                `db:"name" json:"name"`
	Code string                `db:"code" json:"code"` // Unique code for the area

	// Geographic data
	BoundaryPolygon   []valueobjects.Coordinates `db:"boundary_polygon" json:"boundary_polygon"`
	CenterLat         float64                    `db:"center_lat" json:"center_lat"`
	CenterLng         float64                    `db:"center_lng" json:"center_lng"`
	OperationalRadius int                        `db:"operational_radius" json:"operational_radius"` // in kilometers

	// Configuration
	Status              valueobjects.AreaStatus `db:"status" json:"status"`
	ServiceAvailability map[string]bool         `db:"service_availability" json:"service_availability"`
	Settings            map[string]interface{}  `db:"settings" json:"settings"`

	// Area Behavior Settings
	SwitchingMode   AreaSwitchingMode `db:"switching_mode" json:"switching_mode"`
	AllowCrossTrade bool              `db:"allow_cross_trade" json:"allow_cross_trade"`
	NearbyAreaIDs   []uuid.UUID       `db:"nearby_area_ids" json:"nearby_area_ids"`
	TransitionZones []TransitionZone  `db:"transition_zones" json:"transition_zones"`

	// Operational data
	Timezone  string   `db:"timezone" json:"timezone"`
	Languages []string `db:"languages" json:"languages"`
	Currency  string   `db:"currency" json:"currency"`
	TaxRate   float64  `db:"tax_rate" json:"tax_rate"`

	// Service limits
	MinOrderValue        float64                `db:"min_order_value" json:"min_order_value"`
	DeliveryFeeStructure map[string]interface{} `db:"delivery_fee_structure" json:"delivery_fee_structure"`
	CrossAreaFees        map[string]float64     `db:"cross_area_fees" json:"cross_area_fees"`
	OperationalHours     map[string]interface{} `db:"operational_hours" json:"operational_hours"`

	// Metadata
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	ActivatedAt   *time.Time `db:"activated_at" json:"activated_at,omitempty"`
	DeactivatedAt *time.Time `db:"deactivated_at" json:"deactivated_at,omitempty"`

	// Relationships (loaded separately)
	Children []Area `db:"-" json:"children,omitempty"`
	Parent   *Area  `db:"-" json:"parent,omitempty"`
}

type AreaSwitchingMode string

const (
	SwitchingModeAutomatic AreaSwitchingMode = "automatic"
	SwitchingModeManual    AreaSwitchingMode = "manual"
	SwitchingModeHybrid    AreaSwitchingMode = "hybrid"
	SwitchingModeLocked    AreaSwitchingMode = "locked"
)

type TransitionZone struct {
	AreaID           uuid.UUID                  `json:"area_id"`
	OverlapPolygon   []valueobjects.Coordinates `json:"overlap_polygon"`
	TransitionRadius float64                    `json:"transition_radius"` // in kilometers
}

// Domain methods
func (a *Area) IsActive() bool {
	return a.Status == valueobjects.AreaStatusActive
}

func (a *Area) IsServiceAvailable(service string) bool {
	if !a.IsActive() {
		return false
	}
	available, exists := a.ServiceAvailability[service]
	return exists && available
}

func (a *Area) GetFullPath() []string {
	path := []string{a.Name}
	if a.Parent != nil {
		path = append(a.Parent.GetFullPath(), path...)
	}
	return path
}

func (a *Area) ContainsCoordinate(lat, lng float64) bool {
	// Simplified point-in-polygon check using ray casting algorithm
	// In production, use a proper geographic library
	if len(a.BoundaryPolygon) < 3 {
		// If no proper polygon, use center and radius
		center := valueobjects.Coordinates{Latitude: a.CenterLat, Longitude: a.CenterLng}
		point, _ := valueobjects.NewCoordinates(lat, lng)
		distance := center.DistanceTo(point)
		return distance <= float64(a.OperationalRadius)
	}

	// Ray casting algorithm for point in polygon
	inside := false
	p1 := a.BoundaryPolygon[0]
	n := len(a.BoundaryPolygon)

	for i := 1; i <= n; i++ {
		p2 := a.BoundaryPolygon[i%n]
		if lng > min(p1.Longitude, p2.Longitude) {
			if lng <= max(p1.Longitude, p2.Longitude) {
				if lat <= max(p1.Latitude, p2.Latitude) {
					if p1.Longitude != p2.Longitude {
						xinters := (lng-p1.Longitude)*(p2.Latitude-p1.Latitude)/(p2.Longitude-p1.Longitude) + p1.Latitude
						if p1.Latitude == p2.Latitude || lat <= xinters {
							inside = !inside
						}
					}
				}
			}
		}
		p1 = p2
	}
	return inside
}

func (a *Area) CanAddSubArea() bool {
	switch a.Type {
	case valueobjects.AreaTypeCountry:
		return true // Can add provinces
	case valueobjects.AreaTypeProvince:
		return true // Can add towns
	case valueobjects.AreaTypeTown:
		return true // Can add subareas
	case valueobjects.AreaTypeSubarea:
		return false // Cannot add more levels
	default:
		return false
	}
}

func (a *Area) Activate() {
	now := time.Now()
	a.Status = valueobjects.AreaStatusActive
	a.ActivatedAt = &now
	a.DeactivatedAt = nil
	a.UpdatedAt = now
}

func (a *Area) Deactivate() {
	now := time.Now()
	a.Status = valueobjects.AreaStatusInactive
	a.DeactivatedAt = &now
	a.UpdatedAt = now
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
