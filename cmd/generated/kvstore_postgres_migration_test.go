package generated_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/splunk/splunkctl/cmd/generated"
	"github.com/splunk/splunkctl/internal/client"
)

func TestKvstorePostgresMigration_Stop(t *testing.T) {
	c, _ := newTLSTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/services/kvstore/migrateToPostgres/stop" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"entry": []any{map[string]any{"name": "x", "content": map[string]any{}}},
		})
	}))
	results, err := c.Do(context.Background(), client.Request{Method: "POST", Path: "/kvstore/migrateToPostgres/stop"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestKvstorePostgresMigration_Stop_Command(t *testing.T) {
	if generated.KvstorePostgresMigrationStopCmd.Use != "stop" {
		t.Errorf("unexpected Use: %q", generated.KvstorePostgresMigrationStopCmd.Use)
	}
}
