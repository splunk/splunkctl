package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestDataModel_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/models/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "internal_audit_logs", "content": map[string]any{"acceleration.enabled": false}},
				map[string]any{"name": "internal_server", "content": map[string]any{"acceleration.enabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/models/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 data models, got %d", len(results))
	}
}

func TestDataModel_AccelerationShow(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/models/Authentication" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "Authentication", "content": map[string]any{
					"acceleration.enabled":       true,
					"acceleration.earliest_time": "-1mon",
					"acceleration.status":        "Complete",
				}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/models/", ID: "Authentication"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["acceleration.enabled"] != true {
		t.Errorf("expected acceleration enabled, got %v", results[0]["acceleration.enabled"])
	}
}

func TestDataModel_AccelerationEnable(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/models/Authentication" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("acceleration") != "1" {
			t.Errorf("expected acceleration=1, got %s", r.FormValue("acceleration"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "Authentication", "content": map[string]any{"acceleration.enabled": true}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/models/",
		ID:     "Authentication",
		Params: map[string]string{"acceleration": "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestDataModel_AccelerationDisable(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("acceleration") != "0" {
			t.Errorf("expected acceleration=0, got %s", r.FormValue("acceleration"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "Authentication", "content": map[string]any{"acceleration.enabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/models/",
		ID:     "Authentication",
		Params: map[string]string{"acceleration": "0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
