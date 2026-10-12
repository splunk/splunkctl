package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/cmdctx"
)

func TestRootUserAgentGroupsRequestsByInvocation(t *testing.T) {
	var mu sync.Mutex
	var userAgents []string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		userAgents = append(userAgents, r.UserAgent())
		mu.Unlock()
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	parent := &cobra.Command{Use: "telemetry-test", Hidden: true}
	leaf := &cobra.Command{
		Use:  "nested <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := cmdctx.ClientFrom(cmd)
			for _, path := range []string{"/server/info", "/data/indexes/"} {
				if _, _, err := client.RawDo(cmd.Context(), http.MethodGet, path, nil); err != nil {
					return err
				}
			}
			return nil
		},
	}
	parent.AddCommand(leaf)
	rootCmd.AddCommand(parent)

	oldContext := rootCmd.Context()
	type savedFlag struct {
		name    string
		value   string
		changed bool
	}
	var saved []savedFlag
	for _, name := range []string{"host", "token", "insecure", "timeout", "output", flagYesName, flagReadOnlyName} {
		flag := rootCmd.PersistentFlags().Lookup(name)
		saved = append(saved, savedFlag{name: name, value: flag.Value.String(), changed: flag.Changed})
	}
	t.Cleanup(func() {
		rootCmd.RemoveCommand(parent)
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(oldContext)
		for _, state := range saved {
			flag := rootCmd.PersistentFlags().Lookup(state.name)
			if err := flag.Value.Set(state.value); err != nil {
				t.Errorf("restore --%s: %v", state.name, err)
			}
			flag.Changed = state.changed
		}
	})

	for range 2 {
		rootCmd.SetContext(context.Background())
		leaf.SetContext(nil)
		rootCmd.SetArgs([]string{
			"--host", server.URL,
			"--token", "test-token",
			"--insecure=true",
			"--timeout=30",
			"--output=json",
			"--read-only=true",
			"telemetry-test", "nested", "private-argument",
		})
		if _, err := rootCmd.ExecuteC(); err != nil {
			t.Fatal(err)
		}
	}

	mu.Lock()
	got := append([]string(nil), userAgents...)
	mu.Unlock()
	if len(got) != 4 {
		t.Fatalf("got %d requests, want 4", len(got))
	}
	pattern := regexp.MustCompile(`^splunkctl/` + regexp.QuoteMeta(Version) + ` \(cmd=telemetry-test_nested; requestID=([0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})\)$`)
	var ids []string
	for _, ua := range got {
		match := pattern.FindStringSubmatch(ua)
		if match == nil {
			t.Fatalf("unexpected User-Agent %q", ua)
		}
		ids = append(ids, match[1])
	}
	if ids[0] != ids[1] || ids[2] != ids[3] {
		t.Fatalf("requests within one invocation have different IDs: %v", ids)
	}
	if ids[0] == ids[2] {
		t.Fatalf("separate invocations reused ID %q", ids[0])
	}
}
