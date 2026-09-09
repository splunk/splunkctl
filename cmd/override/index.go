package override

import (
	"errors"

	"github.com/spf13/cobra"
)

// IndexCleanCmd is injected into the generated IndexCmd in cmd/register_overrides.go.
var IndexCleanCmd = &cobra.Command{
	Use:     "clean <name>",
	Short:   "Purge all events from an index (requires server-side CLI access)",
	Example: `  splunkctl index clean main`,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		return errors.New(
			"index clean is not available via the REST API.\n" +
				"Run this directly on the Splunk server:\n\n" +
				"  splunk stop\n" +
				"  splunk clean eventdata -index " + name + "\n" +
				"  splunk start",
		)
	},
}

func init() {
	IndexCleanCmd.Flags().String("name", "", "index name (positional or --name)")

	// Hidden from --help — not available via the REST API by design (see RunE above).
	IndexCleanCmd.Hidden = true
}
