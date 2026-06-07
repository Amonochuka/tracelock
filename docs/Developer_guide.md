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

Services are wired together in `main.go` — the only file that knows about all dependencies.

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
    ErrInvalidRole        = errors.New("role must be 'admin' or 'user'")
    ErrAdminExists        = errors.New("an admin account already exists")
    ErrTokenNotFound      = errors.New("refresh token not found")
    ErrTokenRevoked       = errors.New("refresh token has been revoked")
    ErrTokenExpired       = errors.New("refresh token has expired")
)

// internal/access/errors.go
var (
    ErrZoneNotFound       = errors.New("zone not found")
    ErrZoneFull           = errors.New("zone is at capacity")
    ErrUserAlreadyInZone  = errors.New("user already in zone")
    ErrNoActiveSession    = errors.New("no active session")
    ErrNoHashFound        = errors.New("hash does not exist")
    ErrZoneNameExists     = errors.New("zone name already exists")
    ErrZoneHasActivity    = errors.New("zone has active sessions and cannot be deleted")
    ErrAccessDenied       = errors.New("user does not have access to this zone")
    ErrAccessNotFound     = errors.New("access grant not found")
    ErrDeviceNotFound     = errors.New("device not found")
    ErrDeviceSerialExists = errors.New("device serial already exists")
    ErrDeviceInactive     = errors.New("device is not active")
    ErrCredentialExists   = errors.New("credential already exists for this method")
    ErrCredentialNotFound = errors.New("credential not found")
    ErrCredentialRevoked  = errors.New("credential has been revoked")
)
```

### Rules

- `sql.ErrNoRows` and `pq.Error` are handled in the repo/auth layer only — never leak past it
- Sentinel errors pass through the service unwrapped — `errors.Is` handles them up the chain
- Unexpected infrastructure errors are wrapped with context at the layer they are caught
- The service never imports `database/sql` or `github.com/lib/pq`

---

## 3. JWT

### Access token

Short-lived (15 minutes). Used for all authenticated API requests.

```go
claims := jwt.MapClaims{
    "sub":  user.ID,
    "role": user.Role,
    "exp":  time.Now().Add(time.Minute * 15).Unix(),
    "iat":  time.Now().Unix(),
}
```

### Refresh token

Long-lived (7 days). Stored in the `refresh_tokens` DB table. Used only to obtain a new access token via `POST /refresh`. Revoked on logout.

```go
refreshToken, expiresAt, err := j.GenerateRefreshToken()
```

### Token flow

```
POST /login → access token (15min) + refresh token (7 days)
access token expires → POST /refresh → new access token
POST /logout → refresh token revoked
```

### Context usage

Claims are stored in request context by JWT middleware:

```go
claims := auth.GetUserClaims(r)
userID, err := auth.GetUserIDFromContext(r.Context())
role, err := auth.GetUserRoleFromContext(r.Context())
```

---

## 4. Admin Bootstrap

The first admin cannot be created via `/register` (which always creates `role=user`). Use the one-time bootstrap endpoint:

```bash
POST /bootstrap
{"name": "Alice", "email": "alice@company.com", "password": "securepass"}
```

After the first successful call, the endpoint permanently returns `403`. All subsequent admin promotions go through `PUT /admin/users/{id}/role` using an existing admin JWT.

---

## 5. Zone Access Flow

### API-based entry (JWT)

```
POST /zones/enter → HandleZoneEvent
  1. Check user has access to zone (user_zone_access table)
  2. Check zone capacity
  3. Create active session
  4. Generate hash chaining previous event
  5. Write access event (action=enter, status=allowed)
```

### Biometric entry (scanner)

```
POST /devices/authenticate → AuthenticateBiometric
  1. Validate device exists and is active
  2. Match credential hash to enrolled credential
  3. Check credential is not revoked
  4. Resolve user from credential
  5. Delegate to HandleZoneEvent (steps 1-5 above)
  6. Issue JWT for the authenticated user
```

Both flows write to the same `active_sessions` and `access_events` tables.

### Denied events

Denied attempts (no access, zone full, already inside) are also written to `access_events` with `status=denied`. The hash chain covers both allowed and denied events.

---

## 6. Biometric System

### Device types

`fingerprint`, `face`, `iris`, `card`, `pin`

Devices are registered to zones by admin via `POST /admin/zones/{id}/devices`.

### Credential enrollment

Admin enrolls a user's biometric credential via `POST /admin/users/{id}/credentials`. The `credential_hash` is a tokenised hash produced by the scanner SDK — raw biometric data is never stored or transmitted.

For testing, simulate a credential hash:
```bash
openssl rand -hex 32
```

### Runtime authentication

The scanner sends `device_id` + `credential_hash` to `POST /devices/authenticate`. The backend:
1. Validates the device
2. Matches the hash to an enrolled credential
3. Resolves the user
4. Checks zone access
5. Creates session and audit event
6. Returns a JWT

### Interface pattern

`BiometricService` uses Go interfaces to avoid circular package dependencies:

```go
type UserResolver interface {
    VerifyUser(id int) (*models.User, error)
}

type JWTIssuer interface {
    GenerateToken(user *models.User) (string, error)
}
```

`auth.UserAuth` and `auth.JWTService` satisfy these interfaces implicitly — no explicit declaration needed.

---

## 7. Hash Chain

Each access event stores a SHA-256 hash that chains the previous event's hash:

```go
data := fmt.Sprintf("%d:%d:%s:%s:%s:%s",
    userID, zoneID, action, timestamp, previousHash, entryMethod)
hash := sha256.Sum256([]byte(data))
```

The hash includes `entryMethod` — a fingerprint entry and a card entry for the same user produce different hashes.

Verify chain integrity via `GET /admin/zones/{id}/verify-chain`.

---

## 8. Rate Limiting

Login and register are rate limited by IP using a token bucket algorithm:

- 5 requests per minute per IP
- Tokens refill continuously (not on a hard reset)
- Old client state cleaned up every 3 minutes
- `X-Forwarded-For` header used for real IP behind Render's proxy

Rate limit state is in-memory — does not survive server restarts. For multi-instance production use, replace with Redis.

---

## 9. Database Schema

```sql
users               — id, name, email, password_hash, role, created_at
zones               — id, name, description, max_capacity, created_at
active_sessions     — user_id, zone_id, entered_at (PK: user_id, zone_id)
access_events       — id, user_id, zone_id, action, status, timestamp, hash, previous_hash, device_id, entry_method
user_zone_access    — user_id, zone_id, granted_by, granted_at (PK: user_id, zone_id)
refresh_tokens      — id, user_id, token, expires_at, revoked, created_at
devices             — id, zone_id, name, type, serial, active, created_at
biometric_credentials — id, user_id, entry_method, credential_hash, enrolled_at, revoked
```

---

## 10. Graceful Shutdown

The server listens for `SIGTERM` (Render deploy) and `SIGINT` (Ctrl+C) and gives in-flight requests 30 seconds to complete before exiting. DB connection is closed cleanly on shutdown.

---

## 11. Common Pitfalls

- **`sql.ErrNoRows` in service layer** — move it to the repo. The service should not import `database/sql`.
- **`pq.Error` in service layer** — move duplicate key checks to the repo.
- **`CREATE TABLE IF NOT EXISTS`** — if a table was manually created with the wrong schema, the migration skips it silently. Drop and recreate.
- **JWT `sub` claim is `float64`** — JSON numbers decode as `float64` in Go. Always cast: `int(claims["sub"].(float64))`.
- **Rate limiter not triggering on Render free tier** — in-memory state resets on server restart. Test locally.
- **`X-Forwarded-For` can be spoofed** — acceptable for current scale; use Redis + multiple signals for production hardening.
- **Refresh token not found after logout** — correct behavior. Revoked tokens return `ErrTokenRevoked`, not `ErrTokenNotFound`.
