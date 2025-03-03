package transformers

import (
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/indaco/tempo/testutils"
)

func TestEsbuildTransformer_JSMinification(t *testing.T) {
	transformer := &EsbuildTransformer{Loader: api.LoaderJS}
	input := `function test() {
		console.log("Hello World");
	}`

	expected := `function test(){console.log("Hello World")}`

	minified, err := transformer.Transform(input)
	if err != nil {
		t.Fatalf("Unexpected error during JS minification: %v", err)
	}

	// Normalize whitespace
	minified = strings.TrimSpace(minified)

	// Validate minified output
	if minified != expected {
		t.Errorf("Minified JS mismatch.\nExpected:\n%s\nGot:\n%s", expected, minified)
	}
}

func TestEsbuildTransformer_CSSMinification(t *testing.T) {
	transformer := &EsbuildTransformer{Loader: api.LoaderCSS}
	input := `.button {
		color: blue;
		font-size: 16px;
	}`

	expectedOutputs := []string{
		`.button{color:blue;font-size:16px}`, // Expected output
		`.button{color:#00f;font-size:16px}`, // Esbuild normalization
	}

	minified, err := transformer.Transform(input)
	if err != nil {
		t.Fatalf("Unexpected error during CSS minification: %v", err)
	}

	// Check if output matches any expected variant
	matchFound := false
	for _, expected := range expectedOutputs {
		// Normalize whitespace
		minified = strings.TrimSpace(minified)
		if minified == expected {
			matchFound = true
			break
		}
	}

	if !matchFound {
		t.Errorf("Minified CSS mismatch.\nExpected one of:\n%s\nGot:\n%s", strings.Join(expectedOutputs, "\nOR\n"), minified)
	}
}

func TestEsbuildTransformer_InvalidJS(t *testing.T) {
	transformer := &EsbuildTransformer{Loader: api.LoaderJS}
	input := `function test( { console.log("Invalid syntax")`

	_, err := transformer.Transform(input)

	if err == nil {
		t.Fatal("Expected error for invalid JavaScript, but got none")
	}

	expectedErr := "esbuild minification error"
	if !testutils.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error message to contain %q, but got: %v", expectedErr, err)
	}
}

func TestEsbuildTransformer_InvalidCSS(t *testing.T) {
	transformer := &EsbuildTransformer{Loader: api.LoaderCSS}

	// Input with invalid CSS syntax
	input := `.button { color: ; missing_property; }`
	output, err := transformer.Transform(input)
	// Normalize whitespace
	output = strings.TrimSpace(output)

	// Esbuild does NOT throw an error for invalid CSS.
	// Instead, we verify if the minified output is still incorrect.
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedOutput := `.button{color:;missing_property}`
	if output != expectedOutput {
		t.Errorf("Unexpected minified output.\nExpected: %q\nGot: %q", expectedOutput, output)
	}
}

func TestEsbuildTransformer_UnsupportedLoader(t *testing.T) {
	transformer := &EsbuildTransformer{Loader: api.LoaderNone}
	input := `console.log("No minification");`

	output, err := transformer.Transform(input)
	// Normalize whitespace
	output = strings.TrimSpace(output)

	if err != nil {
		t.Fatalf("Unexpected error for unsupported loader: %v", err)
	}

	// Output should remain unchanged
	if output != input {
		t.Errorf("Expected unmodified input for unsupported loader.\nExpected:\n%s\nGot:\n%s", input, output)
	}
}
