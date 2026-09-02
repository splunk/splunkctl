package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestACL_Show_TwoStep(t *testing.T) {
	requestCount := 0
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/services/saved/eventtypes/") && !strings.Contains(r.URL.Path, "servicesNS"):
			json.NewEncoder(w).Encode(map[string]any{
				"entry": []any{
					map[string]any{
						"name": "my_eventtype",
						"acl":  map[string]any{"owner": "admin", "app": "search", "sharing": "global"},
					},
				},
			})
		case strings.Contains(r.URL.Path, "/acl"):
			json.NewEncoder(w).Encode(map[string]any{
				"entry": []any{
					map[string]any{
						"name":    "my_eventtype",
						"content": map[string]any{"sharing": "global", "owner": "admin"},
					},
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/saved/eventtypes/", ID: "my_eventtype"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	aclBody, aclStatus, err := c.RawDoAbs(context.Background(), "GET",
		"/servicesNS/admin/search/saved/eventtypes/my_eventtype/acl", nil)
	if err != nil {
		t.Fatal(err)
	}
	if aclStatus != 200 {
		t.Fatalf("expected 200, got %d: %s", aclStatus, aclBody)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests for ACL show, got %d", requestCount)
	}
}
