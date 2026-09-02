package cmd

import (
	"errors"
	"testing"

	"github.com/splunk/splunkctl/internal/client"
)

func TestErrorToCode_NotFound(t *testing.T) {
	err := &client.SplunkError{Code: "NOT_FOUND"}
	code, exit := errorToCode(err)
	if code != "NOT_FOUND" || exit != 3 {
		t.Errorf("got (%s, %d), want (NOT_FOUND, 3)", code, exit)
	}
}

func TestErrorToCode_AuthError(t *testing.T) {
	err := &client.SplunkError{Code: "AUTH_ERROR"}
	code, exit := errorToCode(err)
	if code != "AUTH_ERROR" || exit != 4 {
		t.Errorf("got (%s, %d), want (AUTH_ERROR, 4)", code, exit)
	}
}

func TestErrorToCode_OtherSplunkError(t *testing.T) {
	err := &client.SplunkError{Code: "QUOTA_EXCEEDED"}
	code, exit := errorToCode(err)
	if code != "QUOTA_EXCEEDED" || exit != 1 {
		t.Errorf("got (%s, %d), want (QUOTA_EXCEEDED, 1)", code, exit)
	}
}

func TestErrorToCode_ConnectionError(t *testing.T) {
	err := errors.New("connection error: dial tcp: connection refused")
	code, exit := errorToCode(err)
	if code != "CONNECTION_ERROR" || exit != 5 {
		t.Errorf("got (%s, %d), want (CONNECTION_ERROR, 5)", code, exit)
	}
}

func TestErrorToCode_UsageError(t *testing.T) {
	err := errors.New("unknown flag: --typo")
	code, exit := errorToCode(err)
	if code != "USAGE_ERROR" || exit != 1 {
		t.Errorf("got (%s, %d), want (USAGE_ERROR, 1)", code, exit)
	}
}
