package override

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

type mockRawDoer struct {
	body   []byte
	status int
	err    error
}

func (m *mockRawDoer) RawDo(_ context.Context, _, _ string, _ url.Values) ([]byte, int, error) {
	return m.body, m.status, m.err
}

func aclFeedJSON(owner, app string) []byte {
	return []byte(fmt.Sprintf(`{"entry":[{"acl":{"owner":%q,"app":%q}}]}`, owner, app))
}

// resolveAclPath tests

func TestResolveAclPath_UnknownType(t *testing.T) {
	_, err := resolveAclPath(context.Background(), &mockRawDoer{}, "badtype", "", "myobj")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unknown --type") {
		t.Errorf("expected 'unknown --type' in error, got %v", err)
	}
}

func TestResolveAclPath_HappyPath(t *testing.T) {
	mock := &mockRawDoer{body: aclFeedJSON("admin", "search"), status: 200}
	got, err := resolveAclPath(context.Background(), mock, "dashboard", "search", "myview")
	if err != nil {
		t.Fatal(err)
	}
	want := "/servicesNS/admin/search/data/ui/views/myview/acl"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveAclPath_DefaultsOwnerToNobody(t *testing.T) {
	mock := &mockRawDoer{body: aclFeedJSON("", "search"), status: 200}
	got, err := resolveAclPath(context.Background(), mock, "dashboard", "search", "myview")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/servicesNS/nobody/") {
		t.Errorf("expected 'nobody' in path when owner is empty, got %q", got)
	}
}

func TestResolveAclPath_DefaultsAppFromParam(t *testing.T) {
	mock := &mockRawDoer{body: aclFeedJSON("admin", ""), status: 200}
	got, err := resolveAclPath(context.Background(), mock, "dashboard", "myapp", "myview")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/myapp/") {
		t.Errorf("expected passed app 'myapp' in path when JSON app is empty, got %q", got)
	}
}

func TestResolveAclPath_HTTPError(t *testing.T) {
	mock := &mockRawDoer{body: []byte("not found"), status: 404}
	_, err := resolveAclPath(context.Background(), mock, "dashboard", "", "myview")
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "object lookup failed") {
		t.Errorf("expected 'object lookup failed' in error, got %v", err)
	}
}

func TestResolveAclPath_EmptyEntry(t *testing.T) {
	mock := &mockRawDoer{body: []byte(`{"entry":[]}`), status: 200}
	_, err := resolveAclPath(context.Background(), mock, "dashboard", "", "myview")
	if err == nil {
		t.Fatal("expected error for empty entry array")
	}
	if !strings.Contains(err.Error(), "cannot resolve namespace") {
		t.Errorf("expected 'cannot resolve namespace' in error, got %v", err)
	}
}

// parseACLResponse tests

func TestParseACLResponse_SingleEntry(t *testing.T) {
	body := []byte(`{"entry":[{"acl":{"sharing":"app","owner":"admin"}}]}`)
	results, err := parseACLResponse(body, "myview", "dashboard")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "myview" {
		t.Errorf("expected name=myview, got %v", results[0]["name"])
	}
	if results[0]["type"] != "dashboard" {
		t.Errorf("expected type=dashboard, got %v", results[0]["type"])
	}
	if results[0]["sharing"] != "app" {
		t.Errorf("expected sharing=app, got %v", results[0]["sharing"])
	}
}

func TestParseACLResponse_InvalidJSON(t *testing.T) {
	_, err := parseACLResponse([]byte(`not json`), "myview", "dashboard")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parsing acl response") {
		t.Errorf("expected 'parsing acl response' in error, got %v", err)
	}
}

func TestParseACLResponse_EmptyEntry(t *testing.T) {
	_, err := parseACLResponse([]byte(`{"entry":[]}`), "myview", "dashboard")
	if err == nil {
		t.Fatal("expected error for empty entry array")
	}
	if !strings.Contains(err.Error(), "parsing acl response") {
		t.Errorf("expected 'parsing acl response' in error, got %v", err)
	}
}
