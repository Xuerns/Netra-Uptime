package jwt

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
			log.Fatal("FATAL SECURITY CONFIGURATION: JWT_SECRET must be at least 32 characters long in production")
		}
		if len(secret) == 0 {
			return []byte("netra-uptime-dev-secret-must-be-at-least-32-chars-long!")
		}
	}
	return []byte(secret)
}

func GenerateToken(userID string, expiryTime time.Duration) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "netra-uptime",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(getJWTSecret())

	if err != nil {
		return "", errors.New("Failed build token")
	}

	return tokenString, nil
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Token Invalid")
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, errors.New("Token Invalid")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("Token Invalid")
}
