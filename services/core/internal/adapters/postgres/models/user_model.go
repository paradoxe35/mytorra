package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// UserModel represents the Bun model for the users table
type UserModel struct {
	database.BaseModel `bun:"table:users,alias:u"`

	Email             string  `bun:"email,notnull,unique"`
	Phone             string  `bun:"phone,unique"`
	Name              string  `bun:"name,notnull"`
	PasswordHash      string  `bun:"password_hash,notnull"`
	ProfileImageURL   string  `bun:"profile_image_url"`
	EmailVerified     bool    `bun:"email_verified,default:false"`
	PhoneVerified     bool    `bun:"phone_verified,default:false"`
	IDVerified        bool    `bun:"id_verified,default:false"`
	VerificationLevel int     `bun:"verification_level,default:0"`
	GoogleID          *string `bun:"google_id,unique"`
	FacebookID        *string `bun:"facebook_id,unique"`
	MFAEnabled        bool    `bun:"mfa_enabled,default:false"`
	MFASecret         *string `bun:"mfa_secret"`

	// Location data
	AreaID       uuid.UUID       `bun:"area_id,type:uuid"`
	Area         *AreaModel      `bun:"rel:belongs-to,join:area_id=id"`
	LastKnownLat *float64        `bun:"last_known_lat,type:decimal(10,8)"`
	LastKnownLng *float64        `bun:"last_known_lng,type:decimal(11,8)"`
	LastLocation *database.Point `bun:"last_location,type:geometry(Point,4326)"`

	// Status and timestamps
	Status      string     `bun:"status,default:'active'"`
	LastLoginAt *time.Time `bun:"last_login_at"`
}

// BeforeAppendModel is called before inserting or updating the model
func (u *UserModel) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		// Sync location fields if lat/lng provided
		if u.LastKnownLat != nil && u.LastKnownLng != nil {
			u.LastLocation = &database.Point{
				Lng: *u.LastKnownLng,
				Lat: *u.LastKnownLat,
			}
		}
	case *bun.UpdateQuery:
		// Sync location fields if lat/lng changed
		if u.LastKnownLat != nil && u.LastKnownLng != nil {
			u.LastLocation = &database.Point{
				Lng: *u.LastKnownLng,
				Lat: *u.LastKnownLat,
			}
		}
	}
	return nil
}

// ToEntity converts the GORM model to domain entity
func (u *UserModel) ToEntity() (*entities.User, error) {
	// Convert Email
	email, err := valueobjects.NewEmail(u.Email)
	if err != nil {
		return nil, err
	}

	// Convert Phone
	phone, err := valueobjects.NewPhone(u.Phone)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		ID:                u.ID,
		Email:             email,
		Phone:             phone,
		Name:              u.Name,
		PasswordHash:      u.PasswordHash,
		ProfileImageURL:   u.ProfileImageURL,
		EmailVerified:     u.EmailVerified,
		PhoneVerified:     u.PhoneVerified,
		IDVerified:        u.IDVerified,
		VerificationLevel: u.VerificationLevel,
		GoogleID:          u.GoogleID,
		FacebookID:        u.FacebookID,
		MFAEnabled:        u.MFAEnabled,
		MFASecret:         u.MFASecret,
		AreaID:            u.AreaID,
		LastKnownLat:      u.LastKnownLat,
		LastKnownLng:      u.LastKnownLng,
		Status:            entities.UserStatus(u.Status),
		LastLoginAt:       u.LastLoginAt,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}

	// Handle deleted timestamp
	if u.DeletedAt.Valid {
		user.DeletedAt = &u.DeletedAt.Time
	}

	return user, nil
}

// FromEntity creates a GORM model from domain entity
func UserFromEntity(user *entities.User) *UserModel {
	model := &UserModel{
		BaseModel: database.BaseModel{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Email:             user.Email.String(),
		Phone:             user.Phone.String(),
		Name:              user.Name,
		PasswordHash:      user.PasswordHash,
		ProfileImageURL:   user.ProfileImageURL,
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		IDVerified:        user.IDVerified,
		VerificationLevel: user.VerificationLevel,
		GoogleID:          user.GoogleID,
		FacebookID:        user.FacebookID,
		MFAEnabled:        user.MFAEnabled,
		MFASecret:         user.MFASecret,
		AreaID:            user.AreaID,
		LastKnownLat:      user.LastKnownLat,
		LastKnownLng:      user.LastKnownLng,
		Status:            string(user.Status),
		LastLoginAt:       user.LastLoginAt,
	}

	// Handle deleted timestamp
	if user.DeletedAt != nil {
		model.DeletedAt = sql.NullTime{
			Time:  *user.DeletedAt,
			Valid: true,
		}
	}

	// Create PostGIS point from lat/lng
	if user.LastKnownLat != nil && user.LastKnownLng != nil {
		model.LastLocation = &database.Point{
			Lng: *user.LastKnownLng,
			Lat: *user.LastKnownLat,
		}
	}

	return model
}
