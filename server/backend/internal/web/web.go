// Package web serves the frontend compiled by the server MakeFile.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// Build the frontend before compiling Go: make -f MakeFile build (from server/).
//
//go:embed dist
var content embed.FS

func Handler() http.Handler {
	root, err := fs.Sub(content, "dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Unknown API endpoints must never receive the frontend.
		if r.URL.Path == "/rest" || strings.HasPrefix(r.URL.Path, "/rest/") {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Serve the React entry point for the dashboard, including browser refreshes.
		if r.URL.Path == "/dashboard" || r.URL.Path == "/dashboard/" {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			name = "index.html"
		}
		info, err := fs.Stat(root, name)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}
