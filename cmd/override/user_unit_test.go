package override

import (
	"strings"
	"testing"
)

func TestParseUserFeed_SingleUser(t *testing.T) {
	body := []byte(`{"entry":[{"name":"jdoe","content":{"email":"jdoe@example.com","roles":["power"]}}]}`)
	results, err := parseUserFeed(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["name"] != "jdoe" {
		t.Errorf("expected name=jdoe, got %v", results[0]["name"])
	}
	if results[0]["email"] != "jdoe@example.com" {
		t.Errorf("expected email=jdoe@example.com, got %v", results[0]["email"])
	}
}

func TestParseUserFeed_EmptyNoMessages(t *testing.T) {
	body := []byte(`{"entry":[],"messages":[]}`)
	results, err := parseUserFeed(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestParseUserFeed_EmptyWithSplunkError(t *testing.T) {
	body := []byte(`{"entry":[],"messages":[{"type":"ERROR","text":"Permission denied"}]}`)
	_, err := parseUserFeed(body)
	if err == nil {
		t.Fatal("expected error for Splunk error message, got nil")
	}
	if !strings.Contains(err.Error(), "Permission denied") {
		t.Errorf("expected error to contain 'Permission denied', got %v", err)
	}
}

func TestParseUserFeed_InvalidJSON(t *testing.T) {
	_, err := parseUserFeed([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing user response") {
		t.Errorf("expected error to contain 'parsing user response', got %v", err)
	}
}
