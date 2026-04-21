package media

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var videoExts = map[string]bool{
	".mp4": true, ".avi": true, ".mov": true, ".mkv": true,
	".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	".mpg": true, ".mpeg": true, ".3gp": true, ".ts": true,
}

// IsVideo returns true if the file has a recognized video extension.
func IsVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return videoExts[ext]
}

// FindVideos walks a directory and returns paths to all video files.
func FindVideos(dir string) ([]string, error) {
	var videos []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && IsVideo(path) {
			videos = append(videos, path)
		}
		return nil
	})
	return videos, err
}

// DurationUs returns the video duration in microseconds using ffprobe, or 0 if unknown.
func DurationUs(path string) int64 {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return int64(seconds * 1_000_000)
}

// CheckFFmpeg returns true if ffmpeg is available in PATH.
func CheckFFmpeg() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// SizeInfo holds file size comparison data.
type SizeInfo struct {
	InputMB      float64
	OutputMB     float64
	ReductionPct float64
}

// ComputeSizes returns size comparison between input and output files.
func ComputeSizes(input, output string) SizeInfo {
	var s SizeInfo
	inInfo, err1 := os.Stat(input)
	outInfo, err2 := os.Stat(output)
	if err1 == nil && err2 == nil {
		s.InputMB = float64(inInfo.Size()) / 1024 / 1024
		s.OutputMB = float64(outInfo.Size()) / 1024 / 1024
		if s.InputMB > 0 {
			s.ReductionPct = (1 - s.OutputMB/s.InputMB) * 100
		}
	}
	return s
}

// ProgressFunc is called during an operation with the percentage complete (0-100).
type ProgressFunc func(percent float64)

// DefaultOutputPath returns the default output path with a suffix.
// e.g., DefaultOutputPath("video.mp4", "compressed") → "video-compressed.mp4"
func DefaultOutputPath(input, suffix string) string {
	ext := filepath.Ext(input)
	base := strings.TrimSuffix(input, ext)
	return base + "-" + suffix + ext
}

// DefaultOutputDir returns the default output directory with a suffix.
// e.g., DefaultOutputDir("my-videos", "compressed") → "my-videos-compressed"
func DefaultOutputDir(input, suffix string) string {
	return filepath.Clean(input) + "-" + suffix
}
