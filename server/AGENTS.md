# Styx app guide

This file covers `server/frontend` and `server/backend`. Keep it aligned with the code when routes, API contracts, build steps, or the visual system change. The app is a desktop-first React frontend served by a Go backend. Mobile and tablet layouts are not current targets.

## Structure

- `frontend/index.html` is the Vite entry document.
- `frontend/src/main.jsx` mounts React, renders the login at `/`, and renders the dashboard at `/dashboard`. There is no routing library yet.
- `frontend/src/Dashboard.jsx` contains the dashboard shell, top navigation, campaign selector, and Campaigns view.
- `frontend/src/FilesView.jsx` lists uploads for the selected campaign and implements the upload dialog.
- `frontend/src/authFetch.js` sends protected requests and refreshes expired access tokens once.
- `frontend/src/CampaignContext.jsx` owns the campaign response and current campaign ID for the dashboard. Future campaign-scoped views should consume this context instead of hardcoding a UUID.
- `frontend/src/styles.css` contains the current design tokens and page styles. `frontend/src/mark.svg` is the Styx symbol; preserve it.
- `frontend/api.md` documents backend endpoints and their response quirks. Read it before adding API calls.
- `frontend/vite.config.js` proxies `/rest` to the Go API during development. `API_BASE_URL` overrides the default `http://127.0.0.1:8080`.
- `frontend/scripts/stage-frontend.mjs` copies Vite output to `backend/internal/web/dist` for Go embedding.
- `backend/internal/web/web.go` serves embedded assets and the React entry point for `/dashboard` refreshes. `/rest` remains the API namespace. Update this handler when adding client routes, and extend `web_test.go` accordingly.
- `server/MakeFile` defines the build; `server/Makefile` includes it so plain `make` works.

## Current behavior

- Login posts username and password to `/rest/auth/login`. The returned `access` and `refresh` tokens are saved in local storage, then the browser navigates to `/dashboard`.
- Dashboard loading requests `/rest/campaigns/list` and saves the response in local storage as `campaignsResponse`. Each campaign includes `total_clients` and `total_files`, counted from matching rows in `clients` and `user_uploads`. The API can return `campaigns: null`; treat it as an empty list.
- The top-right campaign selector displays campaign names and stores the chosen UUID as `currentCampaignId`. It remains clickable when the list is empty, and a nearby New button can create a campaign via `POST /rest/campaigns/new` before reloading the list. On load, retain the saved selection only if it still exists in the latest response. A login clears the previous campaign cache and selection.
- `CampaignProvider` exposes `campaignsResponse`, `campaigns`, `currentCampaignId`, `selectCampaign`, loading status, and retry. Use its ID for future campaign-scoped requests. Do not embed the example UUID from `api.md` in code.
- The dashboard initially shows only the client list. Double-click a client row (or focus it and press Enter) to open its shell below. The divider can be dragged or moved with arrow keys to resize the two areas; each area scrolls independently and the shell input stays fixed at the bottom. The client request is `GET /rest/clients/list/{campaignID}` with the access token; `clients: null` means an empty list. With no campaign selected, prompt the user to choose one. The shell currently has no command execution endpoint.
- The Campaigns tab lists campaign name, client and file totals, creation time, and a Delete action. Deletion requires a confirmation dialog and posts to `/rest/campaigns/delete/{id}`. The backend removes related client and upload records and stored upload directories. Reload the campaign list after deletion, which also clears a deleted current selection.
- The Files tab lists uploads for the selected campaign. Its Upload dialog accepts a local file, an editable file name, and an `exe`, `dll`, or `shellcode` type. Uploads call `POST /rest/uploads/start` with campaign ID, file name, byte size, and type, then `PUT /rest/uploads/begin/{id}` with the raw file bytes. Each row has a confirmed Delete action that calls `DELETE /rest/uploads/delete/{id}`. Refresh files and campaign totals after upload or deletion. Builder and Settings remain disabled.
- Protected frontend requests send the stored bearer access token. On a 401, they POST the stored refresh token to `/rest/auth/refresh`, save the new access and refresh pair, and retry once. An invalid refresh clears stored auth and returns to login. Campaign, upload, and client routes use bearer-token auth middleware.

## Visual direction

Use a quiet, clean SaaS interface related to Styx: deep charcoal ground, slightly lighter flat surfaces, pale mist text, and restrained river-teal accents. Current tokens in `src/styles.css`: ground `#101112`, surface `#191b1c`, raised `#232627`, text `#eff2ef`, muted `#a5adaa`, accent `#b9dad2`. Keep data and controls readable against those surfaces. Avoid gradients, glass, glows, large shadows, tinted icon tiles, and busy dashboard cards.

The login is a narrow form on the dark ground, with clear labels, a pale teal primary button, and visible focus. The dashboard has a compact horizontal top bar: preserved Styx SVG symbol, italic serif `styx` wordmark, plain navigation buttons, and the campaign dropdown on the right. There is no sidebar. Use the existing spacing, color tokens, and control geometry when adding views. Desktop is the supported design target. The user specifically removed extra promotional/footer copy from login; do not reintroduce it.

The project includes `frontend/.agents/skills/design-taste-frontend/SKILL.md`. Apply its app-screen guidance to new app views and inspect screenshots when making material visual changes, while honoring the decisions above.

## Build and verification

From `server/`, `make all` (or plain `make`) runs `npm ci`, builds the React frontend, stages the assets, and compiles `backend/server` with them embedded. `make test` stages the frontend and runs Go tests. `make run` builds and starts the server. From `server/frontend/`, use `npm ci`, `npm run dev`, and `npm run build` for local work. A standalone frontend build does not update the embedded binary; run `make all` for that.

Generated `node_modules`, `frontend/dist`, `backend/internal/web/dist`, and binaries are ignored at the repository root. The Go embed package needs staged files before direct `go build` or `go test` on a fresh checkout. Preserve unrelated backend changes; the API and database areas may be under active development.
