package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

type areaRepository struct {
	db *database.DB
}

// NewAreaRepository creates a new Bun-based area repository
func NewAreaRepository(db *database.DB) ports.AreaRepository {
	return &areaRepository{db: db}
}

func (r *areaRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Area, error) {
	var areaModel models.AreaModel

	err := r.db.NewSelect().
		Model(&areaModel).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find area by id: %w", err)
	}

	return areaModel.ToEntity()
}

func (r *areaRepository) FindByCode(ctx context.Context, code string) (*entities.Area, error) {
	var areaModel models.AreaModel

	err := r.db.NewSelect().
		Model(&areaModel).
		Where("code = ?", code).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find area by code: %w", err)
	}

	return areaModel.ToEntity()
}

func (r *areaRepository) FindByParentID(ctx context.Context, parentID *uuid.UUID) ([]*entities.Area, error) {
	var areaModels []models.AreaModel

	query := r.db.NewSelect().
		Model(&areaModels)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Order("name ASC").Scan(ctx)
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

	err := r.db.NewSelect().
		Model(&areaModels).
		Where("type = ?", string(areaType)).
		Order("name ASC").
		Scan(ctx)
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

	err := r.db.NewSelect().
		Model(&areaModels).
		Where("status = ? AND activated_at IS NOT NULL", valueobjects.AreaStatusActive).
		Order("name ASC").
		Scan(ctx)
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
	var areaModel models.AreaModel

	// Using PostGIS ST_Contains to find area containing the point
	err := r.db.NewSelect().
		Model(&areaModel).
		Where("ST_Contains(boundary, ST_SetSRID(ST_MakePoint(?, ?), 4326))", lng, lat).
		WhereOr("ST_DWithin(center_point, ST_SetSRID(ST_MakePoint(?, ?), 4326), operational_radius * 1000)", lng, lat).
		Where("status = ?", valueobjects.AreaStatusActive).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*entities.Area{}, nil
		}
		return nil, fmt.Errorf("failed to find area by location: %w", err)
	}

	area, err := areaModel.ToEntity()
	if err != nil {
		return nil, err
	}
	return []*entities.Area{area}, nil
}

func (r *areaRepository) FindNearbyAreas(ctx context.Context, areaID uuid.UUID, maxDistanceKm float64) ([]*entities.Area, error) {
	var areaModels []models.AreaModel

	// Find nearby areas using PostGIS ST_DWithin
	subQuery := r.db.NewSelect().
		TableExpr("areas").
		Column("center_point").
		Where("id = ?", areaID)

	err := r.db.NewSelect().
		Model(&areaModels).
		Where("id != ?", areaID).
		Where("status = ?", valueobjects.AreaStatusActive).
		Where("ST_DWithin(center_point, (?), ?)", subQuery, maxDistanceKm*1000).
		OrderExpr("ST_Distance(center_point, (SELECT center_point FROM areas WHERE id = ?)) ASC", areaID).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby areas: %w", err)
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
	model := models.AreaFromEntity(area)

	// Use Upsert to handle both insert and update
	_, err := r.db.NewInsert().
		Model(model).
		On("CONFLICT (id) DO UPDATE").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save area: %w", err)
	}

	// Update the entity with any generated values
	savedArea, err := model.ToEntity()
	if err != nil {
		return fmt.Errorf("failed to convert saved area: %w", err)
	}

	*area = *savedArea
	return nil
}

func (r *areaRepository) Update(ctx context.Context, area *entities.Area) error {
	model := models.AreaFromEntity(area)

	_, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", area.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update area: %w", err)
	}

	return nil
}

func (r *areaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.NewDelete().
		Model(&models.AreaModel{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete area: %w", err)
	}

	return nil
}

func (r *areaRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model(&models.AreaModel{}).
		Where("code = ?", code).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check area existence: %w", err)
	}

	return exists, nil
}

func (r *areaRepository) GetHierarchy(ctx context.Context, areaID uuid.UUID) ([]*entities.Area, error) {
	var areaModels []models.AreaModel

	// Recursive CTE to get all parent areas
	err := r.db.NewRaw(`
		WITH RECURSIVE area_hierarchy AS (
			SELECT * FROM areas WHERE id = ?
			UNION ALL
			SELECT a.* FROM areas a
			INNER JOIN area_hierarchy ah ON a.id = ah.parent_id
		)
		SELECT * FROM area_hierarchy ORDER BY type ASC
	`, areaID).
		Scan(ctx, &areaModels)
	if err != nil {
		return nil, fmt.Errorf("failed to get area hierarchy: %w", err)
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
