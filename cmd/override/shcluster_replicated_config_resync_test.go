package override

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
	"github.com/splunk/splunkctl/internal/output"
)

func newShclusterReplicatedConfigResyncTestServer(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return client.New(&config.Config{Host: srv.URL, Token: "test-token", Insecure: true})
}

func runShclusterReplicatedConfigResyncCmd(t *testing.T, c *client.Client, args []string) error {
	t.Helper()
	var buf bytes.Buffer
	ctx := cmdctx.WithClient(context.Background(), c)
	ctx = output.WithWriter(ctx, &buf)

	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("output", "", "")
	root.PersistentFlags().StringArray("param", nil, "")
	parent := &cobra.Command{Use: "shcluster-replicated-config"}
	parent.AddCommand(newShclusterReplicatedConfigResyncCmd())
	root.AddCommand(parent)
	root.SetContext(ctx)
	root.SetArgs(args)
	return root.Execute()
}

func TestShclusterReplicatedConfigResync_SendsResyncDestructiveTrue(t *testing.T) {
	var gotVal string
	c := newShclusterReplicatedConfigResyncTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotVal = r.FormValue("resync_destructive")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	if err := runShclusterReplicatedConfigResyncCmd(t, c, []string{"shcluster-replicated-config", "resync"}); err != nil {
		t.Fatal(err)
	}
	if gotVal != "true" {
		t.Errorf("expected resync_destructive=true, got %q", gotVal)
	}
}

func TestShclusterReplicatedConfigResync_ParamCannotOverrideResyncDestructive(t *testing.T) {
	var gotVal string
	c := newShclusterReplicatedConfigResyncTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotVal = r.FormValue("resync_destructive")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	// A conflicting --param resync_destructive=false must not defeat the
	// command's documented destructive-resync semantics.
	err := runShclusterReplicatedConfigResyncCmd(t, c, []string{"shcluster-replicated-config", "resync", "--param", "resync_destructive=false"})
	if err != nil {
		t.Fatal(err)
	}
	if gotVal != "true" {
		t.Errorf("expected resync_destructive=true (resync semantics must win), got %q", gotVal)
	}
}
