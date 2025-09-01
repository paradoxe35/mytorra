package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type areaRepository struct {
	db *database.DB
}

// NewAreaRepository creates a new area repository
func NewAreaRepository(db *database.DB) ports.AreaRepository {
	return &areaRepository{db: db}
}

func (r *areaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Area, error) {
	var areaModel models.AreaModel
	query := `
		SELECT id, parent_id, type, name, code, boundary_polygon,
		       center_lat, center_lng, operational_radius, status,
		       service_availability, settings, switching_mode,
		       allow_cross_trade, nearby_area_ids, transition_zones,
		       timezone, languages, currency, tax_rate, min_order_value,
		       delivery_fee_structure, cross_area_fees, operational_hours,
		       created_at, updated_at, activated_at, deactivated_at
		FROM areas
		WHERE id = $1`

	err := r.db.GetContext(ctx, &areaModel, query, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find area by id: %w", err)
	}

	return areaModel.ToEntity()
}

func (r *areaRepository) FindByCode(ctx context.Context, code string) (*entities.Area, error) {
	var areaModel models.AreaModel
	query := `
		SELECT id, parent_id, type, name, code, boundary_polygon,
		       center_lat, center_lng, operational_radius, status,
		       service_availability, settings, switching_mode,
		       allow_cross_trade, nearby_area_ids, transition_zones,
		       timezone, languages, currency, tax_rate, min_order_value,
		       delivery_fee_structure, cross_area_fees, operational_hours,
		       created_at, updated_at, activated_at, deactivated_at
		FROM areas
		WHERE code = $1`

	err := r.db.GetContext(ctx, &areaModel, query, code)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find area by code: %w", err)
	}

	return areaModel.ToEntity()
}

func (r *areaRepository) FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error) {
	var areaModels []models.AreaModel

	qb := database.NewQueryBuilder(`
		SELECT id, parent_id, type, name, code, boundary_polygon,
		       center_lat, center_lng, operational_radius, status,
		       service_availability, settings, switching_mode,
		       allow_cross_trade, nearby_area_ids, transition_zones,
		       timezone, languages, currency, tax_rate, min_order_value,
		       delivery_fee_structure, cross_area_fees, operational_hours,
		       created_at, updated_at, activated_at, deactivated_at
		FROM areas`)

	if parentID == nil {
		qb.Where("parent_id IS NULL")
	} else {
		qb.Where("parent_id = $1", *parentID)
	}

	qb.OrderBy("name ASC")

	query, args := qb.Build()
	err := r.db.SelectContext(ctx, &areaModels, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find areas by parent id: %w", err)
	}

	areas := make([]*entities.Area, len(areaModels))
	for i, model := range areaModels {
		area, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert area model: %w", err)
		}
		areas[i] = area
	}

	return areas, nil
}

func (r *areaRepository) FindByType(ctx context.Context, areaType valueobjects.AreaType) ([]*entities.Area, error) {
	var areaModels []models.AreaModel
	query := `
		SELECT id, parent_id, type, name, code, boundary_polygon,
		       center_lat, center_lng, operational_radius, status,
		       service_availability, settings, switching_mode,
		       allow_cross_trade, nearby_area_ids, transition_zones,
		       timezone, languages, currency, tax_rate, min_order_value,
		       delivery_fee_structure, cross_area_fees, operational_hours,
		       created_at, updated_at, activated_at, deactivated_at
		FROM areas
		WHERE type = $1
		ORDER BY name ASC`

	err := r.db.SelectContext(ctx, &areaModels, query, areaType)
	if err != nil {
		return nil, fmt.Errorf("failed to find areas by type: %w", err)
	}

	areas := make([]*entities.Area, len(areaModels))
	for i, model := range areaModels {
		area, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert area model: %w", err)
		}
		areas[i] = area
	}

	return areas, nil
}

func (r *areaRepository) FindActive(ctx context.Context) ([]*entities.Area, error) {
	var areaModels []models.AreaModel
	query := `
		SELECT id, parent_id, type, name, code, boundary_polygon,
		       center_lat, center_lng, operational_radius, status,
		       service_availability, settings, switching_mode,
		       allow_cross_trade, nearby_area_ids, transition_zones,
		       timezone, languages, currency, tax_rate, min_order_value,
		       delivery_fee_structure, cross_area_fees, operational_hours,
		       created_at, updated_at, activated_at, deactivated_at
		FROM areas
		WHERE status = $1
		ORDER BY type, name ASC`

	err := r.db.SelectContext(ctx, &areaModels, query, valueobjects.AreaStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to find active areas: %w", err)
	}

	areas := make([]*entities.Area, len(areaModels))
	for i, model := range areaModels {
		area, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert area model: %w", err)
		}
		areas[i] = area
	}

	return areas, nil
}

func (r *areaRepository) FindByCoordinate(ctx context.Context, lat, lng float64) ([]*entities.Area, error) {
	var areaModels []models.AreaModel

	// Use the PostGIS function created in migration 006
	// This provides better encapsulation and reusability
	query := `
		SELECT
			f.id, f.parent_id, f.type, f.name, f.code,
			a.boundary_polygon,
			f.center_lat, f.center_lng,
			f.operational_radius,
			f.status,
			a.service_availability,
			a.settings, a.switching_mode, a.allow_cross_trade, a.nearby_area_ids,
			a.transition_zones, a.timezone, a.languages, a.currency,
			a.tax_rate, a.min_order_value, a.delivery_fee_structure,
			a.cross_area_fees, a.operational_hours,
			a.created_at, a.updated_at, a.activated_at, a.deactivated_at
		FROM find_areas_by_coordinate($1, $2, 50) f
		JOIN areas a ON a.id = f.id
		ORDER BY
			CASE f.type
				WHEN 'subarea' THEN 1
				WHEN 'town' THEN 2
				WHEN 'province' THEN 3
				WHEN 'country' THEN 4
			END,
			f.distance_km ASC`

	err := r.db.SelectContext(ctx, &areaModels, query, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("failed to find areas by coordinate: %w", err)
	}

	areas := make([]*entities.Area, len(areaModels))
	for i, model := range areaModels {
		area, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert area model: %w", err)
		}
		areas[i] = area
	}

	return areas, nil
}

func (r *areaRepository) Save(ctx context.Context, area *entities.Area) error {
	query := `
		INSERT INTO areas (
			id, parent_id, type, name, code, boundary_polygon,
			center_lat, center_lng, center_point, operational_radius, status,
			service_availability, settings, switching_mode,
			allow_cross_trade, nearby_area_ids, transition_zones,
			timezone, languages, currency, tax_rate, min_order_value,
			delivery_fee_structure, cross_area_fees, operational_hours,
			created_at, updated_at
		) VALUES (
			:id, :parent_id, :type, :name, :code, :boundary_polygon,
			:center_lat, :center_lng,
			ST_SetSRID(ST_MakePoint(:center_lng, :center_lat), 4326),
			:operational_radius, :status,
			:service_availability, :settings, :switching_mode,
			:allow_cross_trade, :nearby_area_ids, :transition_zones,
			:timezone, :languages, :currency, :tax_rate, :min_order_value,
			:delivery_fee_structure, :cross_area_fees, :operational_hours,
			:created_at, :updated_at
		)`

	areaModel := models.FromEntity(area)
	_, err := r.db.NamedExecContext(ctx, query, areaModel)
	if err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}

	return nil
}

func (r *areaRepository) Update(ctx context.Context, area *entities.Area) error {
	query := `
		UPDATE areas SET
			parent_id = :parent_id,
			type = :type,
			name = :name,
			code = :code,
			boundary_polygon = :boundary_polygon,
			center_lat = :center_lat,
			center_lng = :center_lng,
			center_point = ST_SetSRID(ST_MakePoint(:center_lng, :center_lat), 4326),
			operational_radius = :operational_radius,
			status = :status,
			service_availability = :service_availability,
			settings = :settings,
			switching_mode = :switching_mode,
			allow_cross_trade = :allow_cross_trade,
			nearby_area_ids = :nearby_area_ids,
			transition_zones = :transition_zones,
			timezone = :timezone,
			languages = :languages,
			currency = :currency,
			tax_rate = :tax_rate,
			min_order_value = :min_order_value,
			delivery_fee_structure = :delivery_fee_structure,
			cross_area_fees = :cross_area_fees,
			operational_hours = :operational_hours,
			updated_at = :updated_at,
			activated_at = :activated_at,
			deactivated_at = :deactivated_at
		WHERE id = :id`

	areaModel := models.FromEntity(area)
	result, err := r.db.NamedExecContext(ctx, query, areaModel)
	if err != nil {
		return fmt.Errorf("failed to update area: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("area not found")
	}

	return nil
}

func (r *areaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM areas WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete area: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("area not found")
	}

	return nil
}

func (r *areaRepository) GetHierarchy(ctx context.Context, areaID uuid.UUID) ([]*entities.Area, error) {
	var areaModels []models.AreaModel

	// Recursive CTE to get the full hierarchy from child to parent
	query := `
		WITH RECURSIVE area_hierarchy AS (
			SELECT id, parent_id, type, name, code, boundary_polygon,
			       center_lat, center_lng, operational_radius, status,
			       service_availability, settings, switching_mode,
			       allow_cross_trade, nearby_area_ids, transition_zones,
			       timezone, languages, currency, tax_rate, min_order_value,
			       delivery_fee_structure, cross_area_fees, operational_hours,
			       created_at, updated_at, activated_at, deactivated_at, 0 as level
			FROM areas
			WHERE id = $1

			UNION ALL

			SELECT a.id, a.parent_id, a.type, a.name, a.code, a.boundary_polygon,
			       a.center_lat, a.center_lng, a.operational_radius, a.status,
			       a.service_availability, a.settings, a.switching_mode,
			       a.allow_cross_trade, a.nearby_area_ids, a.transition_zones,
			       a.timezone, a.languages, a.currency, a.tax_rate, a.min_order_value,
			       a.delivery_fee_structure, a.cross_area_fees, a.operational_hours,
			       a.created_at, a.updated_at, a.activated_at, a.deactivated_at, ah.level + 1
			FROM areas a
			INNER JOIN area_hierarchy ah ON a.id = ah.parent_id
		)
		SELECT id, parent_id, type, name, code, boundary_polygon,
		       center_lat, center_lng, operational_radius, status,
		       service_availability, settings, switching_mode,
		       allow_cross_trade, nearby_area_ids, transition_zones,
		       timezone, languages, currency, tax_rate, min_order_value,
		       delivery_fee_structure, cross_area_fees, operational_hours,
		       created_at, updated_at, activated_at, deactivated_at
		FROM area_hierarchy
		ORDER BY level ASC`

	err := r.db.SelectContext(ctx, &areaModels, query, areaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get area hierarchy: %w", err)
	}

	// Check if any areas were found
	if len(areaModels) == 0 {
		return nil, fmt.Errorf("area not found")
	}

	areas := make([]*entities.Area, len(areaModels))
	for i, model := range areaModels {
		area, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert area model: %w", err)
		}
		areas[i] = area
	}

	return areas, nil
}

// InTransaction executes a function within a database transaction
func (r *areaRepository) InTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return r.db.InTransaction(ctx, fn)
}
