package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const firedAlertPath = "/alerts/fired_alerts/"

// FiredAlertCmd is the top-level fired-alert command. Registered in cmd/register_overrides.go.
var FiredAlertCmd = &cobra.Command{
	Use:   "fired-alert",
	Short: "View fired alerts",
}

var firedAlertListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fired alerts",
	Example: `  splunkctl fired-alert list
  splunkctl fired-alert list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   firedAlertPath,
			ID:     name,
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var firedAlertClearCmd = &cobra.Command{
	Use:     "clear <name>",
	Short:   "Clear fired alerts for a saved search",
	Example: `  splunkctl fired-alert clear "My Alert"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   firedAlertPath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "cleared": true}}, outFmt)
	},
}

func init() {
	firedAlertListCmd.Flags().String("name", "", "filter to a specific saved search name")
	firedAlertClearCmd.Flags().String("name", "", "saved search name (positional or --name)")
	FiredAlertCmd.AddCommand(firedAlertListCmd, firedAlertClearCmd)
}
