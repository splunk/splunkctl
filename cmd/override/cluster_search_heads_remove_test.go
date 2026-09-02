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

func newClusterSearchHeadsRemoveTestServer(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return client.New(&config.Config{Host: srv.URL, Token: "test-token", Insecure: true})
}

func runClusterSearchHeadsRemoveCmd(t *testing.T, c *client.Client, args []string) error {
	t.Helper()
	var buf bytes.Buffer
	ctx := cmdctx.WithClient(context.Background(), c)
	ctx = output.WithWriter(ctx, &buf)

	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("output", "", "")
	root.PersistentFlags().StringArray("param", nil, "")
	parent := &cobra.Command{Use: "cluster-search-heads"}
	parent.AddCommand(newClusterSearchHeadsRemoveCmd())
	root.AddCommand(parent)
	root.SetContext(ctx)
	root.SetArgs(args)
	return root.Execute()
}

func TestClusterSearchHeadsRemove_SendsPOST(t *testing.T) {
	var gotMethod, gotPath, gotGuids string
	c := newClusterSearchHeadsRemoveTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		r.ParseForm()
		gotGuids = r.FormValue("guids")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	err := runClusterSearchHeadsRemoveCmd(t, c, []string{"cluster-search-heads", "remove", "GUID1,GUID2"})
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %q", gotMethod)
	}
	if gotPath != "/services/cluster/master/control/default/remove_search_heads" {
		t.Errorf("unexpected path %q", gotPath)
	}
	if gotGuids != "GUID1,GUID2" {
		t.Errorf("expected guids=GUID1,GUID2, got %q", gotGuids)
	}
}

func TestClusterSearchHeadsRemove_GuidsFlag(t *testing.T) {
	var gotGuids string
	c := newClusterSearchHeadsRemoveTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotGuids = r.FormValue("guids")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	err := runClusterSearchHeadsRemoveCmd(t, c, []string{"cluster-search-heads", "remove", "--guids", "GUID3"})
	if err != nil {
		t.Fatal(err)
	}
	if gotGuids != "GUID3" {
		t.Errorf("expected guids=GUID3, got %q", gotGuids)
	}
}
