package bundler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/plaid/llm-context-bundler/internal/walker"
)

const (
	// MaxFileSize is the maximum size for a single output file (100 MB).
	MaxFileSize = 100 * 1024 * 1024
)

// Bundler assembles markdown files into output file(s).
type Bundler struct {
	rootDirs   []string
	outputPath string
	verbose    bool
}

// New creates a new Bundler.
func New(rootDirs []string, outputPath string, verbose bool) *Bundler {
	return &Bundler{
		rootDirs:   rootDirs,
		outputPath: outputPath,
		verbose:    verbose,
	}
}

// fileContent holds a file's path and its content.
type fileContent struct {
	path        string // Display path (may include source label prefix for multi-dir)
	sourceLabel string // Source directory label for grouping in TOC
	content     []byte
}

// Bundle processes all files and writes them to output file(s).
// Returns the list of output files created.
func (b *Bundler) Bundle(files []walker.FileInfo) ([]string, error) {
	// Read all file contents first
	var contents []fileContent
	multiDir := len(b.rootDirs) > 1

	for _, file := range files {
		// Read using absolute path from FileInfo
		absPath := filepath.Join(file.SourceDir, file.Path)
		content, err := os.ReadFile(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not read %s: %v\n", absPath, err)
			continue
		}

		// For multi-directory mode, prefix display path with source label
		displayPath := file.Path
		if multiDir {
			displayPath = file.SourceLabel + "/" + file.Path
		}

		contents = append(contents, fileContent{
			path:        displayPath,
			sourceLabel: file.SourceLabel,
			content:     content,
		})
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("no readable files to bundle")
	}

	// Calculate sizes and determine how to split
	parts := b.splitIntoParts(contents)

	// Write each part
	var outputFiles []string
	for i, part := range parts {
		outputPath := b.getPartPath(i, len(parts))
		if err := b.writePart(outputPath, part); err != nil {
			return outputFiles, err
		}
		outputFiles = append(outputFiles, outputPath)
	}

	return outputFiles, nil
}

// splitIntoParts divides files into parts that fit within MaxFileSize.
func (b *Bundler) splitIntoParts(contents []fileContent) [][]fileContent {
	var parts [][]fileContent
	var currentPart []fileContent
	var currentSize int

	for _, fc := range contents {
		// Estimate size: content + header overhead (~100 bytes per file)
		fileSize := len(fc.content) + 100

		// If adding this file would exceed the limit, start a new part
		// (unless current part is empty - always include at least one file)
		if currentSize+fileSize > MaxFileSize && len(currentPart) > 0 {
			parts = append(parts, currentPart)
			currentPart = nil
			currentSize = 0
		}

		currentPart = append(currentPart, fc)
		currentSize += fileSize
	}

	// Don't forget the last part
	if len(currentPart) > 0 {
		parts = append(parts, currentPart)
	}

	return parts
}

// getPartPath returns the output path for a given part number.
func (b *Bundler) getPartPath(partIndex, totalParts int) string {
	if totalParts == 1 {
		return b.outputPath
	}

	// Split filename and extension
	ext := filepath.Ext(b.outputPath)
	base := strings.TrimSuffix(b.outputPath, ext)

	return fmt.Sprintf("%s_part%d%s", base, partIndex+1, ext)
}

// writePart writes a single part file.
func (b *Bundler) writePart(outputPath string, contents []fileContent) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", outputPath, err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintln(file, "# Bundled Context")
	fmt.Fprintln(file)

	// Write table of contents
	fmt.Fprintln(file, "## Table of Contents")

	multiDir := len(b.rootDirs) > 1
	if multiDir {
		// Group by source label
		groups := make(map[string][]fileContent)
		var order []string
		for _, fc := range contents {
			if _, exists := groups[fc.sourceLabel]; !exists {
				order = append(order, fc.sourceLabel)
			}
			groups[fc.sourceLabel] = append(groups[fc.sourceLabel], fc)
		}

		for _, label := range order {
			fmt.Fprintf(file, "\n### %s\n", label)
			for _, fc := range groups[label] {
				anchor := pathToAnchor(fc.path)
				fmt.Fprintf(file, "- [%s](#%s)\n", fc.path, anchor)
			}
		}
	} else {
		// Single directory: flat TOC
		for _, fc := range contents {
			anchor := pathToAnchor(fc.path)
			fmt.Fprintf(file, "- [%s](#%s)\n", fc.path, anchor)
		}
	}
	fmt.Fprintln(file)

	// Write each file
	for _, fc := range contents {
		// Write separator
		fmt.Fprintln(file, "---")
		fmt.Fprintln(file)

		// Write source header
		fmt.Fprintf(file, "<!-- SOURCE: %s -->\n", fc.path)

		// Write content
		file.Write(fc.content)

		// Ensure newline at end
		if len(fc.content) > 0 && fc.content[len(fc.content)-1] != '\n' {
			fmt.Fprintln(file)
		}
		fmt.Fprintln(file)
	}

	return nil
}

// pathToAnchor converts a file path to a GitHub-compatible markdown anchor.
func pathToAnchor(path string) string {
	anchor := strings.ToLower(path)
	anchor = strings.ReplaceAll(anchor, "/", "")
	anchor = strings.ReplaceAll(anchor, "\\", "")
	reg := regexp.MustCompile(`[^a-z0-9\-_]`)
	anchor = reg.ReplaceAllString(anchor, "")
	return anchor
}
