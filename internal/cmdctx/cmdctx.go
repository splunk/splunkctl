// Package cmdctx lets different command packages retrieve the same REST client
// and write check from a command's context. This keeps the packages from having
// to import each other.
package cmdctx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
)

type key string

const (
	clientKey     key = "client"
	writeCheckKey key = "write_check"
)

// WithClient stores c in the context and returns the new context.
func WithClient(ctx context.Context, c *client.Client) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

// WithWriteCheck attaches the write-check function to the command's context.
// Local-file commands can retrieve it later even though they do not use the
// REST client.
func WithWriteCheck(ctx context.Context, check client.WriteCheck) context.Context {
	return context.WithValue(ctx, writeCheckKey, check)
}

// CheckWrite retrieves and runs the write-check function before a command changes
// a local file. If no function was attached, return an error instead of allowing
// the write without protection.
func CheckWrite(ctx context.Context) error {
	check, _ := ctx.Value(writeCheckKey).(client.WriteCheck)
	if check == nil {
		return fmt.Errorf("write safety check not configured")
	}
	return check(ctx)
}

// ClientFrom retrieves the client from the command context.
// Panics with a descriptive message if the client was not initialized.
func ClientFrom(cmd *cobra.Command) *client.Client {
	v := cmd.Context().Value(clientKey)
	if v == nil {
		panic("ClientFrom called on command that skipped client initialization (check no_client annotation)")
	}
	return v.(*client.Client)
}

// RunDescribe fetches and prints available --param keys for the given path if --describe is set.
// Returns (true, nil) if --describe was requested (caller should return immediately),
// (false, nil) otherwise.
func RunDescribe(cmd *cobra.Command, c *client.Client, path string) (bool, error) {
	describe, err := cmd.Flags().GetBool("describe")
	if err != nil || !describe {
		return false, nil
	}
	fields, err := c.Describe(cmd.Context(), path)
	if err != nil {
		return true, fmt.Errorf("describe: %w", err)
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return true, enc.Encode(fields)
}

// ExtraParams returns any key=value pairs passed via --param flags, merged into dst.
// dst is modified in place and returned for convenience.
func ExtraParams(cmd *cobra.Command, dst map[string]string) map[string]string {
	vals, err := cmd.Flags().GetStringArray("param")
	if err != nil || len(vals) == 0 {
		return dst
	}
	for _, kv := range vals {
		k, v, ok := strings.Cut(kv, "=")
		if ok && k != "" {
			dst[k] = v
		}
	}
	return dst
}
