package override

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

const lookupFilePath = "/data/lookup-table-files/"

// LookupFileCmd is the top-level lookup-file command. Registered in cmd/register_overrides.go.
var LookupFileCmd = &cobra.Command{
	Use:   "lookup-file",
	Short: "Manage lookup table CSV files",
}

var lookupFileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List lookup table files",
	Example: `  splunkctl lookup-file list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   lookupFilePath,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var lookupFileUploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a lookup table CSV file",
	Long: `Upload a lookup table CSV file to Splunk.

Requires SPLUNK_HOME to be set (the Splunk server must be local). The file is
first staged in $SPLUNK_HOME/var/run/splunk/lookup_tmp, then registered via REST.`,
	Example: `  splunkctl lookup-file upload ./my_lookup.csv
  splunkctl lookup-file upload ./my_lookup.csv --name custom_name.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" && len(args) > 0 {
			filePath = args[0]
		}
		if filePath == "" {
			return fmt.Errorf("file path required (positional or --file)")
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			name = filepath.Base(filePath)
		}

		splunkHome := os.Getenv("SPLUNK_HOME")
		if splunkHome == "" {
			return fmt.Errorf("SPLUNK_HOME environment variable is not set; lookup file upload requires local Splunk access")
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}

		// Stage the file in Splunk's lookup_tmp directory.
		stagingDir := filepath.Join(splunkHome, "var", "run", "splunk", "lookup_tmp")
		if err := os.MkdirAll(stagingDir, 0755); err != nil {
			return fmt.Errorf("creating staging dir: %w", err)
		}
		stagingPath := filepath.Join(stagingDir, name)
		if err := os.WriteFile(stagingPath, data, 0644); err != nil {
			return fmt.Errorf("staging file: %w", err)
		}

		// Register the staged file with Splunk.
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "POST",
			Path:   lookupFilePath,
			Params: map[string]string{
				"name":     name,
				"eai:data": stagingPath,
			},
		})
		if err != nil {
			_ = os.Remove(stagingPath)
			return err
		}
		if len(results) == 0 {
			results = []map[string]any{{"name": name, "uploaded": true}}
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var lookupFileRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Delete a lookup table file",
	Example: `  splunkctl lookup-file remove my_lookup.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		name, _ := cmd.Flags().GetString("name")
		if name == "" && len(args) > 0 {
			name = args[0]
		}
		_, err := c.Do(cmd.Context(), client.Request{
			Method: "DELETE",
			Path:   lookupFilePath,
			ID:     name,
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), []map[string]any{{"name": name, "deleted": true}}, outFmt)
	},
}

func init() {
	lookupFileUploadCmd.Flags().String("file", "", "path to CSV file (positional or --file)")
	lookupFileUploadCmd.Flags().String("name", "", "destination filename (default: basename of file)")
	lookupFileRemoveCmd.Flags().String("name", "", "filename to delete (positional or --name)")
	LookupFileCmd.AddCommand(lookupFileListCmd, lookupFileUploadCmd, lookupFileRemoveCmd)
}
