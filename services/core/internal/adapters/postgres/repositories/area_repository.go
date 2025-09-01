package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/paradoxe35/torra/packages/database"
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
	var area entities.Area
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

	err := r.db.GetContext(ctx, &area, query, id)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find area by id: %w", err)
	}

	return &area, nil
}

func (r *areaRepository) FindByCode(ctx context.Context, code string) (*entities.Area, error) {
	var area entities.Area
	query := `
		SELECT id, name, code, type, parent_id, boundaries,
		       center_lat, center_lng, radius_km, status, metadata,
		       created_at, updated_at
		FROM areas
		WHERE code = $1`

	err := r.db.GetContext(ctx, &area, query, code)
	if err != nil {
		if database.IsNoRowsError(err) {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find area by code: %w", err)
	}

	return &area, nil
}

func (r *areaRepository) FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error) {
	var areas []*entities.Area

	qb := database.NewQueryBuilder(`
		SELECT id, name, code, type, parent_id, boundaries,
		       center_lat, center_lng, radius_km, status, metadata,
		       created_at, updated_at
		FROM areas`)

	if parentID == nil {
		qb.Where("parent_id IS NULL")
	} else {
		qb.Where("parent_id = $1", *parentID)
	}

	qb.OrderBy("name ASC")

	query, args := qb.Build()
	err := r.db.SelectContext(ctx, &areas, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find areas by parent id: %w", err)
	}

	return areas, nil
}

func (r *areaRepository) FindByType(ctx context.Context, areaType valueobjects.AreaType) ([]*entities.Area, error) {
	var areas []*entities.Area
	query := `
		SELECT id, name, code, type, parent_id, boundaries,
		       center_lat, center_lng, radius_km, status, metadata,
		       created_at, updated_at
		FROM areas
		WHERE type = $1
		ORDER BY name ASC`

	err := r.db.SelectContext(ctx, &areas, query, areaType)
	if err != nil {
		return nil, fmt.Errorf("failed to find areas by type: %w", err)
	}

	return areas, nil
}

func (r *areaRepository) FindActive(ctx context.Context) ([]*entities.Area, error) {
	var areas []*entities.Area
	query := `
		SELECT id, name, code, type, parent_id, boundaries,
		       center_lat, center_lng, radius_km, status, metadata,
		       created_at, updated_at
		FROM areas
		WHERE status = $1
		ORDER BY type, name ASC`

	err := r.db.SelectContext(ctx, &areas, query, valueobjects.AreaStatusActive)
	if err != nil {
		return nil, fmt.Errorf("failed to find active areas: %w", err)
	}

	return areas, nil
}

func (r *areaRepository) FindByCoordinate(ctx context.Context, lat, lng float64) ([]*entities.Area, error) {
	var areas []*entities.Area

	// Use the PostGIS function created in migration 006
	// This provides better encapsulation and reusability
	query := `
		SELECT
			f.id, f.name, f.code, f.type, f.parent_id,
			a.boundary_polygon AS boundaries,
			f.center_lat, f.center_lng,
			f.operational_radius AS radius_km,
			f.status,
			a.service_availability AS metadata,
			a.switching_mode, a.allow_cross_trade, a.nearby_area_ids,
			a.transition_zones, a.timezone, a.languages, a.currency,
			a.tax_rate, a.min_order_value, a.delivery_fee_structure,
			a.cross_area_fees, a.operational_hours,
			a.created_at, a.updated_at, a.activated_at, a.deactivated_at,
			f.distance_km
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

	err := r.db.SelectContext(ctx, &areas, query, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("failed to find areas by coordinate: %w", err)
	}

	return areas, nil
}

func (r *areaRepository) Save(ctx context.Context, area *entities.Area) error {
	query := `
		INSERT INTO areas (
			id, name, code, type, parent_id, boundaries,
			center_lat, center_lng, radius_km, status, metadata,
			created_at, updated_at
		) VALUES (
			:id, :name, :code, :type, :parent_id, :boundaries,
			:center_lat, :center_lng, :radius_km, :status, :metadata,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, area)
	if err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}

	return nil
}

func (r *areaRepository) Update(ctx context.Context, area *entities.Area) error {
	query := `
		UPDATE areas SET
			name = :name,
			code = :code,
			type = :type,
			parent_id = :parent_id,
			boundaries = :boundaries,
			center_lat = :center_lat,
			center_lng = :center_lng,
			radius_km = :radius_km,
			status = :status,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, area)
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
	var areas []*entities.Area

	// Recursive CTE to get the full hierarchy from child to parent
	query := `
		WITH RECURSIVE area_hierarchy AS (
			SELECT id, name, code, type, parent_id, boundaries,
			       center_lat, center_lng, radius_km, status, metadata,
			       created_at, updated_at, 0 as level
			FROM areas
			WHERE id = $1

			UNION ALL

			SELECT a.id, a.name, a.code, a.type, a.parent_id, a.boundaries,
			       a.center_lat, a.center_lng, a.radius_km, a.status, a.metadata,
			       a.created_at, a.updated_at, ah.level + 1
			FROM areas a
			INNER JOIN area_hierarchy ah ON a.id = ah.parent_id
		)
		SELECT id, name, code, type, parent_id, boundaries,
		       center_lat, center_lng, radius_km, status, metadata,
		       created_at, updated_at
		FROM area_hierarchy
		ORDER BY level DESC`

	err := r.db.SelectContext(ctx, &areas, query, areaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get area hierarchy: %w", err)
	}

	return areas, nil
}

// Note: calculateDistance function removed as it's no longer needed
// PostGIS ST_Distance function provides more accurate distance calculations

// InTransaction executes a function within a database transaction
func (r *areaRepository) InTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return r.db.InTransaction(ctx, fn)
}
