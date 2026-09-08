package override

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
)

func TestLookupUploadChecksBeforeStaging(t *testing.T) {
	input := filepath.Join(t.TempDir(), "lookup.csv")
	if err := os.WriteFile(input, []byte("key,value\na,b\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	splunkHome := filepath.Join(t.TempDir(), "splunk")
	t.Setenv("SPLUNK_HOME", splunkHome)

	denied := errors.New("denied")
	ctx := cmdctx.WithWriteCheck(context.Background(), func(context.Context) error { return denied })
	ctx = cmdctx.WithClient(ctx, client.New(&config.Config{Host: "https://splunk.invalid", Token: "tok"}))

	cmd := &cobra.Command{}
	cmd.Flags().String("file", input, "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("output", "", "")
	cmd.SetContext(ctx)

	if err := lookupFileUploadCmd.RunE(cmd, nil); !errors.Is(err, denied) {
		t.Fatalf("error = %v, want denied", err)
	}
	stagingDir := filepath.Join(splunkHome, "var", "run", "splunk", "lookup_tmp")
	if _, err := os.Stat(stagingDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("staging directory was created: %v", err)
	}
}
