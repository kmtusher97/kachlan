package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Unit tests
// ---------------------------------------------------------------------------

func TestCLIRequiresArg(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no arguments provided")
	}
}

func TestCLIRejectsExtraArgs(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"a.mp4", "b.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments provided")
	}
}

func TestCLIMissingInput(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"nonexistent_file.mp4"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent input")
	}
}

// ---------------------------------------------------------------------------
// Integration tests — require ffmpeg
// ---------------------------------------------------------------------------

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found, skipping integration test")
	}
}

// generateTestVideo creates a tiny 1-second test video and returns its path.
func generateTestVideo(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "testsrc=duration=1:size=64x64:rate=1",
		"-f", "lavfi", "-i", "sine=frequency=220:duration=1",
		"-c:v", "libx264", "-c:a", "aac",
		"-shortest", "-y", path,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to generate test video %s: %v\n%s", name, err, out)
	}
	return path
}

func TestCompressSingleFile(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	input := generateTestVideo(t, dir, "sample.mp4")

	err := compressSingle(input, 28, "ultrafast", "")
	if err != nil {
		t.Fatalf("compressSingle failed: %v", err)
	}

	expected := filepath.Join(dir, "sample-compressed.mp4")
	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("expected output file %s not found: %v", expected, err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
}

func TestCompressSingleFileCustomOutput(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	input := generateTestVideo(t, dir, "sample.mp4")
	output := filepath.Join(dir, "my-output.mp4")

	err := compressSingle(input, 28, "ultrafast", output)
	if err != nil {
		t.Fatalf("compressSingle failed: %v", err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("expected output file %s not found: %v", output, err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
}

func TestCompressSingleFileWithSpaces(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	input := generateTestVideo(t, dir, "my vacation video.mp4")

	err := compressSingle(input, 28, "ultrafast", "")
	if err != nil {
		t.Fatalf("compressSingle failed: %v", err)
	}

	expected := filepath.Join(dir, "my vacation video-compressed.mp4")
	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("expected output file %s not found: %v", expected, err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
}

func TestCompressFolderWithSpacesInNames(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "my home videos")
	subDir := filepath.Join(srcDir, "summer trip")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	generateTestVideo(t, srcDir, "family dinner.mp4")
	generateTestVideo(t, subDir, "beach day 2024.mov")

	err := compressFolder(srcDir, 28, "ultrafast", "", 1)
	if err != nil {
		t.Fatalf("compressFolder failed: %v", err)
	}

	outDir := srcDir + "-compressed"
	for _, rel := range []string{"family dinner.mp4", "summer trip/beach day 2024.mov"} {
		p := filepath.Join(outDir, rel)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected output %s not found: %v", p, err)
		}
	}
}

func TestCompressFolderCreatesCompressedFolder(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "my-videos")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	generateTestVideo(t, srcDir, "a.mp4")
	generateTestVideo(t, srcDir, "b.mov")

	err := compressFolder(srcDir, 28, "ultrafast", "", 1)
	if err != nil {
		t.Fatalf("compressFolder failed: %v", err)
	}

	outDir := srcDir + "-compressed"
	for _, name := range []string{"a.mp4", "b.mov"} {
		p := filepath.Join(outDir, name)
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("expected output %s not found: %v", p, err)
		} else if info.Size() == 0 {
			t.Errorf("output %s is empty", p)
		}
	}
}

func TestCompressFolderCustomOutput(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "videos")
	outDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	generateTestVideo(t, srcDir, "clip.mp4")

	err := compressFolder(srcDir, 28, "ultrafast", outDir, 1)
	if err != nil {
		t.Fatalf("compressFolder failed: %v", err)
	}

	p := filepath.Join(outDir, "clip.mp4")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected output %s not found: %v", p, err)
	}
}

func TestCompressFolderPreservesSubdirs(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "media")
	subDir := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	generateTestVideo(t, srcDir, "top.mp4")
	generateTestVideo(t, subDir, "nested.mp4")

	err := compressFolder(srcDir, 28, "ultrafast", "", 1)
	if err != nil {
		t.Fatalf("compressFolder failed: %v", err)
	}

	outDir := srcDir + "-compressed"
	for _, rel := range []string{"top.mp4", "sub/nested.mp4"} {
		p := filepath.Join(outDir, rel)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected output %s not found: %v", p, err)
		}
	}
}

func TestCompressFolderSkipsNonVideoFiles(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "mixed")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	generateTestVideo(t, srcDir, "real.mp4")
	if err := os.WriteFile(filepath.Join(srcDir, "notes.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "photo.jpg"), []byte{0xFF, 0xD8}, 0644); err != nil {
		t.Fatal(err)
	}

	err := compressFolder(srcDir, 28, "ultrafast", "", 1)
	if err != nil {
		t.Fatalf("compressFolder failed: %v", err)
	}

	outDir := srcDir + "-compressed"

	// Video should exist
	if _, err := os.Stat(filepath.Join(outDir, "real.mp4")); err != nil {
		t.Error("expected compressed video not found")
	}
	// Non-videos should NOT exist
	if _, err := os.Stat(filepath.Join(outDir, "notes.txt")); err == nil {
		t.Error("non-video file notes.txt should not be in output")
	}
	if _, err := os.Stat(filepath.Join(outDir, "photo.jpg")); err == nil {
		t.Error("non-video file photo.jpg should not be in output")
	}
}

func TestCompressFolderEmptyReturnsError(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "empty")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add non-video files only
	if err := os.WriteFile(filepath.Join(srcDir, "readme.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	err := compressFolder(srcDir, 28, "ultrafast", "", 1)
	if err == nil {
		t.Fatal("expected error for folder with no video files")
	}
}

func TestCompressFolderParallelWorkers(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "parallel")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	generateTestVideo(t, srcDir, "v1.mp4")
	generateTestVideo(t, srcDir, "v2.mp4")
	generateTestVideo(t, srcDir, "v3.mp4")

	err := compressFolder(srcDir, 28, "ultrafast", "", 3)
	if err != nil {
		t.Fatalf("compressFolder with workers=3 failed: %v", err)
	}

	outDir := srcDir + "-compressed"
	for _, name := range []string{"v1.mp4", "v2.mp4", "v3.mp4"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Errorf("expected output %s not found: %v", name, err)
		}
	}
}

func TestCompressFolderOriginalUntouched(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "originals")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	input := generateTestVideo(t, srcDir, "keep.mp4")
	before, _ := os.Stat(input)

	err := compressFolder(srcDir, 28, "ultrafast", "", 1)
	if err != nil {
		t.Fatalf("compressFolder failed: %v", err)
	}

	after, err := os.Stat(input)
	if err != nil {
		t.Fatal("original file was deleted")
	}
	if before.Size() != after.Size() {
		t.Error("original file was modified")
	}
}
