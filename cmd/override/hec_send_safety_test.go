package override

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
)

func TestHECSendUsesGuardedClient(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	defer server.Close()
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	denied := errors.New("denied")
	clientConfig := &config.Config{Host: server.URL, Token: "test-token"}
	ctx := cmdctx.WithClient(context.Background(), client.NewWithWriteCheck(
		clientConfig,
		func(context.Context) error { return denied },
	))

	cmd := &cobra.Command{}
	cmd.Flags().String("event", "test event", "")
	cmd.Flags().String("index", "", "")
	cmd.Flags().String("sourcetype", "", "")
	cmd.Flags().String("source", "", "")
	cmd.Flags().String("hec-token", "", "")
	cmd.Flags().String("hec-port", serverURL.Port(), "")
	cmd.Flags().String("host", server.URL, "")
	cmd.Flags().String("token", "test-token", "")
	cmd.Flags().Bool("insecure", false, "")
	cmd.SetContext(ctx)

	if err := hecSendCmd.RunE(cmd, nil); !errors.Is(err, denied) {
		t.Fatalf("error = %v, want denied", err)
	}
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}
