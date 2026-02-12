package auth

import "os"

var jwtSecret []byte

// load it from the environment
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set")
	}

	jwtSecret = []byte(secret)
}
