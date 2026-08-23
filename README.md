# TraceLock

A full-stack **physical access control system** — a biometric security platform for managing zone access, tracking personnel movement, and maintaining tamper-evident audit trails in real time.

🌐 **Live Demo:** [https://tracelock-fe.onrender.com](https://tracelock-fe.onrender.com)

---

## Demo Login

This is a public demo — sign in with the shared administrator account:

> **Admin portal:** [tracelock-fe.onrender.com/admin/login](https://tracelock-fe.onrender.com/admin/login)
>
> - **Email:** `admin@tracelock.io`
> - **Password:** `admin123`

These credentials grant full admin access to the demo instance only (zones, users, devices, audit logs). They are intentionally published for evaluation — never use them outside the demo.

The backend runs a **demo guardian** job that automatically restores this account (role, lock state, password) within ~30 seconds if a visitor changes or locks it, so the demo can't be taken over permanently.

---

## What It Is

TraceLock lets a security operations team:
- Define physical zones (server rooms, lobbies, restricted areas) with capacity limits
- Grant/revoke personnel access to specific zones
- Track every entry and exit with a **SHA-256 hash chain** (tamper-evident audit log)
- Monitor live zone occupancy via a **WebSocket dashboard**
- Authenticate via biometrics — fingerprint, face scan, iris, card, or PIN — from physical devices

---

## Structure

```
tracelock/
├── backend/    # Go REST API + PostgreSQL
└── frontend/   # Next.js 16 admin + user portal
```

| Component | README |
|---|---|
| **Backend** | [backend/docs/README.md](backend/docs/README.md) |
| **Frontend** | [frontend/README.md](frontend/README.md) |

---

## Quick Start

**1. Start the database**
```bash
cd backend
docker compose up -d
```

**2. Start the backend API**
```bash
cd backend
go run ./cmd/api
```

**3. Start the frontend**
```bash
cd frontend
npm install
npm run dev
```

**4. Open your browser**
- `http://localhost:3000/bootstrap` — first-time setup (creates the admin account)
- `http://localhost:3000/login` — login for all subsequent uses

---

## Role Overview

| Role | Lands on | Can access |
|---|---|---|
| `admin` | `/admin` dashboard | Zone management, user management, audit logs, analytics |
| `user` | `/dashboard` | Personal access history and authorized zone list |

---

## Tech Stack

| Layer | Technology |
|---|---|
| API | Go, chi router, PostgreSQL |
| Auth | JWT (access + refresh tokens), bcrypt |
| Security | SHA-256 hash chain, rate limiting, account lockout |
| Frontend | Next.js 16, TypeScript, React Context |
| Realtime | WebSocket live occupancy feed |
| Dev | Docker Compose, godotenv, golang-migrate |
