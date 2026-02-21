package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/indaco/tempo/internal/processor/transformers"
)

const (
	benchCSS = `.button {
	color: blue;
	background-color: white;
	border: 1px solid #ccc;
	padding: 8px 16px;
	font-size: 14px;
	border-radius: 4px;
	cursor: pointer;
}
.button:hover {
	background-color: #f0f0f0;
	border-color: #aaa;
}
.button:active {
	background-color: #e0e0e0;
}`

	benchJS = `function greet(name) {
	var greeting = "Hello, " + name + "!";
	console.log(greeting);
	return greeting;
}
function add(a, b) {
	return a + b;
}
function multiply(a, b) {
	return a * b;
}
greet("World");`

	benchTemplContent = `/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */`
)

func BenchmarkPassthroughProcessor_CSS(b *testing.B) {
	dir := b.TempDir()
	inputPath := filepath.Join(dir, "input.css")
	outputPath := filepath.Join(dir, "output.templ")

	if err := os.WriteFile(inputPath, []byte(benchCSS), 0644); err != nil {
		b.Fatalf("failed to write input: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
		b.Fatalf("failed to write output: %v", err)
	}

	p := &PassthroughProcessor{}
	b.ResetTimer()

	for b.Loop() {
		// Restore output file between iterations so markers are always present.
		if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
			b.Fatalf("failed to restore output: %v", err)
		}
		if err := p.Process(inputPath, outputPath, "tempo"); err != nil {
			b.Fatalf("Process failed: %v", err)
		}
	}
}

func BenchmarkPassthroughProcessor_JS(b *testing.B) {
	dir := b.TempDir()
	inputPath := filepath.Join(dir, "input.js")
	outputPath := filepath.Join(dir, "output.templ")

	if err := os.WriteFile(inputPath, []byte(benchJS), 0644); err != nil {
		b.Fatalf("failed to write input: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
		b.Fatalf("failed to write output: %v", err)
	}

	p := &PassthroughProcessor{}
	b.ResetTimer()

	for b.Loop() {
		if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
			b.Fatalf("failed to restore output: %v", err)
		}
		if err := p.Process(inputPath, outputPath, "tempo"); err != nil {
			b.Fatalf("Process failed: %v", err)
		}
	}
}

func BenchmarkMinifierProcessor_CSS(b *testing.B) {
	dir := b.TempDir()
	inputPath := filepath.Join(dir, "input.css")
	outputPath := filepath.Join(dir, "output.templ")

	if err := os.WriteFile(inputPath, []byte(benchCSS), 0644); err != nil {
		b.Fatalf("failed to write input: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
		b.Fatalf("failed to write output: %v", err)
	}

	transformer := &transformers.EsbuildTransformer{Loader: api.LoaderCSS}
	p := &MinifierProcessor{Transform: transformer.Transform}
	b.ResetTimer()

	for b.Loop() {
		if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
			b.Fatalf("failed to restore output: %v", err)
		}
		if err := p.Process(inputPath, outputPath, "tempo"); err != nil {
			b.Fatalf("Process failed: %v", err)
		}
	}
}

func BenchmarkMinifierProcessor_JS(b *testing.B) {
	dir := b.TempDir()
	inputPath := filepath.Join(dir, "input.js")
	outputPath := filepath.Join(dir, "output.templ")

	if err := os.WriteFile(inputPath, []byte(benchJS), 0644); err != nil {
		b.Fatalf("failed to write input: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
		b.Fatalf("failed to write output: %v", err)
	}

	transformer := &transformers.EsbuildTransformer{Loader: api.LoaderJS}
	p := &MinifierProcessor{Transform: transformer.Transform}
	b.ResetTimer()

	for b.Loop() {
		if err := os.WriteFile(outputPath, []byte(benchTemplContent), 0644); err != nil {
			b.Fatalf("failed to restore output: %v", err)
		}
		if err := p.Process(inputPath, outputPath, "tempo"); err != nil {
			b.Fatalf("Process failed: %v", err)
		}
	}
}
