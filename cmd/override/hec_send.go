package override

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/cmdctx"
	"github.com/splunk/splunkctl/internal/config"
)

func init() {
	hecSendCmd.Flags().String("event", "", "event string or JSON to send (required)")
	hecSendCmd.Flags().String("index", "", "target index")
	hecSendCmd.Flags().String("sourcetype", "", "sourcetype")
	hecSendCmd.Flags().String("source", "", "source")
	hecSendCmd.Flags().String("hec-token", "", "HEC token (overrides --token for the collector endpoint)")
	hecSendCmd.Flags().String("hec-port", "8088", "HEC port (default 8088)")
	_ = hecSendCmd.MarkFlagRequired("event")
	HecCmd.AddCommand(hecSendCmd)
}

var hecSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a test event via HEC (/services/collector)",
	Example: `  splunkctl hec send --event "hello world" --index main
  splunkctl hec send --event '{"msg":"test","level":"info"}' --sourcetype json --hec-token abc123`,
	RunE: func(cmd *cobra.Command, args []string) error {
		eventStr, _ := cmd.Flags().GetString("event")
		index, _ := cmd.Flags().GetString("index")
		sourcetype, _ := cmd.Flags().GetString("sourcetype")
		source, _ := cmd.Flags().GetString("source")
		hecToken, _ := cmd.Flags().GetString("hec-token")
		hecPort, _ := cmd.Flags().GetString("hec-port")

		flagHost, _ := cmd.Flags().GetString("host")
		flagToken, _ := cmd.Flags().GetString("token")
		flagInsecure, _ := cmd.Flags().GetBool("insecure")
		cfg, err := config.Resolve(flagHost, flagToken, flagInsecure, "")
		if err != nil {
			return err
		}

		// replace port in host URL with hec-port
		host := cfg.Host
		if idx := strings.LastIndex(host, ":"); idx > 8 {
			host = host[:idx] + ":" + hecPort
		}
		hecURL := strings.TrimRight(host, "/") + "/services/collector/event"

		token := hecToken
		if token == "" {
			token = cfg.Token
		}

		payload := map[string]any{}
		var eventJSON any
		if json.Unmarshal([]byte(eventStr), &eventJSON) == nil {
			payload["event"] = eventJSON
		} else {
			payload["event"] = eventStr
		}
		if index != "" {
			payload["index"] = index
		}
		if sourcetype != "" {
			payload["sourcetype"] = sourcetype
		}
		if source != "" {
			payload["source"] = source
		}

		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(cmd.Context(), "POST", hecURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Splunk "+token)
		req.Header.Set("Content-Type", "application/json")

		// HEC needs this custom request because it uses a different port and token.
		// DoHTTP still runs the shared write check before sending the POST.
		resp, err := cmdctx.ClientFrom(cmd).DoHTTP(req)
		if err != nil {
			return fmt.Errorf("hec send: %w", err)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode >= 400 {
			return fmt.Errorf("hec send failed (%d): %s", resp.StatusCode, string(respBody))
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(respBody))
		return nil
	},
}
