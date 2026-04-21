package media

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsVideo(t *testing.T) {
	yes := []string{
		"clip.mp4", "clip.avi", "clip.mov", "clip.mkv",
		"clip.wmv", "clip.flv", "clip.webm", "clip.m4v",
		"clip.mpg", "clip.mpeg", "clip.3gp", "clip.ts",
	}
	for _, f := range yes {
		if !IsVideo(f) {
			t.Errorf("expected %q to be recognized as video", f)
		}
	}

	no := []string{
		"photo.jpg", "doc.pdf", "song.mp3", "notes.txt",
		"archive.zip", "noext", "",
	}
	for _, f := range no {
		if IsVideo(f) {
			t.Errorf("expected %q to NOT be recognized as video", f)
		}
	}
}

func TestIsVideoCaseInsensitive(t *testing.T) {
	cases := []string{"video.MP4", "video.Mp4", "video.MOV", "video.Mkv"}
	for _, f := range cases {
		if !IsVideo(f) {
			t.Errorf("expected %q to be recognized as video (case insensitive)", f)
		}
	}
}

func TestDefaultOutputPath(t *testing.T) {
	tests := []struct {
		input, suffix, want string
	}{
		{"video.mp4", "compressed", "video-compressed.mp4"},
		{"/tmp/my video.mov", "compressed", "/tmp/my video-compressed.mov"},
		{"a.mkv", "cropped", "a-cropped.mkv"},
	}
	for _, tt := range tests {
		got := DefaultOutputPath(tt.input, tt.suffix)
		if got != tt.want {
			t.Errorf("DefaultOutputPath(%q, %q) = %q, want %q", tt.input, tt.suffix, got, tt.want)
		}
	}
}

func TestDefaultOutputDir(t *testing.T) {
	tests := []struct {
		input, suffix, want string
	}{
		{"my-videos", "compressed", "my-videos-compressed"},
		{"/tmp/clips/", "compressed", "/tmp/clips-compressed"},
	}
	for _, tt := range tests {
		got := DefaultOutputDir(tt.input, tt.suffix)
		if got != tt.want {
			t.Errorf("DefaultOutputDir(%q, %q) = %q, want %q", tt.input, tt.suffix, got, tt.want)
		}
	}
}

func TestFindVideos(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"a.mp4", "b.mov"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("fake"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("text"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "c.mkv"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	videos, err := FindVideos(dir)
	if err != nil {
		t.Fatalf("FindVideos: %v", err)
	}
	if len(videos) != 3 {
		t.Errorf("got %d videos, want 3", len(videos))
	}
}

func TestFindVideosEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	videos, err := FindVideos(dir)
	if err != nil {
		t.Fatalf("FindVideos: %v", err)
	}
	if len(videos) != 0 {
		t.Errorf("got %d videos, want 0", len(videos))
	}
}

func TestComputeSizes(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.bin")
	output := filepath.Join(dir, "output.bin")

	if err := os.WriteFile(input, make([]byte, 1024*1024), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, make([]byte, 512*1024), 0644); err != nil {
		t.Fatal(err)
	}

	s := ComputeSizes(input, output)
	if s.InputMB < 0.99 || s.InputMB > 1.01 {
		t.Errorf("InputMB = %.2f, want ~1.0", s.InputMB)
	}
	if s.OutputMB < 0.49 || s.OutputMB > 0.51 {
		t.Errorf("OutputMB = %.2f, want ~0.5", s.OutputMB)
	}
	if s.ReductionPct < 49 || s.ReductionPct > 51 {
		t.Errorf("ReductionPct = %.1f%%, want ~50%%", s.ReductionPct)
	}
}

func TestComputeSizesMissing(t *testing.T) {
	s := ComputeSizes("/nonexistent/a", "/nonexistent/b")
	if s.InputMB != 0 || s.OutputMB != 0 || s.ReductionPct != 0 {
		t.Error("expected zero values for nonexistent files")
	}
}

func TestCheckFFmpeg(t *testing.T) {
	_ = CheckFFmpeg()
}

// requireFFmpeg skips the test if ffmpeg is not installed.
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

func TestDurationUs(t *testing.T) {
	requireFFmpeg(t)

	dir := t.TempDir()
	generateTestVideo(t, dir, "test.mp4")

	dur := DurationUs(filepath.Join(dir, "test.mp4"))
	if dur <= 0 {
		t.Fatalf("expected positive duration, got %d", dur)
	}
	if dur < 500_000 || dur > 2_000_000 {
		t.Errorf("duration = %d us, expected ~1_000_000 us for 1s video", dur)
	}
}

func TestDurationUsNonexistent(t *testing.T) {
	dur := DurationUs("/nonexistent/file.mp4")
	if dur != 0 {
		t.Errorf("expected 0 for nonexistent file, got %d", dur)
	}
}
