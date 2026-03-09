package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey = contextKey("user")

var ErrTokenInvalidMethod = errors.New("invalid jwt signing method")

// struct to define users and their roles within
type UserClaims struct {
	UserID int
	Role   string
}

func JWTMiddleware(j *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing or invalid token authorization header", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := j.VerifyToken(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(float64)
			role, ok2 := claims["role"].(string)
			if !ok || !ok2 {
				http.Error(w, "invalid token payload", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, &UserClaims{
				UserID: int(userID),
				Role:   role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// add helper to get claims from context
func GetUserClaims(r *http.Request) *UserClaims {
	claims, ok := r.Context().Value(UserContextKey).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}
