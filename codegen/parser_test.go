package main

import (
	"os"
	"strings"
	"testing"
	"text/template"
)

const sampleXML = `<root>
  <item obj="index">
    <common>
      <uri>/data/indexes/</uri>
      <argsmap>
        <arg cliname="name" eainame="name"/>
      </argsmap>
    </common>
    <cmd name="list">
      <implied_arg_name>name</implied_arg_name>
      <help>
        <title>list all indexes</title>
        <optional><arg name="name">filter by index name</arg></optional>
        <examples><ex>./splunk list index</ex></examples>
      </help>
    </cmd>
    <cmd name="add">
      <implied_arg_name>name</implied_arg_name>
      <help>
        <title>add an index</title>
        <required><arg name="name">the index name</arg></required>
      </help>
    </cmd>
  </item>
  <item obj="master" synonym="cluster-manager"/>
</root>`

func writeTempXML(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "splunkrc*.xml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := f.WriteString(sampleXML); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestParseXML(t *testing.T) {
	items, err := ParseXML(writeTempXML(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Obj != "index" {
		t.Errorf("expected obj=index, got %s", items[0].Obj)
	}
	if items[0].Common.URI != "/data/indexes/" {
		t.Errorf("expected uri=/data/indexes/, got %s", items[0].Common.URI)
	}
	if len(items[0].Cmds) != 2 {
		t.Errorf("expected 2 cmds, got %d", len(items[0].Cmds))
	}
	if items[1].Synonym != "cluster-manager" {
		t.Errorf("expected synonym=cluster-manager, got %s", items[1].Synonym)
	}
}

func TestHTTPMethod(t *testing.T) {
	cases := []struct{ verb, want string }{
		{"list", "GET"}, {"show", "GET"}, {"display", "GET"},
		{"add", "POST"}, {"edit", "POST"}, {"enable", "POST"},
		{"remove", "DELETE"},
	}
	for _, tc := range cases {
		if got := HTTPMethod(tc.verb); got != tc.want {
			t.Errorf("HTTPMethod(%q) = %q, want %q", tc.verb, got, tc.want)
		}
	}
}

func TestBuildGeneratedFile(t *testing.T) {
	items, err := ParseXML(writeTempXML(t))
	if err != nil {
		t.Fatal(err)
	}
	gf := BuildGeneratedFile(items[0])

	if gf.ObjectName != "index" {
		t.Errorf("expected ObjectName=index, got %s", gf.ObjectName)
	}
	if gf.SafeName != "Index" {
		t.Errorf("expected SafeName=Index, got %s", gf.SafeName)
	}
	if gf.CommonURI != "/data/indexes/" {
		t.Errorf("expected CommonURI=/data/indexes/, got %s", gf.CommonURI)
	}
	if len(gf.Verbs) != 2 {
		t.Fatalf("expected 2 verbs, got %d", len(gf.Verbs))
	}
	if gf.Verbs[0].Name != "list" || gf.Verbs[0].HTTPMethod != "GET" {
		t.Errorf("expected list/GET, got %s/%s", gf.Verbs[0].Name, gf.Verbs[0].HTTPMethod)
	}
	if gf.Verbs[1].Name != "add" || gf.Verbs[1].HTTPMethod != "POST" {
		t.Errorf("expected add/POST, got %s/%s", gf.Verbs[1].Name, gf.Verbs[1].HTTPMethod)
	}
}

func TestToSafeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"index", "Index"},
		{"cluster-manager", "ClusterManager"},
		{"shc-status", "ShcStatus"},
		{"saved_search", "SavedSearch"},
	}
	for _, tc := range cases {
		if got := toSafeName(tc.in); got != tc.want {
			t.Errorf("toSafeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildGeneratedFile_SynonymSkipped(t *testing.T) {
	items, err := ParseXML(writeTempXML(t))
	if err != nil {
		t.Fatal(err)
	}
	// items[1] is the synonym alias — BuildGeneratedFile should return empty verbs
	gf := BuildGeneratedFile(items[1])
	if len(gf.Verbs) != 0 {
		t.Errorf("expected 0 verbs for synonym item, got %d", len(gf.Verbs))
	}
}

func TestStripEaiIDField(t *testing.T) {
	cases := []struct{ in, want string }{
		{"{username}", "username"},
		{"{name}", "name"},
		{"plain", "plain"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := stripEaiIDField(tc.in); got != tc.want {
			t.Errorf("stripEaiIDField(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseXML_FileNotFound(t *testing.T) {
	_, err := ParseXML("/nonexistent/path/file.xml")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestParseXML_InvalidXML(t *testing.T) {
	f, err := os.CreateTemp("", "invalid*.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("this is not xml!!!")
	f.Close()

	_, err = ParseXML(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
}

func TestRewriteExample_VerbWithFlags(t *testing.T) {
	got := rewriteExample("splunk list -app search", "index", "list")
	if !strings.Contains(got, "splunkctl") {
		t.Errorf("expected splunkctl in output, got %q", got)
	}
}

func TestRewriteExample_FallbackDotSlash(t *testing.T) {
	got := rewriteExample("./splunk something-else", "index", "list")
	if !strings.Contains(got, "splunkctl") {
		t.Errorf("expected splunkctl in output, got %q", got)
	}
}

func TestObjectToFilename_ReplacesHyphensAndAddsExtension(t *testing.T) {
	if got := objectToFilename("shc-status"); got != "shc_status.go" {
		t.Errorf("expected shc_status.go, got %q", got)
	}
}

func TestObjectToFilename_LowercasesName(t *testing.T) {
	if got := objectToFilename("Index"); got != "index.go" {
		t.Errorf("expected index.go, got %q", got)
	}
}

func TestRemoveFlagByCliName_RemovesMatchingFlag(t *testing.T) {
	flags := []FlagDef{{CliName: "output"}, {CliName: "format"}, {CliName: "count"}}
	result := removeFlagByCliName(flags, "format")
	if len(result) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(result))
	}
	for _, f := range result {
		if f.CliName == "format" {
			t.Error("expected 'format' flag to be removed")
		}
	}
}

func TestRemoveFlagByCliName_NoMatch(t *testing.T) {
	flags := []FlagDef{{CliName: "output"}, {CliName: "count"}}
	result := removeFlagByCliName(flags, "format")
	if len(result) != 2 {
		t.Errorf("expected 2 flags unchanged, got %d", len(result))
	}
}

func TestHasAnyVerbURI_ReturnsTrueWhenURIExists(t *testing.T) {
	item := Item{Cmds: []Cmd{{URI: "/services/data/indexes"}}}
	if !hasAnyVerbURI(item) {
		t.Error("expected true when a cmd has a URI")
	}
}

func TestHasAnyVerbURI_ReturnsFalseWhenNoURI(t *testing.T) {
	item := Item{Cmds: []Cmd{{URI: ""}}}
	if hasAnyVerbURI(item) {
		t.Error("expected false when no cmd has a URI")
	}
}

func TestRewriteExample_Empty(t *testing.T) {
	if got := rewriteExample("", "index", "list"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestRewriteExample_SplunkPrefix(t *testing.T) {
	got := rewriteExample("splunk list index", "index", "list")
	if got != "splunkctl index list" {
		t.Errorf("expected 'splunkctl index list', got %q", got)
	}
}

func TestRewriteExample_DotSlashPrefix(t *testing.T) {
	got := rewriteExample("./splunk list index", "index", "list")
	if got != "splunkctl index list" {
		t.Errorf("expected 'splunkctl index list', got %q", got)
	}
}

func TestRenderFile_ProducesValidGo(t *testing.T) {
	tmpl := template.Must(template.New("t").Parse("package {{.PackageName}}\n"))
	gf := &GeneratedFile{PackageName: "generated", ObjectName: "index"}
	out, err := RenderFile(tmpl, gf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "package generated") {
		t.Errorf("expected 'package generated' in output, got: %s", out)
	}
}

func TestWriteFile_WritesGoFile(t *testing.T) {
	tmpl := template.Must(template.New("t").Parse("package {{.PackageName}}\n"))
	gf := &GeneratedFile{PackageName: "generated", ObjectName: "myobj"}
	src, err := RenderFile(tmpl, gf)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := WriteFile(dir, gf, src); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir + "/myobj.go"); err != nil {
		t.Errorf("expected myobj.go to exist, got: %v", err)
	}
}
