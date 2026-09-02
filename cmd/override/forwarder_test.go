package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestForwarder_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/deployment/server/clients/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/deployment/server/clients/"})
	if err != nil {
		t.Fatal(err)
	}
	if results == nil {
		t.Error("expected non-nil results")
	}
}
