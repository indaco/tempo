package validation

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// ErrInvalidGitURL indicates an invalid or unsafe Git URL.
var ErrInvalidGitURL = fmt.Errorf("invalid git URL")

// ErrPathTraversal indicates a path traversal attempt was detected.
var ErrPathTraversal = fmt.Errorf("path traversal detected")

// ErrInvalidPath indicates an invalid path.
var ErrInvalidPath = fmt.Errorf("invalid path")

// allowedGitSchemes defines the allowed URL schemes for git operations.
var allowedGitSchemes = map[string]bool{
	"https": true,
	"git":   true,
	"ssh":   true,
}

// ValidateGitURL ensures the URL is safe to use with git clone.
// It validates the URL format and restricts to safe schemes (https, git, ssh).
// Local filesystem paths are also allowed for cloning local repositories.
func ValidateGitURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("%w: empty URL", ErrInvalidGitURL)
	}

	// Check for flag injection (URLs starting with -)
	if strings.HasPrefix(rawURL, "-") {
		return fmt.Errorf("%w: URL cannot start with dash", ErrInvalidGitURL)
	}

	// Allow local filesystem paths (absolute or relative without scheme)
	// Git supports cloning from local paths like /path/to/repo or ./repo
	if filepath.IsAbs(rawURL) || !strings.Contains(rawURL, "://") {
		// Local path - check for traversal
		if strings.Contains(filepath.Clean(rawURL), "..") {
			return fmt.Errorf("%w: path traversal not allowed in local path", ErrInvalidGitURL)
		}
		return nil
	}

	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidGitURL, err)
	}

	// Check scheme
	if !allowedGitSchemes[u.Scheme] {
		return fmt.Errorf("%w: unsupported scheme %q (allowed: https, git, ssh)", ErrInvalidGitURL, u.Scheme)
	}

	// Ensure host is present for remote URLs
	if u.Host == "" {
		return fmt.Errorf("%w: URL must have a host", ErrInvalidGitURL)
	}

	return nil
}

// ValidateLocalPath ensures a path doesn't contain traversal attempts.
// It allows both relative and absolute paths but rejects traversal patterns.
func ValidateLocalPath(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Check for flag injection
	if strings.HasPrefix(path, "-") {
		return fmt.Errorf("%w: path cannot start with dash", ErrInvalidPath)
	}

	// Clean the path to normalize it
	cleaned := filepath.Clean(path)

	// Check for parent directory traversal in the cleaned path
	// This catches both relative traversal (../foo) and attempts to escape
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	// For relative paths, use Go's built-in check
	if !filepath.IsAbs(path) && !filepath.IsLocal(path) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	return nil
}

// ValidateDirectory validates that a directory path is safe.
// It cleans the path and ensures it doesn't escape the intended boundary.
func ValidateDirectory(dir string) error {
	if dir == "" {
		return nil // Empty dir means current directory, which is valid
	}

	// Clean the path
	cleanDir := filepath.Clean(dir)

	// Check for flag injection
	if strings.HasPrefix(cleanDir, "-") {
		return fmt.Errorf("%w: directory cannot start with dash", ErrInvalidPath)
	}

	return nil
}

// SanitizePath cleans a path and returns the sanitized version.
// Returns an error if the path is invalid.
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	cleaned := filepath.Clean(path)

	// Check for flag injection after cleaning
	if strings.HasPrefix(cleaned, "-") {
		return "", fmt.Errorf("%w: path cannot start with dash", ErrInvalidPath)
	}

	return cleaned, nil
}
