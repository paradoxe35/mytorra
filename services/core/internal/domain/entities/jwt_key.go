package entities

import (
	"time"

	"github.com/google/uuid"
)

// JWTKey represents a JWT signing key stored in the database
type JWTKey struct {
	ID         uuid.UUID  `db:"id"`
	KeyID      string     `db:"key_id"`      // Unique identifier for the key (kid in JWK)
	PrivateKey string     `db:"private_key"` // PEM-encoded private key (should be encrypted)
	PublicKey  string     `db:"public_key"`  // PEM-encoded public key
	Algorithm  string     `db:"algorithm"`   // RS256, RS384, RS512
	IsActive   bool       `db:"is_active"`   // Whether this key is currently active for signing
	CreatedAt  time.Time  `db:"created_at"`
	ExpiresAt  time.Time  `db:"expires_at"`
	RotatedAt  *time.Time `db:"rotated_at"` // When this key was rotated
}

// IsValid checks if the JWT key is still valid for use
func (k *JWTKey) IsValid() bool {
	return k.IsActive && time.Now().Before(k.ExpiresAt)
}

// Rotate marks this key as rotated
func (k *JWTKey) Rotate() {
	now := time.Now()
	k.IsActive = false
	k.RotatedAt = &now
}

// JWTKeyAlgorithm represents supported JWT signing algorithms
type JWTKeyAlgorithm string

const (
	JWTKeyAlgorithmRS256 JWTKeyAlgorithm = "RS256"
	JWTKeyAlgorithmRS384 JWTKeyAlgorithm = "RS384"
	JWTKeyAlgorithmRS512 JWTKeyAlgorithm = "RS512"
)

// IsValidAlgorithm checks if the algorithm is supported
func (a JWTKeyAlgorithm) IsValid() bool {
	switch a {
	case JWTKeyAlgorithmRS256, JWTKeyAlgorithmRS384, JWTKeyAlgorithmRS512:
		return true
	default:
		return false
	}
}
