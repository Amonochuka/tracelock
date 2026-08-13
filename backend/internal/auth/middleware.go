package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type contextKey string

const UserContextKey = contextKey("user")

// struct to define users and their roles within
type UserClaims struct {
	UserID int
	Role   string
}

func JWTMiddleware(j *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}

			if tokenString == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid token")
				return
			}

			claims, err := j.VerifyToken(tokenString)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "session expired, please log in again")
				return
			}

			userID, ok := claims["sub"].(float64)
			role, ok2 := claims["role"].(string)
			if !ok || !ok2 {
				writeJSONError(w, http.StatusUnauthorized, "invalid token payload")
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

// add helper to get userID from context
func GetUserIDFromContext(ctx context.Context) (int, error) {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	if !ok || claims == nil {
		return 0, ErrUserNotFound
	}
	return claims.UserID, nil
}

// add helper to get role from context
func GetUserRoleFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	if !ok || claims == nil {
		return "", ErrUserNotFound
	}
	return claims.Role, nil
}
