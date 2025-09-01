package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// JWTKeyPair represents an RSA key pair for JWT signing
type JWTKeyPair struct {
	KeyID      string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// GenerateRSAKeyPair generates a new RSA key pair for JWT signing
func GenerateRSAKeyPair(keyID string, validityDays int) (*JWTKeyPair, error) {
	// Generate RSA key pair with 2048 bits (industry standard)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	now := time.Now()
	return &JWTKeyPair{
		KeyID:      keyID,
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		CreatedAt:  now,
		ExpiresAt:  now.AddDate(0, 0, validityDays),
	}, nil
}

// EncodePrivateKeyPEM encodes an RSA private key to PEM format
func EncodePrivateKeyPEM(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return string(pem.EncodeToMemory(privateKeyPEM)), nil
}

// DecodePrivateKeyPEM decodes a PEM-encoded RSA private key
func DecodePrivateKeyPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1 format for backward compatibility
		rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		return rsaKey, nil
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not RSA")
	}

	return rsaKey, nil
}

// EncodePublicKeyPEM encodes an RSA public key to PEM format
func EncodePublicKeyPEM(publicKey *rsa.PublicKey) (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return string(pem.EncodeToMemory(publicKeyPEM)), nil
}

// DecodePublicKeyPEM decodes a PEM-encoded RSA public key
func DecodePublicKeyPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	return rsaKey, nil
}

// ConvertRSAPublicKeyToJWK converts an RSA public key to JWK format
func ConvertRSAPublicKeyToJWK(keyID string, publicKey *rsa.PublicKey) JWK {
	// Encode modulus (n) and exponent (e) to base64url
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	return JWK{
		Kid: keyID,
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   n,
		E:   e,
	}
}

// CreateJWKS creates a JWKS from multiple key pairs
func CreateJWKS(keyPairs []*JWTKeyPair) JWKS {
	jwks := JWKS{
		Keys: make([]JWK, 0, len(keyPairs)),
	}

	for _, kp := range keyPairs {
		// Only include non-expired keys
		if time.Now().Before(kp.ExpiresAt) {
			jwk := ConvertRSAPublicKeyToJWK(kp.KeyID, kp.PublicKey)
			jwks.Keys = append(jwks.Keys, jwk)
		}
	}

	return jwks
}

// MarshalJWKS converts JWKS to JSON
func MarshalJWKS(jwks JWKS) ([]byte, error) {
	return json.MarshalIndent(jwks, "", "  ")
}

// GenerateKeyID generates a unique key ID
func GenerateKeyID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate key ID: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateRefreshToken generates a cryptographically secure refresh token
func GenerateRefreshToken() (string, error) {
	// Generate 32 bytes of random data for a 256-bit token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// HashRefreshToken creates a SHA256 hash of a refresh token for storage
func HashRefreshToken(token string) string {
	// We'll use bcrypt for refresh token hashing too for consistency
	// This provides protection even if the database is compromised
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.MinCost)
	return string(hash)
}

// VerifyRefreshToken verifies a refresh token against its hash
func VerifyRefreshToken(token, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	return err == nil
}
