# Styx: modules, command/task model, output parsing — implementation plan

Goal: new modules can be dropped in without recompiling the server, commands are
validated reliably, tasks are unambiguous (built-in vs module command), and raw
agent output is turned into final, properly formatted text the server just sends.

## Decisions (locked)

- **Output formatting uses an embedded Lua interpreter** (`gopher-lua`, pure Go,
  `CGO_ENABLED=0`, keeps the static single binary). This is the only option that
  gives arbitrary per-command formatting without a server rebuild, without native
  plugins, and without a binary-format JSON spec.
- The server's only involvement in formatting is:
  `output, err := parsers.Parse(module, command, data)` → `ws.SendOutput(...)`.
- **Parsers are discovered from the filesystem by convention**, never compiled in:
  `~/.local/share/styx/modules/<module>/<command>.lua`.
- **Tasks are modeled as `kind + module + command + args`**, never a single
  ambiguous `CmdCode`. Built-ins and module commands live in separate namespaces.
- **Input is tokenized with a quote-aware lexer**, not `strings.Split`. Positional
  args are primary; optional named flags (`--flag value`) are allowed when clarity
  helps. No JSON schema of the binary format.
- Old `endpoint/modules/*.c` modules are legacy and will be replaced; do not design
  the new output contract around their current binary layout.

---

## 1. Task DB schema and lifecycle

Files: `server/backend/internal/db/db.go` (migrations),
`server/backend/internal/db/db_tasks.go`.

- [ ] Replace the `tasks` table columns. Proposed:
  ```sql
  CREATE TABLE IF NOT EXISTS tasks(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    TaskID      INTEGER NOT NULL UNIQUE,
    ClientID    INTEGER NOT NULL,
    CampaignID  TEXT NOT NULL,
    Kind        TEXT NOT NULL CHECK (Kind IN
                 ('load-module','unload-module','list-modules','run-command')),
    Module      TEXT,      -- module name, NULL for non-module builtins
    ModuleID    INTEGER,   -- module id/bit, NULL for non-module builtins
    Command     TEXT,      -- module command name ('ls','mv',...), NULL for builtins
    Code        INTEGER,   -- module command code, NULL for builtins
    Args        TEXT NOT NULL CHECK (json_valid(Args)),
    Status      TEXT NOT NULL DEFAULT 'pending'
                 CHECK (Status IN ('pending','sent','done','failed')),
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    sent_at     DATETIME,
    completed_at DATETIME
  );
  ```
- [ ] Add a migration path for existing DBs (old table has only `CmdCode`); either
  drop/recreate `tasks` (acceptable, they are transient) or migrate columns.
- [ ] Fix the FK: `ClientID` currently references `clients(ClientID)`, which is not
  unique. Reference `clients.id`, or add `UNIQUE(ClientID)` to `clients`.
- [ ] Add `task_output` table for `MSG_OUTPUT` (chunked/streamed):
  ```sql
  CREATE TABLE task_output(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    TaskID INTEGER NOT NULL,
    seq INTEGER NOT NULL,
    stream TEXT NOT NULL DEFAULT 'stdout',
    data BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
  ```
- [ ] Implement `GetTasks(agentID)` properly: select `Status='pending'` for the
  agent, return them, and flip to `sent` in the same transaction (prevents replay
  on every poll).
- [ ] Implement task completion/failure transitions on `MSG_OUTPUT`.
- [ ] `InsertTask`: resolve + validate the command through the command catalog
  (section 2), store canonical typed args JSON, reject unknown/malformed commands
  before insert.
- [ ] Remove the duplicate/ambiguous code map. Wire codes belong to one catalog;
  the DB stores semantic fields (`Kind`, `Module`, `Command`, `Code`), and dispatch
  translates to wire values.

## 2. Command catalog and module registry

Files: new `server/backend/internal/commands/` (or extend
`internal/modules/modules.go`), `~/.local/share/styx/modules/modules.json`.

- [ ] One source of truth for every command: built-ins and module commands.
  Each entry: name, kind, module, wire code, arg schema, usage/help text.
- [ ] Built-ins:
  - `help` / `help <command>` — handled **entirely server-side, no task row**.
  - `load-module <name>` — `Kind=load-module`, resolves id/file from registry.
  - `unload-module <name>` — `Kind=unload-module` (needs new wire op, see §5).
  - `list-modules` — `Kind=list-modules` (`0xb2`).
- [ ] Module commands come from the registry (`commands` map), e.g.
  `ls, cp, pwd, cat, mv`. Resolve `(module, command) -> code`, validate that the
  module exists and (later) is loaded.
- [ ] Align wire codes with the agent (`core.h`):
  `TASK_LOAD_MODULE=0xb1`, `TASK_LIST_MODULES=0xb2`, `TASK_RUN_CMD=0xb3`.
  The current `0x1/0x2` mapping in `db_tasks.go` / `ws/types.go` is wrong.
- [ ] Decide `unload-module` wire op (e.g. `0xb4`) and implement it in the agent
  (currently `moduleUnload()` is a stub).
- [ ] Fix registry location: `modules.New()` already reads
  `~/.local/share/styx/modules/modules.json`; move the repo-root `modules.json`
  there (via `setup`) and keep it the minimal `id/file/commands` map. No parser
  entries required — parsers are found by path convention.
- [ ] Decide auto-load behavior: if a module command is issued for a module that is
  not loaded, either reject with a clear error or emit `load-module` first.
  Document the choice.
- [ ] Sanitize module/command names before using them in filesystem paths.

## 3. Command input parsing (frontend + backend)

Files: `server/frontend/src/Dashboard.jsx` (ShellPanel submit),
backend task API + catalog validation.

- [ ] Replace `value.split(/\s+/)` with a quote-aware tokenizer.
- [ ] Lexer rules (document them):
  - whitespace separates tokens outside quotes;
  - double quotes group, including spaces and empty string `""`;
  - backslash is **not** a general escape (Windows paths must survive:
    `C:\Users\foo`); only `\"` (or `""`) yields a literal quote;
  - optional single quotes may group as-is.
- [ ] Frontend builds typed args: `[{type:"string", value:"..."}]`; supports the
  optional `--flag value` / `--flag=value` forms.
- [ ] POST `/rest/tasks/new` body stays:
  `{cmd, args:[{type,value}], client_id, campaign_id}`.
- [ ] Backend validates `cmd` + arity/types against the catalog, returns structured
  errors (`unknown command`, `missing argument: path`, `module not loaded`), not a
  generic 500.
- [ ] `help` returns usage text directly (HTTP response / WS), never a task row.
- [ ] After `load-module`, the client should learn the module is available (see
  next point) so module commands can validate.

## 4. Output parsing (embedded Lua)

Files: new `server/backend/internal/parsers/` (host + reader + render helpers),
parser scripts under each module directory.

- [ ] Add `github.com/yuin/gopher-lua` (pure Go; no cgo; static build preserved).
- [ ] `Parse(module, command string, data []byte) (string, error)`:
  - path = `filepath.Join(modulesDir, module, command+".lua")`;
  - deterministic lookup, derives from the task row — no compile-time knowledge;
  - missing file → `ErrNoParser` (caller falls back to hex/text);
  - sanitize `module`/`command` (reject `/`, `..`) before joining.
- [ ] Host reader API over the bounded output buffer:
  `read_u32le`, `read_u64le`, `read_cstr`, `read_bytes(n)`, `remaining`, `skip`.
  Every read is bounds-checked; overrun returns an error, never panics.
- [ ] Host render helpers so scripts do **not** hand-pad:
  `render.table(rows, columns, align)`, `render.columns(rows)`, `render.text(s)`.
  These implement padding/newlines/column widths once, shared by all scripts.
- [ ] Sandbox: `SkipOpenLibs: true` (no `io`/`os`), instruction/time budget,
  max output size, max input size. Run with `context` for timeout/cancel.
- [ ] Script contract: `function Parse(r) return <final string> end` (may use
  `render.*`). Example:
  ```lua
  function Parse(r)
    local rows = {}
    while r.remaining() > 0 do
      rows[#rows+1] = { name = r.str(), size = r.u64() }
    end
    return render.columns(rows, {"name", "size"})
  end
  ```
- [ ] Server call site is exactly:
  ```go
  output, err := parsers.Parse(task.Module, task.Command, data)
  if err != nil { output = fallback(data) } // utf-8 text or hex+ascii
  ws.SendOutput(task.ClientID, output)
  ```
- [ ] Optional: at startup/setup, scan `modules/*/*.lua` to populate `help`
  (not required for execution).
- [ ] No parser → generic fallback (UTF-8 text, else hex+ascii dump). A module
  without a parser must still work.

## 5. Agent / wire protocol

Files: `endpoint/ldr/includes/agent/core.h`,
`endpoint/ldr/src/agent/dispatch.cc`,
`endpoint/ldr/src/agent/tasks/mods/mods.cc`,
`endpoint/ldr/src/agent/core.cc`,
`server/backend/internal/agentsrv/agent.go`,
`server/backend/internal/packager/unpacker.go`.

- [ ] Fix the agent `Run()` loop: the packager is created once and appended to each
  iteration without resetting (`NewPacker`/`PrepareMetaData` must run per request).
  Currently responses accumulate into one buffer.
- [ ] Implement server-side `MSG_GET_TASK` response serialization: send `RESP_NO_TASKS`
  or the task count + each task (`TaskID`, kind code, module id, command code, args).
- [ ] Implement server-side `MSG_OUTPUT` handling: parse output type + payload,
  write `task_output`, update `Status` (`done`/`failed`), then push to WS.
- [ ] Agent dispatch: handle `TASK_RUN_CMD` (0xb3) — read TaskID, module id,
  command code, args; call the module entry; emit `MSG_OUTPUT`.
- [ ] Agent `unload-module` (new op) — implement `moduleUnload()` (currently stub).
- [ ] `moduleLoad`/`moduleList` arg layout must match the server serializer.
- [ ] Note: `InsertAgent` currently hardcodes campaign `"11"` and name `"tes-test"`;
  the campaign ID comes from the URL path — thread it through to the insert.

## 6. WebSocket output delivery

Files: `server/backend/internal/ws/ws.go`, `server/backend/internal/ws/types.go`,
`server/backend/internal/agentsrv/agent.go`.

- [ ] On `MSG_OUTPUT`, after formatting via Lua, call
  `ws.Send(clientID, WSEvent{Event: "task.Output", Meta: {TaskID, TabID}, Body: {output}})`
  so all open tabs for that client receive it.
- [ ] Ensure `tab_id` metadata matches the frontend subscription (the frontend
  already filters `message.metadata.tab_id`).
- [ ] Frontend already listens for `event === 'task.Output'` and reads
  `body.output` / `body.message` — keep that contract.
- [ ] Decide: broadcast to all tabs of a client, or only the tab that created the
  task. The current `Send(clientID, data)` broadcasts; document behavior.

## 7. Frontend

Files: `server/frontend/src/Dashboard.jsx`, `server/frontend/src/authFetch.js`.

- [ ] Quote-aware command tokenizer; show usage/validation errors from the backend.
- [ ] Optional local `help` rendering (or rely on server `help`).
- [ ] Keep posting `client_id` + `campaign_id` from the selected context.
- [ ] Render `task.Output` text as-is (possibly `white-space: pre`).

## 8. Setup / packaging / docs

Files: `server/backend/internal/setup/setup.go`, docs.

- [ ] Ensure `setup` creates `~/.local/share/styx/modules/<module>/` and copies/seeds
  `modules.json` there.
- [ ] Module layout convention:
  ```
  ~/.local/share/styx/modules/<module>/
      <module>.mod      # module binary container
      <command>.lua     # optional parser, e.g. ls.lua
  ```
- [ ] Ship parsers alongside modules; parser is optional (fallback if absent).
- [ ] Update `server/AGENTS.md` / frontend docs when the task contract and output
  delivery are implemented.

## 9. Open questions

- `unload-module` wire op number and semantics (unload arena? mark bitmap only?).
- Should module commands require the module to be loaded, or auto-load on first use?
- Single final output per task vs streamed chunks (the `task_output` table supports
  chunks; decide UI behavior).
- Do we need `list-modules` results routed back to the client (loaded bitmap)?
- Whether `load-module <id> <name>` or `load-module <name>` is the operator syntax
  (recommended: `load-module <name>`; id resolved from registry).

## 10. Verification

- [ ] Unit tests: lexer (quotes, empty args, Windows paths), arg validation,
      task insert/resolve, parser path lookup + path-traversal rejection.
- [ ] Unit tests: Lua reader bounds (overrun → error), render helpers (column
      alignment/newlines), missing-parser fallback.
- [ ] DB tests: `pending -> sent -> done/failed` transitions, no task replay.
- [ ] Integration: agent register → `load-module` → module command → raw output →
      Lua format → WS `task.Output` delivered to frontend.
- [ ] `make test` from `server/` passes; frontend build stages and embeds.

---

### Explicitly rejected

- Native Go `plugin` `.so` / cgo parsers (Go-version coupling, no isolation, no
  static binary).
- Fork/exec parser subprocess per output (fragile: perms, platform, timeout plumbing).
- A big JSON/binary-format spec per module.
- Hardcoded `parseXOutput` functions in the server (recompile per module).
- Making every module hand-format text with padding/newlines (duplication).
