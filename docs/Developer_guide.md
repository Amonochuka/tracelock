
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

## PostgreSQL Gotcha You Encountered
 1. The Problem;

When you first tried running your app with tracelock_user, you got an error:

pq: permission denied for table users

This happened because:

The tables were created as the postgres superuser.
Your app connects as tracelock_user, a regular DB user.

GRANTs on the database itself do not automatically give table/sequence permissions.
So even though tracelock_user could connect to the database, it could not read or write to tables.

2. How it happened in practice
```
CREATE TABLE users(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Executed as postgres superuser

tracelock_user tried to access users → got permission denied

3. The Fix

Run as postgres and grant table and sequence permissions to tracelock_user:

sudo -u postgres psql
\c tracelock

Then run:
```
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tracelock_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO tracelock_user;
```

After this, tracelock_user could read/write tables and your app worked.

4. Developer Notes

 - This is one of the most common PostgreSQL gotchas when creating an app user.

 - Database-level access ≠ table-level access. Always explicitly grant privileges on tables and sequences.

 - In production, you usually create a limited-permission DB user and run migrations as a superuser or admin user.
