package testutil

import (
	"os"
	"testing"
)

func WithEnvVar(t *testing.T, key, value string, fn func()) {
	if value != "" {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
		defer func() {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("Failed to unset %s: %v", key, err)
			}
		}()
	} else {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Failed to unset %s: %v", key, err)
		}
	}
	fn()
}
