package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/paradoxe35/torra/packages/database"
	"github.com/paradoxe35/torra/packages/errors"
	"github.com/paradoxe35/torra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/torra/services/core/internal/application/ports"
	"github.com/paradoxe35/torra/services/core/internal/domain/entities"
	"github.com/paradoxe35/torra/services/core/internal/domain/valueobjects"
)

type userRepository struct {
	db *database.GormDB
}

// NewUserRepository creates a new GORM-based user repository
func NewUserRepository(db *database.GormDB) ports.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.WithContext(ctx).
		Where("email = ?", email.String()).
		First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.WithContext(ctx).
		Where("phone = ?", phone.String()).
		First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) Save(ctx context.Context, user *entities.User) error {
	model := models.UserFromEntity(user)

	err := r.db.WithContext(ctx).Create(model).Error
	if err != nil {
		if r.isDuplicateError(err) {
			return fmt.Errorf("user already exists")
		}
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Update the entity with any generated values
	savedUser, err := model.ToEntity()
	if err != nil {
		return fmt.Errorf("failed to convert saved user: %w", err)
	}

	*user = *savedUser
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	model := models.UserFromEntity(user)

	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ?", user.ID).
		Updates(model)

	if result.Error != nil {
		if r.isDuplicateError(result.Error) {
			return fmt.Errorf("duplicate key error")
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.UserModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.WithContext(ctx).
		Where("google_id = ?", googleID).
		First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by google id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByFacebookID(ctx context.Context, facebookID string) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.WithContext(ctx).
		Where("facebook_id = ?", facebookID).
		First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by facebook id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ?", userID).
		Update("last_login_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error {
	updates := map[string]interface{}{
		"last_known_lat": lat,
		"last_known_lng": lng,
		"last_location":  gorm.Expr("ST_SetSRID(ST_MakePoint(?, ?), 4326)", lng, lat),
	}

	result := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("id = ?", userID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update location: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) FindByAreaID(ctx context.Context, areaID uuid.UUID, pagination *database.Pagination) ([]*entities.User, error) {
	var userModels []models.UserModel

	query := r.db.WithContext(ctx).
		Where("area_id = ?", areaID)

	if pagination != nil {
		query = query.Scopes(database.Paginate(pagination.Page, pagination.PerPage))
	}

	err := query.Find(&userModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find users by area: %w", err)
	}

	users := make([]*entities.User, len(userModels))
	for i, model := range userModels {
		user, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert user model: %w", err)
		}
		users[i] = user
	}

	return users, nil
}

func (r *userRepository) CountByAreaID(ctx context.Context, areaID uuid.UUID) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("area_id = ?", areaID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users by area: %w", err)
	}

	return count, nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("email = ?", email.String()).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return count > 0, nil
}

func (r *userRepository) ExistsByPhone(ctx context.Context, phone valueobjects.Phone) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Where("phone = ?", phone.String()).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}

	return count > 0, nil
}

// isDuplicateError checks if the error is a unique constraint violation
func (r *userRepository) isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL unique violation error code is 23505
	errStr := err.Error()
	return contains(errStr, "23505") || contains(errStr, "duplicate key") || contains(errStr, "unique constraint")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (r *userRepository) ListByArea(ctx context.Context, area string, limit, offset int) ([]*entities.User, error) {
	var userModels []models.UserModel

	query := r.db.WithContext(ctx)

	// Assuming area is an area code or name
	areaSubQuery := r.db.Table("areas").
		Select("id").
		Where("code = ? OR name = ?", area, area)

	err := query.
		Where("area_id IN (?)", areaSubQuery).
		Limit(limit).
		Offset(offset).
		Find(&userModels).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list users by area: %w", err)
	}

	users := make([]*entities.User, len(userModels))
	for i, model := range userModels {
		user, err := model.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert user model: %w", err)
		}
		users[i] = user
	}

	return users, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.UserModel{}).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}
