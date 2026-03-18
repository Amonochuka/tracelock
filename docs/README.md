# TraceLock – Backend (Go + PostgreSQL)

TraceLock is a backend service for tracking access events and zone activity in real time.  
It is designed as a clean, production-style Go API with PostgreSQL and a modular internal architecture.

This project is being built incrementally with professional backend practices: small features, clear commits, and environment-based configuration.

---

## Tech Stack

- Go
- PostgreSQL
- chi router
- bcrypt for password hashing
- JWT (JSON Web Token) for authentication

---

## Current Features

- PostgreSQL database schema
- Go project layout using `cmd/` and `internal/`
- Database connection using environment variables
- User registration endpoint
- Secure password hashing using bcrypt
- Health endpoint (`GET /health`)

---

## Project Structure

```
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
```

## Environment Variables

Configure the app using environment variables (do not commit real secrets):

```
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
JWT_SECRET=


- `DB_*` → database connection  
- `JWT_SECRET` → used to sign and verify JWT tokens

```

## Running the API

From the project root:

```bash
go run ./cmd/api
```

If successful, the server logs:

Connected to database successfully!

## Endpoints
 - Health Check

 - GET /health → Response: ok

 - User Registration

 - POST /register

```
Example body:

{
  "name": "Amon",
  "email": "amon@example.com",
  "password": "mypassword"
}
```
```
Example curl:

curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Amon","email":"amon@example.com","password":"mypassword"}'
```

## Login & JWT

 - POST /login → returns JWT token

 - Protected endpoints:

 - /me → returns authenticated user data

 - /testjwt → tests JWT validation

JWT payload is signed, not encrypted. Do not commit secrets.

 ***For full developer setup and deeper explanation of JWT, database, and authentication, see Developer Guide***


why use fmt.Errorf

Now the HTTP response actually tells the client something went wrong.

This is how Go handles errors idiomatically — you return errors, don’t just print.

Rule of thumb:

Use fmt.Print only for debugging or logging.

Use return fmt.Errorf(...) (or errors.New) for real errors that the caller must handle.