package override

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/splunk/splunkctl/internal/client"
	"github.com/splunk/splunkctl/internal/config"
)

func newDashboardClient(t *testing.T, handler http.HandlerFunc) (*client.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	c := client.New(&config.Config{Host: srv.URL, Token: "tok", Insecure: true})
	return c, srv
}

func dashboardFeedJSON(owner, app string) string {
	return fmt.Sprintf(`{"entry":[{"acl":{"owner":%q,"app":%q}}]}`, owner, app)
}

func newDashboardContentCmd(content, file string) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("content", content, "")
	cmd.Flags().String("file", file, "")
	return cmd
}

func TestResolveDashboardPath_HappyPath(t *testing.T) {
	c, srv := newDashboardClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, dashboardFeedJSON("admin", "search"))
	})
	defer srv.Close()

	got, err := resolveDashboardPath(context.Background(), c, "myview")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(got, "/servicesNS/admin/search/data/ui/views/myview") {
		t.Errorf("got %q, want suffix /servicesNS/admin/search/data/ui/views/myview", got)
	}
}

func TestResolveDashboardPath_DefaultsOwnerToNobody(t *testing.T) {
	c, srv := newDashboardClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, dashboardFeedJSON("", "search"))
	})
	defer srv.Close()

	got, err := resolveDashboardPath(context.Background(), c, "myview")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/nobody/") {
		t.Errorf("expected 'nobody' in path when owner is empty, got %q", got)
	}
}

func TestResolveDashboardPath_DefaultsAppToSearch(t *testing.T) {
	c, srv := newDashboardClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, dashboardFeedJSON("admin", ""))
	})
	defer srv.Close()

	got, err := resolveDashboardPath(context.Background(), c, "myview")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/search/") {
		t.Errorf("expected 'search' in path when app is empty, got %q", got)
	}
}

func TestResolveDashboardPath_HTTPError(t *testing.T) {
	c, srv := newDashboardClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"messages":[{"type":"ERROR","text":"not found"}]}`)
	})
	defer srv.Close()

	_, err := resolveDashboardPath(context.Background(), c, "myview")
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "dashboard lookup failed") {
		t.Errorf("expected 'dashboard lookup failed' in error, got %v", err)
	}
}

func TestResolveDashboardPath_EmptyEntry(t *testing.T) {
	c, srv := newDashboardClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"entry":[]}`)
	})
	defer srv.Close()

	_, err := resolveDashboardPath(context.Background(), c, "myview")
	if err == nil {
		t.Fatal("expected error for empty entry array")
	}
	if !strings.Contains(err.Error(), "cannot resolve namespace") {
		t.Errorf("expected 'cannot resolve namespace' in error, got %v", err)
	}
}

func TestParseDashboardFeed_SingleEntry(t *testing.T) {
	body := []byte(`{"entry":[{"name":"soc-overview","content":{"label":"SOC Overview"},"acl":null}]}`)
	results, err := parseDashboardFeed(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "soc-overview" {
		t.Errorf("expected name=soc-overview, got %v", results[0]["name"])
	}
	if results[0]["label"] != "SOC Overview" {
		t.Errorf("expected label=SOC Overview, got %v", results[0]["label"])
	}
}

func TestParseDashboardFeed_WithACL(t *testing.T) {
	body := []byte(`{"entry":[{"name":"soc-overview","content":{},"acl":{"sharing":"app","owner":"admin","app":"search"}}]}`)
	results, err := parseDashboardFeed(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["sharing"] != "app" {
		t.Errorf("expected sharing=app, got %v", results[0]["sharing"])
	}
	if results[0]["owner"] != "admin" {
		t.Errorf("expected owner=admin, got %v", results[0]["owner"])
	}
	if results[0]["app"] != "search" {
		t.Errorf("expected app=search, got %v", results[0]["app"])
	}
}

func TestParseDashboardFeed_Empty(t *testing.T) {
	body := []byte(`{"entry":[]}`)
	results, err := parseDashboardFeed(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseDashboardFeed_InvalidJSON(t *testing.T) {
	_, err := parseDashboardFeed([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing dashboard response") {
		t.Errorf("expected 'parsing dashboard response' in error, got %v", err)
	}
}

func TestDashboardPath_WithApp(t *testing.T) {
	got := dashboardPath("search")
	want := "/servicesNS/nobody/search/data/ui/views/"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDashboardPath_NoApp(t *testing.T) {
	got := dashboardPath("")
	want := "/services/data/ui/views/"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDashboardContent_InlineContent(t *testing.T) {
	cmd := newDashboardContentCmd("<dashboard/>", "")
	got, err := dashboardContent(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if got != "<dashboard/>" {
		t.Errorf("got %q, want <dashboard/>", got)
	}
}

func TestDashboardContent_NeitherFlagSet(t *testing.T) {
	cmd := newDashboardContentCmd("", "")
	_, err := dashboardContent(cmd)
	if err == nil {
		t.Fatal("expected error when neither --content nor --file is set")
	}
	if !strings.Contains(err.Error(), "--content or --file") {
		t.Errorf("expected '--content or --file' in error, got %v", err)
	}
}

func TestDashboardContent_FromFile(t *testing.T) {
	f, err := os.CreateTemp("", "dashboard-*.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("<dashboard><label>Test</label></dashboard>")
	f.Close()

	cmd := newDashboardContentCmd("", f.Name())
	got, err := dashboardContent(cmd)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "<dashboard>") {
		t.Errorf("expected dashboard XML in output, got %q", got)
	}
}
