package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedFrontend(t *testing.T) {
	h := Handler()
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/login", 200}, {"HEAD", "/login", 200}, {"GET", "/login/", 200},
		{"GET", "/dashboard", 200}, {"HEAD", "/dashboard", 200}, {"GET", "/dashboard/", 200},
		{"GET", "/", 200}, {"HEAD", "/", 200}, {"GET", "/missing.js", 404},
		{"GET", "/rest", 404}, {"GET", "/rest/unknown", 404},
		{"GET", "/assets/", 404}, {"POST", "/", 405}, {"GET", "/../go.mod", 404},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("got %d, want %d", w.Code, tc.status)
			}
			if tc.method == "HEAD" && w.Body.Len() != 0 {
				t.Fatal("HEAD returned a body")
			}
			if (tc.path == "/" || tc.path == "/login" || tc.path == "/login/" || tc.path == "/dashboard" || tc.path == "/dashboard/") && tc.method == "GET" {
				if !strings.Contains(w.Body.String(), `id="root"`) {
					t.Fatal("missing React root")
				}
				if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
					t.Fatal("wrong HTML MIME type")
				}
			}
		})
	}
	assets, err := fs.Glob(content, "dist/assets/*.js")
	if err != nil || len(assets) == 0 {
		t.Fatal("missing compiled JavaScript")
	}
	for _, asset := range assets {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/"+strings.TrimPrefix(asset, "dist/"), nil))
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatal("embedded asset not served")
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "javascript") {
			t.Fatal("wrong JS MIME type")
		}
		if !strings.Contains(w.Header().Get("Cache-Control"), "immutable") {
			t.Fatal("asset caching missing")
		}
	}
}
