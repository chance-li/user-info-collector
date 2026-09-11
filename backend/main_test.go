package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestListenAddr(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ADDR", "")
	os.Unsetenv("PORT")
	os.Unsetenv("ADDR")
	if got := listenAddr(); got != ":8080" {
		t.Fatalf("default = %q, want :8080", got)
	}

	t.Setenv("ADDR", ":9090")
	if got := listenAddr(); got != ":9090" {
		t.Fatalf("ADDR = %q, want :9090", got)
	}

	t.Setenv("PORT", "10000")
	if got := listenAddr(); got != ":10000" {
		t.Fatalf("PORT should win, got %q", got)
	}

	t.Setenv("PORT", ":10001")
	if got := listenAddr(); got != ":10001" {
		t.Fatalf("PORT with colon = %q, want :10001", got)
	}
}

func TestParseCORSOrigins(t *testing.T) {
	if parseCORSOrigins("") != nil {
		t.Fatal("empty should mean allow all")
	}
	if parseCORSOrigins("*") != nil {
		t.Fatal("* should mean allow all")
	}
	got := parseCORSOrigins(" https://a.vercel.app/ ,https://b.vercel.app ")
	if len(got) != 2 || got[0] != "https://a.vercel.app" || got[1] != "https://b.vercel.app" {
		t.Fatalf("got %#v", got)
	}
}

func TestCORSAllowlist(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := corsMiddleware([]string{"https://app.vercel.app"}, inner)

	req := httptest.NewRequest(http.MethodOptions, "/api/submissions", nil)
	req.Header.Set("Origin", "https://app.vercel.app")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://app.vercel.app" {
		t.Fatalf("allowed origin missing: %v", rec.Header())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/submissions", nil)
	req.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unexpected allow origin: %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
