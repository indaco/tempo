package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/tempo/internal/cmdrunner"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/validation"
)

// GitOperations defines the interface for git operations.
// This allows for dependency injection and easier testing.
type GitOperations interface {
	CloneOrUpdate(repoURL, repoPath string, logger logger.LoggerInterface) error
	Update(repoPath string, logger logger.LoggerInterface) error
	Clone(repoURL, repoPath string, logger logger.LoggerInterface) error
	ForceReclone(repoURL, repoPath string, logger logger.LoggerInterface) error
	IsValidRepo(repoPath string) bool
}

// DefaultGitOps is the default implementation of GitOperations.
type DefaultGitOps struct{}

// NewGitOperations creates a new GitOperations instance.
func NewGitOperations() GitOperations {
	return &DefaultGitOps{}
}

// Function variables to allow mocking in tests (kept for backward compatibility)
// Deprecated: Use GitOperations interface instead for new code.
var (
	CloneOrUpdate = DefaultCloneOrUpdate
	UpdateRepo    = DefaultUpdateRepo
	CloneRepoFunc = CloneRepo
)

// Default implementation of CloneOrUpdate
func DefaultCloneOrUpdate(repoURL, repoPath string, logger logger.LoggerInterface) error {
	if IsValidGitRepo(repoPath) {
		logger.Info("Updating existing plugin repository").WithAttrs("repo_path", repoPath)
		return UpdateRepo(repoPath, logger)
	}

	return CloneRepoFunc(repoURL, repoPath, logger)
}

// Default implementation of UpdateRepo
func DefaultUpdateRepo(repoPath string, logger logger.LoggerInterface) error {
	logger.Info("Updating existing repository").WithAttrs("repo_path", repoPath)
	return cmdrunner.RunCommand(repoPath, "git", "pull")
}

// CloneRepo clones a Git repository into the specified path.
// It validates the URL and path to prevent command injection attacks.
func CloneRepo(repoURL, repoPath string, logger logger.LoggerInterface) error {
	// Validate URL to prevent command injection
	if err := validation.ValidateGitURL(repoURL); err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Validate path to prevent path traversal
	if err := validation.ValidateLocalPath(repoPath); err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	logger.Info("Cloning repository").WithAttrs("repo_url", repoURL)
	// Use "--" separator to prevent flag injection
	return cmdrunner.RunCommand(".", "git", "clone", "--", repoURL, repoPath)
}

// ForceReclone removes an existing repository and re-clones it.
// It validates the URL and path before performing any operations.
func ForceReclone(repoURL, repoPath string, logger logger.LoggerInterface) error {
	// Validate URL first (CloneRepo will also validate, but fail fast)
	if err := validation.ValidateGitURL(repoURL); err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	// Validate path before removal
	if err := validation.ValidateLocalPath(repoPath); err != nil {
		return fmt.Errorf("invalid repository path: %w", err)
	}

	logger.Warning("Force-cloning repository. Removing existing folder").WithAttrs("repo_path", repoPath)

	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove existing repository: %w", err)
	}

	return CloneRepo(repoURL, repoPath, logger)
}

// IsValidGitRepo checks if the given path is a valid Git repository.
func IsValidGitRepo(repoPath string) bool {
	gitPath := filepath.Join(repoPath, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}

// Interface implementations for DefaultGitOps

// CloneOrUpdate implements GitOperations.CloneOrUpdate.
func (g *DefaultGitOps) CloneOrUpdate(repoURL, repoPath string, logger logger.LoggerInterface) error {
	return DefaultCloneOrUpdate(repoURL, repoPath, logger)
}

// Update implements GitOperations.Update.
func (g *DefaultGitOps) Update(repoPath string, logger logger.LoggerInterface) error {
	return DefaultUpdateRepo(repoPath, logger)
}

// Clone implements GitOperations.Clone.
func (g *DefaultGitOps) Clone(repoURL, repoPath string, logger logger.LoggerInterface) error {
	return CloneRepo(repoURL, repoPath, logger)
}

// ForceReclone implements GitOperations.ForceReclone.
func (g *DefaultGitOps) ForceReclone(repoURL, repoPath string, logger logger.LoggerInterface) error {
	return ForceReclone(repoURL, repoPath, logger)
}

// IsValidRepo implements GitOperations.IsValidRepo.
func (g *DefaultGitOps) IsValidRepo(repoPath string) bool {
	return IsValidGitRepo(repoPath)
}

// Ensure DefaultGitOps implements GitOperations
var _ GitOperations = (*DefaultGitOps)(nil)
