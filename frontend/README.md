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

## Frontend Progress Report

The current implementation status and feature history are tracked in [FRONTEND_PROGRESS_REPORT.md](FRONTEND_PROGRESS_REPORT.md). Please update this document before each frontend-related commit.

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
| `/login` | Public | Personnel login |
| `/admin/login` | Public | Separate administrator login |
| `/admin` | Admin | Live zone occupancy dashboard (WebSocket) |
| `/admin/zones` | Admin | Zone list, empty state, and zone creation |
| `/admin/zones/:id` | Admin | Zone drilldown — occupancy, active personnel |
| `/admin/users` | Admin | Personnel list with role badges |
| `/dashboard` | User | Personal access history and authorized zones |

### Role-Based Access

Authentication and routing is managed by `AuthContext`:
- **Admin** accounts sign in at `/admin/login` and are redirected to `/admin`
- **Regular user** accounts sign in at `/login` and are redirected to `/dashboard`
- Attempting to visit `/admin` as a non-admin silently redirects to `/dashboard`
- Unauthenticated access to any protected route redirects to `/login`

## Hardware-Free Demo

Until a device is connected, an authorised regular user can use the **Simulate entry** and **Simulate exit** controls on `/dashboard`. These call the real authenticated access endpoints with `entry_method: "web_simulator"`, so they are recorded in the audit history and update the admin occupancy dashboard through the live WebSocket feed.

---

## API

All API calls use the hostname currently open in the browser, so the app works on both `localhost` and local network IPs (`192.168.x.x`) without changes. Set `NEXT_PUBLIC_API_URL` only when the API is hosted on a different server or port.

The backend is expected at port `8080` on the same host.

---

## Network Testing

The development server accepts private `192.168.x.x` LAN addresses through `allowedDevOrigins`, so changing between laptops on that network does not require a config edit. Restart `npm run dev` after changing this setting.

```ts
allowedDevOrigins: ['192.168.*.*'],
```
