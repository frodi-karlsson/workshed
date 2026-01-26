//go:build !integration
// +build !integration

package tui

import (
	"testing"
)

func TestIsHumanMode(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		expect bool
	}{
		{"empty", "", true},
		{"human", "human", true},
		{"json", "json", false},
		{"other", "pretty", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WORKSHED_LOG_FORMAT", tt.env)
			if got := IsHumanMode(); got != tt.expect {
				t.Errorf("IsHumanMode() = %v, want %v", got, tt.expect)
			}
		})
	}
}
