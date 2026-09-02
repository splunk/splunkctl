package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: codegen <xml-path> <template-path> <out-dir>")
		os.Exit(1)
	}
	xmlPath := os.Args[1]
	tmplPath := os.Args[2]
	outDir := os.Args[3]

	items, err := ParseXML(xmlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse XML: %v\n", err)
		os.Exit(1)
	}

	funcMap := template.FuncMap{
		// title converts a verb name like "rolling-restart" to a Go-safe PascalCase
		// identifier component like "RollingRestart".
		"title": func(s string) string {
			return toSafeName(s)
		},
		"goString": func(s string) string {
			// use backtick string unless it contains a backtick; fallback to double-quote
			if !strings.Contains(s, "`") {
				return "`" + s + "`"
			}
			return fmt.Sprintf("%q", s)
		},
		"escapeBacktick": func(s string) string {
			return strings.ReplaceAll(s, "`", "'")
		},
		"not": func(v interface{}) bool {
			if v == nil {
				return true
			}
			switch val := v.(type) {
			case bool:
				return !val
			case string:
				return val == ""
			default:
				return false
			}
		},
	}

	tmplSrc, err := os.ReadFile(tmplPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read template: %v\n", err)
		os.Exit(1)
	}

	tmpl, err := template.New("cmd").Funcs(funcMap).Parse(string(tmplSrc))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse template: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "create out dir: %v\n", err)
		os.Exit(1)
	}

	skipped := 0
	written := 0
	errors := 0

	for _, item := range items {
		// skip synonym/alias items (no children to generate)
		if item.Synonym != "" {
			skipped++
			continue
		}
		// skip verb-only items that have no URI anywhere
		if item.Common.URI == "" && !hasAnyVerbURI(item) {
			skipped++
			continue
		}

		// skip if an override file exists in sibling override/ directory
		overridePath := filepath.Join(filepath.Dir(outDir), "override", objectToFilename(item.Obj))
		if _, err := os.Stat(overridePath); err == nil {
			fmt.Printf("skip (override): %s\n", item.Obj)
			skipped++
			continue
		}

		gf := BuildGeneratedFile(item)
		gf.XMLSource = xmlPath
		if len(gf.Verbs) == 0 {
			skipped++
			continue
		}

		src, err := RenderFile(tmpl, gf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "render %s: %v\n", item.Obj, err)
			errors++
			continue
		}

		if err := WriteFile(outDir, gf, src); err != nil {
			fmt.Fprintf(os.Stderr, "write %s: %v\n", item.Obj, err)
			errors++
			continue
		}
		fmt.Printf("wrote: %s\n", item.Obj)
		written++
	}

	fmt.Printf("\ndone: %d written, %d skipped, %d errors\n", written, skipped, errors)
	if errors > 0 {
		os.Exit(1)
	}
}

func hasAnyVerbURI(item Item) bool {
	for _, cmd := range item.Cmds {
		if strings.TrimSpace(cmd.URI) != "" {
			return true
		}
	}
	return false
}
