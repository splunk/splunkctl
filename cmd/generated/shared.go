// Package generated contains Cobra commands auto-generated from splunkrc_cmds.xml.
// Do not edit generated files — run the codegen binary to regenerate.
package generated

import (
	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
)

// clientFrom retrieves the client from the command context.
// It delegates to cmdctx.ClientFrom to avoid an import cycle between
// the generated package and the cmd package.
func clientFrom(cmd *cobra.Command) *client.Client {
	return cmdctx.ClientFrom(cmd)
}
