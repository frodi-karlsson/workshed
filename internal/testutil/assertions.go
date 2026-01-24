package testutil

import (
	"strings"
	"testing"
)

func AssertNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func AssertError(t *testing.T, err error, msg string) {
	if err == nil {
		t.Fatalf("%s: expected error, got nil", msg)
	}
}

func AssertErrorContains(t *testing.T, err error, substr string) {
	if err == nil {
		t.Fatalf("Expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("Error %q should contain %q", err.Error(), substr)
	}
}

func AssertNotNil(t *testing.T, v interface{}, msg string) {
	if v == nil {
		t.Fatalf("%s: expected non-nil value", msg)
	}
}

func AssertNil(t *testing.T, v interface{}, msg string) {
	if v != nil {
		t.Fatalf("%s: expected nil, got %v", msg, v)
	}
}

func AssertEqual(t *testing.T, got, want interface{}, msg string) {
	if got != want {
		t.Fatalf("%s: got %v, want %v", msg, got, want)
	}
}

func AssertNotEqual(t *testing.T, got, want interface{}, msg string) {
	if got == want {
		t.Fatalf("%s: got %v, did not want %v", msg, got, want)
	}
}

func AssertTrue(t *testing.T, v bool, msg string) {
	if !v {
		t.Fatalf("%s: expected true", msg)
	}
}

func AssertFalse(t *testing.T, v bool, msg string) {
	if v {
		t.Fatalf("%s: expected false", msg)
	}
}

func AssertGreaterThan(t *testing.T, got, want int, msg string) {
	if got <= want {
		t.Fatalf("%s: got %d, want greater than %d", msg, got, want)
	}
}

func AssertLessThan(t *testing.T, got, want int, msg string) {
	if got >= want {
		t.Fatalf("%s: got %d, want less than %d", msg, got, want)
	}
}

func AssertNonEmpty(t *testing.T, s string, msg string) {
	if s == "" {
		t.Fatalf("%s: expected non-empty string", msg)
	}
}

func AssertContains(t *testing.T, s, substr string, msg string) {
	if !strings.Contains(s, substr) {
		t.Fatalf("%s: expected %q to contain %q", msg, s, substr)
	}
}

func AssertNotContains(t *testing.T, s, substr string, msg string) {
	if strings.Contains(s, substr) {
		t.Fatalf("%s: expected %q to not contain %q", msg, s, substr)
	}
}
