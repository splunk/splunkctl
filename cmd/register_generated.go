package cmd

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/cmd/generated"
)

func init() {
	// Generated commands that belong in the core group alongside override commands.
	for _, c := range []*cobra.Command{
		generated.AppCmd,
		generated.UserCmd,
		generated.RoleCmd,
		generated.IndexCmd,
	} {
		c.GroupID = "core"
		rootCmd.AddCommand(c)
	}

	// Generated commands hidden from --help — superseded by override commands or low value.
	// Still fully functional: splunkctl <command> --help
	for _, c := range []*cobra.Command{
		generated.HealthCmd,        // use: splunkctl server health
		generated.MonitorCmd,       // use: splunkctl scripted-input
		generated.ForwardServerCmd, // use: splunkctl forwarder
		generated.TcpCmd,
		generated.UdpCmd,
		generated.KvstoreCmd, // use: splunkctl kv
		generated.KvstoreStatusCmd,
		generated.SchedulerCmd, // use: splunkctl scheduler show
		generated.SchedulerStatusCmd,
		generated.DeployServerCmd, // use: splunkctl serverclass
		generated.DeployClientCmd,
		generated.DeployPollCmd,
	} {
		c.Hidden = true
		rootCmd.AddCommand(c)
	}

	// Clustering commands — visible under "Clustering Commands:" in --help.
	for _, c := range []*cobra.Command{
		generated.ClusterConfigCmd,
		generated.ClusterManagerCmd,
		generated.ClusterPeersCmd,
		generated.ClusterAllPeersCmd,
		generated.ClusterBucketsCmd,
		generated.ClusterBundleCmd,
		generated.ClusterDataCmd,
		generated.ClusterGenerationCmd,
		generated.ClusterRecoveryCmd,
		generated.ClusterSearchHeadsCmd,
		generated.ClusterStatusCmd,
		generated.ClusterUsageCmd,
		generated.ShclusterConfigCmd,
		generated.ShclusterCaptainCmd,
		generated.ShclusterCaptainInfoCmd,
		generated.ShclusterMembersCmd,
		generated.ShclusterMemberCmd,
		generated.ShclusterMemberInfoCmd,
		generated.ShclusterStatusCmd,
		generated.ShclusterBundleCmd,
		generated.ShclusterArtifactsCmd,
		generated.ShclusterMemberArtifactsCmd,
		generated.ShclusterMaintenanceModeCmd,
		generated.ShclusterIndexingReadyCmd,
		generated.ShclusterKvmigrationStatusCmd,
		generated.ShclusterKvupgradeStatusCmd,
		generated.ShclusterReplicatedConfigCmd,
		generated.ShclusterSchedulerCacheCmd,
		generated.ShclusterSchedulerJobsCmd,
		generated.ShclusterSplunkSecretCmd,
		generated.ShclusterConfigurationSetCmd,
		generated.ShclusterCommonEncryptCmd,
		generated.ShcStatusCmd,
		generated.ShcRecoveryCmd,
		generated.ShcUpgradeCmd,
		generated.PeerInfoCmd,
		generated.PeerBucketsCmd,
		generated.IndexerDiscoveryCmd,
		generated.ManagerInfoCmd,
		generated.MaintenanceModeCmd,
		generated.IndexingReadyCmd,
		generated.SearchAdmissionControlCmd,
		generated.WorkloadManagementCmd,
		generated.WorkloadManagementStatusCmd,
		generated.WorkloadPoolCmd,
		generated.WorkloadPolicyCmd,
		generated.WorkloadRuleCmd,
		generated.WorkloadCategoryCmd,
		generated.WorkloadConfigCmd,
	} {
		c.GroupID = "cluster"
		rootCmd.AddCommand(c)
	}

	// Advanced commands — hidden from --help but fully functional.
	// Discover with: splunkctl schema   or   splunkctl <command> --help
	for _, c := range []*cobra.Command{
		generated.AdCmd,
		generated.AppserverPortsCmd,
		generated.BundleReplicationConfigCmd,
		generated.BundleReplicationStatusCmd,
		generated.CascadePlansCmd,
		generated.CascadePlanStatusCmd,
		generated.CrlCmd,
		generated.DatastoreDirCmd,
		generated.DefaultHostnameCmd,
		generated.DfsmasterPortCmd,
		generated.EventlogCmd,
		generated.ExcessBucketsCmd,
		generated.ExecCmd,
		generated.FipsModeCmd,
		generated.GuidCmd,
		generated.InputstatusCmd,
		generated.KvstoreMaintenanceModeCmd,
		generated.KvstorePortCmd,
		generated.LicenserGroupsCmd,
		generated.LicenserLocalpeerCmd,
		generated.LicenserMessagesCmd,
		generated.LicenserPeersCmd,
		generated.LicenserPoolsCmd,
		generated.LicenserStacksCmd,
		generated.LicensesCmd,
		generated.ListenCmd,
		generated.LogLevelCmd,
		generated.MinfreembCmd,
		generated.MonitornohandleCmd,
		generated.OnCmd,
		generated.OneshotCmd,
		generated.PerfmonCmd,
		generated.PostgresCmd,
		generated.PostgresHealthCmd,
		generated.PostgresRecoveryStatusCmd,
		generated.PostgresStatusCmd,
		generated.RegistryCmd,
		generated.RemoteInputQueueConfigCmd,
		generated.RemoteInputQueueStatusCmd,
		generated.RemoteOutputQueueConfigCmd,
		generated.RemoteOutputQueueStatusCmd,
		generated.ServernameCmd,
		generated.SplunkdPortCmd,
		generated.SplunkSecretCmd,
		generated.StandaloneCmd,
		generated.StandaloneKvupgradeStatusCmd,
		generated.WebPortCmd,
		generated.WebserverCmd,
		generated.WebSslCmd,
		generated.WinhostmonCmd,
		generated.WinnetmonCmd,
		generated.WinprintmonCmd,
		generated.WmiCmd,
	} {
		c.Hidden = true
		rootCmd.AddCommand(c)
	}
}
