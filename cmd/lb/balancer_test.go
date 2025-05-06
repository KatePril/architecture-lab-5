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
