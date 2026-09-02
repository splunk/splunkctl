package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestWorkloadCategory_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/categories/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "search", "content": map[string]any{"cpu_weight": "50", "mem_weight": "50"}},
				map[string]any{"name": "misc", "content": map[string]any{"cpu_weight": "50", "mem_weight": "50"}},
			},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "GET",
		Path:   "/workloads/categories/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestWorkloadCategory_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/categories/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("cpu_weight") != "60" {
			t.Errorf("expected cpu_weight=60, got %q", r.FormValue("cpu_weight"))
		}
		if r.FormValue("mem_weight") != "40" {
			t.Errorf("expected mem_weight=40, got %q", r.FormValue("mem_weight"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "search", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/workloads/categories/",
		ID:     "search",
		Params: map[string]string{
			"cpu_weight": "60",
			"mem_weight": "40",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
