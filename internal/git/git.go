package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/tempo/internal/cmdrunner"
	"github.com/indaco/tempo/internal/logger"
)

// Function variables to allow mocking in tests
var (
	CloneOrUpdate = DefaultCloneOrUpdate
	UpdateRepo    = DefaultUpdateRepo
)

// Default implementation of CloneOrUpdate
func DefaultCloneOrUpdate(repoURL, repoPath string, logger logger.LoggerInterface) error {
	if IsValidGitRepo(repoPath) {
		logger.Info("Updating existing plugin repository").WithAttrs("repo_path", repoPath)
		return UpdateRepo(repoPath, logger)
	}

	return CloneRepo(repoURL, repoPath, logger)
}

// Default implementation of UpdateRepo
func DefaultUpdateRepo(repoPath string, logger logger.LoggerInterface) error {
	logger.Info("Updating existing repository").WithAttrs("repo_path", repoPath)
	return cmdrunner.RunCommand(repoPath, "git", "pull")
}

// CloneRepo clones a Git repository into the specified path.
func CloneRepo(repoURL, repoPath string, logger logger.LoggerInterface) error {
	logger.Info("Cloning repository").WithAttrs("repo_url", repoURL)
	return cmdrunner.RunCommand(".", "git", "clone", repoURL, repoPath)
}

// ForceReclone removes an existing repository and re-clones it.
func ForceReclone(repoURL, repoPath string, logger logger.LoggerInterface) error {
	logger.Warning("Force-cloning repository. Removing existing folder").WithAttrs("repo_path", repoPath)

	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove existing repository: %s", err)
	}

	return CloneRepo(repoURL, repoPath, logger)
}

// IsValidGitRepo checks if the given path is a valid Git repository.
func IsValidGitRepo(repoPath string) bool {
	gitPath := filepath.Join(repoPath, ".git")
	_, err := os.Stat(gitPath)
	return err == nil
}
