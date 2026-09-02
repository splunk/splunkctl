package cmd

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/cmd/generated"
	"github.com/splunk/splunkctl/cmd/override"
)

func init() {
	// Define the groups shown in --help output.
	rootCmd.AddGroup(
		&cobra.Group{ID: "core", Title: "Core Commands:"},
		&cobra.Group{ID: "cluster", Title: "Clustering Commands:"},
	)

	// ── Core commands ─────────────────────────────────────────────────────────
	// These appear in --help. Advanced/low-level commands are hidden; use
	// 'splunkctl schema' or 'splunkctl <command> --help' to explore them.
	for _, c := range []*cobra.Command{
		override.SearchCmd,
		override.JobsCmd,
		override.SavedSearchCmd,
		override.HecCmd,
		override.DashboardCmd,
		override.LookupCmd,
		override.LookupFileCmd,
		override.MacroCmd,
		override.EventtypeCmd,
		override.FieldExtractionCmd,
		override.DataModelCmd,
		override.ScriptedInputCmd,
		override.FiredAlertCmd,
		override.CredentialCmd,
		override.WhoamiCmd,
		override.ServerCmd,
		override.IntrospectionCmd,
		override.LicenseCmd,
		override.KvCmd,
		override.ConfCmd,
		override.ForwarderCmd,
		override.ServerclassCmd,
		override.DistSearchCmd,
		override.AclCmd,
		override.SkillCmd,
		override.RawCmd,
	} {
		c.GroupID = "core"
		rootCmd.AddCommand(c)
	}

	// Inject override subcommands into generated parent commands.
	// The generated parents live in the "advanced" group (registered in register_generated.go).
	generated.IndexCmd.AddCommand(override.IndexCleanCmd)
	// Replace generated app install with the override that sends filename=true for file paths.
	generated.AppCmd.RemoveCommand(generated.AppInstallCmd)
	generated.AppCmd.AddCommand(override.AppInstallCmd)
	// Replace generated user subcommands with overrides.
	generated.UserCmd.RemoveCommand(generated.UserAddCmd)
	generated.UserCmd.AddCommand(override.UserAddCmd)
	generated.UserCmd.RemoveCommand(generated.UserEditCmd)
	generated.UserCmd.AddCommand(override.UserEditCmd)
	generated.UserCmd.RemoveCommand(generated.UserRemoveCmd)
	generated.UserCmd.AddCommand(override.UserRemoveCmd)
	generated.UserCmd.AddCommand(override.UserPasswordCmd)
	generated.SchedulerCmd.AddCommand(override.SchedulerShowCmd)
	// Replace generated workload-rule add with the override that sends "name"
	// instead of "rule_name", which the /workloads/rules/ REST endpoint requires.
	generated.WorkloadRuleCmd.RemoveCommand(generated.WorkloadRuleAddCmd)
	generated.WorkloadRuleCmd.AddCommand(override.WorkloadRuleAddCmd)
	// Replace generated workload-pool add with the override that sends "name"
	// instead of "pool_name", which the /workloads/pools/ REST endpoint requires.
	generated.WorkloadPoolCmd.RemoveCommand(generated.WorkloadPoolAddCmd)
	generated.WorkloadPoolCmd.AddCommand(override.WorkloadPoolAddCmd)
	// Replace generated workload-config get-base-dirname/check with overrides
	// that use GET instead of POST; both endpoints silently return an empty
	// entry list on POST and only return data on GET.
	generated.WorkloadConfigCmd.RemoveCommand(generated.WorkloadConfigGetBaseDirnameCmd)
	generated.WorkloadConfigCmd.AddCommand(override.WorkloadConfigGetBaseDirnameCmd)
	generated.WorkloadConfigCmd.RemoveCommand(generated.WorkloadConfigCheckCmd)
	generated.WorkloadConfigCmd.AddCommand(override.WorkloadConfigCheckCmd)
	// Replace generated shcluster-maintenance-mode enable/disable with overrides
	// that send the required "mode" param (1/0); the generated versions send an
	// empty body and Splunk rejects them with "required arguments are missing: mode".
	generated.ShclusterMaintenanceModeCmd.RemoveCommand(generated.ShclusterMaintenanceModeEnableCmd)
	generated.ShclusterMaintenanceModeCmd.AddCommand(override.ShclusterMaintenanceModeEnableCmd)
	generated.ShclusterMaintenanceModeCmd.RemoveCommand(generated.ShclusterMaintenanceModeDisableCmd)
	generated.ShclusterMaintenanceModeCmd.AddCommand(override.ShclusterMaintenanceModeDisableCmd)
	// Replace generated cluster-search-heads remove with the override that
	// uses POST instead of DELETE; the remove_search_heads custom action requires POST.
	generated.ClusterSearchHeadsCmd.RemoveCommand(generated.ClusterSearchHeadsRemoveCmd)
	generated.ClusterSearchHeadsCmd.AddCommand(override.ClusterSearchHeadsRemoveCmd)
	// Replace generated maintenance-mode enable/disable with overrides that
	// send the required "mode" param (1/0); the generated versions send an
	// empty body and Splunk rejects them with "required arguments are missing: mode".
	generated.MaintenanceModeCmd.RemoveCommand(generated.MaintenanceModeEnableCmd)
	generated.MaintenanceModeCmd.AddCommand(override.MaintenanceModeEnableCmd)
	generated.MaintenanceModeCmd.RemoveCommand(generated.MaintenanceModeDisableCmd)
	generated.MaintenanceModeCmd.AddCommand(override.MaintenanceModeDisableCmd)
	// Replace generated shcluster-replicated-config resync with the override
	// that sends "resync_destructive=true"; the endpoint requires push_to or
	// resync_destructive and the generated version sends neither.
	generated.ShclusterReplicatedConfigCmd.RemoveCommand(generated.ShclusterReplicatedConfigResyncCmd)
	generated.ShclusterReplicatedConfigCmd.AddCommand(override.ShclusterReplicatedConfigResyncCmd)
}
