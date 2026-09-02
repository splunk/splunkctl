package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newWorkloadConfigGetBaseDirnameCmd() *cobra.Command {
	return &cobra.Command{
		Use:         "get-base-dirname",
		Short:       `get the base dir name for splunk workload pools`,
		Annotations: map[string]string{"describe_path": "/workloads/config/get-base-dirname"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "GET",
				Path:   "/workloads/config/get-base-dirname",
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

func newWorkloadConfigCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: `Preflight checks of workload management.`,
		Example: `
  'splunkctl workload-config check'`,
		Annotations: map[string]string{"describe_path": "/workloads/config/preflight-checks"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "GET",
				Path:   "/workloads/config/preflight-checks",
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

// WorkloadConfigGetBaseDirnameCmd overrides the generated command to use GET
// instead of POST; the /workloads/config/get-base-dirname endpoint only
// returns data on GET and silently returns an empty entry list on POST.
var WorkloadConfigGetBaseDirnameCmd = newWorkloadConfigGetBaseDirnameCmd()

// WorkloadConfigCheckCmd overrides the generated command to use GET instead
// of POST; the /workloads/config/preflight-checks endpoint only returns data
// on GET and silently returns an empty entry list on POST.
var WorkloadConfigCheckCmd = newWorkloadConfigCheckCmd()
