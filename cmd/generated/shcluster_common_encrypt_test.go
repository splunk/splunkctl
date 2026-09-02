package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestShclusterCommonEncrypt_Edit(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/shcluster/member/control/control/reencrypt" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		r.ParseForm()
		if r.FormValue("config") == "" {
			t.Error("config param required")
		}
		if r.FormValue("prefix") == "" {
			t.Error("prefix param required")
		}
		if r.FormValue("key") == "" {
			t.Error("key param required")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "ok", "content": map[string]any{}}},
		})
	}))

	results, err := c.Do(context.Background(), client.Request{
		Method: "POST",
		Path:   "/shcluster/member/control/control/reencrypt",
		Params: map[string]string{
			"config": "app",
			"prefix": "credential::",
			"key":    "password",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
