package cmdctx_test

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
)

func TestWithClient_StoresAndRetrieves(t *testing.T) {
	c := client.New(&config.Config{Host: "https://localhost:8089", Token: "tok"})
	ctx := cmdctx.WithClient(context.Background(), c)

	cmd := &cobra.Command{}
	cmd.SetContext(ctx)

	got := cmdctx.ClientFrom(cmd)
	if got != c {
		t.Error("ClientFrom returned a different client than the one stored by WithClient")
	}
}

func TestClientFrom_PanicsWhenMissing(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected ClientFrom to panic when no client is in context")
		}
	}()
	cmdctx.ClientFrom(cmd)
}

func newParamCmd(params ...string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringArray("param", params, "")
	return cmd
}

func TestExtraParams_NoFlags(t *testing.T) {
	cmd := newParamCmd()
	dst := map[string]string{"name": "myindex"}
	result := cmdctx.ExtraParams(cmd, dst)
	if result["name"] != "myindex" {
		t.Errorf("expected name=myindex to be preserved, got %v", result["name"])
	}
	if len(result) != 1 {
		t.Errorf("expected no extra keys, got %v", result)
	}
}

func TestExtraParams_SingleParam(t *testing.T) {
	cmd := newParamCmd("count=10")
	dst := map[string]string{}
	result := cmdctx.ExtraParams(cmd, dst)
	if result["count"] != "10" {
		t.Errorf("expected count=10, got %v", result["count"])
	}
}

func TestExtraParams_MultipleParams(t *testing.T) {
	cmd := newParamCmd("count=10", "output_mode=json")
	dst := map[string]string{}
	result := cmdctx.ExtraParams(cmd, dst)
	if result["count"] != "10" {
		t.Errorf("expected count=10, got %v", result["count"])
	}
	if result["output_mode"] != "json" {
		t.Errorf("expected output_mode=json, got %v", result["output_mode"])
	}
}

func TestExtraParams_MalformedParamIgnored(t *testing.T) {
	cmd := newParamCmd("noequals")
	dst := map[string]string{}
	result := cmdctx.ExtraParams(cmd, dst)
	if _, ok := result["noequals"]; ok {
		t.Error("malformed param without '=' should be silently ignored")
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestRunDescribe_ReturnsFalseWhenNotSet(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("describe", false, "")
	cmd.SetContext(context.Background())

	c := client.New(&config.Config{Host: "https://localhost:8089", Token: "tok"})
	called, err := cmdctx.RunDescribe(cmd, c, "/data/indexes/")
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("expected RunDescribe to return false when --describe is not set")
	}
}
