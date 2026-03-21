package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"tracelock/internal/httpdir"
)

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
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				httpdir.WriteError(w,http.StatusUnauthorized, "missing or invalid token authorization header")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := j.VerifyToken(tokenString)
			if err != nil {
				httpdir.WriteError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			userID, ok := claims["sub"].(float64)
			role, ok2 := claims["role"].(string)
			if !ok || !ok2 {
				httpdir.WriteError(w, http.StatusUnauthorized, "invalid token payload")
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

//add helper to get userID from context
func GetUserIDFromContext(ctx context.Context)(int, error){
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	if !ok || claims == nil{
		return 0, fmt.Errorf("user not found")
	}
	return claims.UserID, nil
}
