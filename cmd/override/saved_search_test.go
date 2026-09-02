package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestSavedSearch_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/saved/searches/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "My Alert", "content": map[string]any{"search": "index=main error | stats count", "is_scheduled": "1"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/saved/searches/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "My Alert" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestSavedSearch_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		r.ParseForm()
		if r.FormValue("name") == "" || r.FormValue("search") == "" {
			t.Error("name and search required for saved-search add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": r.FormValue("name"), "content": map[string]any{"search": r.FormValue("search")}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/saved/searches/",
		Params: map[string]string{"name": "My Alert", "search": "index=main error | stats count"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "My Alert" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestSavedSearch_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Go's HTTP server decodes %20 → space in r.URL.Path
		if r.URL.Path != "/services/saved/searches/My Alert" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "My Alert", "content": map[string]any{"search": r.FormValue("search")}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/saved/searches/",
		ID:     "My Alert",
		Params: map[string]string{"search": "index=main fatal | stats count"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSavedSearch_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		// Go's HTTP server decodes %20 → space in r.URL.Path
		if r.URL.Path != "/services/saved/searches/My Alert" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{Method: "DELETE", Path: "/saved/searches/", ID: "My Alert"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSavedSearch_Dispatch(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Go's HTTP server decodes %20 → space in r.URL.Path
		if r.URL.Path != "/services/saved/searches/My Alert/dispatch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "sid_12345", "content": map[string]any{"sid": "sid_12345"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/saved/searches/My Alert/dispatch",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
