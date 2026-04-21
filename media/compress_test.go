package media

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCompressVideo(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	input := generateTestVideo(t, dir, "sample.mp4")
	output := filepath.Join(dir, "out.mp4")

	err := CompressVideo(context.Background(), input, output, 28, "ultrafast", nil)
	if err != nil {
		t.Fatalf("CompressVideo() failed: %v", err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("output not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}
}

func TestCompressVideoWithProgress(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	input := generateTestVideo(t, dir, "sample.mp4")
	output := filepath.Join(dir, "out.mp4")

	var maxPct float64
	err := CompressVideo(context.Background(), input, output, 28, "ultrafast", func(pct float64) {
		if pct > maxPct {
			maxPct = pct
		}
	})
	if err != nil {
		t.Fatalf("CompressVideo() with progress failed: %v", err)
	}

	if maxPct < 50 {
		t.Errorf("expected progress to reach at least 50%%, got %.1f%%", maxPct)
	}
}

func TestCompressVideoCancellation(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	input := generateTestVideo(t, dir, "sample.mp4")
	output := filepath.Join(dir, "out.mp4")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := CompressVideo(ctx, input, output, 28, "ultrafast", nil)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
