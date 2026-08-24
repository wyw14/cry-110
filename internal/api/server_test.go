package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	system, err := NewSystem(filepath.Join(t.TempDir(), "state"), time.Now().UTC())
	if err != nil {
		t.Fatalf("NewSystem: %v", err)
	}
	return NewServer(system)
}

func TestProbeRoutes(t *testing.T) {
	server := newTestServer(t)
	paths := []string{
		"/healthz",
		"/api/isolation",
		"/api/pitch",
		"/api/yaw",
		"/api/crane",
		"/api/turning-gear",
		"/api/incidents",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", path, response.Code, http.StatusOK)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("GET %s content type = %q", path, contentType)
			}
		})
	}
}

func TestIsolationAndIncidentPersist(t *testing.T) {
	server := newTestServer(t)
	isolationBody := bytes.NewBufferString(`{"turbine_id":"WTG-17","reason":"gearbox inspection"}`)
	isolationRequest := httptest.NewRequest(http.MethodPost, "/api/isolation", isolationBody)
	isolationResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(isolationResponse, isolationRequest)
	if isolationResponse.Code != http.StatusAccepted {
		t.Fatalf("POST isolation status = %d, body = %s", isolationResponse.Code, isolationResponse.Body.String())
	}

	incidentBody := bytes.NewBufferString(`{"kind":"icing","message":"blade ice detected"}`)
	incidentRequest := httptest.NewRequest(http.MethodPost, "/api/incidents", incidentBody)
	incidentResponse := httptest.NewRecorder()
	server.Handler().ServeHTTP(incidentResponse, incidentRequest)
	if incidentResponse.Code != http.StatusAccepted {
		t.Fatalf("POST incident status = %d, body = %s", incidentResponse.Code, incidentResponse.Body.String())
	}

	events, err := server.system.Journal.Events()
	if err != nil {
		t.Fatalf("read journal: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("journal events = %d, want 2", len(events))
	}
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil || len(data) == 0 {
			t.Fatalf("event is not serializable: %v", err)
		}
	}
}
