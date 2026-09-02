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

func newMaintenanceModeTestServer(t *testing.T, handler http.Handler) *client.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return client.New(&config.Config{Host: srv.URL, Token: "test-token", Insecure: true})
}

func runMaintenanceModeCmd(t *testing.T, c *client.Client, args []string) error {
	t.Helper()
	var buf bytes.Buffer
	ctx := cmdctx.WithClient(context.Background(), c)
	ctx = output.WithWriter(ctx, &buf)

	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("output", "", "")
	root.PersistentFlags().StringArray("param", nil, "")
	parent := &cobra.Command{Use: "maintenance-mode"}
	parent.AddCommand(newMaintenanceModeEnableCmd())
	parent.AddCommand(newMaintenanceModeDisableCmd())
	root.AddCommand(parent)
	root.SetContext(ctx)
	root.SetArgs(args)
	return root.Execute()
}

func TestMaintenanceModeEnable_SendsMode1(t *testing.T) {
	var gotMode string
	c := newMaintenanceModeTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotMode = r.FormValue("mode")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	if err := runMaintenanceModeCmd(t, c, []string{"maintenance-mode", "enable"}); err != nil {
		t.Fatal(err)
	}
	if gotMode != "1" {
		t.Errorf("expected mode=1, got %q", gotMode)
	}
}

func TestMaintenanceModeDisable_SendsMode0(t *testing.T) {
	var gotMode string
	c := newMaintenanceModeTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotMode = r.FormValue("mode")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	if err := runMaintenanceModeCmd(t, c, []string{"maintenance-mode", "disable"}); err != nil {
		t.Fatal(err)
	}
	if gotMode != "0" {
		t.Errorf("expected mode=0, got %q", gotMode)
	}
}

func TestMaintenanceModeEnable_ParamCannotOverrideMode(t *testing.T) {
	var gotMode string
	c := newMaintenanceModeTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotMode = r.FormValue("mode")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	// A conflicting --param mode=0 must not turn "enable" into a disable request.
	err := runMaintenanceModeCmd(t, c, []string{"maintenance-mode", "enable", "--param", "mode=0"})
	if err != nil {
		t.Fatal(err)
	}
	if gotMode != "1" {
		t.Errorf("expected mode=1 (enable semantics must win), got %q", gotMode)
	}
}

func TestMaintenanceModeDisable_ParamCannotOverrideMode(t *testing.T) {
	var gotMode string
	c := newMaintenanceModeTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		gotMode = r.FormValue("mode")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{}})
	}))

	// A conflicting --param mode=1 must not turn "disable" into an enable request.
	err := runMaintenanceModeCmd(t, c, []string{"maintenance-mode", "disable", "--param", "mode=1"})
	if err != nil {
		t.Fatal(err)
	}
	if gotMode != "0" {
		t.Errorf("expected mode=0 (disable semantics must win), got %q", gotMode)
	}
}
