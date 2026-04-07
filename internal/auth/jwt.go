// Package auth handles JWT signing/verification and OAuth2 flows.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the JWT payload.
type Claims struct {
	UserID        uuid.UUID  `json:"sub"`
	Role          string     `json:"role"`
	Impersonating *uuid.UUID `json:"imp,omitempty"` // set when admin assumes a user
	AdminID       *uuid.UUID `json:"adm,omitempty"` // original admin ID during impersonation
	jwt.RegisteredClaims
}

// JWTService signs and verifies JWTs.
type JWTService struct {
	secret   []byte
	duration time.Duration
}

func NewJWTService(secret string, duration time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), duration: duration}
}

// Sign returns a signed JWT for the given user.
func (j *JWTService) Sign(userID uuid.UUID, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// SignImpersonation returns a short-lived JWT where an admin acts as another user.
func (j *JWTService) SignImpersonation(adminID, targetUserID uuid.UUID, targetRole string) (string, error) {
	claims := Claims{
		UserID:        targetUserID,
		Role:          targetRole,
		Impersonating: &targetUserID,
		AdminID:       &adminID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// Verify parses and validates a JWT, returning its claims.
func (j *JWTService) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing jwt: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
