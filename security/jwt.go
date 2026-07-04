package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthClaims represents the JWT claims for authentication
type AuthClaims struct {
	Sub   string   `json:"sub"`
	Roles []string `json:"roles"` // roles UUIDs
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token JWT
func GenerateAccessToken(
	sub string,
	roles []string,
	secret string,
	expireTime int,
) (string, error) {
	claims := AuthClaims{
		Sub:   sub,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT and returns the claims
func ValidateToken(tokenString string, secret string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}
