# TraceLock – Developer Guide

This document is for developers contributing to TraceLock.
It covers setup, code patterns, JWT handling, error handling, and database considerations.

---

## 1. Architecture Overview

TraceLock uses a three-layer architecture per package:

```
Handler → Service → Repo/Auth
```

Each layer has a single responsibility:

- **Repo/Auth** — talks to the database. Translates raw DB errors (`sql.ErrNoRows`, `pq.Error`) into domain sentinel errors. Never leaks DB concerns upward.
- **Service** — business logic. Receives domain sentinels, passes them through clean. Only wraps unexpected infrastructure errors with context.
- **Handler** — HTTP concerns. Maps domain sentinels to HTTP status codes. The only layer that knows about `http.ResponseWriter`.

---

## 2. Error Handling Pattern

### Sentinel errors

Each package defines its own sentinel errors:

```go
// internal/auth/errors.go
var (
    ErrTokenInvalidMethod = errors.New("invalid jwt signing method")
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailExists        = errors.New("email already exists")
    ErrInvalidCredentials = errors.New("invalid email or password")
)

// internal/access/errors.go
var (
    ErrZoneNotFound       = errors.New("zone not found")
    ErrZoneFull           = errors.New("zone is at capacity")
    ErrUserAlreadyInZone  = errors.New("user already in zone")
    ErrNoActiveSession    = errors.New("no active session")
    ErrNoHashFound        = errors.New("no previous hash found")
)
```

### Rules

- `sql.ErrNoRows` and `pq.Error` are handled in the repo/auth layer only — never leak past it
- Sentinel errors pass through the service unwrapped — `errors.Is` handles them up the chain
- Unexpected infrastructure errors (DB down, timeouts) are wrapped with context at the layer they are caught
- The service never imports `database/sql` or `github.com/lib/pq`

### Example flow

```go
// UserAuth — translates DB error to sentinel
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrInvalidCredentials
}

// UserService — passes sentinel through clean
return s.auth.Authenticate(email, password)

// Handler — maps sentinel to HTTP response
if errors.Is(err, auth.ErrInvalidCredentials) {
    http.Error(w, "invalid credentials", http.StatusUnauthorized)
}
```

---

## 3. JWT

### Initialization

```go
jwtService := auth.NewJWTService(cfg.JWTSecret)
```

Reads `JWT_SECRET` from config and stores it as `[]byte` for signing and verification.

### Token generation

```go
claims := jwt.MapClaims{
    "sub":  user.ID,
    "role": user.Role,
    "exp":  time.Now().Add(24 * time.Hour).Unix(),
    "iat":  time.Now().Unix(),
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
```

Tokens expire after 24 hours. The payload is signed, not encrypted — do not store sensitive data in claims.

### Parsing and middleware

```go
token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
    if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, ErrTokenInvalidMethod
    }
    return jwtSecret, nil
})
```

Claims are extracted as:

```go
claims, ok := token.Claims.(jwt.MapClaims)
userID := int(claims["sub"].(float64))  // JSON numbers decode as float64
role := claims["role"].(string)
```

### Context usage

User claims are stored in request context by the middleware:

```go
userClaims, ok := r.Context().Value(UserContextKey).(*UserClaims)
```

Access them in handlers via:

```go
claims := auth.GetUserClaims(r)
userID, err := auth.GetUserIDFromContext(r.Context())
```

---

## 4. Zone Access Flow

When a user enters or exits a zone, `HandleZoneEvent` runs the following:

**Enter:**
1. Fetch zone max capacity
2. Count active users in zone
3. Reject if at capacity (`ErrZoneFull`)
4. Create active session — unique constraint on `(user_id, zone_id)` prevents duplicates
5. Fetch last event hash for the zone (empty string if first event)
6. Generate new SHA-256 hash chaining the previous hash
7. Write access event with `action=enter`, `status=allowed`

**Exit:**
1. Delete active session — returns `ErrNoActiveSession` if none exists
2. Fetch last event hash
3. Generate new hash
4. Write access event with `action=exit`, `status=allowed`

The hash chain provides tamper-evidence — each event references the hash of the previous one, making it detectable if records are altered or deleted.

---

## 5. Database Schema

```sql
users           — id, name, email, password_hash, role, created_at
zones           — id, name, description, max_capacity, created_at
access_events   — id, user_id, zone_id, action, status, timestamp, hash, previous_hash
active_sessions — user_id, zone_id, entered_at  (PRIMARY KEY: user_id, zone_id)
```

`active_sessions` uses a composite primary key on `(user_id, zone_id)` — this enforces that a user can only have one active session per zone at the database level.

---

## 6. PostgreSQL Permissions

Tables must have explicit privileges for the app DB user. This is a common PostgreSQL gotcha — database-level access does not grant table-level access.

After running migrations as a superuser, grant privileges:

```sql
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tracelock_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO tracelock_user;
```

Without this, the app will get `pq: permission denied for table users`.

---

## 7. Common Pitfalls

- **`sql.ErrNoRows` in service layer** — move it to the repo. The service should not import `database/sql`.
- **`pq.Error` in service layer** — same, move duplicate key checks to the repo.
- **`fmt.Errorf` wrapping nil** — always guard with `if err != nil` before wrapping, otherwise a successful operation returns a non-nil error.
- **`CREATE TABLE IF NOT EXISTS`** — if a table was manually created with the wrong schema before migrations ran, the migration skips it silently. Drop and recreate.
- **JWT `sub` claim is `float64`** — JSON numbers always decode as `float64` in Go. Always cast: `int(claims["sub"].(float64))`.
- **Token split across terminal lines** — when curling with a Bearer token, keep the full token on one line or use a shell variable.
- **Dropped all tables** — re-register users and recreate sessions; foreign key constraints will reject references to non-existent users.
