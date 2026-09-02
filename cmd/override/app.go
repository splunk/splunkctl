package override

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/output"
)

func newInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <name>",
		Short: "install specified app name",
		Example: `
  splunkctl app install foo.tgz
  splunkctl app install foo.tar.gz
  splunkctl app install foo.spl
  splunkctl app install /home/username/sos_1.0.tgz --update 1`,
		Annotations: map[string]string{"describe_path": "/apps/local/"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cmdctx.ClientFrom(cmd)
			if done, err := cmdctx.RunDescribe(cmd, c, "/apps/local/"); done {
				return err
			}
			name, _ := cmd.Flags().GetString("name")
			if name == "" && len(args) > 0 {
				name = args[0]
			}
			params := cmdctx.ExtraParams(cmd, map[string]string{})
			if name != "" {
				params["name"] = name
			}
			if isFilePath(name) {
				params["filename"] = "true"
			}
			if v, _ := cmd.Flags().GetString("update"); v != "" {
				params["update"] = v
			}
			results, err := c.Do(cmd.Context(), client.Request{
				Method: "POST",
				Path:   "/apps/local/",
				Params: params,
			})
			if err != nil {
				return err
			}
			outFmt, _ := cmd.Flags().GetString("output")
			return output.Print(cmd.Context(), results, outFmt)
		},
	}
	cmd.Flags().String("name", "", "name (positional or --name)")
	cmd.Flags().String("update", "", "update specified app during install if it already exists. Defaults to false or 0.")
	return cmd
}

// AppInstallCmd overrides the generated install to send filename=true when the
// argument is a file path rather than an app-store name.
var AppInstallCmd = newInstallCmd()

// isFilePath reports whether name refers to a local file rather than an app store name.
// Splunk requires filename=true when the name is a file path on the Splunk server.
func isFilePath(name string) bool {
	if strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://") {
		return false
	}
	if strings.ContainsAny(name, `/\`) {
		return true
	}
	for _, ext := range []string{".tgz", ".spl", ".tar.gz", ".tar"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
