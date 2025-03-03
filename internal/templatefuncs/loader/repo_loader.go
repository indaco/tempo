package loader

import (
	"path/filepath"

	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/git"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/utils"
)

// InstallFunctionPackageFromRepo fetches and installs an external function package from a repository.
func InstallFunctionPackageFromRepo(
	dataDir, repoURL string,
	forceClone bool,
	logger logger.LoggerInterface,
) error {
	logger.Info("Checking repository state").WithAttrs("repo_url", repoURL)

	// Extract repo name and define clone path
	repoName := utils.ExtractNameFromURL(repoURL)
	clonePath := filepath.Join(dataDir, "tempo_functions", repoName)

	exists, err := utils.DirExists(clonePath)
	if err != nil {
		return err
	}

	if exists {
		// If it's a valid Git repo
		if git.IsValidGitRepo(clonePath) {
			if forceClone {
				return git.ForceReclone(repoURL, clonePath, logger)
			}
			return git.UpdateRepo(clonePath, logger)
		}

		// If it's not a valid Git repo
		return errors.Wrap("Function provider folder already exists but is not a valid repository. Use --force to re-clone or manually remove the folder.")
	}

	// Step 1: Clone the repo
	if err := git.CloneRepo(repoURL, clonePath, logger); err != nil {
		return err
	}

	// Step 2: Register the functions from the cloned repo
	return RegisterFunctionsFromPath(clonePath, logger)
}
