package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

func setupAreaRepo(t *testing.T) (*areaRepository, func()) {
	// Setup test database with Bun
	testDB := database.NewTestDB(t)
	ctx := context.Background()

	// Create tables for required models
	err := testDB.CreateTables(ctx, (*models.AreaModel)(nil))
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	repo := NewAreaRepository(testDB.DB)

	cleanup := func() {
		testDB.Close()
	}

	return repo.(*areaRepository), cleanup
}

func createTestArea(t *testing.T, areaType valueobjects.AreaType) *entities.Area {
	return &entities.Area{
		ID:                   uuid.New(),
		ParentID:             nil,
		Type:                 areaType,
		Name:                 "Test Area",
		Code:                 "TEST001",
		BoundaryPolygon:      nil,
		CenterLat:            40.7128,
		CenterLng:            -74.0060,
		OperationalRadius:    10,
		Status:               valueobjects.AreaStatusActive,
		ServiceAvailability:  make(map[string]bool),
		Settings:             make(map[string]interface{}),
		SwitchingMode:        entities.SwitchingModeManual,
		AllowCrossTrade:      false,
		NearbyAreaIDs:        []uuid.UUID{},
		TransitionZones:      []entities.TransitionZone{},
		Timezone:             "UTC",
		Languages:            []string{"en"},
		Currency:             "USD",
		TaxRate:              0.0,
		MinOrderValue:        0.0,
		DeliveryFeeStructure: make(map[string]interface{}),
		CrossAreaFees:        make(map[string]float64),
		OperationalHours:     make(map[string]interface{}),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func TestAreaRepository_Save(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Test: Save area
	err := repo.Save(ctx, area)
	assert.NoError(t, err)

	// Test: Verify area was saved
	savedArea, err := repo.FindByID(ctx, area.ID)
	assert.NoError(t, err)
	assert.NotNil(t, savedArea)
	assert.Equal(t, area.ID, savedArea.ID)
	assert.Equal(t, area.Name, savedArea.Name)
	assert.Equal(t, area.Code, savedArea.Code)
}

func TestAreaRepository_FindByID(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Create area first
	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Test: Find existing area
	foundArea, err := repo.FindByID(ctx, area.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundArea)
	assert.Equal(t, area.ID, foundArea.ID)
	assert.Equal(t, area.Name, foundArea.Name)

	// Test: Find non-existent area
	_, err = repo.FindByID(ctx, uuid.New())
	assert.Error(t, err)
}

func TestAreaRepository_FindByCode(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Create area first
	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Test: Find existing area by code
	foundArea, err := repo.FindByCode(ctx, area.Code)
	assert.NoError(t, err)
	assert.NotNil(t, foundArea)
	assert.Equal(t, area.ID, foundArea.ID)
	assert.Equal(t, area.Code, foundArea.Code)

	// Test: Find non-existent area by code
	_, err = repo.FindByCode(ctx, "NONEXISTENT")
	assert.Error(t, err)
}

func TestAreaRepository_FindByParentID(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create parent area
	parentArea := createTestArea(t, valueobjects.AreaTypeCountry)
	parentArea.Name = "Parent Country"
	parentArea.Code = "PARENT001"
	err := repo.Save(ctx, parentArea)
	require.NoError(t, err)

	// Create child areas
	childArea1 := createTestArea(t, valueobjects.AreaTypeProvince)
	childArea1.ParentID = &parentArea.ID
	childArea1.Name = "Child Province 1"
	childArea1.Code = "CHILD001"

	childArea2 := createTestArea(t, valueobjects.AreaTypeProvince)
	childArea2.ParentID = &parentArea.ID
	childArea2.Name = "Child Province 2"
	childArea2.Code = "CHILD002"

	err = repo.Save(ctx, childArea1)
	require.NoError(t, err)
	err = repo.Save(ctx, childArea2)
	require.NoError(t, err)

	// Test: Find children by parent ID
	children, err := repo.FindByParentID(ctx, &parentArea.ID)
	assert.NoError(t, err)
	assert.Len(t, children, 2)

	// Verify children belong to parent
	for _, child := range children {
		assert.NotNil(t, child.ParentID)
		assert.Equal(t, parentArea.ID, *child.ParentID)
	}

	// Test: Find root areas (no parent)
	rootAreas, err := repo.FindByParentID(ctx, nil)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(rootAreas), 1) // At least the parent area
}

func TestAreaRepository_FindByType(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create areas of different types
	countryArea := createTestArea(t, valueobjects.AreaTypeCountry)
	countryArea.Name = "Test Country"
	countryArea.Code = "COUNTRY001"

	townArea := createTestArea(t, valueobjects.AreaTypeTown)
	townArea.Name = "Test Town"
	townArea.Code = "TOWN001"

	err := repo.Save(ctx, countryArea)
	require.NoError(t, err)
	err = repo.Save(ctx, townArea)
	require.NoError(t, err)

	// Test: Find areas by type
	countries, err := repo.FindByType(ctx, valueobjects.AreaTypeCountry)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(countries), 1)

	towns, err := repo.FindByType(ctx, valueobjects.AreaTypeTown)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(towns), 1)

	// Verify types are correct
	for _, country := range countries {
		assert.Equal(t, valueobjects.AreaTypeCountry, country.Type)
	}
	for _, town := range towns {
		assert.Equal(t, valueobjects.AreaTypeTown, town.Type)
	}
}

func TestAreaRepository_FindActive(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create active area
	activeArea := createTestArea(t, valueobjects.AreaTypeTown)
	activeArea.Status = valueobjects.AreaStatusActive
	activeArea.Name = "Active Area"
	activeArea.Code = "ACTIVE001"

	// Create inactive area
	inactiveArea := createTestArea(t, valueobjects.AreaTypeTown)
	inactiveArea.Status = valueobjects.AreaStatusInactive
	inactiveArea.Name = "Inactive Area"
	inactiveArea.Code = "INACTIVE001"

	err := repo.Save(ctx, activeArea)
	require.NoError(t, err)
	err = repo.Save(ctx, inactiveArea)
	require.NoError(t, err)

	// Test: Find only active areas
	activeAreas, err := repo.FindActive(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(activeAreas), 1)

	// Verify all returned areas are active
	for _, area := range activeAreas {
		assert.Equal(t, valueobjects.AreaStatusActive, area.Status)
	}
}

func TestAreaRepository_FindByCoordinate(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create area with specific coordinates
	area := createTestArea(t, valueobjects.AreaTypeTown)
	area.CenterLat = 40.7128 // New York coordinates
	area.CenterLng = -74.0060
	area.OperationalRadius = 50 // Large radius to ensure it covers the test point
	area.Name = "New York Area"
	area.Code = "NYC001"

	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Test: Find areas by coordinate (within the area)
	areas, err := repo.FindByCoordinate(ctx, 40.7128, -74.0060)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(areas), 1)

	// Test: Find areas by coordinate (far away - should return empty or distant areas)
	_, err = repo.FindByCoordinate(ctx, 0.0, 0.0)
	assert.NoError(t, err)
	// This might return areas ordered by distance, so we don't assert on count
}

func TestAreaRepository_Update(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Create area first
	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Update area
	area.Name = "Updated Area Name"
	area.Status = valueobjects.AreaStatusInactive
	area.UpdatedAt = time.Now()

	// Test: Update area
	err = repo.Update(ctx, area)
	assert.NoError(t, err)

	// Test: Verify area was updated
	updatedArea, err := repo.FindByID(ctx, area.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Area Name", updatedArea.Name)
	assert.Equal(t, valueobjects.AreaStatusInactive, updatedArea.Status)

	// Test: Update non-existent area
	nonExistentArea := createTestArea(t, valueobjects.AreaTypeTown)
	nonExistentArea.ID = uuid.New()
	err = repo.Update(ctx, nonExistentArea)
	assert.Error(t, err)
}

func TestAreaRepository_Delete(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Create area first
	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Test: Delete area
	err = repo.Delete(ctx, area.ID)
	assert.NoError(t, err)

	// Test: Verify area is deleted (should not be found)
	_, err = repo.FindByID(ctx, area.ID)
	assert.Error(t, err)

	// Test: Delete non-existent area
	err = repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
}

func TestAreaRepository_GetHierarchy(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create hierarchy: Country -> Province -> Town
	country := createTestArea(t, valueobjects.AreaTypeCountry)
	country.Name = "Test Country"
	country.Code = "COUNTRY001"
	country.ParentID = nil

	province := createTestArea(t, valueobjects.AreaTypeProvince)
	province.Name = "Test Province"
	province.Code = "PROVINCE001"
	province.ParentID = &country.ID

	town := createTestArea(t, valueobjects.AreaTypeTown)
	town.Name = "Test Town"
	town.Code = "TOWN001"
	town.ParentID = &province.ID

	// Save in order
	err := repo.Save(ctx, country)
	require.NoError(t, err)
	err = repo.Save(ctx, province)
	require.NoError(t, err)
	err = repo.Save(ctx, town)
	require.NoError(t, err)

	// Test: Get hierarchy for town (should return town -> province -> country)
	hierarchy, err := repo.GetHierarchy(ctx, town.ID)
	assert.NoError(t, err)
	assert.Len(t, hierarchy, 3)

	// Verify hierarchy order (should be from child to parent)
	assert.Equal(t, town.ID, hierarchy[0].ID)
	assert.Equal(t, province.ID, hierarchy[1].ID)
	assert.Equal(t, country.ID, hierarchy[2].ID)

	// Test: Get hierarchy for country (should return only country)
	countryHierarchy, err := repo.GetHierarchy(ctx, country.ID)
	assert.NoError(t, err)
	assert.Len(t, countryHierarchy, 1)
	assert.Equal(t, country.ID, countryHierarchy[0].ID)

	// Test: Get hierarchy for non-existent area
	_, err = repo.GetHierarchy(ctx, uuid.New())
	assert.Error(t, err)
}

func TestAreaRepository_ComplexCoordinateSearch(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create overlapping areas with different types and sizes
	country := createTestArea(t, valueobjects.AreaTypeCountry)
	country.CenterLat = 40.0
	country.CenterLng = -74.0
	country.OperationalRadius = 1000 // Very large radius
	country.Name = "Large Country"
	country.Code = "LARGE001"

	province := createTestArea(t, valueobjects.AreaTypeProvince)
	province.CenterLat = 40.5
	province.CenterLng = -74.2
	province.OperationalRadius = 100 // Medium radius
	province.Name = "Medium Province"
	province.Code = "MEDIUM001"

	town := createTestArea(t, valueobjects.AreaTypeTown)
	town.CenterLat = 40.7128
	town.CenterLng = -74.0060
	town.OperationalRadius = 25 // Small radius
	town.Name = "Small Town"
	town.Code = "SMALL001"

	// Save all areas
	err := repo.Save(ctx, country)
	require.NoError(t, err)
	err = repo.Save(ctx, province)
	require.NoError(t, err)
	err = repo.Save(ctx, town)
	require.NoError(t, err)

	// Test: Search at town center (should find all three, ordered by distance/type)
	areas, err := repo.FindByCoordinate(ctx, 40.7128, -74.0060)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(areas), 1)

	// The results should be ordered by type priority and distance
	// Exact ordering depends on the SQL implementation, but we can verify basic properties
	for _, area := range areas {
		assert.NotEmpty(t, area.Name)
		assert.NotEmpty(t, area.Code)
	}
}

func TestAreaRepository_StatusTransitions(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Create area as active
	area.Status = valueobjects.AreaStatusActive
	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Test transition to testing
	area.Status = valueobjects.AreaStatusTesting
	err = repo.Update(ctx, area)
	assert.NoError(t, err)

	// Verify status change
	updatedArea, err := repo.FindByID(ctx, area.ID)
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.AreaStatusTesting, updatedArea.Status)

	// Test transition to suspended
	area.Status = valueobjects.AreaStatusSuspended
	err = repo.Update(ctx, area)
	assert.NoError(t, err)

	// Verify it's no longer in active areas
	activeAreas, err := repo.FindActive(ctx)
	assert.NoError(t, err)
	for _, activeArea := range activeAreas {
		assert.NotEqual(t, area.ID, activeArea.ID)
	}
}

func TestAreaRepository_ConcurrentOperations(t *testing.T) {
	repo, cleanup := setupAreaRepo(t)
	defer cleanup()

	ctx := context.Background()
	area := createTestArea(t, valueobjects.AreaTypeTown)

	// Create area first
	err := repo.Save(ctx, area)
	require.NoError(t, err)

	// Test concurrent updates
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			// Get fresh area data to avoid conflicts
			freshArea, err := repo.FindByID(ctx, area.ID)
			if err == nil {
				freshArea.Name = "Updated Name"
				freshArea.UpdatedAt = time.Now()
				repo.Update(ctx, freshArea)
			}
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			// Get fresh area data to avoid conflicts
			freshArea, err := repo.FindByID(ctx, area.ID)
			if err == nil {
				if i%2 == 0 {
					freshArea.Status = valueobjects.AreaStatusActive
				} else {
					freshArea.Status = valueobjects.AreaStatusTesting
				}
				freshArea.UpdatedAt = time.Now()
				repo.Update(ctx, freshArea)
			}
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify area still exists and is in a consistent state
	finalArea, err := repo.FindByID(ctx, area.ID)
	assert.NoError(t, err)
	assert.NotNil(t, finalArea)
	assert.NotEmpty(t, finalArea.Name)
	assert.True(t, finalArea.Status.IsValid())
}
