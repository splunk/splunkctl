package cmd_test

import (
	"testing"

	"github.com/spf13/cobra"
	_ "github.com/splunk/splunkctl/cmd"
	"github.com/splunk/splunkctl/cmd/generated"
)

func hasSubcommand(parent *cobra.Command, use string) bool {
	for _, c := range parent.Commands() {
		if c.Use == use {
			return true
		}
	}
	return false
}

func TestRegisterOverrides_IndexCleanInjected(t *testing.T) {
	if !hasSubcommand(generated.IndexCmd, "clean <name>") {
		t.Error("IndexCmd: expected override subcommand 'clean <name>' to be injected")
	}
}

func TestRegisterOverrides_UserOverridesInjected(t *testing.T) {
	for _, use := range []string{"add <username>", "edit <username>", "remove <username>", "password <username>"} {
		if !hasSubcommand(generated.UserCmd, use) {
			t.Errorf("UserCmd: expected subcommand %q to be present after override injection", use)
		}
	}
}

func TestRegisterOverrides_SchedulerShowInjected(t *testing.T) {
	if !hasSubcommand(generated.SchedulerCmd, "show") {
		t.Error("SchedulerCmd: expected override subcommand 'show' to be injected")
	}
}

func TestRegisterOverrides_GroupsDefined(t *testing.T) {
	groups := map[string]bool{}
	for _, g := range generated.AppCmd.Root().Groups() {
		groups[g.ID] = true
	}
	for _, id := range []string{"core", "cluster"} {
		if !groups[id] {
			t.Errorf("expected group %q to be registered on rootCmd", id)
		}
	}
}
