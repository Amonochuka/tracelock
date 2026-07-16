# TraceLock – Backend (Go + PostgreSQL)

TraceLock is a biometric access control backend — a production-style Go API that tracks physical zone access events in real time, enforces permissions, and maintains a tamper-evident SHA-256 hash chain on every event.

Built incrementally with professional backend practices: small features, clear commits, and environment-based configuration.

---

## Live API

**Base URL:** https://https://tracelock.onrender.com

**GitHub:** https://github.com/Amonochuka/tracelock

> The API does not expose a root route — opening the base URL returns 404. Use the endpoints below.
> Hosted on Render's free tier, so the first request may take a few seconds while the backend spins up.

---

## Tech Stack

- Go
- PostgreSQL
- chi router
- bcrypt for password hashing
- JWT (JSON Web Token) for authentication
- SHA-256 hash chain for tamper-evident audit trail
- godotenv for local environment loading

---

## Current Features

- User registration, login and bcrypt password hashing
- JWT authentication (15min access token + 7-day refresh token)
- Role-based access control (admin / user)
- One-time bootstrap endpoint for first admin creation
- Zone management (CRUD with capacity enforcement)
- User-zone access control (admin grants/revokes per user)
- Zone entry and exit tracking with device and entry method attribution
- Tamper-evident access event hash chain using SHA-256
- Active session management (one session per user per zone)
- Biometric device management (fingerprint, face, iris, card, pin)
- Biometric credential enrollment and revocation per user
- Runtime biometric authentication — device scan resolves user, verifies access, creates session and issues JWT
- IP-based rate limiting on login and register (token bucket algorithm)
- Chi request logging and request ID middleware
- Graceful shutdown with 30-second drain and DB connection cleanup
- PostgreSQL migrations run automatically on startup

---

## Project Structure

```
tracelock/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── access/
│   │   ├── access_repo.go
│   │   ├── access_service.go
│   │   ├── device_repo.go
│   │   ├── device_service.go
│   │   ├── credential_repo.go
│   │   ├── credential_service.go
│   │   ├── biometric_service.go
│   │   ├── hash.go
│   │   └── errors.go
│   ├── auth/
│   │   ├── user_auth.go
│   │   ├── user_service.go
│   │   ├── jwt.go
│   │   ├── middleware.go
│   │   └── errors.go
│   ├── db/
│   │   ├── db.go
│   │   └── migrations.go
│   ├── httpdir/
│   │   ├── router.go
│   │   ├── auth_handlers.go
│   │   ├── access_handlers.go
│   │   ├── zone_handlers.go
│   │   ├── permissions_handlers.go
│   │   ├── user_handlers.go
│   │   ├── device_handlers.go
│   │   ├── credential_handlers.go
│   │   ├── biometric_handlers.go
│   │   ├── helpers.go
│   │   ├── response.go
│   │   └── middleware/
│   │       ├── roles.go
│   │       └── ratelimit.go
│   ├── models/
│   │   └── models.go
│   └── config/
│       └── config.go
├── migrations/
│   └── tables.sql
├── docs/
│   ├── README.md
│   ├── Developer_guide.md
│   └── security.md
├── .gitignore
├── go.mod
└── go.sum
```

---

## Environment Variables

Create a `.env` file in the project root (never commit real secrets):

```
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=tracelock
DB_SSLMODE=disable
JWT_SECRET=yoursecretkey
```

godotenv loads `.env` automatically on startup — no need to source it manually.

---

## Running the API

```bash
go run ./cmd/api
```

If successful:

```
Tracelock API running on: 8080
```

---

## Endpoints

### Public

| Method | Route                   | Description                                      |
|--------|-------------------------|--------------------------------------------------|
| GET    | /health                 | Health check                                     |
| POST   | /bootstrap              | Create first admin (self-sealing, one-time only) |
| POST   | /register               | Register a new user                              |
| POST   | /login                  | Login — returns access token + refresh token     |
| POST   | /refresh                | Get new access token using refresh token         |
| POST   | /logout                 | Revoke refresh token                             |
| POST   | /devices/authenticate   | Biometric scanner authentication                 |

### Protected (requires JWT)

| Method | Route                  | Description                          |
|--------|------------------------|--------------------------------------|
| GET    | /me                    | Authenticated user profile           |
| GET    | /me/events             | Current user's access history        |
| GET    | /me/access             | Zones current user can enter         |
| GET    | /protected             | Test JWT — returns user ID and role  |
| GET    | /testjwt               | Confirms JWT middleware is working   |
| POST   | /zones/enter           | Enter a zone                         |
| POST   | /zones/exit            | Exit a zone                          |
| GET    | /zones                 | List all zones with live occupancy   |
| GET    | /zones/{id}            | Zone detail with active users        |

### Admin only (requires role: admin)

| Method | Route                                    | Description                          |
|--------|------------------------------------------|--------------------------------------|
| GET    | /admin/ping                              | Admin access test                    |
| GET    | /admin/users                             | List all users                       |
| PUT    | /admin/users/{id}/role                   | Update user role                     |
| GET    | /users/{id}/events                       | User access history                  |
| GET    | /users/{id}/access                       | Zones a user can enter               |
| POST   | /admin/zones                             | Create zone                          |
| PUT    | /admin/zones/{id}                        | Update zone                          |
| DELETE | /admin/zones/{id}                        | Delete zone                          |
| GET    | /admin/zones/{id}/users                  | Users with access to a zone          |
| GET    | /zones/{id}/events                       | Paginated event log for a zone       |
| GET    | /admin/zones/{id}/verify-chain           | Verify hash chain integrity          |
| POST   | /admin/access                            | Grant user access to a zone          |
| DELETE | /admin/access                            | Revoke user access to a zone         |
| POST   | /admin/zones/{id}/devices                | Register a device to a zone          |
| GET    | /admin/zones/{id}/devices                | List devices in a zone               |
| GET    | /admin/devices/{id}                      | Get a device                         |
| PUT    | /admin/devices/{id}                      | Update a device                      |
| PATCH  | /admin/devices/{id}/deactivate           | Deactivate a device                  |
| DELETE | /admin/devices/{id}                      | Delete a device                      |
| POST   | /admin/users/{id}/credentials            | Enroll biometric credential          |
| GET    | /admin/users/{id}/credentials            | List user credentials                |
| GET    | /admin/users/{id}/credentials/{method}   | Get credential by method             |
| DELETE | /admin/users/{id}/credentials/{method}   | Revoke credential                    |

---

*For full developer setup, JWT internals, database notes, and common pitfalls — see the Developer Guide.*
