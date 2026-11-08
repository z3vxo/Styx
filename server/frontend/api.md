# Styx frontend API guide

## Application context

The app is built around the **currently selected campaign ID**. Keep the selected campaign in shared frontend state, use it for campaign-scoped views and requests, and update those views when the selection changes. Do not hardcode a campaign ID. The current development campaign has ID `01a1a7cc-eaac-7321-9e6a-fb16e611d085`; it is an example, not a permanent default. Retrieve available campaigns from the API.

The backend mounts its routes under `/rest`. The default server address is `http://127.0.0.1:8080`, configurable in the backend. Use a configurable API base URL or a development proxy; the server does not currently configure CORS.

Most endpoints encode JSON, but do not consistently set `Content-Type: application/json`. Check the HTTP status and parse the response according to the endpoint rather than relying on that header. Errors from campaign and file endpoints usually have the shape `{"error":"..."}`; authentication errors are plain text. Some malformed requests currently return HTTP 500.

## Authentication

### `POST /rest/auth/login`

Send JSON:

```json
{"username":"alice","password":"password"}
```

On success, returns HTTP 200 and JSON:

```json
{"access":"<jwt>","refresh":"<jwt>"}
```

The backend issues tokens and requires a bearer access token on campaign, upload, and client routes. An unknown user returns 401; a wrong password currently returns 500.

### `POST /rest/auth/refresh`

Send the stored refresh token as JSON `{"token":"<refresh>"}`. The frontend also sends the current access token in the `Authorization: Bearer <access>` header. The refresh route validates the token in the JSON body and returns HTTP 200 with a new `{"access":"<jwt>","refresh":"<jwt>"}` pair. The frontend saves both and retries the request that received 401 once. Missing or malformed input returns 400; an invalid or expired refresh token returns 401.

## Campaigns

### `GET /rest/campaigns/list`

Returns HTTP 200:

```json
{"total":1,"campaigns":[{"name":"test","id":"01a1a7cc-eaac-7321-9e6a-fb16e611d085","total_clients":2,"total_files":3,"paths":["/whatever"],"created_at":"2026-11-04T18:23:06Z"}]}
```

`total_clients` and `total_files` count matching rows in `clients` and `user_uploads`, respectively, by `CampaignUUID`. `paths` lists the campaign's stored paths and is an empty array for a campaign without paths. When empty, `campaigns` may be `null`, so treat it as an empty list.

### `POST /rest/campaigns/new`

Send `{"name":"Campaign name","paths":["/whatever"]}` as JSON. At least one path is required. The New campaign form can add more path fields with its + button. The backend stores the paths with the campaign. Success returns HTTP 200 and `{"status":"OK"}`; it does not return the new ID. Refresh the campaign list to obtain it.

### `POST /rest/campaigns/delete/{id}`

Deletes the campaign identified by `id`, its client and upload records, and stored upload directories. The endpoint also accepts `DELETE` for existing callers. Success returns HTTP 200 and `{"status":"OK"}`; an unknown ID returns 404. If deleting the selected campaign, reload the campaign list so the shared selection and campaign-scoped views reset.

## Clients

### `GET /rest/clients/list/{campaignID}`

Use the selected campaign ID and send `Authorization: Bearer <access>`. Returns HTTP 200:

```json
{"total":1,"clients":[{"name":"client-a","client_id":42,"hostname":"host-a","domain":"example.com","created_at":"2026-11-04T18:23:06Z"}]}
```

When empty, `clients` may be `null`. The dashboard treats it as an empty list and reloads whenever the selected campaign changes.

## Files

Files belong to the selected campaign. Uploads use two requests:

1. `POST /rest/uploads/start` with JSON metadata:

   ```json
   {"file_name":"example.exe","campaign_uuid":"01a1a7cc-eaac-7321-9e6a-fb16e611d085","file_size":12,"file_type":"exe"}
   ```

   Substitute the selected campaign ID. `file_size` is the file size in bytes and `file_type` is one of `exe`, `dll`, or `shellcode` in the Files tab. The filename must be a single filename, without a path. Success returns HTTP 200 and `{"id":"<file-uuid>"}`.

2. `PUT /rest/uploads/begin/{id}` using the returned file UUID. Send the raw file bytes as the request body, not JSON or multipart form data. Success returns HTTP 200 with an empty body. A missing upload ID in the in-memory registry returns 400; a write or database error returns 500. The registry is in memory, so a server restart loses pending upload IDs. A failed request may already have written bytes to disk.

### `GET /rest/uploads/list/{campaignID}`

Use the currently selected campaign ID. Returns HTTP 200:

```json
{"total":1,"files":[{"file_name":"example.exe","file_size":12,"file_type":"exe","file_uuid":"<file-uuid>","uploaded_at":"2026-11-04T18:23:06Z"}]}
```

When empty, `files` may be `null`. Refresh this list after a successful upload or deletion.

### `DELETE /rest/uploads/delete/{id}`

Use the file UUID, not the campaign ID. Success removes the stored file directory and database row, then returns HTTP 200 and `{"status":"OK"}`; deletion errors return 400. The current API does not expose a file-download endpoint.

## WebSocket

### `GET /rest/ws/connect?client_id={id}&tab_id={tabID}`

The server upgrades this request after validating the numeric client ID, tab ID, and an access token sent as the `styx.jwt.<access>` WebSocket subprotocol. The browser also offers the `styx` subprotocol. The frontend checks its access token with a protected request before connecting, allowing its normal refresh flow to run.

The shell opens a connection for the selected client's `client_id` when its panel
appears and closes it when the panel closes or the selection changes. It creates
a new tab ID for each panel instance. The backend stores connections by client
and tab ID. `Send(clientID, data)` sends JSON to every open tab for that client.
Incoming WebSocket messages are currently read and discarded. Task submission
uses HTTP; live output delivery is not connected yet.

## Tasks

### `POST /rest/tasks/new`

Send a bearer-authenticated JSON request using the selected client and campaign:

```json
{"cmd":"load-module","args":[{"type":"string","value":"example"}],"client_id":42,"campaign_id":"<campaign-id>"}
```

The shell splits input on whitespace: the first word is `cmd`, and the remaining
words become string arguments. The backend currently maps `load-module` and
`unload-module`. Success returns HTTP 200 and `{"task_id":123}`. Errors return
`{"error":"..."}`. A task ID confirms insertion; client execution and output
are not implemented yet.
