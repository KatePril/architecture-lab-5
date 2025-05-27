package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBalancerSuccessful(t *testing.T) {
	dstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock response"))
	}))
	defer dstServer.Close()

	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)

	rec := httptest.NewRecorder()
	dstHost := strings.TrimPrefix(dstServer.URL, "http://")

	written, err := forward(dstHost, rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("got %v, want %v", err, nil)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if string(body) != "mock response" {
		t.Errorf("got %v, want %v", body, "mock response")
	}
	if written != int64(len(body)) {
		t.Errorf("got %d, want %d", len(body), written)
	}
	if resp.Header.Get("X-Mock") != "true" {
		t.Errorf("got %v, want %v", resp.Header.Get("X-Mock"), "true")
	}
}

func TestBalancerError(t *testing.T) {
	dstHost := "invalid.host.local"

	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	rec := httptest.NewRecorder()

	written, err := forward(dstHost, rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("got %d, want %d", resp.StatusCode, http.StatusServiceUnavailable)
	}
	if written != 4 {
		t.Errorf("got %d, want %d", written, 4)
	}
	if len(body) != 0 {
		t.Errorf("expected empty body, got: %s", body)
	}
}

func TestBalancerIntegration(t *testing.T) {
	dbServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key":"s.k.a.m", "value":"2025-05-27"}`))
	}))
	defer dbServer.Close()

	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		resp, err := http.Get(dbServer.URL + "/db/" + key)
		if err != nil || resp.StatusCode != http.StatusOK {
			http.Error(w, "db error", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, resp.Body)
	}))
	defer backendServer.Close()

	dstHost := strings.TrimPrefix(backendServer.URL, "http://")
	req := httptest.NewRequest(http.MethodGet, "http://localhost/api/v1/some-data?key=s.k.a.m", nil)
	rec := httptest.NewRecorder()

	written, err := forward(dstHost, rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
	expectedBody := `{"key":"s.k.a.m", "value":"2025-05-27"}`
	if strings.TrimSpace(string(body)) != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, body)
	}
	if written != int64(len(body)) {
		t.Errorf("expected written %d bytes, got %d", len(body), written)
	}
}
