package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
)

const annotationNoClient = "no_client"

var (
	flagHost     string
	flagToken    string
	flagOutput   string
	flagInsecure bool
)

var rootCmd = &cobra.Command{
	Use:   "splunkctl",
	Short: "Lightweight CLI for managing Splunk via REST API",
	Long: `splunkctl is an agent-native CLI for Splunk.

Getting started:
  1. Configure your Splunk connection:
       splunkctl config set host https://splunk.example.com:8089
       splunkctl config set token <your-api-token>
     Get a token: Splunk Web → Settings → Tokens → New Token

  2. Install the agent skill so Claude or Codex knows how to use this CLI:
       splunkctl skill install --interactive

  3. Start managing Splunk:
       splunkctl server health
       splunkctl index list
       splunkctl search 'index=main error'

Commands follow the pattern: splunkctl <object> <verb>

Low-level and rarely-used commands are hidden from this help output.
To see the full command tree: splunkctl schema
To use a hidden command directly: splunkctl <command> --help

Connection (flag → env var → ~/.splunkctl/config.yaml):
  --host   or SPLUNKCTL_HOST   Splunk REST endpoint (e.g. https://splunk.example.com:8089)
  --token  or SPLUNKCTL_TOKEN  Splunk API bearer token`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// The command-line flags have been read by this point. Create one function
		// that will allow or reject writes using the --yes and --read-only values.
		writeCheck, err := newWriteCheck(cmd)
		if err != nil {
			return err
		}

		// Commands such as "config set" and "skill install" write local files and
		// never call the REST client. Store the function on the command so those
		// commands can call it directly before changing a file.
		cmd.SetContext(cmdctx.WithWriteCheck(cmd.Context(), writeCheck))

		if cmd.Annotations[annotationNoClient] == "true" {
			return nil
		}
		cfg, err := config.Resolve(flagHost, flagToken, flagInsecure, flagOutput)
		if err != nil {
			return err
		}

		// Give the same function to the REST client. Before the client sends a
		// POST, PUT, PATCH, DELETE, or other non-read request, it calls this function.
		c := client.NewWithWriteCheck(cfg, writeCheck)
		cmd.SetContext(cmdctx.WithClient(cmd.Context(), c))
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		code, exitCode := errorToCode(err)
		fmt.Fprintf(os.Stderr, "{\"error\":%q,\"code\":%q}\n", err.Error(), code)
		os.Exit(exitCode)
	}
}

// errorToCode maps an error to a (code string, exit code int) pair.
// CLI exit codes: 0=success; errorToCode emits 1=general/usage fallback,
// 3=not found, 4=auth error, and 5=connection error. 2 is reserved and unused.
func errorToCode(err error) (string, int) {
	var safetyErr *writeSafetyError
	if errors.As(err, &safetyErr) {
		return safetyErr.code, 1
	}
	var splunkErr *client.SplunkError
	if errors.As(err, &splunkErr) {
		switch splunkErr.Code {
		case "NOT_FOUND":
			return "NOT_FOUND", 3
		case "AUTH_ERROR":
			return "AUTH_ERROR", 4
		}
		return splunkErr.Code, 1
	}
	if strings.Contains(err.Error(), "connection error") {
		return "CONNECTION_ERROR", 5
	}
	return "USAGE_ERROR", 1
}

// ClientFrom retrieves the initialized Splunk REST client from the command context.
// For hand-written commands in this package. Generated commands use cmdctx.ClientFrom directly.
func ClientFrom(cmd *cobra.Command) *client.Client {
	return cmdctx.ClientFrom(cmd)
}

func init() {
	config.Init()

	rootCmd.PersistentFlags().StringVar(&flagHost, "host", "", "Splunk REST endpoint (env: SPLUNKCTL_HOST)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Splunk API bearer token (env: SPLUNKCTL_TOKEN)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: json|table|text (default: table on TTY, json otherwise)")
	rootCmd.PersistentFlags().BoolVar(&flagInsecure, "insecure", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().StringArray("param", nil, "Extra Splunk API parameter as key=value (repeatable, e.g. --param frozenTimePeriodInSecs=2592000)")
	rootCmd.PersistentFlags().Bool("describe", false, "Show available --param keys for this command (required and optional) then exit")
	rootCmd.PersistentFlags().Bool(flagYesName, false, "Skip write confirmation")
	rootCmd.PersistentFlags().Bool(flagReadOnlyName, false, "Block operations that can modify state")
}
