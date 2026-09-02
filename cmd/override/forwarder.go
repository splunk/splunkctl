package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// ForwarderCmd is the top-level forwarder command. Registered in cmd/register_overrides.go.
var ForwarderCmd = &cobra.Command{
	Use:   "forwarder",
	Short: "Inspect connected forwarders via the deployment server",
}

var forwarderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connected forwarder clients",
	Example: `  splunkctl forwarder list
  splunkctl forwarder list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/deployment/server/clients/",
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var forwarderReloadCmd = &cobra.Command{
	Use:     "reload",
	Short:   "Reload deployment server configuration",
	Example: `  splunkctl forwarder reload`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   "/deployment/server/config/_reload",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	ForwarderCmd.AddCommand(forwarderListCmd, forwarderReloadCmd)
}
