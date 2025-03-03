package loader

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	tempo_api "github.com/indaco/tempo-api/templatefuncs"
	"github.com/indaco/tempo/internal/logger"
)

// DynamicTemplateFuncProvider implements tempo_api.TemplateFuncProvider
type DynamicTemplateFuncProvider struct {
	functions template.FuncMap
}

// GetFunctions returns registered function names
func (p *DynamicTemplateFuncProvider) GetFunctions() template.FuncMap {
	return p.functions
}

// loadDynamicProvider dynamically loads a Go package, extracts the provider, and registers functions.
func loadDynamicProvider(meta ProviderMetadata, logger logger.LoggerInterface) (tempo_api.TemplateFuncProvider, error) {
	logger.Info("Importing Go package").WithAttrs("package_path", meta.ModuleDir)

	// Step 1: Ensure `provider.go` exists (we still need this for AST parsing)
	if _, err := os.Stat(meta.FilePath); os.IsNotExist(err) {
		fmt.Printf("Cannot find provider.go at: %s", meta.FilePath)
		return nil, fmt.Errorf("invalid function provider: Missing required provider.go file")
	}

	// Step 2: Validate Provider variable presence
	if err := validateProviderPresence(meta.FilePath); err != nil {
		return nil, err
	}

	// Step 3: Build the provider binary
	binPath, err := buildProviderExecutable(meta.ModuleDir, meta.Package)
	if err != nil {
		return nil, err
	}

	// Step 4: Execute provider binary and parse function metadata
	funcMap, err := runProviderExecutable(binPath, meta.ModuleDir)
	if err != nil {
		return nil, err
	}

	logger.Success("Function provider successfully loaded", meta.ModuleDir)
	return &DynamicTemplateFuncProvider{functions: funcMap}, nil
}

// validateProviderPresence checks if "Provider" exists in provider.go using AST parsing.
func validateProviderPresence(providerPath string) error {
	src, err := os.ReadFile(providerPath)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, providerPath, src, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("failed to parse provider.go: %s", err)
	}

	// Inspect AST to find `Provider`
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range valueSpec.Names {
				if name.Name == "Provider" {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("invalid function provider: Missing required exported Provider variable")
}

// buildProviderExecutable compiles the provider package into an executable binary.
func buildProviderExecutable(packagePath, providerPackage string) (string, error) {
	cmdDir := filepath.Join(packagePath, "cmd")
	_ = os.MkdirAll(cmdDir, 0755)

	// Extract the module name from go.mod
	moduleName, err := extractModuleName(packagePath)
	if err != nil {
		return "", fmt.Errorf("failed to determine module name: %s", err)
	}

	// Construct the correct import path
	importPath := moduleName
	if providerPackage != "" && providerPackage != moduleName {
		importPath = moduleName + "/" + providerPackage
	}

	// Construct the correct reference for `Provider`
	providerRef := providerPackage
	if strings.Contains(providerPackage, "/") {
		parts := strings.Split(providerPackage, "/")
		providerRef = parts[len(parts)-1] // Use only the last part
	}

	mainFile := filepath.Join(cmdDir, "main.go")
	mainContent := fmt.Sprintf(`package main

import (
	"encoding/json"
	"fmt"
	"os"
	"%[1]s" // Dynamically determined package
	"github.com/indaco/tempo-api/templatefuncs"
)

func main() {
	if %[2]s.Provider == nil {
		fmt.Println(`+"`"+`{"error": "Provider is nil"}`+"`"+`)
		os.Exit(1)
	}

	provider, ok := %[2]s.Provider.(templatefuncs.TemplateFuncProvider)
	if !ok {
		fmt.Println(`+"`"+`{"error": "Provider does not implement TemplateFuncProvider"}`+"`"+`)
		os.Exit(1)
	}

	functions := provider.GetFunctions()
	funcNames := make([]string, 0, len(functions))
	for name := range functions {
		funcNames = append(funcNames, name)
	}

	data, err := json.Marshal(struct {
		Functions []string `+"`json:\"functions\"`"+`
	}{Functions: funcNames})

	if err != nil {
		fmt.Println(`+"`"+`{"error": "Error encoding JSON"}`+"`"+`)
		os.Exit(1)
	}

	fmt.Println(string(data))
}`, importPath, providerRef)

	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create main.go: %s", err)
	}

	// Ensure go.mod exists
	goModPath := filepath.Join(packagePath, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return "", fmt.Errorf("go.mod not found in provider package: %s", packagePath)
	}

	// Run `go mod tidy`
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = packagePath
	if err := tidyCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run go mod tidy in %s: %s", packagePath, err)
	}

	// Build provider binary
	binPath := filepath.Join(packagePath, "provider_bin")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd")
	buildCmd.Dir = packagePath
	var buildErr bytes.Buffer
	buildCmd.Stderr = &buildErr

	if err := buildCmd.Run(); err != nil {
		if bytes.Contains(buildErr.Bytes(), []byte("missing method GetFunctions")) {
			return "", fmt.Errorf("invalid function provider: Provider does not implement TemplateFuncProvider (missing method GetFunctions)")
		}
		return "", fmt.Errorf("failed to compile provider package: %s", err)
	}

	return binPath, nil
}

// extractModuleName extracts the module name from go.mod.
func extractModuleName(modulePath string) (string, error) {
	goModPath := filepath.Join(modulePath, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", fmt.Errorf("failed to extract module name from go.mod at %s. Check if the file contains a valid module declaration", modulePath)
}

// runProviderExecutable executes the built provider binary and extracts registered functions.
func runProviderExecutable(binPath, packagePath string) (template.FuncMap, error) {
	cmd := exec.Command(binPath)
	cmd.Dir = packagePath
	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		output := out.String()

		switch {
		case bytes.Contains([]byte(output), []byte(`{"error": "Provider is nil"}`)):
			return nil, fmt.Errorf("invalid function provider: Provider is nil")
		case bytes.Contains([]byte(output), []byte(`{"error": "Provider does not implement TemplateFuncProvider"}`)):
			return nil, fmt.Errorf("invalid function provider: Provider does not implement TemplateFuncProvider")
		default:
			return nil, fmt.Errorf("failed to execute provider binary: %s", err)
		}
	}

	var metadata struct {
		Functions []string `json:"functions"`
	}

	if err := json.Unmarshal(out.Bytes(), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse function list from provider: %s", err)
	}

	if len(metadata.Functions) == 0 {
		return nil, errors.New("invalid function provider: No functions found in Provider")
	}

	funcMap := make(template.FuncMap)
	for _, fnName := range metadata.Functions {
		funcMap[fnName] = func() string {
			return fmt.Sprintf("[Function %s Loaded]", fnName)
		}
	}

	return funcMap, nil
}
