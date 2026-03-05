
# Developer's guide


## TraceLock – Developer Guide

This document is for developers contributing to TraceLock.  
It covers detailed setup, code patterns, JWT handling, and database considerations.


## 1. JWT Initialization

```go
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set")
	}
	jwtSecret = []byte(secret)
}
```

1. Reads JWT_SECRET from the environment

2. Converts it to []byte and stores in jwtSecret

3. Used for signing and verifying tokens

## 2. Generating JWT Tokens
```
func GenerateToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
```
 - claims → JWT payload (map[string]interface{})

 - Library encodes as JSON, base64, and signs it

## 3. Parsing & Middleware
```
token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrTokenInvalidMethod
	}
	return jwtSecret, nil
})
```

 - Verifies signature and decodes payload

 - Claims extracted as:
```
claims, ok := token.Claims.(jwt.MapClaims)
userID, ok := claims["sub"].(float64)
role, ok2 := claims["role"].(string)
```

***JSON numbers decode as float64 → convert to int as needed***

## 4. Context Usage

 - Store UserClaims in request context for authenticated handlers:

```userClaims, ok := r.Context().Value(UserContextKey).(*UserClaims)```

 - Allows safe access to user info without globals

## 5. Database Notes

Tables and sequences must have privileges for tracelock_user:
```
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tracelock_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO tracelock_user;
```

 - PostgreSQL handles multiple connections safely

 - Each HTTP request uses its own transaction

 - No mutexes needed for phase 1–2

## 6. Common Pitfalls

 - Environment variables not exported → server cannot read

- JWT sub type → float64 needs int conversion

- Global variables → do not use for per-request JWT claims
