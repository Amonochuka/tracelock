# TraceLock – Security Notes

---

## 1. Secrets Management

- Never commit `.env` or real credentials to version control
- `.env` is in `.gitignore` — verify with `git ls-files | grep .env`
- For production, use environment variables injected by the server (Render dashboard)
- Rotate `JWT_SECRET` and DB password immediately if accidentally exposed

---

## 2. Password Hashing

Passwords are hashed using bcrypt before storage. Plain text passwords are never stored or logged.

```go
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

On login, bcrypt compares the submitted password against the stored hash without ever decrypting it. Hashing is handled in the `UserAuth` layer — the service and handler never handle raw password comparison logic.

---

## 3. JWT Security

- JWT payload is signed with HMAC-SHA256, not encrypted
- Anyone can decode the payload — do not store sensitive data in claims
- Only the server can validate the signature using `JWT_SECRET`
- Access tokens expire after 15 minutes
- Refresh tokens expire after 7 days and are stored in DB
- Refresh tokens are revoked on logout
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

## 4. Refresh Token Security

- Refresh tokens are random 32-byte hex strings generated with `crypto/rand`
- Stored in the `refresh_tokens` table with expiry and revocation flag
- Revoked on logout — subsequent refresh attempts return `403`
- Expired tokens return `401`
- Token cleanup job (purge expired tokens) is planned for a future phase

---

## 5. Biometric Data Security

Raw biometric data (fingerprints, face scans, iris patterns) is **never stored or transmitted** by TraceLock.

The scanner SDK processes the biometric locally and produces a tokenised feature vector. Only the hash of this token is stored in `biometric_credentials.credential_hash`. This means:

- A database breach does not expose biometric data
- Hashes cannot be reversed to reconstruct the original biometric
- Each credential is unique per user per method

For testing, credential hashes are simulated with:
```bash
openssl rand -hex 32
```

---

## 6. User Enumeration Protection

The API returns the same error for both wrong email and wrong password:

```json
{"error": "invalid email or password"}
```

This prevents attackers from discovering which emails are registered. Internally, `ErrInvalidCredentials` covers both cases.

---

## 7. Role-Based Access Control

Routes are protected by JWT middleware. Admin-only routes additionally require `role=admin` in the token claims, enforced by `middleware.RequireRole("admin")`.

Roles are stored in the `users` table and embedded in the JWT at login time. Role changes take effect on next login — existing tokens retain the old role until expiry.

Admin promotion flow:
1. `POST /bootstrap` — creates first admin directly
2. `PUT /admin/users/{id}/role` — all subsequent promotions via API (admin JWT required)

Raw SQL role updates should never be used in production.

---

## 8. Zone Access Integrity

Each access event stores a SHA-256 hash chaining the previous event — providing tamper evidence. If any event record is altered or deleted, the chain breaks and `GET /admin/zones/{id}/verify-chain` will detect it.

The hash includes: `userID`, `zoneID`, `action`, `timestamp`, `previousHash`, `entryMethod` — making each event cryptographically unique.

---

## 9. Rate Limiting

Login and register endpoints are rate limited to 5 requests per minute per IP using a token bucket algorithm. Exceeding the limit returns `429 Too Many Requests`.

Known limitations:
- State is in-memory — resets on server restart
- `X-Forwarded-For` can be spoofed by a sophisticated attacker
- Does not prevent slow distributed brute force attacks

Planned hardening: account lockout after repeated failed attempts, Redis-backed rate limiting for multi-instance deployments.

---

## 10. Bootstrap Security

`POST /bootstrap` is a public endpoint but self-sealing — it checks for any existing admin before creating one. After the first successful call it permanently returns `403`. This prevents privilege escalation on a fresh deploy.

---

## 11. PostgreSQL Authentication

- `peer` → authenticates via Linux username, no password required for local socket connections
- `scram-sha-256` → TCP connections require password
- App DB user should have least privileges — only table and sequence access, not superuser

---

## 12. Production Checklist

- [ ] Rotate `JWT_SECRET` before going live
- [ ] Use a dedicated DB user with least privileges
- [ ] Ensure `.env` is in `.gitignore` and not in git history
- [ ] Run migrations as superuser, app connects as restricted user
- [ ] Use HTTPS — JWT tokens in plain HTTP are exposed in transit
- [ ] Enable account lockout after repeated failed login attempts
- [ ] Set up token cleanup job for expired refresh tokens
- [ ] Replace in-memory rate limiter with Redis for multi-instance deployments
- [ ] Consider 2FA for admin accounts
