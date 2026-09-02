package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

var reBraceVar = regexp.MustCompile(`\{[^}]+\}`)

// GeneratedFile is the data passed to the template.
type GeneratedFile struct {
	PackageName  string
	ObjectName   string // e.g. "index", "cluster-manager"
	SafeName     string // Go-safe identifier, e.g. "Index", "ClusterManager"
	CommonURI    string
	ArgsMap      []ArgMapping
	Verbs        []VerbDef
	XMLSource    string // path to the source XML file
	NeedsStrings bool   // true when any verb has PathHasTemplate
}

// VerbDef is one verb (list, add, etc.) on an object.
type VerbDef struct {
	Name               string
	HTTPMethod         string
	URI                string // effective URI (per-verb override or common)
	PathHasTemplate    bool   // true when URI contains {varname} placeholders
	PathVarPlaceholder string // e.g. "{name}" — the literal placeholder to substitute
	ImpliedArg         string
	EaiID              string // ID field name extracted from e.g. "{username}" → "username"
	EaiIDLiteral       bool   // true = EaiID is a hardcoded path segment, not user-supplied
	OfflineOK          bool
	Short              string
	Long               string
	Examples           []string
	RequiredArgs       []FlagDef
	OptionalArgs       []FlagDef
}

// FlagDef describes one CLI flag for a verb.
type FlagDef struct {
	CliName string
	EaiName string
	Usage   string
}

// toSafeName converts "cluster-manager" → "ClusterManager".
func toSafeName(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	var b strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			b.WriteRune(unicode.ToUpper(rune(p[0])))
			b.WriteString(p[1:])
		}
	}
	return b.String()
}

// stripEaiIDField extracts "username" from "{username}".
func stripEaiIDField(eaiID string) string {
	s := strings.TrimPrefix(eaiID, "{")
	s = strings.TrimSuffix(s, "}")
	return s
}

// deduplicateFlags removes duplicate flag definitions (same CliName) keeping first.
func deduplicateFlags(flags []FlagDef) []FlagDef {
	seen := make(map[string]bool)
	out := make([]FlagDef, 0, len(flags))
	for _, f := range flags {
		if !seen[f.CliName] {
			seen[f.CliName] = true
			out = append(out, f)
		}
	}
	return out
}

// BuildGeneratedFile converts an Item into a GeneratedFile ready for templating.
func BuildGeneratedFile(item Item) *GeneratedFile {
	gf := &GeneratedFile{
		PackageName: "generated",
		ObjectName:  item.Obj,
		SafeName:    toSafeName(item.Obj),
		CommonURI:   strings.TrimSpace(item.Common.URI),
		ArgsMap:     item.Common.ArgsMap.Args,
	}

	// build argsmap lookup: cliname → eainame
	cliToEai := map[string]string{}
	for _, a := range item.Common.ArgsMap.Args {
		cliToEai[a.CliName] = a.EaiName
	}

	for _, cmd := range item.Cmds {
		if cmd.Synonym != "" {
			// verb alias — skip
			continue
		}
		if strings.TrimSpace(cmd.Name) == "" {
			continue
		}

		effectiveURI := strings.TrimSpace(cmd.URI)
		if effectiveURI == "" {
			effectiveURI = gf.CommonURI
		}
		if effectiveURI == "" {
			continue // no URI = can't generate REST call
		}

		// build flag defs from required + optional help args
		var required, optional []FlagDef
		for _, a := range cmd.Help.Required {
			name := strings.TrimLeft(strings.TrimSpace(a.Name), "-") // strip leading --
			if name == "" {
				continue
			}
			eai := cliToEai[name]
			if eai == "" {
				eai = name
			}
			required = append(required, FlagDef{CliName: name, EaiName: eai, Usage: strings.TrimSpace(a.Text)})
		}
		for _, a := range cmd.Help.Optional {
			name := strings.TrimLeft(strings.TrimSpace(a.Name), "-") // strip leading --
			if name == "" {
				continue
			}
			eai := cliToEai[name]
			if eai == "" {
				eai = name
			}
			optional = append(optional, FlagDef{CliName: name, EaiName: eai, Usage: strings.TrimSpace(a.Text)})
		}

		// deduplicate flags
		required = deduplicateFlags(required)
		optional = deduplicateFlags(optional)

		// build Short from help title
		short := strings.TrimSpace(cmd.Help.Title)
		if short == "" {
			short = cmd.Name + " " + item.Obj
		}

		// examples: rewrite "./splunk verb obj ..." → "splunkctl obj verb ..." and strip -auth
		var examples []string
		for _, ex := range cmd.Help.Examples {
			ex = rewriteExample(ex, item.Obj, cmd.Name)
			if ex != "" {
				examples = append(examples, ex)
			}
		}

		eaiIDField := ""
		eaiIDLiteral := false
		if cmd.EaiID != "" {
			raw := strings.TrimSpace(cmd.EaiID)
			if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
				eaiIDField = stripEaiIDField(raw) // user-supplied, strip braces
			} else {
				eaiIDField = raw // hardcoded literal
				eaiIDLiteral = true
			}
		}

		// If eaiIDField matches the implied arg, only use eaiIDField
		// to avoid duplicate flag registration
		impliedArg := strings.TrimSpace(cmd.ImpliedArg)
		if eaiIDField != "" && eaiIDField == impliedArg {
			impliedArg = "" // suppress ImpliedArg; EaiID flag will handle it
		}

		// Also remove flags from required/optional that duplicate eaiIDField or impliedArg
		if eaiIDField != "" {
			required = removeFlagByCliName(required, eaiIDField)
			optional = removeFlagByCliName(optional, eaiIDField)
		}
		if impliedArg != "" {
			required = removeFlagByCliName(required, impliedArg)
			optional = removeFlagByCliName(optional, impliedArg)
		}

		pathVarPlaceholder := ""
		if m := reBraceVar.FindString(effectiveURI); m != "" {
			pathVarPlaceholder = m
		}

		vd := VerbDef{
			Name:               cmd.Name,
			HTTPMethod:         HTTPMethod(cmd.Name),
			URI:                effectiveURI,
			PathHasTemplate:    pathVarPlaceholder != "",
			PathVarPlaceholder: pathVarPlaceholder,
			ImpliedArg:         impliedArg,
			EaiID:              eaiIDField,
			EaiIDLiteral:       eaiIDLiteral,
			OfflineOK:          cmd.OfflineOK != nil,
			Short:              short,
			Examples:           examples,
			RequiredArgs:       required,
			OptionalArgs:       optional,
		}
		gf.Verbs = append(gf.Verbs, vd)
	}

	for _, v := range gf.Verbs {
		if v.PathHasTemplate {
			gf.NeedsStrings = true
			break
		}
	}
	return gf
}

// reAuthRe matches -auth user:pass in examples.
var reAuthRe = regexp.MustCompile(`\s*-auth\s+\S+`)

// rewriteExample converts "./splunk verb obj ..." → "splunkctl obj verb ..." and strips -auth.
func rewriteExample(ex, obj, verb string) string {
	ex = strings.TrimSpace(ex)
	if ex == "" {
		return ""
	}
	// strip -auth flag+value
	ex = reAuthRe.ReplaceAllString(ex, "")
	// replace old ./splunk or splunk prefix with splunkctl obj verb
	for _, prefix := range []string{"./splunk ", "splunk "} {
		if strings.Contains(ex, prefix+verb+" "+obj) {
			ex = strings.ReplaceAll(ex, prefix+verb+" "+obj, "splunkctl "+obj+" "+verb)
			return strings.TrimSpace(ex)
		}
		if strings.Contains(ex, prefix+verb+" -") {
			// verb with flags but no explicit obj in example
			ex = strings.ReplaceAll(ex, prefix+verb+" ", "splunkctl "+obj+" "+verb+" ")
			return strings.TrimSpace(ex)
		}
		if strings.Contains(ex, prefix+verb) {
			ex = strings.ReplaceAll(ex, prefix+verb, "splunkctl "+obj+" "+verb)
			return strings.TrimSpace(ex)
		}
	}
	// fallback: replace bare ./splunk
	ex = strings.ReplaceAll(ex, "./splunk ", "splunkctl ")
	return strings.TrimSpace(ex)
}

// removeFlagByCliName removes all flags whose CliName equals name.
func removeFlagByCliName(flags []FlagDef, name string) []FlagDef {
	out := make([]FlagDef, 0, len(flags))
	for _, f := range flags {
		if f.CliName != name {
			out = append(out, f)
		}
	}
	return out
}

// RenderFile renders the Go source for a GeneratedFile using the template.
func RenderFile(tmpl *template.Template, gf *GeneratedFile) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, gf); err != nil {
		return nil, fmt.Errorf("template execute for %s: %w", gf.ObjectName, err)
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// return raw for debugging
		return buf.Bytes(), fmt.Errorf("gofmt for %s: %w\n--- source ---\n%s", gf.ObjectName, err, buf.String())
	}
	return formatted, nil
}

// WriteFile writes the rendered source to outDir/<objectName>.go.
func WriteFile(outDir string, gf *GeneratedFile, src []byte) error {
	filename := objectToFilename(gf.ObjectName)
	path := filepath.Join(outDir, filename)
	return os.WriteFile(path, src, 0644)
}

func objectToFilename(obj string) string {
	return strings.ToLower(strings.ReplaceAll(obj, "-", "_")) + ".go"
}
