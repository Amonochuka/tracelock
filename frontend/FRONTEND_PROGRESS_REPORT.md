# TraceLock Frontend Progress Report

This is the maintained product and implementation record for the TraceLock frontend. **Update this document before every frontend-related commit.** The change log must name the feature, behaviour, files changed, validation performed, and any unfinished work.

## Product Purpose

TraceLock has two deliberately different web experiences:

- **Personnel portal** at `/login` and `/dashboard` lets a regular user view only their own access, use the temporary browser-based entry simulator, and review their own audit history.
- **Administration portal** at `/admin/login` and `/admin/*` lets an administrator create zones, view live occupancy, manage people and access policy, register hardware devices, enrol biometric credentials, and run authenticated hardware simulations.

The two portals are separated in the navigation and login presentation, but the real security boundary is the backend: every protected endpoint requires a signed JWT and admin endpoints additionally enforce the `admin` role. Hiding a link is useful for a cleaner user experience; it is not used as security.

## Current Frontend Experience

| Area | Status | What it does now |
| --- | --- | --- |
| Personnel sign-in | Complete | `/login` accepts only regular-user accounts for the personnel portal. It has no visible administrator or bootstrap link. |
| Admin sign-in | Complete | `/admin/login` is a direct, separate administration entry point and accepts only admin accounts. |
| Route protection | Complete | The auth context sends unauthenticated visitors away from protected pages and prevents a regular user from using `/admin/*`. |
| Personal dashboard | Complete | Shows the signed-in person’s permitted zones and personal access history. |
| Browser entry simulator | Complete | Lets a permitted user simulate entry/exit while no hardware is connected. It writes a real audited event and updates admin occupancy via WebSocket. |
| Admin occupancy dashboard | Complete | Loads current zone occupancy and receives subsequent changes through the live WebSocket feed. Right-hand column shows Zone Analytics (Peak Entry Times bar chart) with a per-zone selector. |
| Zone management | Complete | `/admin/zones` lists zones, provides a helpful empty state, and creates a first zone without leaving the page. |
| Zone drill-down | Complete | Tabbed view with Overview (active personnel), Event History (paginated log + hash-chain verify), and Settings (inline edit + guarded deletion). |
| Personnel management | Complete | Lists users and roles. Create, Manage Access (zone grant/revoke toggles), and Manage Credentials (enrol/revoke biometric credentials) all available via per-row action menu. |
| Hardware Simulator | Complete | `/admin/simulator` allows an admin to register a mock device, enrol a user credential, and trigger a hardware authentication payload against the backend — all without exposing the `DEVICE_API_KEY`. |

## How the Hardware-Free Demo Works

1. An administrator creates a zone at `/admin/zones`.
2. An administrator opens `/admin/simulator`, registers a mock device in that zone, and enrols a credential hash against a user.
3. The administrator (or a tester using the live link) triggers a **Simulate Scan** on the simulator page — no `DEVICE_API_KEY` copy/paste required.
4. The backend validates device identity, credential, zone capacity, and exit rules; stores a hash-chained audit event; then broadcasts zone occupancy via WebSocket. The admin dashboard updates in real time.
5. Alternatively, a regular user can sign in at `/login`, open `/dashboard`, and use the **Simulate entry/exit** buttons for their authorized zones — using the simpler JWT-based path.

Both paths exercise the real access rules and audit trail. The hardware simulator is suitable for live demos to stakeholders without requiring any physical device.

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

- `src/app/admin/(dashboard)/layout.tsx` — Shared admin shell and navigation. Links to Dashboard, Zones, Users, and the new Simulator page. Visibility is separate from API authority.
- `src/app/admin/(dashboard)/page.tsx` — Occupancy dashboard. Fetches an initial snapshot from `/zones/occupancy`, then merges WebSocket occupancy updates for responsive live cards and capacity warnings.
- `src/app/admin/(dashboard)/zones/page.tsx` — Zone index. Shows zone cards, a meaningful no-zones state, and a create-zone form backed by `POST /admin/zones`.
- `src/app/admin/(dashboard)/zones/[id]/page.tsx` — Operational detail view for a single zone. Loads configuration and the active people list.
- `src/app/admin/(dashboard)/users/page.tsx` — Personnel list with **Create User** modal. Submits to `POST /admin/users` and instantly refreshes the table on success.
- `src/app/admin/(dashboard)/simulator/page.tsx` — **[NEW]** Three-step hardware simulator: (1) register a device in a zone, (2) enrol a user credential hash, (3) trigger a device authentication payload via a secure admin JWT proxy endpoint (`POST /admin/simulate-device`). Displays raw backend JSON output in a terminal-style console.

### Visual system

- `src/app/globals.css` — Defines the dark SOC design tokens and reusable layout, card, button, form, status, navigation, and table classes used across the portal.

## Known Limitations and Next Frontend Priorities

1. ~Replace WebSocket query-string JWTs with a safer production authentication mechanism.~ (Complete)
2. ~Add form-level validation, loading/error polish, and automated frontend tests.~ (Form polish complete; automated tests still pending)
3. Build analytic views, beginning with zone activity history and then heat maps once enough real/simulated event data exists.

## Feature Change Log

### 2026-08-18 — Default Route Redirect to Admin Login

- **Status:** Complete
- **Summary:** Updated the root page redirect (`/`) to point to `/admin/login` instead of `/login` as the primary features are currently focused on the administrative portal.
- **Files touched:** `src/app/page.tsx`, `FRONTEND_PROGRESS_REPORT.md`.
- **Validation:** Navigating to the base route successfully redirects to the admin login page.

### 2026-08-18 — Delete User and Analytics Timezone Fix

- **Status:** Complete
- **Summary:** (1) Implemented full-stack user deletion. Backend: added `DeleteUser` to `UserRepository` interface, `UserAuth` repo, `UserService` (with admin guard), `ErrCannotDeleteAdmin` error, `DeleteUserHandler`, and `DELETE /admin/users/{id}` route. Frontend: added **Delete User** button in the Users page action dropdown — hidden for admin accounts. (2) Fixed the Peak Entry Times bar chart showing UTC hours instead of local hours by applying the browser's timezone offset (`new Date().getTimezoneOffset()`) when mapping backend records to chart buckets.
- **Files touched:** `backend/internal/auth/interfaces.go`, `backend/internal/auth/user_auth.go`, `backend/internal/auth/user_service.go`, `backend/internal/auth/errors.go`, `backend/internal/httpdir/auth_handlers.go`, `backend/internal/httpdir/router.go`, `src/app/admin/(dashboard)/users/page.tsx`, `src/components/ZoneAnalyticsChart.tsx`.
- **Validation:** `go build ./...` clean, `npm run build` exit code 0, all 12 routes generated.

### 2026-08-18 — Simulator: Forward Biometric Mismatch to Backend and Refresh Events

- **Status:** Complete
- **Summary:** Previously, the Hardware Simulator silently blocked mismatch scan tests on the client side and only refreshed the Hash Chain Events log on successful responses. This meant denied events (e.g. `device_type_mismatch`) were invisible in the UI even though the backend was correctly logging them. Fixed by: (1) removing the client-side `activeDeviceType !== activeCredentialMethod` guard so every scan is sent to the server, and (2) triggering the hash chain events refresh after every simulation response, not just successful ones.
- **Files touched:** `src/app/admin/(dashboard)/simulator/page.tsx`.
- **Validation:** `npm run build` exit code 0, TypeScript clean, all 12 routes generated. Mismatched scans now appear correctly in the Hash Chain Events panel with reason `device_type_mismatch`.
- **Follow-up:** None.

### 2026-08-17 — Form Validation and Loading State Polish

- **Status:** Complete
- **Summary:** Applied consistent form validation, loading states, and error handling across all administration portal forms. (1) Added a `.spinner` CSS keyframe animation and `.btn:disabled` / `.input:disabled` styles to `globals.css`. (2) All form inputs and selects are now disabled while their request is in-flight, preventing double submissions. (3) All submit buttons swap their icon for a spinning indicator during loading, giving immediate visual feedback. (4) `pattern=".*\S+.*"` validation added to free-text name fields (zone name, user name, device name, credential hash) to block whitespace-only submissions before they reach the backend.
- **Files touched:** `src/app/globals.css`, `src/components/SignInForm.tsx`, `src/app/bootstrap/page.tsx`, `src/app/admin/(dashboard)/zones/page.tsx`, `src/app/admin/(dashboard)/zones/[id]/page.tsx`, `src/app/admin/(dashboard)/users/page.tsx`, `src/app/admin/(dashboard)/simulator/page.tsx`.
- **Validation:** `npm run build` exit code 0, TypeScript clean, all 12 routes generated.
- **Follow-up:** Add automated frontend tests and day-of-week heatmap analytics view.

### 2026-08-13 — Secure WebSocket Ticket Authentication

- **Status:** Complete
- **Summary:** Replaced the vulnerable query-string JWT WebSocket authentication with a short-lived ticket-based handshake. The frontend now makes an authenticated `POST /ws/ticket` request to retrieve a single-use ticket before opening the WebSocket connection. This prevents the long-lived JWT from being permanently recorded in server or proxy URL access logs.
- **Files touched:** `src/app/admin/(dashboard)/page.tsx`, `backend/internal/access/ws_ticket.go` (new), `backend/internal/access/hub.go`, `backend/internal/auth/middleware.go`.
- **Validation:** Tests pass. Admin dashboard correctly fetches a ticket and establishes a WebSocket connection.
- **Follow-up:** Add form-level validation and automated frontend tests.

### 2026-08-13 — Credential management from Users page

- **Status:** Complete
- **Summary:** Administrators can now manage biometric credentials directly from the Personnel page via a new **Manage Credentials** modal (accessed from the per-row `⋮` menu). The modal shows all enrolled credentials for a user (method, truncated hash) with per-credential **Revoke** buttons, plus an enrol form to add a new credential (method dropdown + raw hash input). Backed by `GET /admin/users/{id}/credentials`, `POST /admin/users/{id}/credentials`, and `DELETE /admin/users/{id}/credentials/{method}`.
- **Files touched:** `src/app/admin/(dashboard)/users/page.tsx`.
- **Validation:** Build clean. Enrol and revoke flows confirmed against backend endpoints.
- **Follow-up:** Consider inline credential status badge (e.g. `REVOKED`) for revoked but retained records.

### 2026-08-13 — Zone Analytics chart on admin dashboard

- **Status:** Complete
- **Summary:** Added a **Peak Entry Times** bar chart to the right column of the admin dashboard. The `ZoneAnalyticsChart` component fetches `GET /admin/zones/{id}/analytics`, groups `(day_of_week, hour, entry_count)` records into 24 hourly buckets, and renders via Recharts `BarChart`. A dropdown selector lets the admin switch between zones; the first available zone is auto-selected on load. Shows a friendly empty state when no entry events exist for a zone.
- **Files touched:** `src/components/ZoneAnalyticsChart.tsx` (new), `src/app/admin/(dashboard)/page.tsx`.
- **Validation:** Build clean. Chart renders with real event data.
- **Follow-up:** Day-of-week heatmap view is a natural next step once event volume grows.

### 2026-08-13 — Fix dashboard analytics layout (vanilla CSS vs Tailwind)

- **Status:** Complete
- **Summary:** The two-column dashboard layout was not rendering because the project uses plain vanilla CSS (no Tailwind installed) but the code contained Tailwind-only arbitrary-value classes (`lg:flex-row`, `min-h-[350px]`, `lg:w-[450px]`, `overflow-y-auto`, `pr-1`). Replaced these with a proper `.dashboard-layout` / `.dashboard-right` CSS rule pair in `globals.css` (with `@media (min-width: 1024px)` breakpoint for desktop side-by-side) and inline `style` props for one-off pixel values.
- **Files touched:** `src/app/admin/(dashboard)/page.tsx`, `src/components/ZoneAnalyticsChart.tsx`, `src/app/globals.css`.
- **Validation:** Build clean. Two-column layout confirmed on desktop; stacks on mobile.
- **Follow-up:** Audit remaining pages for other stray Tailwind-only utility classes.

### 2026-08-11 — Zone detail page: tabbed UI, event history, chain verify, edit and delete

- **Status:** Complete
- **Summary:** Completely overhauled the zone detail page (`/admin/zones/[id]`) with a three-tab layout. (1) **Overview** — active personnel table, now accurately showing users with real active sessions (fixed previous session). (2) **Event History** — full paginated access event log (15 events/page) with prev/next controls, plus a **Verify Chain** button that calls `GET /admin/zones/{id}/verify-chain` and displays a clear pass/fail result with event count. (3) **Settings** — inline edit form backed by `PUT /admin/zones/{id}` with success/error feedback; a Danger Zone section requires the admin to type the exact zone name before the permanent delete button (`DELETE /admin/zones/{id}`) becomes enabled, guarding against accidental removal. Status card updated to reflect OCCUPIED vs SECURE dynamically based on actual occupancy.
- **Files touched:** `src/app/admin/(dashboard)/zones/[id]/page.tsx`, `frontend/FRONTEND_PROGRESS_REPORT.md`.
- **Validation:** `npm run build` exit code 0, TypeScript clean, all 12 routes generated.
- **Follow-up:** Replace WebSocket query-string JWTs with a safer token mechanism before production deployment.

### 2026-08-10 — Admin UI for zone access grants and personnel list fix

- **Status:** Complete
- **Summary:** Added a "Manage Access" modal to the Users page, allowing an admin to toggle user access for individual zones. Also fixed a bug on the Zone details page where the "Active Personnel" list mistakenly displayed all users with access rather than users currently physically in the zone, by adding and switching to a new `/active-users` endpoint.
- **Files touched:** `src/app/admin/(dashboard)/users/page.tsx`, `src/app/admin/(dashboard)/zones/[id]/page.tsx`, `backend/internal/access/access_service.go`, `backend/internal/httpdir/access_handlers.go`, `backend/internal/httpdir/router.go`.
- **Validation:** Verified "Manage Access" modal correctly retrieves and modifies user permissions, reflecting changes instantly in the database via the REST API. Confirmed successful build.
- **Follow-up:** Add zone editing, deletion safeguards, event history, and hash-chain verification controls to the zone detail page.

### 2026-08-06 — Role-separated portals, browser simulator, and zone management

- **Status:** Complete
- **Summary:** Added separate personnel and admin login presentations; added authenticated browser-based entry/exit simulation for authorised users; added the missing zones index with creation and empty state; fixed the personal dashboard’s handling of paginated event responses.
- **Files touched:** `src/context/AuthContext.tsx`, `src/components/SignInForm.tsx`, `src/app/login/page.tsx`, `src/app/admin/login/page.tsx`, `src/app/dashboard/page.tsx`, `src/app/admin/zones/page.tsx`, `FRONTEND_PROGRESS_REPORT.md`, `README.md`.
- **Validation:** Focused ESLint passed for all changed frontend files. Full `npm run lint` is currently blocked by an unrelated `@ts-ignore` lint error in `next.config.ts`. `npm run build` reached the Next.js font stage but could not download Inter from Google Fonts in the restricted/offline environment.
- **Follow-up:** Build admin UI for zone access grants; further refine analytic views.

### 2026-08-07 — Admin user creation UI

- **Status:** Complete
- **Summary:** Added a **Add Personnel** modal to the Users page allowing an admin to create standard user accounts directly from the portal without any API calls. On success, the table refreshes in real time.
- **Files touched:** `src/app/admin/(dashboard)/users/page.tsx`.
- **Validation:** Verified the modal renders, submits to `POST /admin/users` with admin JWT, and newly created user appears in the table.
- **Follow-up:** Extend the row action menu with role changes, access grants, and account unlock.

### 2026-08-08 — Scalable layout: search, scrollable grids, simulator tabs

- **Status:** Complete
- **Summary:** Three pages updated to handle large numbers of zones/records without the page growing infinitely. (1) **Dashboard** — added real-time search filter, compacted occupancy cards, wrapped grid in `max-height` scrollable container, cards now show 4 per row on wide screens. (2) **Zones page** — added search/filter bar, compact zone cards with description clamp and inline lock/unlock icon, scrollable card grid, "New Zone" button toggles inline create form instead of having the form always visible at the bottom. (3) **Simulator** — split into **Setup** and **Records** tabs. Setup tab holds the 4 workflow steps; Records tab holds the Devices list, Credentials list, and Hash Chain Events log. The **Use** button in Records auto-populates Step 4 fields *and* switches back to the Setup tab, preserving the one-click workflow.
- **Files touched:** `src/app/admin/(dashboard)/page.tsx`, `src/app/admin/(dashboard)/zones/page.tsx`, `src/app/admin/(dashboard)/simulator/page.tsx`.
- **Validation:** `npm run build` passed, exit code 0, no TypeScript errors across all 12 routes.
- **Follow-up:** Consider server-side search + pagination at 500+ zones.

### 2026-08-08 — Simulator data panels: device list, credential list, hash chain log

- **Status:** Complete
- **Summary:** Overhauled the Hardware Simulator page with three new live data panels below the setup steps. (1) **Registered Devices** table shows every device across all zones with ID, name, type, zone, serial, and active status — a **Use** button auto-loads the device into Step 4, and a **Delete** button removes it inline (replacing the old standalone Step 5 delete form). (2) **Enrolled Credentials** table shows every user credential with user name, entry method, truncated stored hash, and revoked status — a **Use** button loads it into Step 4. (3) **Hash Chain Events** log shows the last 15 access events for a selected zone including timestamp, user, action, entry method, status, reason, and truncated hash — auto-refreshes after a successful simulation. All three panels have a manual refresh button. Steps 1 and 2 still auto-fill Step 4 on success.
- **Files touched:** `src/app/admin/(dashboard)/simulator/page.tsx`.
- **Validation:** `npm run build` passed with exit code 0, no TypeScript errors.
- **Follow-up:** Consider paginating the events log; add credential revoke action from the credentials table.

### 2026-08-07 — Hardware Device Simulator page

- **Status:** Complete
- **Summary:** Added a three-step Hardware Simulator at `/admin/simulator`. Step 1 registers a mock device in a zone (`POST /admin/zones/{id}/devices`). Step 2 enrols a raw credential hash against a selected user (`POST /admin/users/{id}/credentials`). Step 3 triggers a device authentication payload via a new secure proxy endpoint (`POST /admin/simulate-device`) which calls the same logic as the real hardware route but is protected by an Admin JWT instead of the secret API key — so the admin portal link can be shared without exposing server secrets. A terminal-style console shows the raw backend JSON response.
- **Files touched:** `src/app/admin/(dashboard)/simulator/page.tsx` (new), `src/app/admin/(dashboard)/layout.tsx` (Simulator nav link + Cpu icon), `backend/internal/httpdir/biometric_handlers.go` (AdminSimulateBiometricHandler), `backend/internal/httpdir/router.go` (POST /admin/simulate-device).
- **Validation:** Verified navigation link renders correctly. Device creation, credential enrolment, and simulation payload all route to correct backend endpoints with proper auth headers. Backend validates logic and WebSocket broadcast updates the admin dashboard.
- **Follow-up:** Add zone access grant UI so the full demo flow can be completed without any direct API calls.

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

### 2026-08-06 — Fix admin login layout using Next.js Route Groups

- **Status:** Complete
- **Summary:** Fixed an issue where the unauthenticated `/admin/login` route returned a blank screen due to inheriting the protected `app/admin/layout.tsx`. Moved the protected admin pages and layout into an `(dashboard)` Route Group so the login page renders independently of the dashboard shell.
- **Files touched:** Moved `src/app/admin/layout.tsx`, `page.tsx`, `users/`, and `zones/` into `src/app/admin/(dashboard)/`.
- **Validation:** Verified `/admin/login` correctly mounts without rendering a blank screen and `(dashboard)` routes still display the dashboard shell.

## Pre-Commit Checklist

- [ ] Update this report before each frontend-related commit.
- [ ] Add a feature-log entry with exact files and real behaviour.
- [ ] Record validation performed and any failures.
- [ ] Document new routes, permissions, API calls, and known limitations.
- [ ] Confirm the change does not expose admin-only functionality through the personnel portal.
