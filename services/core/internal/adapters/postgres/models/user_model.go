package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/restayway/gogis"
	"gorm.io/gorm"

	"github.com/paradoxe35/mytorra/packages/database"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/entities"
	"github.com/paradoxe35/mytorra/services/core/internal/domain/valueobjects"
)

// UserModel represents the GORM model for the users table
type UserModel struct {
	database.BaseModel
	Email             string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone             string  `gorm:"type:varchar(20);uniqueIndex"`
	Name              string  `gorm:"type:varchar(255);not null"`
	PasswordHash      string  `gorm:"type:varchar(255);not null"`
	ProfileImageURL   string  `gorm:"type:text"`
	EmailVerified     bool    `gorm:"type:boolean;default:false"`
	PhoneVerified     bool    `gorm:"type:boolean;default:false"`
	IDVerified        bool    `gorm:"type:boolean;default:false"`
	VerificationLevel int     `gorm:"type:int;default:0"`
	GoogleID          *string `gorm:"type:varchar(255);uniqueIndex"`
	FacebookID        *string `gorm:"type:varchar(255);uniqueIndex"`
	MFAEnabled        bool    `gorm:"type:boolean;default:false"`
	MFASecret         *string `gorm:"type:varchar(255)"`

	// Location data
	AreaID       uuid.UUID    `gorm:"type:uuid;index"`
	Area         *AreaModel   `gorm:"foreignKey:AreaID"`
	LastKnownLat *float64     `gorm:"type:decimal(10,8)"`
	LastKnownLng *float64     `gorm:"type:decimal(11,8)"`
	LastLocation *gogis.Point `gorm:"type:geometry(Point,4326)"`

	// Status and timestamps
	Status      string     `gorm:"type:varchar(20);default:'active'"`
	LastLoginAt *time.Time `gorm:"type:timestamp"`
}

// TableName specifies the table name for the UserModel
func (UserModel) TableName() string {
	return "users"
}

// BeforeCreate hook to set default values
func (u *UserModel) BeforeCreate(tx *gorm.DB) error {
	// Call parent BeforeCreate
	if err := u.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	// Sync location fields if lat/lng provided
	if u.LastKnownLat != nil && u.LastKnownLng != nil {
		u.LastLocation = &gogis.Point{
			Lng: *u.LastKnownLng,
			Lat: *u.LastKnownLat,
		}
	}
	return nil
}

// BeforeUpdate hook to update timestamps
func (u *UserModel) BeforeUpdate(tx *gorm.DB) error {
	// Sync location fields if lat/lng changed
	if u.LastKnownLat != nil && u.LastKnownLng != nil {
		u.LastLocation = &gogis.Point{
			Lng: *u.LastKnownLng,
			Lat: *u.LastKnownLat,
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
		model.DeletedAt = gorm.DeletedAt{
			Time:  *user.DeletedAt,
			Valid: true,
		}
	}

	// Create PostGIS point from lat/lng
	if user.LastKnownLat != nil && user.LastKnownLng != nil {
		model.LastLocation = &gogis.Point{
			Lng: *user.LastKnownLng,
			Lat: *user.LastKnownLat,
		}
	}

	return model
}
