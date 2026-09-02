//go:generate go run ./codegen/ ../../cfg/bundles/default/static/splunkrc_cmds.xml ./codegen/templates/command.go.tmpl ./cmd/generated/

package main

import "github.com/splunk/splunkctl/cmd"

func main() {
	cmd.Execute()
}
