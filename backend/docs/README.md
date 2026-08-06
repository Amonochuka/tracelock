# TraceLock – Backend (Go + PostgreSQL

TraceLock is a biometric access control backend — a production-style Go API that tracks physical zone access events in real time, enforces permissions, and maintains a tamper-evident SHA-256 hash chain on every event.

Built incrementally with professional backend practices: small features, clear commits, and environment-based configuration.

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

- Admin-managed user provisioning, login, and bcrypt password hashing
- JWT authentication (15min access token + 7-day refresh token with hashing)
- Refresh token hashing: tokens are hashed before storage, compared by hash on lookup/revocation
- Account lockout: automatic after 5 failed login attempts, 15-minute lockout window
- Role-based access control (admin / user)
- One-time bootstrap endpoint for first admin creation (rate-limited, returns 404 after first use)
- Zone management (CRUD with capacity enforcement)
- User-zone access control (admin grants/revokes per user)
- Zone entry and exit tracking with device and entry method attribution
- Tamper-evident access event hash chain using SHA-256 (includes entry method in hash)
- Active session management (one session per user per zone)
- Biometric device management (fingerprint, face, iris, card, pin)
- Biometric credential enrollment and revocation per user
- Biometric credential hashing: values normalized and hashed server-side before storage/lookup
- Runtime biometric authentication — device scan resolves user, verifies access, creates session and issues JWT
- Live zone occupancy via WebSocket feed (`GET /ws/zones`)
- IP-based rate limiting on login, bootstrap, and admin user creation (token bucket algorithm, 5 req/min per IP)
- Chi request logging with request ID middleware
- HTTP server timeouts (5s read header, 15s read, 30s write, 60s idle)
- Automatic expired token cleanup on startup and every 24 hours
- Graceful shutdown with 30-second drain and DB connection cleanup
- Embedded PostgreSQL migrations that run automatically on startup

---

## Project Structure

```
tracelock/
├── api/                           # legacy or external API clients
├── cmd/
│   └── api/
│       ├── main.go
│       └── main_test.go
├── internal/
│   ├── access/
│   │   ├── access_repo.go
│   │   ├── access_repo_test.go
│   │   ├── access_service.go
│   │   ├── access_service_test.go
│   │   ├── biometric_service.go
│   │   ├── credential_repo.go
│   │   ├── credential_repo_test.go
│   │   ├── credential_service.go
│   │   ├── device_repo.go
│   │   ├── device_service.go
│   │   ├── errors.go
│   │   ├── hash.go
│   │   ├── hash_test.go
│   │   ├── hash_chain_test.go
│   │   ├── hub.go
│   │   └── interfaces.go
│   ├── auth/
│   │   ├── errors.go
│   │   ├── interfaces.go
│   │   ├── jwt.go
│   │   ├── middleware.go
│   │   ├── user_auth.go
│   │   └── user_service.go
│   ├── config/
│   │   └── config.go
│   ├── db/
│   │   ├── db.go
│   │   └── migrations.go
│   ├── httpdir/
│   │   ├── access_handlers.go
│   │   ├── auth_handlers.go
│   │   ├── biometric_handlers.go
│   │   ├── credential_handler.go
│   │   ├── device_handlers.go
│   │   ├── helpers.go
│   │   ├── logger.go
│   │   ├── permissions_handlers.go
│   │   ├── response.go
│   │   ├── router.go
│   │   └── middleware/
│   │       ├── apikey.go
│   │       ├── ratelimit.go
│   │       └── roles.go
│   └── models/
│       └── models.go
├── migrations/
│   ├── embed.go
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_zones_table.up.sql
│   ├── 000002_create_zones_table.down.sql
│   ├── 000003_create_devices_table.up.sql
│   ├── 000003_create_devices_table.down.sql
│   ├── 000004_create_access_events_table.up.sql
│   ├── 000004_create_access_events_table.down.sql
│   ├── 000005_create_active_sessions_table.up.sql
│   ├── 000005_create_active_sessions_table.down.sql
│   ├── 000006_create_user_zone_access_table.up.sql
│   ├── 000006_create_user_zone_access_table.down.sql
│   ├── 000007_create_refresh_tokens_table.up.sql
│   ├── 000007_create_refresh_tokens_table.down.sql
│   ├── 000008_create_biometric_credentials_table.up.sql
│   └── 000008_create_biometric_credentials_table.down.sql
├── docs/
│   ├── README.md
│   ├── Developer_guide.md
│   └── security.md
├── .env
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
DEVICE_API_KEY=your-device-api-key
ALLOWED_ORIGIN=http://localhost:3000
```

godotenv loads `.env` automatically on startup — no need to source it manually.

**Required fields:** `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `DEVICE_API_KEY`  
**Optional fields:** `PORT` (default: 8080), `DB_SSLMODE` (default: disable), `ALLOWED_ORIGIN` (default: *)

`DEVICE_API_KEY` is required for `/devices/authenticate`.  
`ALLOWED_ORIGIN` must be set to the exact frontend origin in production (e.g. `https://app.example.com`). Local development also accepts `localhost` and private `192.168.x.x:3000` frontend origins. A wildcard (`*`) is not valid alongside `Authorization` headers and will cause browser CORS rejections.

---

## Running the API

```bash
go run ./cmd/api
```

If successful:

```
Tracelock API running on: 8080
```

## Recovering an Admin Password

To reset an existing admin password without deleting any TraceLock data, run this command on the server that has the backend `.env` and database access:

```bash
go run ./cmd/reset-admin-password
```

The command prompts for the admin email and new password without echoing the password to the terminal. It resets only that admin account, clears lockout state, and revokes its refresh sessions.

---

## Endpoints

### Public

| Method | Route                   | Description                                      |
|--------|-------------------------|--------------------------------------------------|
| GET    | /health                 | Health check                                     |
| POST   | /bootstrap              | Create first admin (self-sealing, one-time only) |
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
| GET    | /zones/occupancy       | Get current zone occupancy totals    |
| GET    | /ws/zones              | WebSocket feed for live occupancy    |

### Admin only (requires role: admin)

| Method | Route                                    | Description                          |
|--------|------------------------------------------|--------------------------------------|
| GET    | /admin/ping                              | Admin access test                    |
| POST   | /admin/users                             | Create a regular user                |
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

## Notes

- `/bootstrap` is self-sealing: after the initial admin is created, it returns `404 Not Found` for subsequent calls.
- `POST /logout` and `POST /refresh` are public endpoints and do not require a JWT.
- The project currently uses embedded SQL migrations from `migrations/*.sql` and `migrations/embed.go`.
- There is a global CORS middleware (chi `cors.Handler`) as well as WebSocket origin validation, both controlled by `ALLOWED_ORIGIN`.

---

*For full developer setup, JWT internals, database notes, and common pitfalls — see the Developer Guide.*
