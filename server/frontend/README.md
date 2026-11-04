# Styx frontend

React and Vite. Desktop is the supported layout target. The login page currently
shows a confirmation after authentication; session persistence and dashboard
navigation will be added with the dashboard.

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
