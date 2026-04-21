# kachlan

<p align="center">
  <img src="assets/logo.svg" alt="kachlan - sliced lime" width="200"/>
</p>

**kachlan** (কচলান) — Bangla for *squeezing juice from a lime*. Just like wringing every last drop from a lime, kachlan squeezes the excess out of your video files.

Fast video compression CLI powered by ffmpeg. Compress a single file or an entire folder with one command.

## Install

### Homebrew (macOS/Linux)

```bash
brew install kmtusher97/tap/kachlan
```

### Debian/Ubuntu

```bash
# Download the latest .deb from the releases page
curl -sL https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_0.2.0_linux_amd64.deb -o kachlan.deb
sudo dpkg -i kachlan.deb
```

See [Releases](https://github.com/kmtusher97/kachlan/releases) for all versions and architectures (`.deb`, `.rpm`).

### Windows

Download the `.zip` from [Releases](https://github.com/kmtusher97/kachlan/releases), extract it, and add `kachlan.exe` to your PATH.

> Install ffmpeg via `winget install ffmpeg` or [download it here](https://ffmpeg.org/download.html).

### Go install

```bash
go install github.com/kmtusher97/kachlan/cli@latest
```

### From source

```bash
git clone https://github.com/kmtusher97/kachlan.git
cd kachlan/cli
make install
```

### Download binary

Grab a prebuilt binary from [Releases](https://github.com/kmtusher97/kachlan/releases).

> **Prerequisite:** [ffmpeg](https://ffmpeg.org/download.html) must be installed and available in your PATH.

## Usage

### Compress a single video

```bash
kachlan video.mp4
# → video-compressed.mp4
```

### Compress with custom output

```bash
kachlan video.mp4 -o output.mp4
```

### Compress all videos in a folder

```bash
kachlan ./my-videos/
# → ./my-videos-compressed/  (same structure, all videos compressed)
```

### Compress folder with parallel workers

```bash
kachlan ./my-videos/ -w 4
```

### Adjust quality

```bash
# Higher quality (lower CRF = bigger file)
kachlan video.mp4 --crf 23

# Lower quality (higher CRF = smaller file)
kachlan video.mp4 --crf 32

# Slower encoding for better compression
kachlan video.mp4 --preset slow
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--crf` | `28` | Quality level (0-51, lower = better quality) |
| `--preset` | `fast` | Encoding speed (`ultrafast` to `veryslow`) |
| `-o, --output` | auto | Output file or folder path |
| `-w, --workers` | `1` | Parallel compressions (folder mode) |

## Supported formats

`.mp4` `.avi` `.mov` `.mkv` `.wmv` `.flv` `.webm` `.m4v` `.mpg` `.mpeg` `.3gp` `.ts`

## How it works

kachlan wraps ffmpeg with sensible defaults:

- **Video codec:** H.264 (libx264)
- **Audio codec:** AAC
- **CRF:** 28 (good balance of quality and size)
- **Preset:** fast

When compressing a folder, kachlan mirrors the directory structure into a new `<folder>-compressed` directory, preserving the original files untouched.

## License

[MIT](LICENSE)
