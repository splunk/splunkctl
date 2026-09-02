package override_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestScriptedInput_List(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/services/data/inputs/script/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{
				map[string]any{"name": "/opt/splunk/bin/scripts/bandwidth.sh", "content": map[string]any{"interval": "60", "disabled": false}},
			},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/data/inputs/script/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 scripted input, got %d", len(results))
	}
}

func TestScriptedInput_EnableDisable(t *testing.T) {
	for _, action := range []string{"enable", "disable"} {
		action := action
		t.Run(action, func(t *testing.T) {
			c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/"+action) {
					t.Errorf("expected path ending with /%s, got %s", action, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"entry": []any{
						map[string]any{"name": "bandwidth.sh", "content": map[string]any{"disabled": action == "disable"}},
					},
				})
			}))

			results, err := c.Do(context.Background(), client.Request{
				Method: "POST",
				Path:   "/data/inputs/script/bandwidth.sh/" + action,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 result, got %d", len(results))
			}
		})
	}
}
