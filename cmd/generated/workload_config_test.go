package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestWorkloadConfig_GetBaseDirname(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/config/get-base-dirname" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "get-base-dirname", "content": map[string]any{"base_dir": "/opt/splunk"}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/workloads/config/get-base-dirname",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestWorkloadConfig_SetBaseDirname(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/config/set-base-dirname" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("workload_pool_base_dir_name") != "/opt/splunk/var" {
			t.Errorf("unexpected workload_pool_base_dir_name: %q", r.FormValue("workload_pool_base_dir_name"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "set-base-dirname", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/workloads/config/set-base-dirname",
		Params: map[string]string{"workload_pool_base_dir_name": "/opt/splunk/var"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestWorkloadConfig_Check(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/workloads/config/preflight-checks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "preflight-checks", "content": map[string]any{"status": "ok"}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/workloads/config/preflight-checks",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
