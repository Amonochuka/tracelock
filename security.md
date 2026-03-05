
# TraceLock – Security & Deployment Notes


## 1. Secrets Management

- Do **not** commit `.env` or real credentials
- Only commit `.env.example` with placeholders

---

## 2. JWT Security

- JWT payload is signed, **not encrypted**
- Anyone can decode payload, only the server can validate signature
- Secret (`JWT_SECRET`) must be kept private

---

## 3. PostgreSQL Authentication

- `peer` → Linux user authentication, no password shown  
- `scram-sha-256` → TCP connections require password  
- Always ensure `tracelock_user` has table and sequence privileges

---

## 4. Recommended Production Practices

- Use environment variables stored securely in the server/container  
- Rotate JWT secrets before production  
- Ensure database users have least privileges required