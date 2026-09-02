package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestFieldExtraction_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/props/extractions/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "access_combined : EXTRACT-my_ext",
					"content": map[string]any{
						"attribute": "EXTRACT-my_ext",
						"stanza":    "access_combined",
						"type":      "Uses transform",
						"value":     `(?P<user>\w+)`,
					},
				},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/props/extractions/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !strings.Contains(results[0]["name"].(string), "EXTRACT-my_ext") {
		t.Errorf("expected compound name with EXTRACT-my_ext, got %v", results[0]["name"])
	}
}

func TestFieldExtraction_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/props/extractions/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		stanza := r.FormValue("stanza")
		typ := r.FormValue("type")
		value := r.FormValue("value")
		if name == "" || stanza == "" || typ == "" || value == "" {
			t.Error("name, stanza, type, value all required for field-extraction add")
		}
		compoundName := stanza + " : EXTRACT-" + name
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": compoundName, "content": map[string]any{"value": value, "stanza": stanza, "type": "Uses transform"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/props/extractions/",
		Params: map[string]string{
			"name":   "my_ext",
			"stanza": "access_combined",
			"type":   "EXTRACT",
			"value":  `(?P<user>\w+)`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "access_combined : EXTRACT-my_ext" {
		t.Errorf("expected compound name, got %v", results[0]["name"])
	}
}

func TestFieldExtraction_EditByCompoundName(t *testing.T) {
	compoundName := "access_combined : EXTRACT-my_ext"
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "EXTRACT-my_ext") {
			t.Errorf("expected compound name in path, got: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": compoundName, "content": map[string]any{"value": `(?P<user>[a-zA-Z]+)`}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/props/extractions/",
		ID:     compoundName,
		Params: map[string]string{"value": `(?P<user>[a-zA-Z]+)`},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestFieldExtraction_Remove(t *testing.T) {
	compoundName := "access_combined : EXTRACT-my_ext"
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "EXTRACT-my_ext") {
			t.Errorf("expected compound name in delete path, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/data/props/extractions/",
		ID:     compoundName,
	})
	if err != nil {
		t.Fatal(err)
	}
}
