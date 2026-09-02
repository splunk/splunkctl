package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/output"
)

func TestSmoke_IndexListAndRenderJSON(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/indexes/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing or wrong auth header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "main", "content": map[string]any{"maxDataSizeMB": 500}},
				map[string]any{"name": "summary", "content": map[string]any{"maxDataSizeMB": 0}},
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
		t.Errorf("expected first index to be main, got %v", results[0]["name"])
	}

	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	if err := output.Print(ctx, results, "json"); err != nil {
		t.Fatal(err)
	}

	var parsed []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v — got: %s", err, buf.String())
	}
	if len(parsed) != 2 {
		t.Fatalf("expected 2 JSON results, got %d", len(parsed))
	}
}

func TestSmoke_AuthErrorPropagation(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(403)
		json.NewEncoder(w).Encode(map[string]any{
			"messages": []any{map[string]any{"type": "ERROR", "text": "Permission denied"}},
		})
	}))

	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "AUTH_ERROR") {
		t.Errorf("expected AUTH_ERROR in error, got: %v", err)
	}
}

func TestSmoke_TableOutputNameFirst(t *testing.T) {
	data := []map[string]any{
		{"name": "main", "disabled": false},
		{"name": "summary", "disabled": true},
	}
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	if err := output.Print(ctx, data, "table"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "main") || !strings.Contains(out, "summary") {
		t.Errorf("expected both index names in table output, got: %s", out)
	}
	lines := strings.Split(out, "\n")
	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "name") {
		t.Errorf("expected 'name' as first column header, got: %s", lines[0])
	}
}

func TestSmoke_PostParamsRoundTrip(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("could not parse form: %v", err)
		}
		name := r.FormValue("name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": name, "content": map[string]any{"maxDataSizeMB": 500}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/data/indexes/",
		Params: map[string]string{"name": "newindex"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "newindex" {
		t.Errorf("expected name=newindex echoed back, got %v", results[0]["name"])
	}
}

func TestSmoke_404ReturnsNotFound(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{
			"messages": []any{map[string]any{"type": "ERROR", "text": "index not found"}},
		})
	}))

	_, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/", ID: "missing"})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "NOT_FOUND") {
		t.Errorf("expected NOT_FOUND in error, got: %v", err)
	}
}

func TestSmoke_EmptyEntryFeedRendersEmptyJSON(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/indexes/"})
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	if err := output.Print(ctx, results, "json"); err != nil {
		t.Fatal(err)
	}
	trimmed := strings.TrimSpace(buf.String())
	if trimmed != "[]" {
		t.Errorf("expected [] for empty feed, got %q", trimmed)
	}
}
