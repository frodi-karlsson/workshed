package unit

import (
	"testing"

	"github.com/frodi/workshed/internal/cli"
	"github.com/frodi/workshed/internal/cli/captures"
	"github.com/frodi/workshed/internal/cli/export"
)

func TestExtractHandleFromArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		extraFilters  []string
		wantHandle    string
		wantRemaining []string
	}{
		{
			name:          "no args",
			args:          []string{},
			wantHandle:    "",
			wantRemaining: nil,
		},
		{
			name:          "handle only",
			args:          []string{"my-workspace"},
			wantHandle:    "my-workspace",
			wantRemaining: nil,
		},
		{
			name:          "handle with flags after",
			args:          []string{"my-workspace", "--filter", "foo"},
			wantHandle:    "my-workspace",
			wantRemaining: []string{"--filter", "foo"},
		},
		{
			name:          "flags before first positional becomes handle",
			args:          []string{"--filter", "foo", "my-workspace"},
			wantHandle:    "foo",
			wantRemaining: []string{"--filter", "my-workspace"},
		},
		{
			name:          "handle before dashdash",
			args:          []string{"my-workspace", "--", "go", "test"},
			wantHandle:    "my-workspace",
			wantRemaining: []string{"--", "go", "test"},
		},
		{
			name:          "only flags",
			args:          []string{"--filter", "foo"},
			wantHandle:    "foo",
			wantRemaining: []string{"--filter"},
		},
		{
			name:          "capture ID with --name as extra filter",
			args:          []string{"01HVABCDEFG", "--name", "foo"},
			extraFilters:  []string{"--name"},
			wantHandle:    "01HVABCDEFG",
			wantRemaining: []string{"--name", "foo"},
		},
		{
			name:          "extra filter excludes capture ID",
			args:          []string{"01HVABCDEFG"},
			extraFilters:  []string{"01HVABCDEFG"},
			wantHandle:    "",
			wantRemaining: []string{"01HVABCDEFG"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHandle, gotRemaining := cli.ExtractHandleFromArgs(tt.args, tt.extraFilters...)
			if gotHandle != tt.wantHandle {
				t.Errorf("ExtractHandleFromArgs() handle = %q, want %q", gotHandle, tt.wantHandle)
			}
			if len(gotRemaining) != len(tt.wantRemaining) {
				t.Errorf("ExtractHandleFromArgs() remaining = %v, want %v", gotRemaining, tt.wantRemaining)
			} else {
				for i, v := range gotRemaining {
					if v != tt.wantRemaining[i] {
						t.Errorf("ExtractHandleFromArgs() remaining[%d] = %q, want %q", i, v, tt.wantRemaining[i])
					}
				}
			}
		})
	}
}

func TestCommandArgsValidation(t *testing.T) {
	t.Run("captures NoArgs accepts 0 args", func(t *testing.T) {
		cmd := captures.Command()
		err := cmd.Args(nil, []string{})
		if err != nil {
			t.Errorf("captures with no args should be valid: %v", err)
		}
	})
}

func TestExportFormatFlag(t *testing.T) {
	t.Run("export has raw format option", func(t *testing.T) {
		cmd := export.Command()
		f := cmd.Flags().Lookup("format")
		if f == nil {
			t.Error("export should have --format flag")
			return
		}
		if f.DefValue != "table" {
			t.Errorf("export --format default = %q, want %q", f.DefValue, "table")
		}
	})
}

func TestExportCompactFlag(t *testing.T) {
	t.Run("export has compact flag", func(t *testing.T) {
		cmd := export.Command()
		f := cmd.Flags().Lookup("compact")
		if f == nil {
			t.Error("export should have --compact flag")
		}
	})
}
