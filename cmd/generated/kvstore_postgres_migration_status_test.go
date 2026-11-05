package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/cmd/generated"
	"github.com/splunk/splunkctl/internal/client"
)

func TestKvstorePostgresMigrationStatus_Show(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/services/kvstore/migrateToPostgres/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "x", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{Method: "GET", Path: "/kvstore/migrateToPostgres/status"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestKvstorePostgresMigrationStatus_Show_Command(t *testing.T) {
	if generated.KvstorePostgresMigrationStatusShowCmd.Use != "show" {
		t.Errorf("unexpected Use: %q", generated.KvstorePostgresMigrationStatusShowCmd.Use)
	}
	if generated.KvstorePostgresMigrationStatusShowCmd.Flags().Lookup("verbose") == nil {
		t.Errorf("missing flag verbose")
	}
}
