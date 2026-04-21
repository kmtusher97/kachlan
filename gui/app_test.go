package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileResult(t *testing.T) {
	dir := t.TempDir()

	input := filepath.Join(dir, "input.bin")
	output := filepath.Join(dir, "output.bin")

	if err := os.WriteFile(input, make([]byte, 1024*1024), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, make([]byte, 512*1024), 0644); err != nil {
		t.Fatal(err)
	}

	app := &App{}
	r := app.fileResult(input, output)

	if r.InputPath != input {
		t.Errorf("InputPath = %q, want %q", r.InputPath, input)
	}
	if r.OutputPath != output {
		t.Errorf("OutputPath = %q, want %q", r.OutputPath, output)
	}
	if r.InputSize < 0.99 || r.InputSize > 1.01 {
		t.Errorf("InputSize = %.2f MB, want ~1.0 MB", r.InputSize)
	}
	if r.OutputSize < 0.49 || r.OutputSize > 0.51 {
		t.Errorf("OutputSize = %.2f MB, want ~0.5 MB", r.OutputSize)
	}
	if r.Reduction < 49 || r.Reduction > 51 {
		t.Errorf("Reduction = %.1f%%, want ~50%%", r.Reduction)
	}
}

func TestFileResultMissingFiles(t *testing.T) {
	app := &App{}
	r := app.fileResult("/nonexistent/input.mp4", "/nonexistent/output.mp4")

	if r.InputSize != 0 || r.OutputSize != 0 || r.Reduction != 0 {
		t.Error("expected zero sizes for nonexistent files")
	}
}
