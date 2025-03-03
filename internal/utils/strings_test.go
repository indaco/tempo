package utils

import (
	"path/filepath"
	"testing"
)

func TestContainsSubstring(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		substr string
		want   bool
	}{
		{"Substring exists", "hello world", "world", true},
		{"Substring does not exist", "hello world", "golang", false},
		{"Empty substring", "hello world", "", true}, // strings.Contains treats empty substring as always found
		{"Empty string and substring", "", "", true}, // empty substring in empty string is valid
		{"Empty string with non-empty substring", "", "hello", false},
		{"Case sensitive match", "GoLang", "golang", false}, // should be case-sensitive
		{"Substring at the start", "abcdef", "abc", true},
		{"Substring at the end", "abcdef", "def", true},
		{"Single character match", "abcdef", "a", true},
		{"Single character no match", "abcdef", "z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsSubstring(tt.str, tt.substr)
			if got != tt.want {
				t.Errorf("ContainsSubstring(%q, %q) = %v; want %v", tt.str, tt.substr, got, tt.want)
			}
		})
	}
}

func TestExtractNameFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:user/repo", "repo"},
		{"https://bitbucket.org/user/my-repo.git", "my-repo"},
		{"https://gitlab.com/user/project.git", "project"},
		{"https://example.com/custom/path/to/repository.git", "repository"},
		{"", ""},
		{"/absolute/path/to/repo.git", "repo"},
		{"just-a-repo-name.git", "just-a-repo-name"},
	}

	for _, test := range tests {
		result := ExtractNameFromURL(test.url)
		if result != test.expected {
			t.Errorf("ExtractNameFromURL(%q) = %q; expected %q", test.url, result, test.expected)
		}
	}
}

func TestExtractNameFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/project", "project"},
		{"./relative/path/to/module", "module"},
		{"/var/www/html", "html"},
		{filepath.FromSlash("C:/Users/User/Documents/repo"), "repo"}, // Cross-platform Windows path
		{filepath.FromSlash("C:/Users/User/Documents/my-project"), "my-project"},
		{filepath.FromSlash("../parent/folder/package"), "package"},
		{"single-folder", "single-folder"},
		{"", ""},
	}

	for _, test := range tests {
		result := ExtractNameFromPath(test.path)
		if result != test.expected {
			t.Errorf("ExtractNameFromPath(%q) = %q; expected %q", test.path, result, test.expected)
		}
	}
}
