package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newWorkloadRuleAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <rule_name>",
		Short: `adds a workload-rule`,
		Example: `
  splunkctl workload-rule add "my_role_rule" --predicate "role=admin OR app=search" --workload_pool "pool_a" --schedule "always_on"
  splunkctl workload-rule add "my_role_rule" --predicate "runtime>20" --action abort --schedule "every_week" --start_time "10:00" --end_time "15:00" --every_week_days "0,4,6" --user_message "The search is aborted due to long runtime"`,
		Annotations: map[string]string{"describe_path": "/workloads/rules/"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, cmd.Annotations["describe_path"]); done {
				return err
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			if v, _ := cmd.Flags().GetString("predicate"); v != "" {
				params["predicate"] = v
			}
			if v, _ := cmd.Flags().GetString("workload_pool"); v != "" {
				params["workload_pool"] = v
			}
			if v, _ := cmd.Flags().GetString("action"); v != "" {
				params["action"] = v
			}
			if v, _ := cmd.Flags().GetString("schedule"); v != "" {
				params["schedule"] = v
			}
			if v, _ := cmd.Flags().GetString("start_date"); v != "" {
				params["start_date"] = v
			}
			if v, _ := cmd.Flags().GetString("start_time"); v != "" {
				params["start_time"] = v
			}
			if v, _ := cmd.Flags().GetString("end_date"); v != "" {
				params["end_date"] = v
			}
			if v, _ := cmd.Flags().GetString("end_time"); v != "" {
				params["end_time"] = v
			}
			if v, _ := cmd.Flags().GetString("every_week_days"); v != "" {
				params["every_week_days"] = v
			}
			if v, _ := cmd.Flags().GetString("every_month_days"); v != "" {
				params["every_month_days"] = v
			}
			if v, _ := cmd.Flags().GetString("user_message"); v != "" {
				params["user_message"] = v
			}
			impliedVal, _ := cmd.Flags().GetString("rule_name")
			if impliedVal == "" && len(args) > 0 {
				impliedVal = args[0]
			}
			if impliedVal != "" {
				params["name"] = impliedVal
			}
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/workloads/rules/",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFlag, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFlag)
		},
	}
	cmd.Flags().String("predicate", "", `the logical expression of workload-rule with predicate as <type>=<value>. eg: role=admin, app=search AND (NOT index=_internal), runtime>10. Possible values of type are: app, role, user, index, runtime, search_type, search_mode, search_time_range`)
	_ = cmd.MarkFlagRequired("predicate")
	cmd.Flags().String("workload_pool", "", `the name of the workload-pool to associate the rule with, the workload-pool must be defined earlier`)
	cmd.Flags().String("action", "", `the monitoring action to perform. Possible values of type are: abort, move, alert`)
	cmd.Flags().String("schedule", "", `the schedule of the workload-rule. Possible values are: always_on, time_range, every_day, every_week, every_month`)
	cmd.Flags().String("start_date", "", `the start date of the validation period`)
	cmd.Flags().String("start_time", "", `the start time of the validation period`)
	cmd.Flags().String("end_date", "", `the end date of the validation period`)
	cmd.Flags().String("end_time", "", `the end time of the validation period`)
	cmd.Flags().String("every_week_days", "", `the recurring days in a week`)
	cmd.Flags().String("every_month_days", "", `the recurring days in a month`)
	cmd.Flags().String("user_message", "", `the message shown in the search job inspector if the rule is applied to a search`)
	cmd.Flags().String("rule_name", "", "rule_name (positional or --rule_name)")
	return cmd
}

// WorkloadRuleAddCmd overrides the generated add to send the rule identifier as
// "name", which is what the /workloads/rules/ REST endpoint requires; the
// generated version sent "rule_name", which Splunk rejects with
// `Argument "rule_name" is not supported by this handler.`
var WorkloadRuleAddCmd = newWorkloadRuleAddCmd()
