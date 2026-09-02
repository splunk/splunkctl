package cmd_test

import (
	"testing"

	_ "github.com/splunk/splunkctl/cmd"
	"github.com/splunk/splunkctl/cmd/generated"
)

func TestRegisterGenerated_CoreGroupID(t *testing.T) {
	for _, tc := range []struct {
		name string
		use  string
	}{
		{"AppCmd", generated.AppCmd.Use},
		{"UserCmd", generated.UserCmd.Use},
		{"RoleCmd", generated.RoleCmd.Use},
		{"IndexCmd", generated.IndexCmd.Use},
	} {
		_ = tc.use
	}

	if generated.AppCmd.GroupID != "core" {
		t.Errorf("AppCmd: expected GroupID=core, got %q", generated.AppCmd.GroupID)
	}
	if generated.UserCmd.GroupID != "core" {
		t.Errorf("UserCmd: expected GroupID=core, got %q", generated.UserCmd.GroupID)
	}
	if generated.RoleCmd.GroupID != "core" {
		t.Errorf("RoleCmd: expected GroupID=core, got %q", generated.RoleCmd.GroupID)
	}
	if generated.IndexCmd.GroupID != "core" {
		t.Errorf("IndexCmd: expected GroupID=core, got %q", generated.IndexCmd.GroupID)
	}
}

func TestRegisterGenerated_HiddenCommands(t *testing.T) {
	hidden := []struct {
		name string
		got  bool
	}{
		{"HealthCmd", generated.HealthCmd.Hidden},
		{"MonitorCmd", generated.MonitorCmd.Hidden},
		{"ForwardServerCmd", generated.ForwardServerCmd.Hidden},
		{"TcpCmd", generated.TcpCmd.Hidden},
		{"UdpCmd", generated.UdpCmd.Hidden},
		{"KvstoreCmd", generated.KvstoreCmd.Hidden},
		{"KvstoreStatusCmd", generated.KvstoreStatusCmd.Hidden},
		{"SchedulerCmd", generated.SchedulerCmd.Hidden},
		{"SchedulerStatusCmd", generated.SchedulerStatusCmd.Hidden},
		{"DeployServerCmd", generated.DeployServerCmd.Hidden},
		{"DeployClientCmd", generated.DeployClientCmd.Hidden},
		{"DeployPollCmd", generated.DeployPollCmd.Hidden},
	}
	for _, tc := range hidden {
		if !tc.got {
			t.Errorf("%s: expected Hidden=true", tc.name)
		}
	}
}

func TestRegisterGenerated_ClusterGroupID(t *testing.T) {
	cluster := []struct {
		name    string
		groupID string
	}{
		{"ClusterConfigCmd", generated.ClusterConfigCmd.GroupID},
		{"ClusterManagerCmd", generated.ClusterManagerCmd.GroupID},
		{"ClusterPeersCmd", generated.ClusterPeersCmd.GroupID},
		{"ShclusterStatusCmd", generated.ShclusterStatusCmd.GroupID},
		{"ShclusterCaptainCmd", generated.ShclusterCaptainCmd.GroupID},
		{"PeerInfoCmd", generated.PeerInfoCmd.GroupID},
		{"MaintenanceModeCmd", generated.MaintenanceModeCmd.GroupID},
		{"WorkloadManagementCmd", generated.WorkloadManagementCmd.GroupID},
	}
	for _, tc := range cluster {
		if tc.groupID != "cluster" {
			t.Errorf("%s: expected GroupID=cluster, got %q", tc.name, tc.groupID)
		}
	}
}

func TestRegisterGenerated_AdvancedCommandsHidden(t *testing.T) {
	advanced := []struct {
		name string
		got  bool
	}{
		{"AdCmd", generated.AdCmd.Hidden},
		{"GuidCmd", generated.GuidCmd.Hidden},
		{"FipsModeCmd", generated.FipsModeCmd.Hidden},
		{"SplunkdPortCmd", generated.SplunkdPortCmd.Hidden},
		{"WebPortCmd", generated.WebPortCmd.Hidden},
	}
	for _, tc := range advanced {
		if !tc.got {
			t.Errorf("%s: expected Hidden=true", tc.name)
		}
	}
}
