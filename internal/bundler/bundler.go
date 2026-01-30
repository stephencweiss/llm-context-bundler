package bundler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/plaid/llm-context-bundler/internal/walker"
)

// Bundler assembles markdown files into a single output file.
type Bundler struct {
	rootDir string
	output  io.Writer
	verbose bool
}

// New creates a new Bundler.
func New(rootDir string, output io.Writer, verbose bool) *Bundler {
	return &Bundler{
		rootDir: rootDir,
		output:  output,
		verbose: verbose,
	}
}

// Bundle processes all files and writes them to the output.
func (b *Bundler) Bundle(files []walker.FileInfo) error {
	// Write header
	fmt.Fprintln(b.output, "# Bundled Context")
	fmt.Fprintln(b.output)

	// Write table of contents
	fmt.Fprintln(b.output, "## Table of Contents")
	for _, file := range files {
		anchor := pathToAnchor(file.Path)
		fmt.Fprintf(b.output, "- [%s](#%s)\n", file.Path, anchor)
	}
	fmt.Fprintln(b.output)

	// Write each file
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(b.rootDir, file.Path))
		if err != nil {
			// Warn and skip unreadable files
			fmt.Fprintf(os.Stderr, "warning: could not read %s: %v\n", file.Path, err)
			continue
		}

		// Write separator
		fmt.Fprintln(b.output, "---")
		fmt.Fprintln(b.output)

		// Write source header
		fmt.Fprintf(b.output, "<!-- SOURCE: %s -->\n", file.Path)

		// Write content
		b.output.Write(content)

		// Ensure newline at end
		if len(content) > 0 && content[len(content)-1] != '\n' {
			fmt.Fprintln(b.output)
		}
		fmt.Fprintln(b.output)
	}

	return nil
}

// pathToAnchor converts a file path to a GitHub-compatible markdown anchor.
// Example: "chapters/chapter-1.md" -> "chapterschapter-1md"
func pathToAnchor(path string) string {
	// Convert to lowercase
	anchor := strings.ToLower(path)

	// Replace path separators with empty string
	anchor = strings.ReplaceAll(anchor, "/", "")
	anchor = strings.ReplaceAll(anchor, "\\", "")

	// Remove characters that aren't alphanumeric, hyphen, or underscore
	reg := regexp.MustCompile(`[^a-z0-9\-_]`)
	anchor = reg.ReplaceAllString(anchor, "")

	return anchor
}
