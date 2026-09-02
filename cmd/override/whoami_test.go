package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestWhoami_TwoStepRequest(t *testing.T) {
	requestCount := 0
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/services/authentication/current-context/":
			json.NewEncoder(w).Encode(map[string]any{
				"entry": []any{
					map[string]any{"name": "admin", "content": map[string]any{"username": "admin", "capabilities": []any{"admin_all_objects"}}},
				},
			})
		case "/services/authentication/users/admin":
			json.NewEncoder(w).Encode(map[string]any{
				"entry": []any{
					map[string]any{"name": "admin", "content": map[string]any{"roles": []any{"admin"}, "email": "admin@example.com", "realname": "Admin User"}},
				},
			})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))

	ctx := context.Background()
	results, err := c.Do(ctx, client.Request{Method: "GET", Path: "/authentication/current-context/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result from current-context, got %d", len(results))
	}

	userResults, err := c.Do(ctx, client.Request{Method: "GET", Path: "/authentication/users/", ID: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(userResults) != 1 {
		t.Fatalf("expected 1 user result, got %d", len(userResults))
	}
	if userResults[0]["name"] != "admin" {
		t.Errorf("expected admin, got %v", userResults[0]["name"])
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests for whoami, got %d", requestCount)
	}
}
