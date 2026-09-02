package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestCascadePlans_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/replication/cascading/plans" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "9ADCF249-3130-4976-B89D-A6CD6B86426F", "content": map[string]any{}},
			},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "GET",
		Path:   "/replication/cascading/plans",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCascadePlans_ListByName(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/replication/cascading/plans/9ADCF249-3130-4976-B89D-A6CD6B86426F" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "9ADCF249-3130-4976-B89D-A6CD6B86426F", "content": map[string]any{}},
			},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "GET",
		Path:   "/replication/cascading/plans",
		ID:     "9ADCF249-3130-4976-B89D-A6CD6B86426F",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
