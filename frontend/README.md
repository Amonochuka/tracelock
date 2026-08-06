# TraceLock Frontend

A **Next.js 16** admin and user portal for the TraceLock physical access control system. Built with TypeScript and the App Router.

---

## Tech Stack

- **Next.js 16** (App Router, Turbopack)
- **TypeScript**
- **React Context API** for auth state
- **Vanilla CSS** with CSS custom properties (dark SOC theme)
- **lucide-react** for icons
- **WebSocket** for live zone occupancy feed

---

## Getting Started

```bash
cd frontend
npm install
npm run dev
```

App runs at `http://localhost:3000`.

> **Prerequisite:** The backend API must be running on port `8080`. See [`backend/docs/README.md`](../backend/docs/README.md).

---

## Pages

| Route | Role | Description |
|---|---|---|
| `/bootstrap` | Public | First-time admin account creation |
| `/login` | Public | Login for all users |
| `/admin` | Admin | Live zone occupancy dashboard (WebSocket) |
| `/admin/zones/:id` | Admin | Zone drilldown — occupancy, active personnel |
| `/admin/users` | Admin | Personnel list with role badges |
| `/dashboard` | User | Personal access history and authorized zones |

### Role-Based Access

Authentication and routing is managed by `AuthContext`:
- **Admin** accounts (`role: "admin"`) are redirected to `/admin` after login
- **Regular user** accounts (`role: "user"`) are redirected to `/dashboard` after login
- Attempting to visit `/admin` as a non-admin silently redirects to `/dashboard`
- Unauthenticated access to any protected route redirects to `/login`

---

## API

All API calls use `window.location.hostname` so the app works on both `localhost` and local network IPs (192.168.x.x) without changes.

The backend is expected at port `8080` on the same host.

---

## Network Testing

To access the app from other devices on your local network, add your machine's IP to `next.config.ts`:

```ts
allowedDevOrigins: ['192.168.x.x'],
```
