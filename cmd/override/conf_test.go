package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestConf_Get(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/configs/conf-props/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "source::...log", "content": map[string]any{"SHOULD_LINEMERGE": "0"}},
				map[string]any{"name": "default", "content": map[string]any{"CHARSET": "UTF-8"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/configs/conf-props/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 stanzas, got %d", len(results))
	}
}

func TestConf_GetStanza(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/configs/conf-props/myapp" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "myapp", "content": map[string]any{"SHOULD_LINEMERGE": "0"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/configs/conf-props/", ID: "myapp"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "myapp" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestConf_Set_Create(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/configs/conf-props/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		if name == "" {
			t.Error("name required for conf create")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{"SHOULD_LINEMERGE": r.FormValue("SHOULD_LINEMERGE")}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/configs/conf-props/",
		Params: map[string]string{"name": "newstanza", "SHOULD_LINEMERGE": "0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "newstanza" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestConf_Set_Update_409Fallback(t *testing.T) {
	// Splunk upsert: POST to base returns 409 when stanza exists; retry with ID path.
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/services/configs/conf-props/" {
			w.WriteHeader(409)
			json.NewEncoder(w).Encode(map[string]any{
				"messages": []any{map[string]any{"type": "ERROR", "text": "An object with name=existingstanza already exists"}},
			})
			return
		}
		if r.URL.Path == "/services/configs/conf-props/existingstanza" {
			r.ParseForm()
			json.NewEncoder(w).Encode(map[string]any{
				"entry": []any{
					map[string]any{"name": "existingstanza", "content": map[string]any{"SHOULD_LINEMERGE": r.FormValue("SHOULD_LINEMERGE")}},
				},
			})
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
		w.WriteHeader(500)
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/configs/conf-props/",
		Params: map[string]string{"name": "existingstanza", "SHOULD_LINEMERGE": "1"},
	})
	if err == nil {
		t.Error("expected 409 error on first call to base path")
	}

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/configs/conf-props/",
		ID:     "existingstanza",
		Params: map[string]string{"SHOULD_LINEMERGE": "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "existingstanza" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestConf_Delete(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/configs/conf-props/mystanza" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{Method: "DELETE", Path: "/configs/conf-props/", ID: "mystanza"})
	if err != nil {
		t.Fatal(err)
	}
}
