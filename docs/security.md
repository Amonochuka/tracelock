# TraceLock – Security Notes

---

## 1. Secrets Management

- Never commit `.env` or real credentials to version control
- Only commit `.env.example` with placeholder values
- For local dev, store variables in `~/.bashrc` so they persist across sessions
- For production, use environment variables injected by the server or container runtime

---

## 2. Password Hashing

Passwords are hashed using bcrypt before storage. Plain text passwords are never stored or logged.

```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

On login, bcrypt compares the submitted password against the stored hash without ever decrypting it.

Hashing is handled in the `UserAuth` layer — the service and handler never see or handle raw password comparison logic.

---

## 3. JWT Security

- JWT payload is signed with HMAC-SHA256, not encrypted
- Anyone can decode the payload — do not store sensitive data in claims
- Only the server can validate the signature using `JWT_SECRET`
- Tokens expire after 24 hours
- `JWT_SECRET` must be kept private and rotated before production

Current claims stored in token:

```json
{
  "sub": 1,
  "role": "user",
  "exp": 1234567890,
  "iat": 1234567890
}
```

---

## 4. User Enumeration Protection

The API returns the same error for both wrong email and wrong password:

```json
{"error": "invalid email or password"}
```

This prevents attackers from discovering which emails are registered in the system. Internally, `ErrInvalidCredentials` is used for both cases.

---

## 5. Role-Based Access Control

Routes are protected by JWT middleware. Admin-only routes additionally require `role=admin` in the token claims, enforced by `middleware.RequireRole("admin")`.

Roles are stored in the `users` table and embedded in the JWT at login time. To promote a user to admin:

```sql
UPDATE users SET role = 'admin' WHERE email = 'user@example.com';
```

The user must log in again to get a new token reflecting the updated role.

---

## 6. Zone Access Integrity

Each access event stores a SHA-256 hash that chains the previous event's hash — similar in concept to a blockchain. This provides tamper evidence: if any event record is altered or deleted, the hash chain breaks.

The `active_sessions` table uses a composite primary key on `(user_id, zone_id)`, enforced at the database level, preventing duplicate active sessions.

---

## 7. PostgreSQL Authentication

- `peer` → authenticates via Linux username, no password required for local connections
- `scram-sha-256` → TCP connections require password
- App DB user should have least privileges required — only table and sequence access, not superuser

---

## 8. Production Checklist

- [ ] Rotate `JWT_SECRET` before going live
- [ ] Use a dedicated DB user with least privileges
- [ ] Ensure `.env` is in `.gitignore`
- [ ] Run migrations as a superuser, app connects as a restricted user
- [ ] Use HTTPS — JWT tokens in plain HTTP are exposed in transit
- [ ] Set appropriate token expiry for your use case
