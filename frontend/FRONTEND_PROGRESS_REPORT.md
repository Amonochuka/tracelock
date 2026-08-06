# TraceLock Frontend Progress Report

This is the maintained product and implementation record for the TraceLock frontend. **Update this document before every frontend-related commit.** The change log must name the feature, behaviour, files changed, validation performed, and any unfinished work.

## Product Purpose

TraceLock has two deliberately different web experiences:

- **Personnel portal** at `/login` and `/dashboard` lets a regular user view only their own access, use the temporary browser-based entry simulator, and review their own audit history.
- **Administration portal** at `/admin/login` and `/admin/*` lets an administrator create and inspect zones, view live occupancy, and manage people and access policy.

The two portals are separated in the navigation and login presentation, but the real security boundary is the backend: every protected endpoint requires a signed JWT and admin endpoints additionally enforce the `admin` role. Hiding a link is useful for a cleaner user experience; it is not used as security.

## Current Frontend Experience

| Area | Status | What it does now |
| --- | --- | --- |
| Personnel sign-in | Complete | `/login` accepts only regular-user accounts for the personnel portal. It has no visible administrator or bootstrap link. |
| Admin sign-in | Complete | `/admin/login` is a direct, separate administration entry point and accepts only admin accounts. |
| Route protection | Complete | The auth context sends unauthenticated visitors away from protected pages and prevents a regular user from using `/admin/*`. |
| Personal dashboard | Complete | Shows the signed-in person’s permitted zones and personal access history. |
| Browser entry simulator | Complete | Lets a permitted user simulate entry/exit while no hardware is connected. It writes a real audited event and updates admin occupancy via WebSocket. |
| Admin occupancy dashboard | Complete | Loads current zone occupancy and receives subsequent changes through the live WebSocket feed. |
| Zone management | Complete | `/admin/zones` lists zones, provides a helpful empty state, and creates a first zone without leaving the page. |
| Zone drill-down | Partial | Shows zone configuration and currently active people. Editing, deletion, event history, and integrity verification are backend-capable but do not yet have controls in the UI. |
| Personnel management | Partial | Lists users and roles. Creating users, access grants, role changes, credential enrolment, and unlock actions remain future UI work. |

## How the Hardware-Free Demo Works

1. An administrator creates a zone at `/admin/zones`.
2. An administrator grants a regular user access using the existing backend access-control endpoint; a dedicated UI for that step is still pending.
3. The regular user signs in at `/login`, opens `/dashboard`, and selects **Simulate entry** or **Simulate exit** for an authorised zone.
4. The browser posts to the same authenticated `/zones/enter` or `/zones/exit` endpoint used by a future device integration, marked with `entry_method: "web_simulator"`.
5. The backend validates access, capacity, and exit rules; stores a hash-chained audit event; changes the active session; then broadcasts zone occupancy. The admin dashboard updates through its WebSocket connection.

This is intentionally not a fake visual counter. It exercises the real access rules and audit path, so it is suitable for demonstrating the dashboard before hardware is available. It should be treated as a development/demo input rather than a replacement for device authentication in production.

## Security Design Notes

- User and admin passwords are never present in links or route names. Both sign-in forms submit credentials only to the backend login endpoint.
- The personnel login does not advertise the admin login or first-time bootstrap screen. Administrators use the separately communicated `/admin/login` route.
- A correct password alone does not grant admin access. The backend signs the user role into the JWT, validates it on each request, and the admin route group requires the `admin` role server-side.
- Frontend redirects are convenience and usability controls. A person who manually types an admin URL still cannot call the admin API without a valid admin JWT.
- The WebSocket currently sends its JWT in a query string because browser WebSockets cannot attach an `Authorization` header. This works with the current backend middleware, but should move to a short-lived WebSocket ticket or secure cookie before a production deployment.

## File-by-File Implementation Analysis

### Application and authentication

- `src/app/layout.tsx` — Wraps every App Router page in `AuthProvider` and defines the product metadata. It is the application-wide entry point for client authentication state.
- `src/context/AuthContext.tsx` — Restores the signed-in user from browser storage, stores/removes the JWT and profile on login/logout, and applies role-aware redirects. `/admin/login` is explicitly an auth page so a logged-out administrator can reach the separate form without being redirected to the personnel portal.
- `src/components/SignInForm.tsx` — Shared client-side credential form used by both portals. It gets a backend token, verifies the profile at `/me`, and only persists the session when the account role matches the portal being used.
- `src/app/login/page.tsx` — Personnel-only entry point. It intentionally exposes no admin or bootstrap navigation.
- `src/app/admin/login/page.tsx` — Admin-only entry point. This is not linked from the personnel experience; its protection still comes from backend role enforcement.
- `src/app/bootstrap/page.tsx` — One-time first-admin initialisation. The backend is responsible for allowing it only before an admin exists.
- `src/lib/api.ts` — Central source for REST and WebSocket base URLs using `NEXT_PUBLIC_API_URL` with a local fallback.

### Personnel experience

- `src/app/dashboard/page.tsx` — Fetches the authenticated user’s `/me/events` and `/me/access` data, correctly unwraps the paginated event response, and derives the displayed in-zone state from the latest allowed event. The simulator posts an auditable `web_simulator` entry/exit event, then refreshes the user’s view.

### Admin experience

- `src/app/admin/layout.tsx` — Shared admin shell and navigation. It links to dashboard, zones, and users; its visibility is separate from API authority.
- `src/app/admin/page.tsx` — Occupancy dashboard. It fetches an initial snapshot from `/zones/occupancy`, then merges WebSocket occupancy updates for responsive live cards and capacity warnings.
- `src/app/admin/zones/page.tsx` — Newly implemented zone index. It shows zone cards, a meaningful no-zones state, and a create-zone form backed by `POST /admin/zones`.
- `src/app/admin/zones/[id]/page.tsx` — Existing operational detail view for a single zone. It loads configuration and the active people list.
- `src/app/admin/users/page.tsx` — Existing personnel list. It displays the backend records but its action menu is still a visual placeholder.

### Visual system

- `src/app/globals.css` — Defines the dark SOC design tokens and reusable layout, card, button, form, status, navigation, and table classes used across the portal.

## Known Limitations and Next Frontend Priorities

1. Add admin UI to create regular users, grant/revoke zone access, and enrol credentials so the demo does not need direct API calls for setup.
2. Add zone editing, deletion safeguards, event history, and hash-chain verification controls to the zone detail page.
3. Replace WebSocket query-string JWTs with a safer production authentication mechanism.
4. Add form-level validation, loading/error polish, and automated frontend tests.
5. Build analytic views, beginning with zone activity history and then heat maps once enough real/simulated event data exists.

## Feature Change Log

### 2026-08-06 — Role-separated portals, browser simulator, and zone management

- **Status:** Complete
- **Summary:** Added separate personnel and admin login presentations; added authenticated browser-based entry/exit simulation for authorised users; added the missing zones index with creation and empty state; fixed the personal dashboard’s handling of paginated event responses.
- **Files touched:** `src/context/AuthContext.tsx`, `src/components/SignInForm.tsx`, `src/app/login/page.tsx`, `src/app/admin/login/page.tsx`, `src/app/dashboard/page.tsx`, `src/app/admin/zones/page.tsx`, `FRONTEND_PROGRESS_REPORT.md`, `README.md`.
- **Validation:** Focused ESLint passed for all changed frontend files. Full `npm run lint` is currently blocked by an unrelated `@ts-ignore` lint error in `next.config.ts`. `npm run build` reached the Next.js font stage but could not download Inter from Google Fonts in the restricted/offline environment.
- **Follow-up:** Build the UI for user creation and zone access grants, then begin analytic views.

### 2026-08-06 — LAN development access

- **Status:** Complete
- **Summary:** Replaced a stale single-IP Next.js development-origin allowlist with a private-LAN pattern so the development app can be opened from `192.168.x.x` addresses without per-laptop edits.
- **Files touched:** `next.config.ts`, `README.md`, `FRONTEND_PROGRESS_REPORT.md`.
- **Notes:** This setting affects development-server resource and WebSocket protection only; production access must use a fixed HTTPS origin.

### 2026-08-06 — Network-aware API connection

- **Status:** Complete
- **Summary:** The frontend now derives its default API host from the browser address instead of incorrectly sending LAN-browser requests to that browser’s own `localhost:8080`.
- **Files touched:** `src/lib/api.ts`, `README.md`, `FRONTEND_PROGRESS_REPORT.md`.
- **Notes:** `NEXT_PUBLIC_API_URL` remains available when the API is intentionally hosted elsewhere.

### 2026-08-06 — Admin password recovery support

- **Status:** Complete
- **Summary:** Added a server-side, terminal-prompted admin password recovery command so an operator can restore admin access without deleting zones, users, or audit history.
- **Files touched:** `backend/cmd/reset-admin-password/main.go`, `backend/internal/auth/interfaces.go`, `backend/internal/auth/user_auth.go`, `backend/internal/auth/user_service.go`, `backend/docs/README.md`, `backend/docs/security.md`.
- **Notes:** The command is intentionally not exposed as a public web endpoint. It clears lockout state and revokes existing admin refresh sessions.

### 2026-08-06 — Bootstrap admin sign-in handoff

- **Status:** Complete
- **Summary:** Corrected the bootstrap page's existing-admin sign-in link so it leads to the separate administrator portal instead of the personnel portal.
- **Files touched:** `src/app/bootstrap/page.tsx`, `FRONTEND_PROGRESS_REPORT.md`.

## Pre-Commit Checklist

- [ ] Update this report before each frontend-related commit.
- [ ] Add a feature-log entry with exact files and real behaviour.
- [ ] Record validation performed and any failures.
- [ ] Document new routes, permissions, API calls, and known limitations.
- [ ] Confirm the change does not expose admin-only functionality through the personnel portal.
