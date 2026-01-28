package workspace

import (
	"testing"
)

func TestParseRepoFlag(t *testing.T) {
	tests := []struct {
		name      string
		repo      string
		wantURL   string
		wantRef   string
		wantDepth int
	}{
		{
			name:      "simple URL",
			repo:      "https://github.com/org/repo",
			wantURL:   "https://github.com/org/repo",
			wantRef:   "",
			wantDepth: 0,
		},
		{
			name:      "URL with depth",
			repo:      "https://github.com/org/repo::10",
			wantURL:   "https://github.com/org/repo",
			wantRef:   "",
			wantDepth: 10,
		},
		{
			name:      "URL with ref",
			repo:      "https://github.com/org/repo@main",
			wantURL:   "https://github.com/org/repo",
			wantRef:   "main",
			wantDepth: 0,
		},
		{
			name:      "URL with ref and depth",
			repo:      "https://github.com/org/repo@main::5",
			wantURL:   "https://github.com/org/repo",
			wantRef:   "main",
			wantDepth: 5,
		},
		{
			name:      "local path with ref and depth",
			repo:      "./local@branch::1",
			wantURL:   "./local",
			wantRef:   "branch",
			wantDepth: 1,
		},
		{
			name:      "SSH URL with ref",
			repo:      "git@github.com:org/repo@main",
			wantURL:   "git@github.com:org/repo",
			wantRef:   "main",
			wantDepth: 0,
		},
		{
			name:      "SSH URL with ref and depth",
			repo:      "git@github.com:org/repo@main::3",
			wantURL:   "git@github.com:org/repo",
			wantRef:   "main",
			wantDepth: 3,
		},
		{
			name:      "depth zero",
			repo:      "https://github.com/org/repo::0",
			wantURL:   "https://github.com/org/repo",
			wantRef:   "",
			wantDepth: 0,
		},
		{
			name:      "no :: but :: in URL-like path",
			repo:      "https://github.com/org/repo::something",
			wantURL:   "https://github.com/org/repo::something",
			wantRef:   "",
			wantDepth: 0,
		},
		{
			name:      "empty ref with depth using @:: syntax",
			repo:      "https://github.com/org/repo@::10",
			wantURL:   "https://github.com/org/repo",
			wantRef:   "",
			wantDepth: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ref, depth := ParseRepoFlag(tt.repo)
			if url != tt.wantURL {
				t.Errorf("ParseRepoFlag(%q).url = %q, want %q", tt.repo, url, tt.wantURL)
			}
			if ref != tt.wantRef {
				t.Errorf("ParseRepoFlag(%q).ref = %q, want %q", tt.repo, ref, tt.wantRef)
			}
			if depth != tt.wantDepth {
				t.Errorf("ParseRepoFlag(%q).depth = %d, want %d", tt.repo, depth, tt.wantDepth)
			}
		})
	}
}
