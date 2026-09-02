package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// ServerCmd is the top-level server command. Registered in cmd/register_overrides.go.
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Splunk server info, health, and messages",
}

var serverInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show server version, build, GUID, and role",
	Example: `  splunkctl server info
  splunkctl server info --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/server/info/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var serverHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show server health status (fast liveness check)",
	Example: `  splunkctl server health
  splunkctl server health --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/server/health/splunkd/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var serverMessagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Show Splunk system messages and bulletin errors",
	Example: `  splunkctl server messages
  splunkctl server messages --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/messages/",
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	ServerCmd.AddCommand(serverInfoCmd, serverHealthCmd, serverMessagesCmd)
}
