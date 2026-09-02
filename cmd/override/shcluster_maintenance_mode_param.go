package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newShclusterMaintenanceModeEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: `Sets the maintenance mode on members in shclustering. Must be invoked at the captain.`,
		Example: `
  'splunkctl shcluster-maintenance-mode enable'`,
		Annotations: map[string]string{"describe_path": "/shcluster/captain/control/default/maintenance"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			params["mode"] = "1"
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/shcluster/captain/control/default/maintenance",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
}

func newShclusterMaintenanceModeDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: `Disables the maintaince mode on peers in shclustering. Must be invoked at the master.`,
		Example: `
  'splunkctl shcluster-maintenance-mode disable'`,
		Annotations: map[string]string{"describe_path": "/shcluster/captain/control/default/maintenance"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			params["mode"] = "0"
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/shcluster/captain/control/default/maintenance",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
}

// ShclusterMaintenanceModeEnableCmd overrides the generated enable to send
// "mode=1"; the /shcluster/captain/control/default/maintenance endpoint
// requires an explicit "mode" argument and rejects requests without one
// ("The following required arguments are missing: mode.").
var ShclusterMaintenanceModeEnableCmd = newShclusterMaintenanceModeEnableCmd()

// ShclusterMaintenanceModeDisableCmd overrides the generated disable to send
// "mode=0", for the same reason as ShclusterMaintenanceModeEnableCmd.
var ShclusterMaintenanceModeDisableCmd = newShclusterMaintenanceModeDisableCmd()
