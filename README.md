# LLM Context Bundler (`lcb`)

A CLI tool that recursively scans a directory for Markdown files and bundles them into a single `.md` file optimized for LLM context windows (e.g., Gemini Apps).

## Features

- **Multi-Directory Support** - Bundle files from multiple directories with `--dir=./docs,./specs`
- **Recursive Traversal** - Walks the entire directory tree to find all Markdown files
- **Smart Filtering** - Automatically excludes `.git`, `node_modules`, `vendor`, and hidden directories
- **Custom Exclusions** - Support for `.lcbignore` file with gitignore-style patterns (per-directory)
- **Table of Contents** - Generates a navigable TOC at the top of each output file
- **Document Separation** - Uses HTML comments (`<!-- SOURCE: path -->`) and horizontal rules for clean separation
- **Auto-Splitting** - Automatically splits output into multiple files if total size exceeds 100 MB
- **Depth-First Ordering** - Files are ordered by directory depth (shallower first), then alphabetically

## Installation

### Requirements

- Go 1.21 or later

### Quick Install

```bash
# Clone and install with make
git clone https://github.plaid.com/plaid/llm-context-bundler.git
cd llm-context-bundler
make install
```

This builds the binary and installs it to `~/.local/bin/`. Make sure `~/.local/bin` is in your PATH.

### Build Only

```bash
# Build without installing
make build

# Run locally
./lcb --help
```

### Other Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make install` | Build and install to `~/.local/bin/` |
| `make clean` | Remove built binaries |
| `make test` | Run tests |
| `make fmt` | Format code |
| `make lint` | Run linter (requires golangci-lint) |
| `make build-all` | Cross-compile for macOS, Linux, Windows |

### Manual Installation

If you prefer not to use make:

```bash
# Build
go build -o lcb .

# Install to your preferred location
sudo mv lcb /usr/local/bin/    # System-wide
# or
mv lcb $(go env GOPATH)/bin/   # GOPATH
# or
mv lcb ~/.local/bin/           # User local
```

### Verify Installation

```bash
lcb --version
```

## Usage

```bash
# Bundle current directory to context.md (default)
lcb

# Custom output filename
lcb --output=bundle.md

# Bundle a specific directory
lcb --dir=./docs

# Bundle multiple directories
lcb --dir=./docs,./specs,./guides

# Show detailed progress
lcb --verbose

# Show version
lcb --version

# Show help
lcb --help
```

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `--output` | `context.md` | Output file path |
| `--dir` | `.` | Root directory(s) to scan (comma-separated for multiple) |
| `--verbose` | `false` | Enable verbose output |
| `--version` | - | Show version and exit |

## Ignore Patterns

Create a `.lcbignore` file in your root directory to exclude files using gitignore-style patterns:

```gitignore
# Comments start with #
drafts/
*.draft.md
internal-notes.md

# Negate patterns with !
!important-draft.md
```

### Default Exclusions

The following are always excluded:
- `.git` directory
- `node_modules` directory
- `vendor` directory
- Hidden directories (starting with `.`)
- Hidden files (starting with `.`)

## Multi-Directory Bundling

When specifying multiple directories, the output clearly identifies each file's source:

```bash
lcb --dir=./docs,./specs
```

Output format with multiple directories:

```markdown
# Bundled Context

## Table of Contents

### docs
- [docs/README.md](#docs-readmemd)
- [docs/api/endpoints.md](#docs-apiendpointsmd)

### specs
- [specs/overview.md](#specs-overviewmd)

---

<!-- SOURCE: docs/README.md -->
...
```

### Per-Directory Ignore Files

Each directory can have its own `.lcbignore` file. Patterns in each file only apply to that directory's files.

### Overlapping Directories

If you specify directories that overlap (e.g., `./docs` and `./docs/api`), the tool will warn you and deduplicate files automatically.

### Label Collisions

If two directories have the same basename (e.g., `./project1/docs` and `./project2/docs`), parent context is automatically added to create unique labels: `project1-docs` and `project2-docs`.

## Output Format

The bundled output follows this structure:

```markdown
# Bundled Context

## Table of Contents
- [intro.md](#intromd)
- [chapters/chapter-1.md](#chapterschapter-1md)

---

<!-- SOURCE: intro.md -->
# Introduction

[original content of intro.md]

---

<!-- SOURCE: chapters/chapter-1.md -->
# Chapter 1

[original content of chapter-1.md]
```

## Auto-Splitting

If the total bundled content exceeds 100 MB, the output automatically splits into multiple files:

- `context_part1.md`
- `context_part2.md`
- etc.

Each part contains its own table of contents for the files it includes.

## Example Workflow

```bash
# Navigate to your documentation directory
cd ~/docs/my-story

# Bundle all markdown files
lcb --output=context.md

# Upload context.md to Gemini Apps
```

## Project Structure

```
llm-context-bundler/
├── main.go                    # CLI entry point
├── go.mod                     # Go module definition
├── internal/
│   ├── walker/
│   │   └── walker.go          # Directory traversal
│   ├── bundler/
│   │   └── bundler.go         # Content assembly & TOC
│   └── ignore/
│       └── ignore.go          # .lcbignore parsing
├── README.md
├── product-spec.md
└── CLAUDE.md
```

## License

MIT
