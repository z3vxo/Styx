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

The backend currently issues tokens but does not attach authentication middleware to the routes. An unknown user returns 401; a wrong password currently returns 500. `POST /rest/auth/refresh` returns 501 and is not implemented.

## Campaigns

### `GET /rest/campaigns/list`

Returns HTTP 200:

```json
{"total":1,"campaigns":[{"name":"test","id":"01a1a7cc-eaac-7321-9e6a-fb16e611d085","created_at":"2026-11-04T18:23:06Z"}]}
```

When empty, `campaigns` may be `null`, so treat it as an empty list.

### `POST /rest/campaigns/new`

Send `{"name":"Campaign name"}` as JSON. Success returns HTTP 200 and `{"status":"OK"}`; it does not return the new ID. Refresh the campaign list to obtain it.

### `DELETE /rest/campaigns/delete/{id}`

Deletes the campaign identified by `id`. Success returns HTTP 200 and `{"status":"OK"}`. If deleting the selected campaign, update the shared selection and campaign-scoped views.

## Files

Files belong to the selected campaign. Uploads use two requests:

1. `POST /rest/uploads/start` with JSON metadata:

   ```json
   {"file_name":"example.txt","campaign_uuid":"01a1a7cc-eaac-7321-9e6a-fb16e611d085","file_size":12,"file_type":"text/plain"}
   ```

   Substitute the selected campaign ID. `file_size` is the file size in bytes and `file_type` is a client-provided string. The filename must be a single filename, without a path. Success returns HTTP 200 and `{"id":"<file-uuid>"}`.

2. `PUT /rest/uploads/begin/{id}` using the returned file UUID. Send the raw file bytes as the request body, not JSON or multipart form data. Success returns HTTP 200 with an empty body. A missing upload ID in the in-memory registry returns 400; a write or database error returns 500. The registry is in memory, so a server restart loses pending upload IDs. A failed request may already have written bytes to disk.

### `GET /rest/uploads/list/{campaignID}`

Use the currently selected campaign ID. Returns HTTP 200:

```json
{"total":1,"files":[{"file_name":"example.txt","file_size":12,"file_type":"text/plain","file_uuid":"<file-uuid>","uploaded_at":"2026-11-04T18:23:06Z"}]}
```

When empty, `files` may be `null`. Refresh this list after a successful upload or deletion.

### `DELETE /rest/uploads/delete/{id}`

Use the file UUID, not the campaign ID. Success returns HTTP 200 and `{"status":"OK"}`; deletion errors return 400. The current API does not expose a file-download endpoint.
