package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestUser_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/authentication/users/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "admin", "content": map[string]any{"roles": []any{"admin"}, "email": "admin@example.com"}},
				map[string]any{"name": "alice", "content": map[string]any{"roles": []any{"power"}, "email": "alice@example.com"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/authentication/users/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 users, got %d", len(results))
	}
}

func TestUser_Add_RawDo(t *testing.T) {
	// user add uses RawDo with url.Values (multi-value roles), not Do
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/authentication/users/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		password := r.FormValue("password")
		roles := r.Form["roles"]
		if name == "" || password == "" || len(roles) == 0 {
			t.Error("name, password, roles required for user add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{"roles": roles}},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "POST", "/authentication/users/", url.Values{
		"name":     []string{"alice"},
		"password": []string{"s3cr3t"},
		"roles":    []string{"power", "user"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	entries := feed["entry"].([]any)
	if len(entries) != 1 {
		t.Fatalf("expected 1 user, got %d", len(entries))
	}
}

func TestUser_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/authentication/users/alice" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/authentication/users/",
		ID:     "alice",
	})
	if err != nil {
		t.Fatal(err)
	}
}
