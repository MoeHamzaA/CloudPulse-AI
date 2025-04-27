package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test if /logs returns HTTP 200
func TestHandleLogs(t *testing.T) {
	req := httptest.NewRequest("GET", "/logs", nil)
	w := httptest.NewRecorder()

	handleLogs(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}

// Test if /metrics returns HTTP 200
func TestHandleMetrics(t *testing.T) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handleMetrics(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}

// Test if /drift-metrics returns HTTP 200
func TestHandleDriftMetrics(t *testing.T) {
	req := httptest.NewRequest("GET", "/drift-metrics", nil)
	w := httptest.NewRecorder()

	handleDriftMetrics(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
}
