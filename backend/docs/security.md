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
- **Hashed before storage:** SHA-256 hash of the token is stored in `refresh_tokens.token_hash`
- Raw token is sent to the client; the server never persists the raw value
- On lookup or revocation, the incoming token is hashed and compared against the stored hash
- This prevents plaintext token exposure if the database is compromised
- Revoked on logout — subsequent refresh attempts return `401 Unauthorized` with `ErrTokenRevoked`
- Expired tokens return `401 Unauthorized` with `ErrTokenExpired`
- **Automatic token cleanup:** expired tokens are deleted immediately on startup and every 24 hours

---

## 5. Biometric Data Security

Raw biometric data (fingerprints, face scans, iris patterns) is **never stored or transmitted** by TraceLock.

The scanner SDK processes the biometric locally and produces a tokenised feature vector. This token is:
1. **Normalized** on the server (hex values decoded; raw bytes preserved)
2. **Hashed with SHA-256** before storage
3. Stored only as the hash in `biometric_credentials.credential_hash`

This means:
- A database breach does not expose raw biometric data or scanner tokens
- Hashes cannot be reversed to reconstruct the original biometric or token
- Multiple representations of the same credential are normalized to a single canonical hash
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

Login, bootstrap, and admin user-creation endpoints are rate limited to 5 requests per minute per IP using a token bucket algorithm. Exceeding the limit returns `429 Too Many Requests`.

Known limitations:
- State is in-memory — resets on server restart
- `X-Forwarded-For` can be spoofed by a sophisticated attacker
- Does not prevent slow distributed brute force attacks

Planned hardening: Redis-backed rate limiting for multi-instance deployments.

## 10. Account Lockout

After 5 failed login attempts on the same email, the account is automatically locked for 15 minutes:
- Subsequent login attempts return `429 Too Many Requests` with "account is temporarily locked"
- Successful login resets the counter and unlocks the account
- Admin can manually unlock via `PUT /admin/users/{id}/unlock`
- Lockout timestamp is stored in `users.locked_until` and checked on every authentication attempt

## 11. Bootstrap Security

`POST /bootstrap` is a public endpoint but hardened with:
- **Rate limiting:** same 5 req/min per IP limit as login
- **Self-sealing:** checks for any existing admin before creating one
- **After first use:** returns `404 Not Found` (misleading response) instead of `403` to avoid revealing admin existence

## 12. User Provisioning

Regular-user self-registration is disabled. `POST /admin/users` requires both a valid JWT and the `admin` role, so only an administrator can create a user account. The public account-creation path is limited to the self-sealing first-admin bootstrap flow.

## 13. Admin Password Recovery

There is no public password-reset endpoint. A deployment operator with direct server and database access can run `go run ./cmd/reset-admin-password`, which securely prompts for an existing admin email and new password. It resets the password, clears lockout state, and revokes the admin's active refresh sessions without deleting users, zones, or access history.

## 14. PostgreSQL Authentication

- `peer` → authenticates via Linux username, no password required for local socket connections
- `scram-sha-256` → TCP connections require password
- App DB user should have least privileges — only table and sequence access, not superuser

---

## 15. Production Checklist

- [ ] Rotate `JWT_SECRET` before going live
- [ ] Rotate `DEVICE_API_KEY` and secure its distribution to devices
- [ ] Use a dedicated DB user with least privileges
- [ ] Ensure `.env` is in `.gitignore` and not in git history
- [ ] Run migrations as superuser, app connects as restricted user
- [ ] Use HTTPS — JWT tokens and refresh tokens in plain HTTP are exposed in transit
- [ ] Verify account lockout is working (5 failed attempts → 15 min lockout)
- [ ] Verify token cleanup runs on startup and every 24 hours
- [ ] Replace in-memory rate limiter with Redis for multi-instance deployments
- [ ] Consider 2FA for admin accounts
- [ ] Enable WebSocket over WSS (WebSocket Secure) in production
- [ ] Test graceful shutdown (30 sec drain) with in-flight requests

---

## 16. Known Limitation — System Timeout Exits on Strict Zones

When the backend starts up (or on the periodic stale session sweep), `CleanupStaleSessions` force-closes any `active_sessions` older than the configured threshold. For each stale session it writes an `exit / allowed` event with `reason: system_timeout` directly to the hash chain — **bypassing the `requires_exit_scan` check**.

**Why this is a concern:**
- A `system_timeout` exit on a zone with `requires_exit_scan = true` (e.g. Server Room) means the system lost track of the session without a physical exit scan.
- This could mean the person is **still physically inside** the zone — the backend simply cleaned up a stale record.
- The audit log will show an `exit / allowed / system_timeout` entry, which could be misread as a legitimate scan-out.

**Current behaviour:**
- The session is removed and zone occupancy is updated as if the person left.
- No alert is raised, no zone is locked.

**Planned hardening (not yet implemented):**
1. For `requires_exit_scan` zones, log `system_timeout` exits with `status: "flagged"` rather than `"allowed"` to distinguish them from legitimate scan-outs.
2. Notify admins (via WebSocket push or a dedicated alert endpoint) when a flagged timeout occurs on a strict zone.
3. Optionally lock the zone for new entry until an admin acknowledges the anomaly.

Until this is implemented, administrators should treat any `system_timeout` exit on a strict zone as a **security anomaly requiring manual verification**.

