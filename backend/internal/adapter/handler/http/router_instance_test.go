package http_test

import (
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_HealthBodyNamesTheServingInstance(t *testing.T) {
	app := newTestRouter(nil, nil, nil)

	response, err := app.Test(httptest.NewRequest(nethttp.MethodGet, "/health", nil))
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode /health body: %v", err)
	}
	if body["status"] != "healthy" || body["instance"] != testInstanceName {
		t.Fatalf("GET /health body = %v, want status healthy and instance %q", body, testInstanceName)
	}
}

func TestRouter_EveryResponseNamesTheServingInstance(t *testing.T) {
	app := newTestRouter(nil, nil, nil)

	for _, path := range []string{"/health", "/api/v1/sprints"} {
		response, err := app.Test(httptest.NewRequest(nethttp.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if got := response.Header.Get("X-Instance"); got != testInstanceName {
			t.Fatalf("GET %s X-Instance = %q, want %q", path, got, testInstanceName)
		}
	}
}
