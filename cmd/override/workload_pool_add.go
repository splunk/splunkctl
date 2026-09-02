package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newWorkloadPoolAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <pool_name>",
		Short: `adds a workload-pool`,
		Example: `
  splunkctl workload-pool add pool_a --category search --cpu_weight 30 --mem_weight 20`,
		Annotations: map[string]string{"describe_path": "/workloads/pools/"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			if v, _ := cmd.Flags().GetString("cpu_weight"); v != "" {
				params["cpu_weight"] = v
			}
			if v, _ := cmd.Flags().GetString("mem_weight"); v != "" {
				params["mem_weight"] = v
			}
			if v, _ := cmd.Flags().GetString("category"); v != "" {
				params["category"] = v
			}
			if v, _ := cmd.Flags().GetString("default_category_pool"); v != "" {
				params["default_category_pool"] = v
			}
			impliedVal, _ := cmd.Flags().GetString("pool_name")
			if impliedVal == "" && len(args) > 0 {
				impliedVal = args[0]
			}
			if impliedVal != "" {
				params["name"] = impliedVal
			}
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/workloads/pools/",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
	cmd.Flags().String("cpu_weight", "", `The cpu weight to be used for this pool`)
	_ = cmd.MarkFlagRequired("cpu_weight")
	cmd.Flags().String("mem_weight", "", `The memory weight to be used for this pool`)
	_ = cmd.MarkFlagRequired("mem_weight")
	cmd.Flags().String("category", "", `The memory weight to be used for this pool`)
	_ = cmd.MarkFlagRequired("category")
	cmd.Flags().String("default_category_pool", "", `Mark this pool as the default category pool`)
	cmd.Flags().String("pool_name", "", "pool_name (positional or --pool_name)")
	return cmd
}

// WorkloadPoolAddCmd overrides the generated add to send the pool identifier as
// "name", which is what the /workloads/pools/ REST endpoint requires; the
// generated version sent "pool_name", which Splunk rejects with
// `Cannot perform action "POST" without a target name to act on.`
var WorkloadPoolAddCmd = newWorkloadPoolAddCmd()
