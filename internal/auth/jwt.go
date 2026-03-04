package auth

import (
	"fmt"
	"os"
	
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// load it from the environment
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set")
	}

	jwtSecret = []byte(secret)
	fmt.Println("JWT initialized succesfully")
}

func GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"role": user.Role,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	//create a token with HS356 signing
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//sign token using jwtSecret
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", ErrTokenInvalidMethod
	}
	return tokenString, nil
}
