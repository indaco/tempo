package validation

import (
	"errors"
	"testing"
)

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr error
	}{
		// Valid URLs
		{
			name:    "valid https URL",
			url:     "https://github.com/user/repo.git",
			wantErr: nil,
		},
		{
			name:    "valid git URL",
			url:     "git://github.com/user/repo.git",
			wantErr: nil,
		},
		{
			name:    "valid ssh URL",
			url:     "ssh://git@github.com/user/repo.git",
			wantErr: nil,
		},
		{
			name:    "valid https URL without .git suffix",
			url:     "https://github.com/user/repo",
			wantErr: nil,
		},
		{
			name:    "valid local absolute path",
			url:     "/tmp/local/repo",
			wantErr: nil,
		},
		{
			name:    "valid local relative path",
			url:     "./local/repo",
			wantErr: nil,
		},
		{
			name:    "valid local path without prefix",
			url:     "local/repo",
			wantErr: nil,
		},
		// Invalid URLs
		{
			name:    "empty URL",
			url:     "",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "file scheme URL",
			url:     "file:///etc/passwd",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "http scheme URL",
			url:     "http://github.com/user/repo.git",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "flag injection with dash",
			url:     "-u./payload",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "flag injection --upload-pack",
			url:     "--upload-pack=evil",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "URL without host",
			url:     "https:///path/only",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "ftp scheme URL",
			url:     "ftp://example.com/repo.git",
			wantErr: ErrInvalidGitURL,
		},
		{
			name:    "local path with traversal",
			url:     "../../../etc/passwd",
			wantErr: ErrInvalidGitURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.url)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateGitURL(%q) = nil, want error containing %v", tt.url, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateGitURL(%q) error = %v, want %v", tt.url, err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidateGitURL(%q) = %v, want nil", tt.url, err)
			}
		})
	}
}

func TestValidateLocalPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		// Valid paths
		{
			name:    "simple filename",
			path:    "file.txt",
			wantErr: nil,
		},
		{
			name:    "nested path",
			path:    "dir/subdir/file.txt",
			wantErr: nil,
		},
		{
			name:    "current directory",
			path:    ".",
			wantErr: nil,
		},
		{
			name:    "absolute path (allowed)",
			path:    "/tmp/test/repo",
			wantErr: nil,
		},
		// Invalid paths
		{
			name:    "empty path",
			path:    "",
			wantErr: ErrInvalidPath,
		},
		{
			name:    "parent directory traversal",
			path:    "../etc/passwd",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "flag injection",
			path:    "-rf",
			wantErr: ErrInvalidPath,
		},
		{
			name:    "double dot traversal",
			path:    "foo/../../bar",
			wantErr: ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocalPath(tt.path)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateLocalPath(%q) = nil, want error containing %v", tt.path, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateLocalPath(%q) error = %v, want %v", tt.path, err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidateLocalPath(%q) = %v, want nil", tt.path, err)
			}
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr error
	}{
		{
			name:    "empty directory (current dir)",
			dir:     "",
			wantErr: nil,
		},
		{
			name:    "current directory",
			dir:     ".",
			wantErr: nil,
		},
		{
			name:    "nested directory",
			dir:     "path/to/dir",
			wantErr: nil,
		},
		{
			name:    "flag injection",
			dir:     "-rf",
			wantErr: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectory(tt.dir)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateDirectory(%q) = nil, want error containing %v", tt.dir, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateDirectory(%q) error = %v, want %v", tt.dir, err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("ValidateDirectory(%q) = %v, want nil", tt.dir, err)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr error
	}{
		{
			name:    "simple path",
			path:    "file.txt",
			want:    "file.txt",
			wantErr: nil,
		},
		{
			name:    "path with dots",
			path:    "./dir/../file.txt",
			want:    "file.txt",
			wantErr: nil,
		},
		{
			name:    "nested path",
			path:    "a/b/c",
			want:    "a/b/c",
			wantErr: nil,
		},
		{
			name:    "empty path",
			path:    "",
			want:    "",
			wantErr: ErrInvalidPath,
		},
		{
			name:    "flag injection",
			path:    "-rf",
			want:    "",
			wantErr: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("SanitizePath(%q) = (%q, nil), want error %v", tt.path, got, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("SanitizePath(%q) error = %v, want %v", tt.path, err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizePath(%q) error = %v, want nil", tt.path, err)
					return
				}
				if got != tt.want {
					t.Errorf("SanitizePath(%q) = %q, want %q", tt.path, got, tt.want)
				}
			}
		})
	}
}
