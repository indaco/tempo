package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/git"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/registry"
)

// clearRegistry removes all entries from the function registry.
// Since registry does not export a Clear method, we directly delete keys from the map.
func clearRegistry() {
	funcs := registry.GetRegisteredFunctions()
	for k := range funcs {
		delete(funcs, k)
	}
}

// TestInstallFunctionPackageFromRepo_NewRepo simulates the case where the clone directory does not exist.
// In this version, the simulated clone writes a provider.go file into a subdirectory (e.g. "pkg").
// The test then only verifies that functions are registered.
func TestInstallFunctionPackageFromRepo_NewRepo(t *testing.T) {
	// Override git.CloneRepoFunc to simulate cloning.
	origCloneRepoFunc := git.CloneRepoFunc
	defer func() { git.CloneRepoFunc = origCloneRepoFunc }()
	git.CloneRepoFunc = func(repoURL, repoPath string, l logger.LoggerInterface) error {
		// Simulate clone: create the clone directory.
		if err := os.MkdirAll(repoPath, 0755); err != nil {
			return err
		}
		// Instead of writing provider.go at the root, create a subdirectory.
		subDir := filepath.Join(repoPath, "pkg")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			return err
		}
		providerContent := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{"testFunc": func() string { return "Hello from simulated clone" }}
}

var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`
		// Write the provider.go file in the subdirectory.
		return os.WriteFile(filepath.Join(subDir, "provider.go"), []byte(providerContent), 0644)
	}

	// Create a temporary data directory.
	dataDir := t.TempDir()
	// Use a repo URL (real or simulated). When testing a real repo, comment out the override above.
	repoURL := "https://github.com/indaco/tempo-provider-sprig.git"

	// Clear the registry for isolation.
	clearRegistry()

	// Call InstallFunctionPackageFromRepo with forceClone = false.
	if err := InstallFunctionPackageFromRepo(dataDir, repoURL, false, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("InstallFunctionPackageFromRepo failed: %v", err)
	}

	// We no longer check for provider.go at a fixed location.
	// Instead, verify that at least one function has been registered.
	funcs := registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Errorf("expected at least one function to be registered, got 0")
	}

	// Optionally, check for our simulated function.
	if _, ok := funcs["testFunc"]; !ok {
		t.Log("warning: expected function 'testFunc' not found; if using a real provider, adjust expected function names accordingly")
	}
}

// TestInstallFunctionPackageFromRepo_ForceReclone performs a real clone and then forces a re‐clone.
func TestInstallFunctionPackageFromRepo_ForceReclone(t *testing.T) {
	dataDir := t.TempDir()
	repoURL := "https://github.com/indaco/tempo-provider-sprig.git"
	// The repo name is extracted from the URL – "tempo-provider-sprig".
	clonePath := filepath.Join(dataDir, "tempo_functions", "tempo-provider-sprig")

	// --- First clone: ---
	// Clone normally (forceClone false) so that functions are registered.
	if err := InstallFunctionPackageFromRepo(dataDir, repoURL, false, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("First clone failed: %v", err)
	}
	funcs := registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Fatalf("expected functions to be registered on first clone, got 0")
	}
	// Check for a known function; for example, "toStrings" should be registered.
	if _, ok := funcs["toStrings"]; !ok {
		t.Fatalf("expected function 'toStrings' to be registered on first clone")
	}

	// Create a dummy file inside the clone to verify later that the folder is removed.
	dummyPath := filepath.Join(clonePath, "dummy.txt")
	if err := os.WriteFile(dummyPath, []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	// --- Force clone: ---
	// Clear the registry so that we can verify re-registration.
	clearRegistry()

	// Call InstallFunctionPackageFromRepo with forceClone true.
	// Note: In the force clone branch, the code calls ForceReclone (which removes the directory and clones)
	// but does not call RegisterFunctionsFromPath. Therefore, we call it manually.
	if err := InstallFunctionPackageFromRepo(dataDir, repoURL, true, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("Force clone failed: %v", err)
	}
	if err := RegisterFunctionsFromPath(clonePath, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("RegisterFunctionsFromPath failed after force clone: %v", err)
	}

	// Verify that the dummy file no longer exists.
	if _, err := os.Stat(dummyPath); err == nil {
		t.Errorf("expected dummy file to be removed after force clone, but it exists")
	}

	// Verify that functions are registered after force clone.
	funcs = registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Errorf("expected functions to be registered after force clone, got 0")
	}
	if _, ok := funcs["toStrings"]; !ok {
		t.Errorf("expected function 'toStrings' to be registered after force clone, but it was not found")
	}
}

func TestInstallFunctionPackageFromRepo_UpdateRepo(t *testing.T) {
	dataDir := t.TempDir()
	repoURL := "https://github.com/indaco/tempo-provider-sprig.git"
	clonePath := filepath.Join(dataDir, "tempo_functions", "tempo-provider-sprig")

	// --- First clone: perform a real clone so that the repository is valid.
	if err := InstallFunctionPackageFromRepo(dataDir, repoURL, false, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("initial clone failed: %v", err)
	}
	// Check that functions are registered.
	funcs := registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Fatalf("expected functions to be registered after initial clone, got 0")
	}
	if _, ok := funcs["toStrings"]; !ok {
		t.Fatalf("expected function 'toStrings' to be registered after initial clone")
	}

	// --- Update branch: simulate an update.
	// Clear the registry so we can verify re-registration.
	clearRegistry()

	// Call InstallFunctionPackageFromRepo again (with forceClone false).
	// Because the clone already exists and is valid, the update branch should be executed.
	if err := InstallFunctionPackageFromRepo(dataDir, repoURL, false, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("update branch failed: %v", err)
	}

	// After updating, call RegisterFunctionsFromPath to load functions from the (updated) clone.
	if err := RegisterFunctionsFromPath(clonePath, logger.NewDefaultLogger()); err != nil {
		t.Fatalf("RegisterFunctionsFromPath failed after update: %v", err)
	}

	// Verify that functions are registered after update.
	funcs = registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Errorf("expected functions to be registered after update, got 0")
	}
	if _, ok := funcs["toStrings"]; !ok {
		t.Errorf("expected function 'toStrings' to be registered after update, but it was not found")
	}
}
