
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

# TraceLock – Backend README

This document summarizes how environment variables, PostgreSQL access, and JWT authentication work in this project (based on the questions discussed during development).

---

## 1. Environment variables

You configure the app using environment variables (usually from your shell or a `.env` file):

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=tracelock_user
DB_PASSWORD=your_password
DB_NAME=tracelock
JWT_SECRET=your_production_secret
```

### What each one is for

* `DB_*` → used to open the PostgreSQL connection.
* `JWT_SECRET` → secret key used to sign and verify JWT tokens.

Important:

* These values are read by your Go process using `os.Getenv(...)`.
* They belong to the **server**, not to the client.

---

## 2. Do not commit real secrets

You should commit only an example file:

```
.env.example
```

Example:

```
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
JWT_SECRET=
```

Your real `.env` (or shell exports) must NOT be committed.

---

## 3. Where do these variables live in production?

When the app is used by a company or deployed to a server:

* The server (VM / container / hosting platform) stores the environment variables.
* The Go binary reads them at runtime.

The users of your API never see them.

---

## 4. JWT initialization

```go
func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set")
	}
	jwtSecret = []byte(secret)
}
```

What this does:

* Reads `JWT_SECRET` from the environment.
* Converts it to `[]byte`.
* Stores it in the global variable `jwtSecret`.

`jwtSecret` is the key used for signing and verifying tokens.

---

## 5. Generating a JWT

```go
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

Important facts:

* `claims` is a Go map (`map[string]interface{}`).
* It becomes the JWT **payload**.
* The library:

  * encodes it as JSON
  * base64-encodes it
  * signs it using `jwtSecret`

The result is a string:

```
header.payload.signature
```

---

## 6. What is inside the JWT string

A JWT has three parts:

* header (algorithm info)
* payload (your claims: sub, role, exp, iat)
* signature

The payload is base64 encoded. That is why you do not see `role` directly when you look at the raw token.

---

## 7. Parsing and verifying a JWT (middleware)

```go
token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrTokenInvalidMethod
	}
	return jwtSecret, nil
})
```

This function does NOT return claims.

It only returns the key to the JWT library.

The library:

* splits the token
* verifies the signature using `jwtSecret`
* decodes the payload

After that you read the claims from the parsed token:

```go
claims, ok := token.Claims.(jwt.MapClaims)
```

---

## 8. Why the callback returns `interface{}`

The JWT library supports multiple algorithms:

* HMAC → `[]byte`
* RSA → `*rsa.PublicKey`
* ECDSA → `*ecdsa.PublicKey`

So the API uses:

```
(interface{}, error)
```

In this project you always return:

```
[]byte
```

because you use HMAC (HS256).

`jwtSecret` itself is NOT an interface.

It is:

```
[]byte
```

It is only returned through an interface.

---

## 9. Extracting values from claims

```go
userID, ok := claims["sub"].(float64)
role, ok2 := claims["role"].(string)
```

Notes:

* JSON numbers become `float64` when decoded into `map[string]interface{}`.
* That is why `sub` is read as `float64` and then converted to `int`.

---

## 10. Important correction from earlier bug

Your token must be created using the authenticated user object:

```
GenerateToken(user)
```

not from a global variable.

Otherwise `role` will be empty in the token.

---

## 11. Typical request flow

1. Client sends login request.
2. Server authenticates user from the database.
3. Server generates JWT using `GenerateToken(user)`.
4. Client stores the token.
5. Client sends the token in requests:

```
Authorization: Bearer <token>
```

6. Middleware parses the token.
7. Middleware extracts `sub` and `role`.
8. Handlers use those values for authorization.

---

## 12. About PostgreSQL authentication and sudo

When you run:

```
sudo -u postgres psql
```

and your `pg_hba.conf` contains:

```
local   all   postgres   peer
```

PostgreSQL uses Linux user authentication (peer).

That means:

* your Linux user identity is trusted
* no password prompt is shown

When you connect using TCP:

```
host  all  all  127.0.0.1/32  scram-sha-256
```

PostgreSQL uses password authentication.

So the password is still used — you just do not see it when using peer auth.

---

## 13. Should real secrets appear in README?

No.

The README should only describe the variable names.

Example:

```
JWT_SECRET=...
DB_PASSWORD=...
```

Never place real secrets in README, even for a portfolio project.

---

## 14. Final important points

* JWT payload is not encrypted. It is only signed.
* Anyone can decode the payload.
* Only the server can create a valid signature (because of `JWT_SECRET`).
* The secret must be changed before real deployment.

---

## 15. Quick mental model

* `jwtSecret` → signing / verification key
* JWT string → encoded header + encoded payload + signature
* `jwt.Parse` → verifies signature and extracts payload
* `token.Claims` → your decoded claims

Role-Based Access Control,