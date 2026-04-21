package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/kmtusher97/kachlan/media"
	"github.com/spf13/cobra"
)

var version = "dev"

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
		return fmt.Errorf("ffmpeg is not installed. Install it first:\n  macOS:   brew install ffmpeg\n  Ubuntu:  sudo apt install ffmpeg\n  Fedora:  sudo dnf install ffmpeg\n  Windows: winget install ffmpeg")
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
		output = media.DefaultOutputPath(input, "compressed")
	}

	fmt.Printf("Compressing: %s → %s\n", input, output)

	if err := media.CompressVideo(context.Background(), input, output, crf, preset, nil); err != nil {
		return err
	}

	printSizeComparison(input, output)
	return nil
}

func compressFolder(input string, crf int, preset, outDir string, workers int) error {
	input = filepath.Clean(input)

	if outDir == "" {
		outDir = media.DefaultOutputDir(input, "compressed")
	}

	videos, err := media.FindVideos(input)
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

			fmt.Printf("Compressing: %s → %s\n", src, dst)

			if err := media.CompressVideo(context.Background(), src, dst, crf, preset, nil); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			printSizeComparison(src, dst)
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

func printSizeComparison(input, output string) {
	s := media.ComputeSizes(input, output)
	fmt.Printf("  ✓ %s: %.1f MB → %.1f MB (%.1f%% smaller)\n", filepath.Base(input), s.InputMB, s.OutputMB, s.ReductionPct)
}
