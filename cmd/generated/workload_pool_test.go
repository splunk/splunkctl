package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestWorkloadPool_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/pools/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "default", "content": map[string]any{"cpu_weight": "50", "mem_weight": "50"}},
			},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "GET",
		Path:   "/workloads/pools/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestWorkloadPool_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/pools/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("pool_name") != "pool_a" {
			t.Errorf("expected pool_name=pool_a, got %q", r.FormValue("pool_name"))
		}
		if r.FormValue("cpu_weight") != "30" {
			t.Errorf("expected cpu_weight=30, got %q", r.FormValue("cpu_weight"))
		}
		if r.FormValue("mem_weight") != "20" {
			t.Errorf("expected mem_weight=20, got %q", r.FormValue("mem_weight"))
		}
		if r.FormValue("category") != "search" {
			t.Errorf("expected category=search, got %q", r.FormValue("category"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "pool_a", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/workloads/pools/",
		Params: map[string]string{
			"pool_name":  "pool_a",
			"cpu_weight": "30",
			"mem_weight": "20",
			"category":   "search",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestWorkloadPool_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/pools/pool_a" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/workloads/pools/",
		ID:     "pool_a",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = results
}

func TestWorkloadPool_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/pools/pool_a" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("cpu_weight") != "40" {
			t.Errorf("expected cpu_weight=40, got %q", r.FormValue("cpu_weight"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "pool_a", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/workloads/pools/",
		ID:     "pool_a",
		Params: map[string]string{"cpu_weight": "40"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
