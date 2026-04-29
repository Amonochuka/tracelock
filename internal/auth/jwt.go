package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"tracelock/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (j *JWTService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Minute * 15).Unix(),
		"iat":  time.Now().Unix(),
	}

	// create a token with HS256 signing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign token using jwtSecret
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", ErrTokenInvalidMethod
	}
	return tokenString, nil
}

func (j *JWTService) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalidMethod
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}
	return claims, nil
}

func (j *JWTService) GenerateRefreshToken() (string, time.Time, error) {
	// generate random 32 byte token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}
	token := hex.EncodeToString(b)
	expiresAt := time.Now().Add(time.Hour * 24 * 7)
	return token, expiresAt, nil
}
