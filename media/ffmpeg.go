package media

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetFFmpegPath returns the path to ffmpeg, checking bundled binary first (for app bundles),
// then system PATH, then the local installation directory.
func GetFFmpegPath() (string, error) {
	// Check for bundled ffmpeg (macOS .app bundle)
	if bundledPath := getBundledFFmpegPath(); bundledPath != "" {
		if _, err := os.Stat(bundledPath); err == nil {
			return bundledPath, nil
		}
	}

	// Check if ffmpeg is in PATH
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path, nil
	}

	// Check local installation
	localPath := getLocalFFmpegPath()
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	return "", fmt.Errorf("ffmpeg not found")
}

// InstallFFmpeg downloads and installs ffmpeg to ~/.kachlan/bin/
// Returns the path to the installed ffmpeg binary.
func InstallFFmpeg(onProgress ProgressFunc) (string, error) {
	// Check if already installed locally
	localPath := getLocalFFmpegPath()
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Create installation directory
	binDir := getLocalBinDir()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("creating install directory: %w", err)
	}

	// Download ffmpeg
	downloadURL, filename := getDownloadURL()
	if downloadURL == "" {
		return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	tmpFile := filepath.Join(os.TempDir(), filename)
	if err := downloadFile(tmpFile, downloadURL, onProgress); err != nil {
		return "", fmt.Errorf("downloading ffmpeg: %w", err)
	}
	defer os.Remove(tmpFile)

	// Extract ffmpeg binary
	if err := extractFFmpeg(tmpFile, binDir); err != nil {
		return "", fmt.Errorf("extracting ffmpeg: %w", err)
	}

	// Verify installation
	if _, err := os.Stat(localPath); err != nil {
		return "", fmt.Errorf("ffmpeg installation failed")
	}

	// Make executable on Unix-like systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(localPath, 0755); err != nil {
			return "", fmt.Errorf("making ffmpeg executable: %w", err)
		}
	}

	return localPath, nil
}

func getLocalBinDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kachlan", "bin")
}

func getLocalFFmpegPath() string {
	binName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binName = "ffmpeg.exe"
	}
	return filepath.Join(getLocalBinDir(), binName)
}

// getBundledFFmpegPath returns the path to bundled ffmpeg if running from a macOS .app bundle
func getBundledFFmpegPath() string {
	if runtime.GOOS != "darwin" {
		return ""
	}

	// Get executable path
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}

	// Check if running from .app bundle (path contains .app/Contents/MacOS/)
	if !strings.Contains(exePath, ".app/Contents/MacOS/") {
		return ""
	}

	// ffmpeg should be in the same directory as the executable
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "ffmpeg")
}

func getDownloadURL() (url, filename string) {
	// Using static builds from official sources
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
				"ffmpeg-windows-amd64.zip"
		}
	case "darwin":
		// macOS universal binary
		return "https://evermeet.cx/ffmpeg/getrelease/zip",
			"ffmpeg-macos.zip"
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz",
				"ffmpeg-linux-amd64.tar.xz"
		}
	}
	return "", ""
}

func downloadFile(filepath string, url string, onProgress ProgressFunc) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Track progress
	totalSize := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024) // 32KB chunks
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			if onProgress != nil && totalSize > 0 {
				onProgress(float64(downloaded) / float64(totalSize) * 100)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func extractFFmpeg(archivePath, destDir string) error {
	ext := filepath.Ext(archivePath)

	if ext == ".zip" {
		return extractZip(archivePath, destDir)
	} else if strings.HasSuffix(archivePath, ".tar.xz") || strings.HasSuffix(archivePath, ".tar.gz") {
		return extractTarGz(archivePath, destDir)
	}

	return fmt.Errorf("unsupported archive format: %s", ext)
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find and extract ffmpeg binary
	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name == "ffmpeg" || name == "ffmpeg.exe" {
			return extractZipFile(f, filepath.Join(destDir, name))
		}
	}

	return fmt.Errorf("ffmpeg binary not found in archive")
}

func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func extractTarGz(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var tarReader *tar.Reader

	// Handle both .tar.gz and .tar.xz (simplified - treat as gzip)
	gzr, err := gzip.NewReader(file)
	if err != nil {
		// If not gzip, try reading as plain tar
		file.Seek(0, 0)
		tarReader = tar.NewReader(file)
	} else {
		defer gzr.Close()
		tarReader = tar.NewReader(gzr)
	}

	// Find and extract ffmpeg binary
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		name := filepath.Base(header.Name)
		if name == "ffmpeg" {
			destPath := filepath.Join(destDir, name)
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			if _, err := io.Copy(out, tarReader); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("ffmpeg binary not found in archive")
}
