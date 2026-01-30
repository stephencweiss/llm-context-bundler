package bundler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	for i, file := range files {
		content, err := os.ReadFile(filepath.Join(b.rootDir, file.Path))
		if err != nil {
			// Warn and skip unreadable files
			fmt.Fprintf(os.Stderr, "warning: could not read %s: %v\n", file.Path, err)
			continue
		}

		// Write separator (except for first file)
		if i > 0 {
			fmt.Fprintln(b.output, "---")
			fmt.Fprintln(b.output)
		}

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
