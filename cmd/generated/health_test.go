package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestHealth_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// HealthListCmd uses path "/server/health/splunkd/details" (no trailing slash, no ID).
		if r.URL.Path != "/services/server/health/splunkd/details" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "splunkd",
					"content": map[string]any{
						"health":  "green",
						"reasons": map[string]any{},
					},
				},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/server/health/splunkd/details"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 health entry, got %d", len(results))
	}
	if results[0]["name"] != "splunkd" {
		t.Errorf("expected splunkd, got %v", results[0]["name"])
	}
	if results[0]["health"] != "green" {
		t.Errorf("expected health=green, got %v", results[0]["health"])
	}
}

func TestHealth_List_WithFeaturesParam(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/server/health/splunkd/details" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		features := r.URL.Query().Get("features")
		if features != "BatchReader, TailReader" {
			t.Errorf("expected features param, got %q", features)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "splunkd",
					"content": map[string]any{
						"health": "yellow",
						"reasons": map[string]any{
							"BatchReader": "indexing queue fill ratio is high",
						},
					},
				},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "GET",
		Path:   "/server/health/splunkd/details",
		Params: map[string]string{"features": "BatchReader, TailReader"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 health entry, got %d", len(results))
	}
	if results[0]["health"] != "yellow" {
		t.Errorf("expected health=yellow, got %v", results[0]["health"])
	}
}

func TestHealth_List_Unhealthy(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "splunkd",
					"content": map[string]any{
						"health": "red",
						"reasons": map[string]any{
							"IndexProcessor": "disk usage is critical",
						},
					},
				},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/server/health/splunkd/details"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 health entry, got %d", len(results))
	}
	if results[0]["health"] != "red" {
		t.Errorf("expected health=red, got %v", results[0]["health"])
	}
}
