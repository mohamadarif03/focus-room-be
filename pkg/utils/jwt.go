package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWT_SECRET = []byte("ini_rahasia_banget_ganti_nanti")

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, role string) (string, error) {
	userIDStr := fmt.Sprint(userID)

	claims := JWTClaims{
		UserID: userIDStr,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_SECRET)
}
