# kachlan

<p align="center">
  <img src="assets/logo.svg" alt="kachlan - sliced lime" width="200"/>
</p>

**kachlan** (কচলান) — Bangla for *squeezing juice from a lime*. Just like wringing every last drop from a lime, kachlan squeezes the excess out of your video files.

Fast video compression CLI powered by ffmpeg. Compress a single file or an entire folder with one command.

## Download

> **Prerequisite:** [ffmpeg](https://ffmpeg.org/download.html) must be installed and available in your PATH.
>
> macOS: `brew install ffmpeg` | Ubuntu: `sudo apt install ffmpeg` | Windows: `winget install ffmpeg`

### Desktop App (GUI)

A graphical desktop app with drag-and-drop, progress bar, and quality controls.

<table>
<tr>
<td align="center" width="33%"><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan-gui_windows_amd64.zip">Windows (x64)</a></td>
<td align="center" width="33%"><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan-gui_darwin_universal.zip">macOS (Universal)</a></td>
<td align="center" width="33%"><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan-gui_linux_amd64.tar.gz">Linux (x64)</a></td>
</tr>
</table>

### Command Line (CLI)

<table>
<tr>
<td align="center" width="33%">
<h3>Windows</h3>
<p><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_windows_amd64.zip">Download for Windows (x64)</a></p>
<p><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_windows_arm64.zip">Download for Windows (ARM)</a></p>
<p><sub>Extract the zip, double-click <code>kachlan.exe</code></sub></p>
</td>
<td align="center" width="33%">
<h3>macOS</h3>
<p><code>brew install kmtusher97/tap/kachlan</code></p>
<p>or <a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_darwin_arm64.tar.gz">Apple Silicon</a> | <a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_darwin_amd64.tar.gz">Intel</a></p>
</td>
<td align="center" width="33%">
<h3>Linux</h3>
<p><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_linux_amd64.deb">Download .deb (x64)</a></p>
<p><a href="https://github.com/kmtusher97/kachlan/releases/latest/download/kachlan_linux_amd64.rpm">Download .rpm (x64)</a></p>
<p><sub>Install: <code>sudo dpkg -i kachlan_*.deb</code></sub></p>
</td>
</tr>
</table>

<p align="center"><a href="https://github.com/kmtusher97/kachlan/releases">All releases and architectures</a></p>

### Other install methods

```bash
# Go install
go install github.com/kmtusher97/kachlan/cli@latest

# From source
git clone https://github.com/kmtusher97/kachlan.git
cd kachlan/cli
make install
```

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
