package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultBcryptCost = 12
	MinPasswordLength = 12
)

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

func CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid password")
		}
		return fmt.Errorf("failed to compare password: %w", err)
	}
	return nil
}

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateResetToken() (string, error) {
	return GenerateSecureToken(32)
}

func GenerateAPIKey() (string, error) {
	return GenerateSecureToken(48)
}

// GenerateMFASecret generates a base32 secret for MFA/TOTP
func GenerateMFASecret() (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Torra",
		AccountName: "",
		SecretSize:  32,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}
	return key.Secret(), nil
}

// VerifyMFACode verifies a TOTP code against a secret
func VerifyMFACode(code, secret string) (bool, error) {
	if len(code) != 6 {
		return false, errors.New("invalid code length")
	}

	// Validate the TOTP code using the library
	valid := totp.Validate(code, secret)
	return valid, nil
}
