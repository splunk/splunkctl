package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// IntrospectionCmd is the top-level introspection command. Registered in cmd/register_overrides.go.
var IntrospectionCmd = &cobra.Command{
	Use:   "introspection",
	Short: "Splunk pipeline introspection — queues, indexer throughput, process info",
}

var introspectionQueuesCmd = &cobra.Command{
	Use:   "queues",
	Short: "Show pipeline queue fill rates (ingestion health)",
	Example: `  splunkctl introspection queues
  splunkctl introspection queues --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/server/introspection/queues/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var introspectionIndexerCmd = &cobra.Command{
	Use:   "indexer",
	Short: "Show indexer throughput and performance stats",
	Example: `  splunkctl introspection indexer
  splunkctl introspection indexer --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/server/introspection/indexer/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var introspectionProcessorsCmd = &cobra.Command{
	Use:     "processors",
	Short:   "Show Splunk pipeline processor info",
	Example: `  splunkctl introspection processors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/server/introspection/processors/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var introspectionDiskUsageCmd = &cobra.Command{
	Use:     "disk-usage",
	Short:   "Show Splunk disk usage by index and bucket type",
	Example: `  splunkctl introspection disk-usage`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/server/introspection/disk-usage/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	IntrospectionCmd.AddCommand(introspectionQueuesCmd, introspectionIndexerCmd, introspectionProcessorsCmd, introspectionDiskUsageCmd)
}
