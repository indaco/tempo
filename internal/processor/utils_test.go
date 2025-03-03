package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/processor/transformers"
	"github.com/indaco/tempo/testutils"
)

func TestProcessWithTransformation_Success(t *testing.T) {
	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "output.templ")

	// Create an output file with guard markers
	outputContent := `package button

func Button() {
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */
}`
	testutils.CreateFile(t, outputFilePath, outputContent)

	// Define a mock transformation function
	mockTransform := func(input string) (string, error) {
		return "minified-content", nil
	}

	// Define transformation config
	cfg := transformers.TransformationConfig{
		RawData:    ".button { color: blue; }",
		Transform:  mockTransform,
		MarkerName: "tempo",
	}

	// Execute transformation
	err := processWithTransformation(cfg, outputFilePath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read updated output file
	resultContent, err := os.ReadFile(outputFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Expected content after transformation
	expectedContent := `package button

func Button() {
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
minified-content
/* [tempo] END */
}`

	// Validate output
	if string(resultContent) != expectedContent {
		t.Errorf("Expected output:\n%s\nGot:\n%s", expectedContent, string(resultContent))
	}
}

func TestProcessWithTransformation_MissingGuardMarkers(t *testing.T) {
	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "output.templ")

	// Create an output file **without** guard markers
	outputContent := `package button

func Button() {
console.log("Hello world");
}`
	testutils.CreateFile(t, outputFilePath, outputContent)

	// Define a mock transformation function
	mockTransform := func(input string) (string, error) {
		return "minified-content", nil
	}

	// Define transformation config
	cfg := transformers.TransformationConfig{
		RawData:    ".button { color: blue; }",
		Transform:  mockTransform,
		MarkerName: "tempo",
	}

	// Execute transformation (should fail)
	err := processWithTransformation(cfg, outputFilePath)
	if err != nil {
		t.Fatal("Expected error due to missing guard markers, but got none")
	}
}

func TestProcessWithTransformation_TransformError(t *testing.T) {
	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "output.templ")

	// Create an output file with guard markers
	outputContent := `package button

func Button() {
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */
}`
	testutils.CreateFile(t, outputFilePath, outputContent)

	// Define a mock transformation function that returns an error
	mockTransform := func(input string) (string, error) {
		return "", os.ErrInvalid
	}

	// Define transformation config
	cfg := transformers.TransformationConfig{
		RawData:    ".button { color: blue; }",
		Transform:  mockTransform,
		MarkerName: "tempo",
	}

	// Execute transformation (should fail)
	err := processWithTransformation(cfg, outputFilePath)
	if err == nil {
		t.Fatal("Expected error due to transformation failure, but got none")
	}
	expectedErr := "failed to transform content"
	if !testutils.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error message to contain %q, but got: %v", expectedErr, err)
	}
}

func TestProcessWithTransformation_ReadFailure(t *testing.T) {
	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "missing_output.templ")

	// Define a mock transformation function
	mockTransform := func(input string) (string, error) {
		return "minified-content", nil
	}

	// Define transformation config
	cfg := transformers.TransformationConfig{
		RawData:    ".button { color: blue; }",
		Transform:  mockTransform,
		MarkerName: "tempo",
	}

	// Execute transformation (should fail due to missing file)
	err := processWithTransformation(cfg, outputFilePath)
	if err == nil {
		t.Fatal("Expected error due to missing output file, but got none")
	}
	expectedErr := "failed to read output file"
	if !testutils.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error message to contain %q, but got: %v", expectedErr, err)
	}
}
