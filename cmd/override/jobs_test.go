package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestJobs_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/search/jobs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "abc.1", "content": map[string]any{"sid": "abc.1", "dispatchState": "DONE", "isDone": true, "resultCount": 10}},
			},
		})
	}))

	body, status, err := c.RawDo(context.Background(), "GET", "/search/jobs", nil)
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
		t.Fatalf("expected 1 job, got %d", len(entries))
	}
}

func TestJobs_Cancel(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/services/search/jobs/abc.1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{Method: "DELETE", Path: "/search/jobs/abc.1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestJobs_Control_Pause(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/search/jobs/abc.1/control" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("action") != "pause" {
			t.Errorf("expected action=pause, got %s", r.FormValue("action"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	_, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/search/jobs/abc.1/control",
		Params: map[string]string{"action": "pause"},
	})
	if err != nil {
		t.Fatal(err)
	}
}
