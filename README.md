
Do you need goroutines, mutexes for tracelock?

Most likely: NO, at least for phase 1-2

Why:

Database access is already safe:

PostgreSQL handles multiple connections itself

Each HTTP request gets its own DB transaction

JWT / login / register:

Each request is independent

No shared in-memory state needing locks

HTTP server:

net/http + chi already handles concurrent requests safely

At your stage, no mutexes or special concurrency code needed

Database + Go’s HTTP server concurrency is enough

For massive scale, the key is stateless design, DB consistency, and load balancing

# TraceLock – Backend (Go + PostgreSQL)

TraceLock is a backend service for tracking access events and zone activity in real time.  
It is designed as a clean, production-style Go API with PostgreSQL and a modular internal
architecture.

This project is being built incrementally with professional backend practices:
small features, clear commits, and environment-based configuration.

---

## Current Features (so far)

- PostgreSQL database schema

- Go project layout using `cmd/` and `internal/`

- Database connection using environment variables

- User registration endpoint

- Secure password hashing using bcrypt

- Health endpoint

---

## Project Structure
tracelock/
├── cmd/
│ └── api/
│ └── main.go
├── internal/
│ ├── auth/
│ │ └── auth.go
│ ├── db/
│ │ └── db.go
│ └── httpapi/
│ └── router.go
└── migrations/
└── 01_init.sql


---

## Environment Variables

The API uses environment variables for database configuration:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=tracelock_user
export DB_PASSWORD=yourpassword
export DB_NAME=tracelock

## Database Schema

  - The database is created using SQL migrations in:

    - migrations/tables.sql

  - Current tables:

    - users

    - zones

    - access_events

    - active_sessions

## Running the API

  - From the project root:

    - go run ./cmd/api

If the connection is successful, the server will start and log:

  - Connected to database successfully!

## Health Check

GET /health

Response:
ok

## User Registration
Endpoint
POST /register

Body
{
  "name": "Amon",
  "email": "amon@example.com",
  "password": "mypassword"
}

Example
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Amon","email":"amon@example.com","password":"mypassword"}'

## Password Handling

Passwords are never stored in plain text.

They are hashed using:

golang.org/x/crypto/bcrypt

before being stored in the database.

## Important PostgreSQL Permission Issue (Solved)

While testing registration, the following error occurred:

pq: permission denied for table users

What this meant

The database connection was working correctly, but the database user did not have
permission to access the tables.

In this setup:

Tables were created using the PostgreSQL superuser:

postgres


The application connects using:

tracelock_user


In PostgreSQL, granting access to a database does not automatically grant access
to tables inside the database.

This is a very common PostgreSQL pitfall.
## How it was fixed

The solution was to grant privileges on all existing tables and sequences
to the application user.

Login as the PostgreSQL superuser:

sudo -u postgres psql


Connect to the TraceLock database:

\c tracelock


Then run:

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO tracelock_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO tracelock_user;


After this, the API was able to insert users successfully.

Why both TABLES and SEQUENCES are needed

PostgreSQL uses sequences for SERIAL / auto-increment columns.

Without permissions on sequences, inserts may fail even if table permissions exist.
psql -U tracelock_user -d tracelock -h 127.0.0.1,  to prompt password in DB, to avoid sudo masking

🛠 Tech Stack

Go

PostgreSQL

chi router

bcrypt

standard database/sql package

install JWT
Login Endpoint + JWT
go get github.com/golang-jwt/jwt/v5


JWT = JSON Web Token

That’s the full name.

Break it down simply:

✅ JSON

Because the data inside the token is a JSON object.

Example payload (conceptually):

{
  "sub": 12,
  "role": "admin"
}

✅ Web

Because it’s designed to be used over HTTP / web APIs.

✅ Token

Because it is a small string that represents:

“this user is authenticated”

and is sent with every request.

So:

JSON Web Token = a signed JSON object used as an authentication token for web APIs.


# Tracelock

**Tracelock** is a production-ready Go backend project implementing secure authentication with JWT and PostgreSQL. The goal is to build a minimal yet professional backend service demonstrating user authentication, JWT middleware, and protected routes.

---

## Features Implemented

- PostgreSQL database setup with proper user privileges
- User registration (`/register`) with password hashing using bcrypt
- User login (`/login`) generating JWT tokens
- JWT middleware validating tokens and storing user info in request context
- Protected endpoint `/testjwt` to test JWT validation
- Protected endpoint `/me` returning the authenticated user’s data
- Environment-based configuration for DB credentials and JWT secret

---

## How JWT Works in Tracelock

- JWT tokens include the `sub` claim (user ID), `exp` (expiry), and `iat` (issued at)
- Tokens are signed using HMAC with a secret stored in the environment
- JWT middleware extracts `sub` from token claims and stores it in context for handlers
- Handlers retrieve the user ID safely via:

```go
userClaims, ok := r.Context().Value(UserContextKey).(*UserClaims)
Challenges & Lessons Learned

Environment variables

JWT secret and DB credentials must be set in the environment.

Common pitfall: variables only live in the current shell unless exported in ~/.bashrc.

JWT token type casting

sub from JWT claims is decoded as float64. Type assertion needed to convert to int.

Database permissions

PostgreSQL requires granting privileges to the user (GRANT ALL PRIVILEGES ON ALL TABLES…) — creating tables as postgres alone is insufficient.

Context-based user info

Storing a struct in context (UserClaims) is the professional way to pass authenticated user info to handlers.