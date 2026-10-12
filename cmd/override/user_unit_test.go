package override

import (
	"errors"
	"strings"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
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

func TestRawUserResult_ClassifiesAuthorizationError(t *testing.T) {
	body := []byte(`{"messages":[{"type":"ERROR","text":"You do not have permission to perform this operation (edit_user)."}]}`)
	_, err := rawUserResult(body, 403)

	var splunkErr *client.SplunkError
	if !errors.As(err, &splunkErr) {
		t.Fatalf("expected *client.SplunkError, got %T: %v", err, err)
	}
	if splunkErr.Status != 403 {
		t.Errorf("expected Status 403, got %d", splunkErr.Status)
	}
	if splunkErr.Code != "AUTH_ERROR" {
		t.Errorf("expected Code AUTH_ERROR (errorToCode maps this to exit 4), got %q", splunkErr.Code)
	}
	if splunkErr.Message != "You do not have permission to perform this operation (edit_user)." {
		t.Errorf("expected Splunk's original message preserved, got %q", splunkErr.Message)
	}
}

func TestRawUserResult_SuccessStillParsesFeed(t *testing.T) {
	body := []byte(`{"entry":[{"name":"jdoe","content":{"email":"jdoe@example.com"}}]}`)
	results, err := rawUserResult(body, 200)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0]["name"] != "jdoe" {
		t.Fatalf("unexpected results: %v", results)
	}
}
