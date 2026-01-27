package create

import "testing"

func TestValidateRepoFlag_Protocols(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"git:// valid", "git://github.com/user/repo", false},
		{"git:// empty host", "git://", true},
		{"git:// with path", "git://github.com/user/repo", false},
		{"ssh:// valid", "ssh://github.com/user/repo", false},
		{"ssh:// with user", "ssh://git@github.com/user/repo", false},
		{"ssh:// empty", "ssh://", true},
		{"ssh:// root only", "ssh:///", true},
		{"https:// valid", "https://github.com/user/repo", false},
		{"https:// empty host", "https://", true},
		{"http:// valid", "http://github.com/user/repo", false},
		{"http:// empty host", "http://", true},
		{"git@ SSH format", "git@github.com:user/repo", false},
		{"file:// valid absolute path", "file:///absolute/path", false},
		{"file:// empty path", "file://", true},
		{"file:// relative path rejected", "file://relative/path", true},
		{"unsupported scheme", "ftp://github.com/user/repo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepoFlag(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("validateRepoFlag(%q) = nil, want error", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateRepoFlag(%q) = %v, want nil", tt.url, err)
			}
		})
	}
}

func TestValidateRepoFlag_LocalPaths(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"existing directory", ".", false},
		{"existing subdirectory", "./testdata", false},
		{"non-existent path", "/non/existent/path", false},
		{"*.git suffix not found", "nonexistent.git", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepoFlag(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("validateRepoFlag(%q) = nil, want error", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateRepoFlag(%q) = %v, want nil", tt.url, err)
			}
		})
	}
}
