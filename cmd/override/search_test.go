package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

func TestSearch_SubmitJob(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/search/jobs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("search") == "" {
			t.Error("search param required")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "1234567890.123", "content": map[string]any{"sid": "1234567890.123"}},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "POST", "/search/jobs", url.Values{
		"search":      []string{"search index=main error"},
		"output_mode": []string{"json"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d: %s", status, body)
	}
}

func TestSearch_PollJob(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/search/jobs/1234567890.123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{
					"name": "1234567890.123",
					"content": map[string]any{
						"sid":           "1234567890.123",
						"isDone":        true,
						"dispatchState": "DONE",
						"resultCount":   3,
					},
				},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "GET", "/search/jobs/1234567890.123", nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	entries := feed["entry"].([]any)
	content := entries[0].(map[string]any)["content"].(map[string]any)
	if content["isDone"] != true {
		t.Errorf("expected isDone=true")
	}
}

func TestSearch_FetchResults(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/search/jobs/1234567890.123/results" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"results": []any{
				map[string]any{"host": "myhost", "source": "/var/log/app.log", "_raw": "error occurred"},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "GET", "/search/jobs/1234567890.123/results", url.Values{
		"count": []string{"100"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("expected 200, got %d", status)
	}
	var feed map[string]any
	json.Unmarshal(body, &feed)
	results := feed["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
