package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestShclusterConfig_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/shcluster/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "config", "content": map[string]any{}}},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/shcluster/config"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestShclusterConfig_Init(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/shcluster/config/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("replication_port") == "" {
			t.Error("replication_port param required")
		}
		if r.FormValue("mgmt_uri") == "" {
			t.Error("mgmt_uri param required")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "ok", "content": map[string]any{}}},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/shcluster/config/config",
		Params: map[string]string{
			"replication_port": "9100",
			"mgmt_uri":         "https://sh1.example.com:8089",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestShclusterConfig_Disable(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/shcluster/config/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "ok", "content": map[string]any{}}},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/shcluster/config/config",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestShclusterConfig_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/shcluster/config/config" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "ok", "content": map[string]any{}}},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/shcluster/config/config",
		Params: map[string]string{"replication_port": "9200"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
