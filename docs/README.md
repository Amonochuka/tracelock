# TraceLock – Backend (Go + PostgreSQL)

TraceLock is a backend service for tracking physical access events and zone activity in real time.
It is designed as a clean, production-style Go API with PostgreSQL and a modular internal architecture.

Built incrementally with professional backend practices: small features, clear commits, and environment-based configuration.

---

## Live API

**Base URL:** https://tracelock-db.onrender.com

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

---

## Current Features

- User registration and login with bcrypt password hashing
- JWT authentication with role-based middleware
- Zone entry and exit tracking
- Tamper-evident access event chain using SHA-256 hashing
- Active session management (one session per user per zone)
- Zone capacity enforcement
- Health endpoint
- PostgreSQL database schema with foreign key constraints

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
│   │   ├── auth_handler.go
│   │   ├── access_handler.go
│   │   ├── response.go
│   │   └── middleware/
│   │       └── roles.go
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
JWT_SECRET=yoursecretkey
```

Load and run:

```bash
source .env && go run ./cmd/api/main.go
```

To persist variables across terminal sessions, add them to `~/.bashrc`.

---

## Running the API

```bash
go run ./cmd/api/main.go
```

If successful:

```
Tracelock API running on: 8080
```

---

## Endpoints

### Public

| Method | Route       | Description          |
|--------|-------------|----------------------|
| GET    | /health     | Health check         |
| POST   | /register   | Register a new user  |
| POST   | /login      | Login, returns JWT   |

### Protected (requires JWT)

| Method | Route          | Description                        |
|--------|----------------|------------------------------------|
| GET    | /me            | Returns authenticated user profile |
| GET    | /protected     | Test JWT — returns user ID and role |
| GET    | /testjwt       | Confirms JWT middleware is working  |
| POST   | /zones/enter   | Enter a zone                        |
| POST   | /zones/exit    | Exit a zone                         |

### Admin only (requires role: admin)

| Method | Route         | Description      |
|--------|---------------|------------------|
| GET    | /admin/ping   | Admin access test |

---

## Live Testing

**Health check**
```bash
curl https://tracelock-db.onrender.com/health
```

**Register**
```bash
curl -X POST https://tracelock-db.onrender.com/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Amon","email":"amon@example.com","password":"password123"}'
```

**Login (returns JWT token)**
```bash
curl -X POST https://tracelock-db.onrender.com/login \
  -H "Content-Type: application/json" \
  -d '{"email":"amon@example.com","password":"password123"}'
```

Save the token for subsequent requests:
```bash
TOKEN="your_jwt_token_here"
```

**Get authenticated user**
```bash
curl https://tracelock-db.onrender.com/me \
  -H "Authorization: Bearer $TOKEN"
```

**Test JWT middleware**
```bash
curl https://tracelock-db.onrender.com/testjwt \
  -H "Authorization: Bearer $TOKEN"
```

**Protected route**
```bash
curl https://tracelock-db.onrender.com/protected \
  -H "Authorization: Bearer $TOKEN"
```

**Enter a zone**
```bash
curl -X POST https://tracelock-db.onrender.com/zones/enter \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"zone_id":1}'
```

**Exit a zone**
```bash
curl -X POST https://tracelock-db.onrender.com/zones/exit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"zone_id":1}'
```

**Admin endpoint (requires admin role)**
```bash
curl https://tracelock-db.onrender.com/admin/ping \
  -H "Authorization: Bearer $TOKEN"
```

---

*For full developer setup, JWT internals, database notes, and common pitfalls — see the Developer Guide.*
