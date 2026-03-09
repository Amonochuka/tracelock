package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (j *JWTService) GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"iat":  time.Now().Unix(),
	}

	//create a token with HS356 signing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//sign token using jwtSecret
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", ErrTokenInvalidMethod
	}
	return tokenString, nil
}
