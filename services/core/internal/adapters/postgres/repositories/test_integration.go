//go:build integration
// +build integration

package gorm_repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/restayway/gogis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/torra/packages/config"
	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

// TestGORMWithPostGIS tests GORM integration with PostGIS spatial features
func TestGORMWithPostGIS(t *testing.T) {
	// Setup test database
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "torra_test",
		Username: "postgres",
		Password: "postgres",
	}

	db, err := database.NewGormDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	// Auto-migrate test tables
	err = db.AutoMigrate(&models.Area{}, &models.User{})
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Area with PostGIS", func(t *testing.T) {
		repo := NewAreaRepository(db)

		// Create test area with PostGIS data
		area := &entities.Area{
			ID:                uuid.New(),
			Type:              valueobjects.AreaTypeCity,
			Name:              "Test City",
			Code:              "TEST001",
			CenterLat:         40.7128,
			CenterLng:         -74.0060,
			OperationalRadius: 10,
			Status:            valueobjects.AreaStatusActive,
		}

		// Save area
		err := repo.Save(ctx, area)
		assert.NoError(t, err)

		// Find by location (within operational radius)
		areas, err := repo.FindByCoordinate(ctx, 40.7200, -74.0100)
		assert.NoError(t, err)
		assert.Len(t, areas, 1)
		assert.Equal(t, "TEST001", areas[0].Code)

		// Find by location (outside operational radius)
		areas, err = repo.FindByCoordinate(ctx, 41.0000, -75.0000)
		assert.NoError(t, err)
		assert.Len(t, areas, 0)

		// Find nearby areas
		nearbyAreas, err := repo.FindNearbyAreas(ctx, area.ID, 50)
		assert.NoError(t, err)
		// Should not include itself
		for _, nearby := range nearbyAreas {
			assert.NotEqual(t, area.ID, nearby.ID)
		}

		// Clean up
		err = repo.Delete(ctx, area.ID)
		assert.NoError(t, err)
	})

	t.Run("User with PostGIS location", func(t *testing.T) {
		repo := NewUserRepository(db)

		// Create test user with location
		email, _ := valueobjects.NewEmail("test@example.com")
		lat := 40.7128
		lng := -74.0060

		user := &entities.User{
			ID:           uuid.New(),
			Email:        email,
			Name:         "Test User",
			PasswordHash: "hashed",
			LastKnownLat: &lat,
			LastKnownLng: &lng,
			Status:       entities.UserStatusActive,
		}

		// Save user
		err := repo.Save(ctx, user)
		assert.NoError(t, err)

		// Update location
		err = repo.UpdateLocation(ctx, user.ID, 40.7500, -73.9900)
		assert.NoError(t, err)

		// Verify location was updated
		updatedUser, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser.LastKnownLat)
		assert.NotNil(t, updatedUser.LastKnownLng)
		assert.InDelta(t, 40.7500, *updatedUser.LastKnownLat, 0.0001)
		assert.InDelta(t, -73.9900, *updatedUser.LastKnownLng, 0.0001)

		// Clean up
		err = repo.Delete(ctx, user.ID)
		assert.NoError(t, err)
	})

	t.Run("Spatial queries with gogis", func(t *testing.T) {
		// Create multiple areas
		areas := []models.Area{
			{
				ID:          uuid.New(),
				Type:        "city",
				Name:        "New York",
				Code:        "NYC",
				CenterPoint: gogis.Point{Lat: 40.7128, Lng: -74.0060},
				Status:      "active",
			},
			{
				ID:          uuid.New(),
				Type:        "city",
				Name:        "Philadelphia",
				Code:        "PHL",
				CenterPoint: gogis.Point{Lat: 39.9526, Lng: -75.1652},
				Status:      "active",
			},
			{
				ID:          uuid.New(),
				Type:        "city",
				Name:        "Boston",
				Code:        "BOS",
				CenterPoint: gogis.Point{Lat: 42.3601, Lng: -71.0589},
				Status:      "active",
			},
		}

		// Bulk insert
		err = db.Create(&areas).Error
		assert.NoError(t, err)

		// Find nearest city to a point
		var nearest models.Area
		testLat, testLng := 40.0000, -74.5000

		err = db.
			Where("status = ?", "active").
			Order(db.Expr("ST_Distance(center_point, ST_SetSRID(ST_MakePoint(?, ?), 4326))", testLng, testLat)).
			First(&nearest).Error

		assert.NoError(t, err)
		assert.Equal(t, "NYC", nearest.Code) // NYC should be nearest

		// Find all cities within 200km
		var withinRange []models.Area
		err = db.
			Where("ST_DWithin(center_point::geography, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
				testLng, testLat, 200000). // 200km in meters
			Find(&withinRange).Error

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(withinRange), 2) // Should include NYC and PHL

		// Clean up
		for _, area := range areas {
			db.Delete(&area)
		}
	})
}
