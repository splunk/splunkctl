package output_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/output"
)

func TestPrint_JSON(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)

	data := []map[string]any{
		{"name": "main", "maxDataSizeMB": 500},
	}
	if err := output.Print(ctx, data, "json"); err != nil {
		t.Fatal(err)
	}

	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nGot: %s", err, buf.String())
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0]["name"] != "main" {
		t.Errorf("expected name=main, got %v", result[0]["name"])
	}
}

func TestPrint_Table(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)

	data := []map[string]any{
		{"name": "main", "disabled": false},
	}
	if err := output.Print(ctx, data, "table"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "name") {
		t.Errorf("expected 'name' column header, got: %s", out)
	}
	if !strings.Contains(out, "main") {
		t.Errorf("expected 'main' in table output, got: %s", out)
	}
}

func TestPrint_TableEmpty(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)

	if err := output.Print(ctx, []map[string]any{}, "table"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "no results") {
		t.Errorf("expected 'no results' for empty data, got: %s", buf.String())
	}
}

func TestPrint_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	err := output.Print(ctx, []map[string]any{}, "xml")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestPrint_NameColumnFirst(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)

	data := []map[string]any{
		{"zzz": "last", "name": "first", "aaa": "middle"},
	}
	if err := output.Print(ctx, data, "table"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	firstLine := strings.SplitN(out, "\n", 2)[0]
	if !strings.HasPrefix(strings.TrimSpace(firstLine), "name") {
		t.Errorf("expected 'name' to be first column in header, got: %s", firstLine)
	}
}

func TestPrint_JSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	if err := output.Print(ctx, []map[string]any{}, "json"); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if out != "[]" {
		t.Errorf("expected [] for empty slice, got %q", out)
	}
}

func TestPrint_Text(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	data := []map[string]any{{"name": "main"}}
	if err := output.Print(ctx, data, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "main") {
		t.Errorf("expected 'main' in text output, got: %s", buf.String())
	}
}

func TestPrint_DefaultsToJSON_WhenFormatEmpty(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	data := []map[string]any{{"name": "main"}}
	if err := output.Print(ctx, data, ""); err != nil {
		t.Fatal(err)
	}
	var result []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
}

func TestPrint_UsesStdoutWhenNoWriterInContext(t *testing.T) {
	err := output.Print(context.Background(), []map[string]any{{"name": "main"}}, "json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrintError_WritesJSONToStderr(t *testing.T) {
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w

	output.PrintError(errors.New("index not found"), "NOT_FOUND", 404)

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	got := buf.String()

	if !strings.Contains(got, "index not found") {
		t.Errorf("expected error message in output, got: %s", got)
	}
	if !strings.Contains(got, "NOT_FOUND") {
		t.Errorf("expected error code in output, got: %s", got)
	}
	if !strings.Contains(got, "404") {
		t.Errorf("expected status 404 in output, got: %s", got)
	}
}

// sortKeysNameFirst must leave order unchanged when there is no "name" key.
func TestPrint_Table_NoNameKey(t *testing.T) {
	var buf bytes.Buffer
	ctx := output.WithWriter(context.Background(), &buf)
	data := []map[string]any{{"host": "myhost", "status": "up"}}
	if err := output.Print(ctx, data, "table"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "host") {
		t.Errorf("expected 'host' column in output, got: %s", buf.String())
	}
}
