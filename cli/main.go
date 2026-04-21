package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var version = "dev"

var videoExts = map[string]bool{
	".mp4": true, ".avi": true, ".mov": true, ".mkv": true,
	".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	".mpg": true, ".mpeg": true, ".3gp": true, ".ts": true,
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		crf     int
		preset  string
		output  string
		workers int
	)

	cmd := &cobra.Command{
		Use:     "kachlan <file-or-folder>",
		Short:   "Fast video compression powered by ffmpeg",
		Long:    "kachlan compresses video files using ffmpeg with sensible defaults.\nPass a single file or an entire folder to compress everything at once.",
		Version: version,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args[0], crf, preset, output, workers)
		},
		SilenceUsage: true,
	}

	cmd.Flags().IntVar(&crf, "crf", 28, "quality level (0-51, lower = better quality, bigger file)")
	cmd.Flags().StringVar(&preset, "preset", "fast", "encoding speed (ultrafast|superfast|veryfast|faster|fast|medium|slow|slower|veryslow)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file or folder path")
	cmd.Flags().IntVarP(&workers, "workers", "w", 1, "number of parallel compressions (folder mode)")

	return cmd
}

func run(input string, crf int, preset, output string, workers int) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is not installed. Install it first:\n  macOS:  brew install ffmpeg\n  Ubuntu: sudo apt install ffmpeg\n  Fedora: sudo dnf install ffmpeg")
	}

	info, err := os.Stat(input)
	if err != nil {
		return fmt.Errorf("cannot access %q: %w", input, err)
	}

	if info.IsDir() {
		return compressFolder(input, crf, preset, output, workers)
	}
	return compressSingle(input, crf, preset, output)
}

func compressSingle(input string, crf int, preset, output string) error {
	if output == "" {
		ext := filepath.Ext(input)
		base := strings.TrimSuffix(input, ext)
		output = base + "-compressed" + ext
	}
	return compressVideo(input, output, crf, preset)
}

func compressFolder(input string, crf int, preset, outDir string, workers int) error {
	input = filepath.Clean(input)

	if outDir == "" {
		outDir = input + "-compressed"
	}

	var videos []string
	err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isVideo(path) {
			videos = append(videos, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("scanning folder: %w", err)
	}

	if len(videos) == 0 {
		return fmt.Errorf("no video files found in %q", input)
	}

	fmt.Printf("Found %d video(s) in %s\n\n", len(videos), input)

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, v := range videos {
		rel, _ := filepath.Rel(input, v)
		dst := filepath.Join(outDir, rel)

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		wg.Add(1)
		go func(src, dst string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := compressVideo(src, dst, crf, preset); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(v, dst)
	}

	wg.Wait()

	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d file(s) failed:\n", len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		return fmt.Errorf("compression completed with %d error(s)", len(errs))
	}

	fmt.Printf("\nAll %d video(s) compressed → %s\n", len(videos), outDir)
	return nil
}

func compressVideo(input, output string, crf int, preset string) error {
	fmt.Printf("Compressing: %s → %s\n", input, output)

	cmd := exec.Command("ffmpeg",
		"-i", input,
		"-vcodec", "libx264",
		"-crf", fmt.Sprintf("%d", crf),
		"-preset", preset,
		"-acodec", "aac",
		"-y",
		output,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to compress %s: %w", input, err)
	}

	printSizeComparison(input, output)
	return nil
}

func printSizeComparison(input, output string) {
	inInfo, err1 := os.Stat(input)
	outInfo, err2 := os.Stat(output)
	if err1 != nil || err2 != nil {
		return
	}

	inMB := float64(inInfo.Size()) / 1024 / 1024
	outMB := float64(outInfo.Size()) / 1024 / 1024
	reduction := (1 - outMB/inMB) * 100

	fmt.Printf("  ✓ %s: %.1f MB → %.1f MB (%.1f%% smaller)\n", filepath.Base(input), inMB, outMB, reduction)
}

func isVideo(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return videoExts[ext]
}
