# Styx frontend

React and Vite. Desktop is the supported layout target. Login saves access and
refresh tokens in local storage, then opens `/dashboard`. The dashboard lists
clients for the selected campaign. Double-click a client row to open a resizable
shell below the list. The shell opens a WebSocket for that client and submits
commands through `POST /rest/tasks/new`; client output is not connected yet.
Campaigns can be created from the New button beside the dashboard selector.
The Campaigns tab lists each campaign's client and file totals and provides a confirmed delete action.
The Files tab lists uploads for the selected campaign, uploads local files through the two-step API, and confirms deletions. Protected requests refresh an expired access token and retry once.
The backend accepts connections at `/rest/ws/connect?client_id=...&tab_id=...`
for future live output. See `api.md` for the current limits.

## Development

```sh
npm ci
npm run dev
```

Vite proxies `/rest` to `http://127.0.0.1:8080`. Set `API_BASE_URL` to use another
backend address. The production page uses the same origin as the Go API.

## Single-binary build

From `server/`:

```sh
make all
make test
./backend/server run
```

The build installs locked frontend dependencies, produces Vite's `frontend/dist`,
copies it into `backend/internal/web/dist`, then compiles `backend/server` with
those files embedded. No Node process or external frontend files are needed at
runtime. Existing Styx setup/configuration is still required to start the server.

The generated embed directory is ignored by Git. On a fresh checkout, use the
MakeFile before running Go commands; Go needs these files at compile time.
`make run` builds and starts the server. The standard `Makefile` includes `MakeFile`, so plain `make` also builds everything.
