package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestRole_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/authorization/roles/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "admin", "content": map[string]any{}},
				map[string]any{"name": "power", "content": map[string]any{}},
				map[string]any{"name": "user", "content": map[string]any{}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/authorization/roles/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 roles, got %d", len(results))
	}
	if results[0]["name"] != "admin" {
		t.Errorf("expected admin, got %v", results[0]["name"])
	}
}

func TestRole_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/authorization/roles/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		rolename := r.FormValue("rolename")
		if name == "" {
			t.Error("name param required for role add")
		}
		if rolename == "" {
			t.Error("rolename param required for role add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{}},
			},
		})
	}))

	// role add --rolename managers: rolename falls through to params["name"] when no positional given
	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/authorization/roles/",
		Params: map[string]string{"name": "managers", "rolename": "managers"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "managers" {
		t.Errorf("expected managers, got %v", results[0]["name"])
	}
}

func TestRole_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/authorization/roles/managers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("rolename") == "" {
			t.Error("rolename param expected for role edit")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "managers", "content": map[string]any{}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/authorization/roles/",
		ID:     "managers",
		Params: map[string]string{"rolename": "managers"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "managers" {
		t.Errorf("expected managers, got %v", results[0]["name"])
	}
}

func TestRole_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/authorization/roles/managers" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/authorization/roles/",
		ID:     "managers",
	})
	if err != nil {
		t.Fatal(err)
	}
}
