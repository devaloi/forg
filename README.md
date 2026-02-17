# forg — Smart File Organizer

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A fast, configurable CLI tool that organizes files into directories based on rules you define in YAML.

## Features

- **Declarative YAML config** — define source directory, rules, and destinations in a single file
- **Rich matching** — filter by file extension, glob pattern, size range, and file age
- **Dry-run preview** — see exactly what will happen before any files move
- **One-command undo** — reverse the last run instantly
- **Conflict strategies** — choose `skip`, `rename`, or `overwrite` when a destination file already exists
- **Recursive scanning** — optionally walk subdirectories
- **Hidden file support** — opt in to organizing dotfiles

## Install

```bash
go install github.com/jasonaloi/forg@latest
```

Or build from source:

```bash
git clone https://github.com/jasonaloi/forg.git
cd forg
make build
```

## Quick Start

**1. Generate a starter config:**

```bash
forg init
```

This creates a `.forg.yaml` in the current directory with sensible defaults.

**2. Edit the config to match your needs:**

```bash
$EDITOR .forg.yaml
```

**3. Preview what will happen:**

```bash
forg preview
```

**4. Run it:**

```bash
forg run
```

**5. Made a mistake? Undo it:**

```bash
forg undo
```

## Configuration Reference

```yaml
# Source directory to scan
source: ~/Downloads

# What to do when a destination file already exists
# Options: skip | rename | overwrite
conflict: rename

rules:
  - name: images
    match:
      # File extensions (case-insensitive)
      extensions: [.jpg, .jpeg, .png, .gif, .webp]
    destination: ~/Pictures/Sorted

  - name: large-videos
    match:
      extensions: [.mp4, .mov, .avi]
      # Size filters — supports B, KB, MB, GB, TB
      min_size: 100MB
      max_size: 4GB
    destination: ~/Videos/Large

  - name: old-archives
    match:
      extensions: [.zip, .tar.gz, .rar]
      # Age filters — supports d (days), w (weeks), m (months), y (years)
      older_than: 30d
    destination: ~/Archives/Old

  - name: recent-logs
    match:
      # Glob pattern matching
      pattern: "*.log"
      newer_than: 2w
    destination: ~/Logs/Recent
```

### Match criteria

All criteria within a single rule are combined with AND logic. A file must satisfy every specified criterion to match.

| Field | Description | Format |
|---|---|---|
| `extensions` | List of file extensions | `[.jpg, .png]` |
| `pattern` | Glob pattern against filename | `*.log`, `report-*` |
| `min_size` | Minimum file size | `100MB`, `1.5GB` |
| `max_size` | Maximum file size | `500KB`, `2TB` |
| `older_than` | Minimum file age | `30d`, `6m`, `1y` |
| `newer_than` | Maximum file age | `2w`, `7d` |

## Commands

| Command | Description |
|---|---|
| `forg init` | Generate a sample `.forg.yaml` config file |
| `forg preview` | Show planned moves without touching any files |
| `forg run` | Execute rules and move files |
| `forg undo` | Reverse the most recent run |

### Flags for `run` and `preview`

| Flag | Short | Description |
|---|---|---|
| `--dry-run` | | Show what would happen without moving files (`run` only) |
| `--recursive` | `-r` | Scan directories recursively |
| `--include-hidden` | | Include hidden files and directories |

### Global flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--config` | `-c` | `.forg.yaml` | Path to configuration file |
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--quiet` | `-q` | `false` | Suppress all non-error output |

## Examples

### Organize a messy Downloads folder

```yaml
source: ~/Downloads
conflict: rename

rules:
  - name: images
    match:
      extensions: [.jpg, .jpeg, .png, .gif, .webp, .svg]
    destination: ~/Pictures/Downloads

  - name: documents
    match:
      extensions: [.pdf, .doc, .docx, .txt, .xlsx, .pptx]
    destination: ~/Documents/Downloads

  - name: installers
    match:
      extensions: [.dmg, .pkg, .exe, .msi]
    destination: ~/Downloads/Installers

  - name: archives
    match:
      extensions: [.zip, .tar.gz, .rar, .7z]
    destination: ~/Downloads/Archives
```

### Archive old large files

```yaml
source: ~/Projects
conflict: skip

rules:
  - name: old-large-files
    match:
      min_size: 500MB
      older_than: 6m
    destination: ~/Archive/OldLarge

  - name: ancient-logs
    match:
      pattern: "*.log"
      older_than: 1y
    destination: ~/Archive/Logs
```

### Sort by type and size

```yaml
source: ~/Media
conflict: rename

rules:
  - name: small-images
    match:
      extensions: [.jpg, .png, .webp]
      max_size: 5MB
    destination: ~/Media/Images/Small

  - name: large-images
    match:
      extensions: [.jpg, .png, .webp]
      min_size: 5MB
    destination: ~/Media/Images/Large

  - name: short-videos
    match:
      extensions: [.mp4, .mov]
      max_size: 100MB
    destination: ~/Media/Videos/Short

  - name: long-videos
    match:
      extensions: [.mp4, .mov]
      min_size: 100MB
    destination: ~/Media/Videos/Long
```

## Architecture

```
internal/
├── scanner/     Walks source directories and collects file metadata
├── rules/       Matcher interface with extension, pattern, size, and age matchers
├── organizer/   Builds a move plan, executes file operations, manages undo log
├── config/      Parses and validates .forg.yaml configuration
cmd/             Cobra CLI commands (init, preview, run, undo)
```

The pipeline flows as: **config → scanner → rules engine → plan → executor → undo log**.

Each rule compiles into a chain of `Matcher` implementations. A file matches a rule only when all of its matchers pass. The first matching rule determines the file's destination.

## Development

```bash
make build       # Build the binary
make test        # Run tests with race detection
make lint        # Run golangci-lint
make vet         # Run go vet
make fmt         # Format code
make coverage    # Generate coverage report
```

## License

[MIT](LICENSE)
