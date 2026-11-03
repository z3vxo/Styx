#!/usr/bin/env python3
import json
from http.server import BaseHTTPRequestHandler, HTTPServer
from uuid import uuid4


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/rest/user/register":
            self.send_error(404)
            return

        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)
        print(f"POST {self.path}: {body.decode('utf-8', errors='replace')}", flush=True)

        response = json.dumps({"user_id": str(uuid4())}).encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(response)))
        self.end_headers()
        self.wfile.write(response)


if __name__ == "__main__":
    print("Listening on http://127.0.0.1:8080", flush=True)
    HTTPServer(("127.0.0.1", 8080), Handler).serve_forever()
