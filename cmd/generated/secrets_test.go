package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/cmd/generated"
	"github.com/splunk/splunkctl/internal/client"
)

func TestSecrets_Reencrypt(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/server/security/secrets/reencrypt" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "x", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{Method: "POST", Path: "/server/security/secrets/reencrypt"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestSecrets_Reencrypt_Command(t *testing.T) {
	if generated.SecretsReencryptCmd.Use != "reencrypt" {
		t.Errorf("unexpected Use: %q", generated.SecretsReencryptCmd.Use)
	}
	if generated.SecretsReencryptCmd.Flags().Lookup("cipher") == nil {
		t.Errorf("missing flag cipher")
	}
}
