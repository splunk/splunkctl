package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/cmd/generated"
	"github.com/splunk/splunkctl/internal/client"
)

func TestShclusterSecrets_Reencrypt(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/shcluster/captain/control/default/reencrypt-secrets" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "reencrypt-secrets", "content": map[string]any{"success": true}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{Method: "POST", Path: "/shcluster/captain/control/default/reencrypt-secrets"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestShclusterSecrets_Reencrypt_Command(t *testing.T) {
	if generated.ShclusterSecretsReencryptCmd.Use != "reencrypt" {
		t.Errorf("unexpected Use: %q", generated.ShclusterSecretsReencryptCmd.Use)
	}
	if generated.ShclusterSecretsReencryptCmd.Flags().Lookup("cipher") == nil {
		t.Errorf("missing flag cipher")
	}
}
