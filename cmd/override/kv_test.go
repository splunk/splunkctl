package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestKV_CollectionList(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/storage/collections/config") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "my_collection", "content": map[string]any{"field.status": "string"}},
			},
		})
	}))

	body, status, err := c.RawDoAbs(context.Background(), "GET", "/servicesNS/nobody/search/storage/collections/config/", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	entries := feed["entry"].([]any)
	if len(entries) != 1 {
		t.Fatalf("expected 1 collection, got %d", len(entries))
	}
}

func TestKV_CollectionAdd(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		r.ParseForm()
		if r.FormValue("name") == "" {
			t.Error("name required for collection add")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": r.FormValue("name"), "content": map[string]any{}},
			},
		})
	}))

	body := url.Values{"name": []string{"my_collection"}}
	respBody, status, err := c.RawDoAbs(context.Background(), "POST", "/servicesNS/nobody/search/storage/collections/config/", body)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, respBody)
	}
}

func TestKV_DataQuery(t *testing.T) {
	// KV Store data endpoint returns a JSON array (not Atom feed)
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/storage/collections/data/my_collection") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{
			map[string]any{"_key": "abc123", "status": "active", "host": "myhost"},
		})
	}))

	body, status, err := c.RawDoAbs(context.Background(), "GET", "/servicesNS/nobody/search/storage/collections/data/my_collection", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var results []map[string]any
	if err := json.Unmarshal(body, &results); err != nil {
		t.Fatalf("expected JSON array, got: %s", body)
	}
	if len(results) != 1 || results[0]["_key"] != "abc123" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestKV_DataPut_JSONBody(t *testing.T) {
	// KV data PUT uses JSON body, not form-encoded
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["status"] == nil {
			t.Error("expected status field in JSON body")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"_key": "new_key_123"})
	}))

	payload := map[string]any{"status": "active", "host": "myhost"}
	body, status, err := c.RawDoJSON(context.Background(), "POST", "/servicesNS/nobody/search/storage/collections/data/my_collection", payload)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
	var result map[string]any
	json.Unmarshal(body, &result)
	if result["_key"] == nil {
		t.Error("expected _key in response")
	}
}

func TestKV_DataDelete(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/storage/collections/data/my_collection/abc123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))

	_, status, err := c.RawDoAbs(context.Background(), "DELETE", "/servicesNS/nobody/search/storage/collections/data/my_collection/abc123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
}
