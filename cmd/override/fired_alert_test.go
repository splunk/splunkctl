package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestFiredAlert_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/alerts/fired_alerts/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "My Alert", "content": map[string]any{"triggered_alerts": 3, "severity": "high"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/alerts/fired_alerts/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 fired alert, got %d", len(results))
	}
}

func TestFiredAlert_Clear(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		// Go's HTTP server decodes %20 → space in r.URL.Path
		if r.URL.Path != "/services/alerts/fired_alerts/My Alert" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{Method: "DELETE", Path: "/alerts/fired_alerts/", ID: "My Alert"})
	if err != nil {
		t.Fatal(err)
	}
}
