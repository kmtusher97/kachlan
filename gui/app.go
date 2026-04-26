package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/kmtusher97/kachlan/media"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	busy   atomic.Bool
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Check and auto-install ffmpeg if needed
	go a.ensureFFmpeg()
}

func (a *App) ensureFFmpeg() {
	ffmpegPath, err := media.GetFFmpegPath()
	if err != nil {
		// ffmpeg not found, install it
		runtime.EventsEmit(a.ctx, "ffmpeg:installing", map[string]any{
			"message": "Installing ffmpeg...",
		})

		ffmpegPath, err = media.InstallFFmpeg(func(pct float64) {
			runtime.EventsEmit(a.ctx, "ffmpeg:progress", map[string]any{
				"percent": pct,
			})
		})

		if err != nil {
			runtime.EventsEmit(a.ctx, "ffmpeg:error", map[string]any{
				"error": fmt.Sprintf("Failed to install ffmpeg: %v", err),
			})
			return
		}

		media.SetFFmpegPath(ffmpegPath)
		runtime.EventsEmit(a.ctx, "ffmpeg:ready", map[string]any{
			"path": ffmpegPath,
		})
	} else if ffmpegPath != "ffmpeg" {
		// Using local installation
		media.SetFFmpegPath(ffmpegPath)
		runtime.EventsEmit(a.ctx, "ffmpeg:ready", map[string]any{
			"path": ffmpegPath,
		})
	} else {
		// System ffmpeg found
		runtime.EventsEmit(a.ctx, "ffmpeg:ready", map[string]any{
			"path": "system",
		})
	}
}

type FileResult struct {
	InputPath  string  `json:"inputPath"`
	OutputPath string  `json:"outputPath"`
	InputSize  float64 `json:"inputSize"`
	OutputSize float64 `json:"outputSize"`
	Reduction  float64 `json:"reduction"`
}

type CompressResponse struct {
	Results []FileResult `json:"results"`
	Errors  []string     `json:"errors"`
}

type VideoInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relativePath"`
}

type VideoListResponse struct {
	Videos []VideoInfo `json:"videos"`
	Error  string      `json:"error,omitempty"`
}

// CheckFFmpeg returns true if ffmpeg is available in PATH.
func (a *App) CheckFFmpeg() bool {
	return media.CheckFFmpeg()
}

// SelectFile opens a native file picker dialog.
func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Video File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Video Files",
				Pattern:     "*.mp4;*.avi;*.mov;*.mkv;*.wmv;*.flv;*.webm;*.m4v;*.mpg;*.mpeg;*.3gp;*.ts",
			},
		},
	})
}

// SelectFolder opens a native folder picker dialog.
func (a *App) SelectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Video Folder",
	})
}

// SelectOutputPath opens a save dialog for choosing output file location.
func (a *App) SelectOutputPath() (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Select Output Location",
		DefaultFilename: "compressed.mp4",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Video Files",
				Pattern:     "*.mp4;*.avi;*.mov;*.mkv;*.wmv;*.flv;*.webm;*.m4v",
			},
		},
	})
}

// GetFolderVideos returns metadata about video files in a folder.
func (a *App) GetFolderVideos(folderPath string) (*VideoListResponse, error) {
	folderPath = filepath.Clean(folderPath)

	videos, err := media.FindVideos(folderPath)
	if err != nil {
		return &VideoListResponse{
			Error: fmt.Sprintf("Failed to scan folder: %v", err),
		}, nil
	}

	var videoInfos []VideoInfo
	for _, v := range videos {
		relPath, err := filepath.Rel(folderPath, v)
		if err != nil {
			relPath = filepath.Base(v) // Fallback to filename only
		}

		videoInfos = append(videoInfos, VideoInfo{
			Name:         filepath.Base(v),
			Path:         v,
			RelativePath: relPath,
		})
	}

	return &VideoListResponse{
		Videos: videoInfos,
	}, nil
}

// CompressFile compresses a single video file.
func (a *App) CompressFile(input string, crf int, preset string, customOutput string) (*CompressResponse, error) {
	if !a.busy.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("compression already in progress")
	}
	defer a.busy.Store(false)

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	defer cancel()

	// Determine output path
	output := customOutput
	if output == "" {
		output = media.DefaultOutputPath(input, "compressed")
	}

	a.emitProgress(filepath.Base(input), 0, 1, "compressing")

	err := media.CompressVideo(ctx, input, output, crf, preset, func(pct float64) {
		a.emitPercent(filepath.Base(input), pct)
	})
	if err != nil {
		a.emitProgress(filepath.Base(input), 0, 1, "error")
		return nil, err
	}

	result := a.fileResult(input, output)
	a.emitProgress(filepath.Base(input), 1, 1, "done")

	return &CompressResponse{Results: []FileResult{result}}, nil
}

// CompressFolder compresses all videos in a folder.
func (a *App) CompressFolder(input string, crf int, preset string, workers int, customOutput string) (*CompressResponse, error) {
	if !a.busy.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("compression already in progress")
	}
	defer a.busy.Store(false)

	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	defer cancel()

	input = filepath.Clean(input)

	// Determine output directory
	outDir := customOutput
	if outDir == "" {
		outDir = media.DefaultOutputDir(input, "compressed")
	}

	videos, err := media.FindVideos(input)
	if err != nil {
		return nil, fmt.Errorf("scanning folder: %w", err)
	}
	if len(videos) == 0 {
		return nil, fmt.Errorf("no video files found in %q", input)
	}

	if workers < 1 {
		workers = 1
	}

	total := len(videos)
	var completed int64

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []FileResult
	var errs []string

	for _, v := range videos {
		rel, _ := filepath.Rel(input, v)
		dst := filepath.Join(outDir, rel)

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return nil, fmt.Errorf("creating output directory: %w", err)
		}

		wg.Add(1)
		go func(src, dst string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			a.emitProgress(filepath.Base(src), int(atomic.LoadInt64(&completed)), total, "compressing")

			if err := media.CompressVideo(ctx, src, dst, crf, preset, func(pct float64) {
				a.emitPercent(filepath.Base(src), pct)
			}); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s: %v", filepath.Base(src), err))
				mu.Unlock()
			} else {
				r := a.fileResult(src, dst)
				mu.Lock()
				results = append(results, r)
				mu.Unlock()
			}

			done := int(atomic.AddInt64(&completed, 1))
			a.emitProgress(filepath.Base(src), done, total, "done")
		}(v, dst)
	}

	wg.Wait()

	return &CompressResponse{Results: results, Errors: errs}, nil
}

// Cancel stops the current compression.
func (a *App) Cancel() {
	if a.cancel != nil {
		a.cancel()
	}
}

func (a *App) fileResult(input, output string) FileResult {
	s := media.ComputeSizes(input, output)
	return FileResult{
		InputPath:  input,
		OutputPath: output,
		InputSize:  s.InputMB,
		OutputSize: s.OutputMB,
		Reduction:  s.ReductionPct,
	}
}

func (a *App) emitProgress(file string, done, total int, status string) {
	runtime.EventsEmit(a.ctx, "compress:progress", map[string]any{
		"file":   file,
		"done":   done,
		"total":  total,
		"status": status,
	})
}

func (a *App) emitPercent(file string, percent float64) {
	runtime.EventsEmit(a.ctx, "compress:percent", map[string]any{
		"file":    file,
		"percent": percent,
	})
}
