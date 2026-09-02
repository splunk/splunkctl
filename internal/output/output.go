package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"golang.org/x/term"
)

// contextKey is unexported so only this package can store/retrieve the writer.
type contextKey string

const writerKey contextKey = "output_writer"

// WithWriter returns a context that uses w for all output (used in tests to capture output).
func WithWriter(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, writerKey, w)
}

// Print writes data to stdout (or the writer from context) in the requested format.
// format: "json" | "table" | "text" — if empty, auto-detects: json when not a TTY, table when TTY.
func Print(ctx context.Context, data []map[string]any, format string) error {
	w := writerFrom(ctx)

	if format == "" {
		if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
			format = "table"
		} else {
			format = "json"
		}
	}

	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	case "table":
		return printTable(w, data)
	case "text":
		for _, row := range data {
			fmt.Fprintf(w, "%v\n", row)
		}
		return nil
	default:
		return fmt.Errorf("unknown output format %q: must be json, table, or text", format)
	}
}

// PrintError writes a structured error to stderr (always JSON regardless of format).
func PrintError(err error, code string, status int) {
	fmt.Fprintf(os.Stderr, `{"error":%q,"code":%q,"status":%d}`+"\n",
		err.Error(), code, status)
}

func printTable(w io.Writer, data []map[string]any) error {
	if len(data) == 0 {
		fmt.Fprintln(w, "(no results)")
		return nil
	}

	// collect ordered keys from first row
	keys := make([]string, 0, len(data[0]))
	for k := range data[0] {
		keys = append(keys, k)
	}
	// ensure "name" is always first column if present
	sortKeysNameFirst(keys)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	// header
	for i, k := range keys {
		if i > 0 {
			fmt.Fprint(tw, "\t")
		}
		fmt.Fprint(tw, k)
	}
	fmt.Fprintln(tw)
	// rows
	for _, row := range data {
		for i, k := range keys {
			if i > 0 {
				fmt.Fprint(tw, "\t")
			}
			fmt.Fprintf(tw, "%v", row[k])
		}
		fmt.Fprintln(tw)
	}
	return tw.Flush()
}

func sortKeysNameFirst(keys []string) {
	for i, k := range keys {
		if k == "name" && i != 0 {
			keys[0], keys[i] = keys[i], keys[0]
			return
		}
	}
}

func writerFrom(ctx context.Context) io.Writer {
	if v := ctx.Value(writerKey); v != nil {
		return v.(io.Writer)
	}
	return os.Stdout
}
