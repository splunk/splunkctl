package override

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

// LicenseCmd is the top-level license command. Registered in cmd/register_overrides.go.
var LicenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Inspect Splunk license usage, pools, and slaves",
}

var licenseUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show daily license GB consumed (last 30 days by default)",
	Example: `  splunkctl license usage
  splunkctl license usage --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/licenser/usage/",
			Params: cmdctx.ExtraParams(cmd, map[string]string{}),
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var licensePoolsCmd = &cobra.Command{
	Use:     "pools",
	Short:   "List license pools and their quota/used bytes",
	Example: `  splunkctl license pools`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/licenser/pools/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var licenseStacksCmd = &cobra.Command{
	Use:     "stacks",
	Short:   "List license stacks",
	Example: `  splunkctl license stacks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/licenser/stacks/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var licenseSlavesCmd = &cobra.Command{
	Use:     "slaves",
	Short:   "List license slaves connected to this license master",
	Example: `  splunkctl license slaves`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/licenser/slaves/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

var licenseMessagesCmd = &cobra.Command{
	Use:     "messages",
	Short:   "Show license warning and violation messages",
	Example: `  splunkctl license messages`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdctx.ClientFrom(cmd)
		results, err := c.Do(cmd.Context(), client.Request{
			Method: "GET",
			Path:   "/licenser/messages/",
		})
		if err != nil {
			return err
		}
		outFmt, _ := cmd.Flags().GetString("output")
		return output.Print(cmd.Context(), results, outFmt)
	},
}

func init() {
	LicenseCmd.AddCommand(licenseUsageCmd, licensePoolsCmd, licenseStacksCmd, licenseSlavesCmd, licenseMessagesCmd)
}
