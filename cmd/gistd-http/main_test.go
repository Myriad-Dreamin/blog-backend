package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestShouldServeIndex(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/", want: true},
		{path: "/h", want: true},
		{path: "/h/", want: true},
		{path: "/h/room-id", want: true},
		{path: "/assets/index.js", want: false},
		{path: "/favicon.ico", want: false},
		{path: "/health", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := shouldServeIndex(tt.path); got != tt.want {
				t.Fatalf("shouldServeIndex(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestStaticHandlerServesIndexForAppRoutes(t *testing.T) {
	root := makeStaticRoot(t)
	handler := newStaticHandler(root)

	for _, path := range []string{"/", "/h/example"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if rec.Body.String() != "<!doctype html><title>Paseo</title>" {
				t.Fatalf("body = %q, want index.html", rec.Body.String())
			}
		})
	}
}

func TestStaticHandlerServesResources(t *testing.T) {
	root := makeStaticRoot(t)
	handler := newStaticHandler(root)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/assets/index.js", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "console.log('paseo');" {
		t.Fatalf("body = %q, want asset contents", rec.Body.String())
	}
}

func TestStaticHandlerReturnsNotFoundForMissingResources(t *testing.T) {
	root := makeStaticRoot(t)
	handler := newStaticHandler(root)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/missing.js", nil)

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func makeStaticRoot(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, indexFile), []byte("<!doctype html><title>Paseo</title>"), 0644); err != nil {
		t.Fatal(err)
	}

	assetsDir := filepath.Join(root, "assets")
	if err := os.Mkdir(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "index.js"), []byte("console.log('paseo');"), 0644); err != nil {
		t.Fatal(err)
	}

	return root
}
