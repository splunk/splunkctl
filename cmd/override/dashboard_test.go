package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestDashboard_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/ui/views/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "my_dashboard", "content": map[string]any{"eai:data": "<dashboard><label>My Dashboard</label></dashboard>"}},
			},
		})
	}))

	body, status, err := c.RawDoAbs(context.Background(), "GET", "/services/data/ui/views/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	if err := json.Unmarshal(body, &feed); err != nil {
		t.Fatal(err)
	}
	entries := feed["entry"].([]any)
	if len(entries) != 1 {
		t.Fatalf("expected 1 dashboard, got %d", len(entries))
	}
}

func TestDashboard_Show(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/ui/views/my_dashboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "my_dashboard", "content": map[string]any{"eai:data": "<dashboard></dashboard>"}},
			},
		})
	}))

	body, status, err := c.RawDoAbs(context.Background(), "GET", "/services/data/ui/views/my_dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	_ = body
}

func TestDashboard_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/ui/views/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("name") == "" || r.FormValue("eai:data") == "" {
			t.Error("name and eai:data required for dashboard add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": r.FormValue("name"), "content": map[string]any{"eai:data": r.FormValue("eai:data")}},
			},
		})
	}))

	body := url.Values{}
	body.Set("name", "my_dashboard")
	body.Set("eai:data", "<dashboard><label>Test</label></dashboard>")
	respBody, status, err := c.RawDoAbs(context.Background(), "POST", "/services/data/ui/views/", body)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, respBody)
	}
}

func TestDashboard_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/ui/views/my_dashboard" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, status, err := c.RawDoAbs(context.Background(), "DELETE", "/services/data/ui/views/my_dashboard", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
}

func TestDashboard_AppNamespace(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/servicesNS/nobody/search/") {
			t.Errorf("expected servicesNS path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, _, err := c.RawDoAbs(context.Background(), "GET", "/servicesNS/nobody/search/data/ui/views/", nil)
	if err != nil {
		t.Fatal(err)
	}
}
