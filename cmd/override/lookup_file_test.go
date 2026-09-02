package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestLookupFile_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/lookup-table-files/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "my_lookup.csv", "content": map[string]any{"eai:data": "/opt/splunk/etc/users/admin/search/lookups/my_lookup.csv"}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/lookup-table-files/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "my_lookup.csv" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestLookupFile_Upload_TwoStep(t *testing.T) {
	// Splunk rejects multipart/raw content; eai:data must be a pre-staged local path.
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/lookup-table-files/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		name := r.FormValue("name")
		eaiData := r.FormValue("eai:data")
		if name == "" {
			t.Error("name required for lookup-file upload")
		}
		if eaiData == "" {
			t.Error("eai:data (staging path) required for lookup-file upload")
		}
		if !strings.HasPrefix(eaiData, "/") {
			t.Errorf("eai:data must be an absolute file path, got: %s", eaiData)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{"eai:data": "/opt/splunk/etc/users/admin/search/lookups/" + name}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/lookup-table-files/",
		Params: map[string]string{
			"name":     "test.csv",
			"eai:data": "/tmp/splunk_lookup_tmp/test.csv",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "test.csv" {
		t.Errorf("unexpected result: %v", results)
	}
}

func TestLookupFile_Remove(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/data/lookup-table-files/my_lookup.csv" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "DELETE",
		Path:   "/data/lookup-table-files/",
		ID:     "my_lookup.csv",
	})
	if err != nil {
		t.Fatal(err)
	}
}
