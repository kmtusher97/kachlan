package media

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CompressVideo compresses a single video file using ffmpeg.
// If onProgress is non-nil and ffprobe can determine the duration,
// real-time progress is reported via the callback.
func CompressVideo(ctx context.Context, input, output string, crf int, preset string, onProgress ProgressFunc) error {
	args := []string{
		"-i", input,
		"-vcodec", "libx264",
		"-crf", fmt.Sprintf("%d", crf),
		"-preset", preset,
		"-acodec", "aac",
	}

	var durationUs int64
	if onProgress != nil {
		durationUs = DurationUs(input)
		if durationUs > 0 {
			args = append(args, "-progress", "pipe:1")
		}
	}

	args = append(args, "-y", output)

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)

	if onProgress != nil && durationUs > 0 {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start ffmpeg: %w", err)
		}

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if val, ok := strings.CutPrefix(line, "out_time_us="); ok {
				if us, err := strconv.ParseInt(val, 10, 64); err == nil && us > 0 {
					pct := min(float64(us)/float64(durationUs)*100, 100)
					onProgress(pct)
				}
			}
		}
	} else {
		if onProgress == nil {
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start ffmpeg: %w", err)
		}
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("cancelled")
		}
		return fmt.Errorf("failed to compress %s: %w", filepath.Base(input), err)
	}
	return nil
}
