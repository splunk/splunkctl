package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRaw_Get(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/server/info" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "server_info", "content": map[string]any{"version": "9.3.0"}},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "GET", "/server/info", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	if feed["entry"] == nil {
		t.Error("expected entry in response")
	}
}

func TestRaw_StripServicesPrefix(t *testing.T) {
	// raw normalizes /services/... → /... before RawDo prepends /services
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/server/info" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	normalizedPath := strings.TrimPrefix("/services/server/info", "/services")
	body, status, err := c.RawDo(context.Background(), "GET", normalizedPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
}

func TestRaw_Post(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, status, err := c.RawDo(context.Background(), "POST", "/data/indexes/", url.Values{"name": []string{"testidx"}})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
}

func TestRaw_Delete(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))

	body, status, err := c.RawDo(context.Background(), "DELETE", "/data/indexes/testidx", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
}
