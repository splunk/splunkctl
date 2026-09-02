package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestIndex_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/indexes/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "main", "content": map[string]any{"disabled": false, "maxTotalDataSizeMB": 500000}},
				map[string]any{"name": "internal", "content": map[string]any{"disabled": false, "maxTotalDataSizeMB": 300000}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 indexes, got %d", len(results))
	}
	if results[0]["name"] != "main" {
		t.Errorf("expected main, got %v", results[0]["name"])
	}
	if results[1]["name"] != "internal" {
		t.Errorf("expected internal, got %v", results[1]["name"])
	}
}

func TestIndex_List_ByName(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		// When an ID is supplied the client appends it to the path.
		if r.URL.Path != "/services/data/indexes/main" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "main", "content": map[string]any{"disabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/", ID: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 index, got %d", len(results))
	}
	if results[0]["name"] != "main" {
		t.Errorf("expected main, got %v", results[0]["name"])
	}
}

func TestIndex_List_IndexNameParam(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/indexes/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		indexName := r.URL.Query().Get("indexName")
		if indexName != "main" {
			t.Errorf("expected indexName=main, got %q", indexName)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "main", "content": map[string]any{"disabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "GET",
		Path:   "/data/indexes/",
		Params: map[string]string{"indexName": "main"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestIndex_Add(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/indexes/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		if name == "" {
			t.Error("name param required for index add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{"disabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/",
		Params: map[string]string{"name": "myindex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "myindex" {
		t.Errorf("expected myindex, got %v", results[0]["name"])
	}
}

func TestIndex_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// edit sends ID so path is .../indexes/myindex
		if r.URL.Path != "/services/data/indexes/myindex" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("maxTotalDataSizeMB") == "" {
			t.Error("maxTotalDataSizeMB param expected for index edit")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "myindex", "content": map[string]any{"maxTotalDataSizeMB": r.FormValue("maxTotalDataSizeMB")}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/",
		ID:     "myindex",
		Params: map[string]string{"maxTotalDataSizeMB": "1000000"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestIndex_Enable(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// IndexEnableCmd builds the path via strings.ReplaceAll and passes it directly
		// (no ID field used), so the server sees /services/data/indexes/myindex/enable/
		if !strings.Contains(r.URL.Path, "myindex") || !strings.HasSuffix(r.URL.Path, "/enable/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "myindex", "content": map[string]any{"disabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/myindex/enable/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["disabled"] != false {
		t.Errorf("expected disabled=false after enable, got %v", results[0]["disabled"])
	}
}

func TestIndex_Disable(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "myindex") || !strings.HasSuffix(r.URL.Path, "/disable/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "myindex", "content": map[string]any{"disabled": true}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/myindex/disable/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["disabled"] != true {
		t.Errorf("expected disabled=true after disable, got %v", results[0]["disabled"])
	}
}

func TestIndex_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/indexes/myindex" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/data/indexes/",
		ID:     "myindex",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIndex_Reload(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/indexes/_reload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/_reload",
	})
	if err != nil {
		t.Fatal(err)
	}
}
