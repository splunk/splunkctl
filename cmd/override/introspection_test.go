package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestIntrospection_Queues(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/server/introspection/queues/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "indexQueue", "content": map[string]any{"current_size_kb": 0, "max_size_kb": 500}},
				map[string]any{"name": "typingQueue", "content": map[string]any{"current_size_kb": 10, "max_size_kb": 500}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/server/introspection/queues/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 queues, got %d", len(results))
	}
}

func TestIntrospection_Indexer(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/server/introspection/indexer/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "indexer", "content": map[string]any{"status": "normal", "average_KBps": 100.5}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/server/introspection/indexer/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["status"] != "normal" {
		t.Errorf("unexpected result: %v", results)
	}
}
