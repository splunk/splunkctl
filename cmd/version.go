package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"
var Commit = "none"
var Built = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print splunkctl version information",
	Long:  "Print version, commit hash, and build date. Use --output json for machine-readable output.",
	Annotations: map[string]string{
		annotationNoClient: "true",
	},
	Example: `  splunkctl version
  splunkctl version --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out, _ := cmd.Flags().GetString("output")
		if out == "" {
			out = flagOutput
		}
		info := map[string]string{
			"version": Version,
			"commit":  Commit,
			"built":   Built,
		}
		if out == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(info)
		}
		fmt.Printf("splunkctl %s (commit: %s, built: %s)\n", Version, Commit, Built)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
