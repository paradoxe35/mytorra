package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

func setupUserRepo(t *testing.T) (*userRepository, *database.DB, func()) {
	// Setup test database with Bun
	testDB := database.NewTestDB(t)
	ctx := context.Background()

	// Create tables for required models
	err := testDB.CreateTables(ctx, (*models.UserModel)(nil), (*models.AreaModel)(nil))
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	repo := NewUserRepository(testDB.DB)

	cleanup := func() {
		testDB.Close()
	}

	return repo.(*userRepository), testDB.DB, cleanup
}

func createTestUser(t *testing.T, areaID uuid.UUID) *entities.User {
	email, err := valueobjects.NewEmail("test@example.com")
	require.NoError(t, err)

	phone, err := valueobjects.NewPhone("+1234567890")
	require.NoError(t, err)

	return &entities.User{
		ID:                uuid.New(),
		Email:             email,
		Phone:             phone,
		Name:              "Test User",
		PasswordHash:      "hashed_password",
		ProfileImageURL:   "",
		EmailVerified:     false,
		PhoneVerified:     false,
		IDVerified:        false,
		VerificationLevel: 0,
		GoogleID:          nil,
		FacebookID:        nil,
		MFAEnabled:        false,
		MFASecret:         nil,
		AreaID:            areaID,
		LastKnownLat:      nil,
		LastKnownLng:      nil,
		Status:            entities.UserStatusActive,
		LastLoginAt:       nil,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		DeletedAt:         nil,
	}
}

func createTestAreaForUser(t *testing.T, db *database.DB) uuid.UUID {
	areaID := uuid.New()
	ctx := context.Background()

	// Create test area using Bun
	area := &entities.Area{
		ID:                   areaID,
		Type:                 valueobjects.AreaTypeTown,
		Name:                 "Test Area",
		Code:                 "TEST001",
		CenterLat:            0,
		CenterLng:            0,
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

	areaModel := models.AreaFromEntity(area)
	_, err := db.NewInsert().Model(areaModel).Exec(ctx)
	require.NoError(t, err)

	return areaID
}

func TestUserRepository_Save(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Test: Save user
	err := repo.Save(ctx, user)
	assert.NoError(t, err)

	// Test: Verify user was saved
	savedUser, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, savedUser)
	assert.Equal(t, user.ID, savedUser.ID)
	assert.Equal(t, user.Email.String(), savedUser.Email.String())
	assert.Equal(t, user.Name, savedUser.Name)
}

func TestUserRepository_FindByID(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Find existing user
	foundUser, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Email.String(), foundUser.Email.String())

	// Test: Find non-existent user
	_, err = repo.FindByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Find existing user by email
	foundUser, err := repo.FindByEmail(ctx, user.Email)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Email.String(), foundUser.Email.String())

	// Test: Find non-existent user by email
	nonExistentEmail, err := valueobjects.NewEmail("nonexistent@example.com")
	require.NoError(t, err)
	_, err = repo.FindByEmail(ctx, nonExistentEmail)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_FindByPhone(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Find existing user by phone
	foundUser, err := repo.FindByPhone(ctx, user.Phone)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Phone.String(), foundUser.Phone.String())

	// Test: Find non-existent user by phone
	nonExistentPhone, err := valueobjects.NewPhone("+9876543210")
	require.NoError(t, err)
	_, err = repo.FindByPhone(ctx, nonExistentPhone)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_Update(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Update user
	user.Name = "Updated Name"
	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	// Test: Update user
	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// Test: Verify user was updated
	updatedUser, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
	assert.True(t, updatedUser.EmailVerified)

	// Test: Update non-existent user
	nonExistentUser := createTestUser(t, areaID)
	nonExistentUser.ID = uuid.New()
	err = repo.Update(ctx, nonExistentUser)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_Delete(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Delete user (soft delete)
	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Test: Verify user is soft deleted (should not be found in normal queries)
	_, err = repo.FindByID(ctx, user.ID)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)

	// Test: Delete non-existent user
	err = repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_ListByArea(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create test areas
	areaID := createTestAreaForUser(t, db)
	areaID2 := createTestAreaForUser(t, db)

	// Create multiple users in the same area
	user1 := createTestUser(t, areaID)
	user1.Email, _ = valueobjects.NewEmail("user1@example.com")

	user2 := createTestUser(t, areaID)
	user2.Email, _ = valueobjects.NewEmail("user2@example.com")
	user2.Phone, _ = valueobjects.NewPhone("+1234567891")

	// Create user in different area
	user3 := createTestUser(t, areaID2)
	user3.Email, _ = valueobjects.NewEmail("user3@example.com")
	user3.Phone, _ = valueobjects.NewPhone("+1234567892")

	// Save all users
	require.NoError(t, repo.Save(ctx, user1))
	require.NoError(t, repo.Save(ctx, user2))
	require.NoError(t, repo.Save(ctx, user3))

	// Test: List users by area
	users, err := repo.ListByArea(ctx, areaID.String(), 10, 0)
	assert.NoError(t, err)
	assert.Len(t, users, 2) // Only users in the specified area

	// Verify the users are from the correct area
	for _, user := range users {
		assert.Equal(t, areaID, user.AreaID)
	}

	// Test: List with pagination
	users, err = repo.ListByArea(ctx, areaID.String(), 1, 0)
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	users, err = repo.ListByArea(ctx, areaID.String(), 1, 1)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestUserRepository_ExistsByEmail(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Test: Email doesn't exist initially
	exists, err := repo.ExistsByEmail(ctx, user.Email)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create user
	err = repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Email exists after creation
	exists, err = repo.ExistsByEmail(ctx, user.Email)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test: Different email doesn't exist
	differentEmail, err := valueobjects.NewEmail("different@example.com")
	require.NoError(t, err)
	exists, err = repo.ExistsByEmail(ctx, differentEmail)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_ExistsByPhone(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Test: Phone doesn't exist initially
	exists, err := repo.ExistsByPhone(ctx, user.Phone)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Create user
	err = repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Phone exists after creation
	exists, err = repo.ExistsByPhone(ctx, user.Phone)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test: Different phone doesn't exist
	differentPhone, err := valueobjects.NewPhone("+9876543210")
	require.NoError(t, err)
	exists, err = repo.ExistsByPhone(ctx, differentPhone)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Update last login using dedicated method
	err = repo.UpdateLastLogin(ctx, user.ID)
	assert.NoError(t, err)

	// Test: Verify last login was updated
	updatedUser, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser.LastLoginAt)
	assert.WithinDuration(t, time.Now(), *updatedUser.LastLoginAt, 5*time.Second)

	// Test: Update last login for non-existent user
	err = repo.UpdateLastLogin(ctx, uuid.New())
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_UpdateLocation(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test: Update location using dedicated method
	lat := 40.7128
	lng := -74.0060
	err = repo.UpdateLocation(ctx, user.ID, lat, lng)
	assert.NoError(t, err)

	// Test: Verify location was updated
	updatedUser, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser.LastKnownLat)
	assert.NotNil(t, updatedUser.LastKnownLng)
	assert.Equal(t, lat, *updatedUser.LastKnownLat)
	assert.Equal(t, lng, *updatedUser.LastKnownLng)

	// Test: Update location for non-existent user
	err = repo.UpdateLocation(ctx, uuid.New(), 0.0, 0.0)
	assert.Error(t, err)
	assert.Equal(t, errors.ErrNotFound, err)
}

func TestUserRepository_ConcurrentOperations(t *testing.T) {
	repo, db, cleanup := setupUserRepo(t)
	defer cleanup()

	ctx := context.Background()
	// Create test area first
	areaID := createTestAreaForUser(t, db)
	user := createTestUser(t, areaID)

	// Create user first
	err := repo.Save(ctx, user)
	require.NoError(t, err)

	// Test concurrent updates (this tests database locking/transactions)
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			// Get fresh user data to avoid conflicts
			freshUser, err := repo.FindByID(ctx, user.ID)
			if err == nil {
				freshUser.Name = "Updated Name"
				freshUser.UpdatedAt = time.Now()
				repo.Update(ctx, freshUser)
			}
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			repo.UpdateLastLogin(ctx, user.ID)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify user still exists and is in a consistent state
	finalUser, err := repo.FindByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, finalUser)
	assert.NotEmpty(t, finalUser.Name)
}
