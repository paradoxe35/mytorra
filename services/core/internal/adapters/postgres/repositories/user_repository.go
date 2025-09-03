package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/packages/errors"
	"github.com/paradoxe35/mytorra/services/core/internal/adapters/postgres/models"
	"github.com/paradoxe35/mytorra/services/core/internal/application/ports"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

type userRepository struct {
	db *database.DB
}

// NewUserRepository creates a new Bun-based user repository
func NewUserRepository(db *database.DB) ports.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.NewSelect().
		Model(&userModel).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.NewSelect().
		Model(&userModel).
		Where("email = ?", email.String()).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByPhone(ctx context.Context, phone valueobjects.Phone) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.NewSelect().
		Model(&userModel).
		Where("phone = ?", phone.String()).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) Save(ctx context.Context, user *entities.User) error {
	model := models.UserFromEntity(user)

	_, err := r.db.NewInsert().
		Model(model).
		Exec(ctx)
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

	result, err := r.db.NewUpdate().
		Model(model).
		Where("id = ?", user.ID).
		OmitZero().
		Exec(ctx)
	if err != nil {
		if r.isDuplicateError(err) {
			return fmt.Errorf("duplicate key error")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model(&models.UserModel{}).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.NewSelect().
		Model(&userModel).
		Where("google_id = ?", googleID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by google id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) FindByFacebookID(ctx context.Context, facebookID string) (*entities.User, error) {
	var userModel models.UserModel

	err := r.db.NewSelect().
		Model(&userModel).
		Where("facebook_id = ?", facebookID).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find user by facebook id: %w", err)
	}

	return userModel.ToEntity()
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model(&models.UserModel{}).
		Where("id = ?", userID).
		Set("last_login_at = ?", now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) UpdateLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error {
	result, err := r.db.NewUpdate().
		Model(&models.UserModel{}).
		Where("id = ?", userID).
		Set("last_known_lat = ?", lat).
		Set("last_known_lng = ?", lng).
		Set("last_location = ST_SetSRID(ST_MakePoint(?, ?), 4326)", lng, lat).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *userRepository) FindByAreaID(ctx context.Context, areaID uuid.UUID, pagination *database.Pagination) ([]*entities.User, error) {
	var userModels []models.UserModel

	query := r.db.NewSelect().
		Model(&userModels).
		Where("area_id = ?", areaID)

	if pagination != nil {
		paginate := database.NewPaginate(pagination.Page, pagination.PerPage)
		query = paginate.Apply(query)
	}

	err := query.Scan(ctx)
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
	count, err := r.db.NewSelect().
		Model(&models.UserModel{}).
		Where("area_id = ?", areaID).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by area: %w", err)
	}

	return int64(count), nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	exists, err := r.db.NewSelect().
		Model(&models.UserModel{}).
		Where("email = ?", email.String()).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

func (r *userRepository) ExistsByPhone(ctx context.Context, phone valueobjects.Phone) (bool, error) {
	exists, err := r.db.NewSelect().
		Model(&models.UserModel{}).
		Where("phone = ?", phone.String()).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check phone existence: %w", err)
	}

	return exists, nil
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

	// Create subquery for area
	areaSubQuery := r.db.NewSelect().
		TableExpr("areas").
		Column("id").
		Where("code = ? OR name = ?", area, area)

	err := r.db.NewSelect().
		Model(&userModels).
		Where("area_id IN (?)", areaSubQuery).
		Limit(limit).
		Offset(offset).
		Scan(ctx)
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
	count, err := r.db.NewSelect().
		Model(&models.UserModel{}).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return int64(count), nil
}
