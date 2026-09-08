package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSchemaListsGlobalSafetyFlagsOnce(t *testing.T) {
	root := &cobra.Command{Use: "splunkctl"}
	root.PersistentFlags().String("host", "", "")
	root.PersistentFlags().String("output", "", "")
	root.PersistentFlags().Bool(flagYesName, false, "")
	root.PersistentFlags().Bool(flagReadOnlyName, false, "")
	leaf := &cobra.Command{Use: "list", Run: func(*cobra.Command, []string) {}}
	leaf.Flags().String("name", "", "")
	root.AddCommand(leaf)

	var entries []CommandInfo
	collectCommands(root, "", &entries)
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want root and leaf", len(entries))
	}

	globalFlags := make(map[string]FlagInfo)
	for _, flag := range entries[0].Flags {
		globalFlags[flag.Name] = flag
	}
	for _, name := range []string{flagYesName, flagReadOnlyName} {
		flag, ok := globalFlags[name]
		if !ok {
			t.Errorf("global schema entry is missing --%s", name)
		} else if flag.Type != "bool" {
			t.Errorf("--%s type = %q, want bool", name, flag.Type)
		} else if flag.Env != "" {
			t.Errorf("--%s env = %q, want none", name, flag.Env)
		}
	}
	if got := globalFlags["host"].Env; got != "SPLUNKCTL_HOST" {
		t.Errorf("--host env = %q, want SPLUNKCTL_HOST", got)
	}
	if got := globalFlags["output"].Env; got != "" {
		t.Errorf("--output env = %q, want none", got)
	}
	if len(entries[1].Flags) != 1 || entries[1].Flags[0].Name != "name" {
		t.Errorf("leaf flags = %#v, want only local --name", entries[1].Flags)
	}
	if got := len(compactEntries(entries)[0].Flags); got != len(globalFlags) {
		t.Errorf("compact global flags = %d, want %d", got, len(globalFlags))
	}
}
