package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestEventtype_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/saved/eventtypes/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "error_events", "content": map[string]any{"search": "error OR fatal", "priority": "1"}},
				map[string]any{"name": "web_events", "content": map[string]any{"search": "sourcetype=access_combined", "priority": "5"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/saved/eventtypes/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 eventtypes, got %d", len(results))
	}
	if results[0]["name"] != "error_events" {
		t.Errorf("expected error_events first, got %v", results[0]["name"])
	}
}

func TestEventtype_ListByName(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/saved/eventtypes/error_events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "error_events", "content": map[string]any{"search": "error OR fatal"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/saved/eventtypes/", ID: "error_events"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "error_events" {
		t.Errorf("expected error_events, got %v", results)
	}
}

func TestEventtype_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/saved/eventtypes/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		search := r.FormValue("search")
		if name == "" || search == "" {
			t.Error("name and search are required")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{"search": search}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/saved/eventtypes/",
		Params: map[string]string{"name": "my_events", "search": "index=main error"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "my_events" {
		t.Errorf("expected my_events, got %v", results)
	}
}

func TestEventtype_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/saved/eventtypes/my_events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "my_events", "content": map[string]any{"search": r.FormValue("search")}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/saved/eventtypes/",
		ID:     "my_events",
		Params: map[string]string{"search": "index=main error OR fatal"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestEventtype_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/saved/eventtypes/my_events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/saved/eventtypes/",
		ID:     "my_events",
	})
	if err != nil {
		t.Fatal(err)
	}
}
