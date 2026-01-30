package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/plaid/llm-context-bundler/internal/bundler"
	"github.com/plaid/llm-context-bundler/internal/ignore"
	"github.com/plaid/llm-context-bundler/internal/walker"
)

var version = "0.1.0"

func main() {
	var (
		outputPath  string
		rootDir     string
		verbose     bool
		showVersion bool
	)

	flag.StringVar(&outputPath, "output", "context.md", "output file path")
	flag.StringVar(&rootDir, "dir", ".", "root directory to scan")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose output")
	flag.BoolVar(&showVersion, "version", false, "show version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "lcb - LLM Context Bundler\n\n")
		fmt.Fprintf(os.Stderr, "Recursively bundle Markdown files into a single file for LLM context windows.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  lcb [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  lcb                           # Bundle current directory to context.md\n")
		fmt.Fprintf(os.Stderr, "  lcb --output=bundle.md        # Custom output filename\n")
		fmt.Fprintf(os.Stderr, "  lcb --dir=./docs              # Bundle specific directory\n")
		fmt.Fprintf(os.Stderr, "  lcb --verbose                 # Show detailed progress\n\n")
		fmt.Fprintf(os.Stderr, "Ignore patterns:\n")
		fmt.Fprintf(os.Stderr, "  Create a .lcbignore file with gitignore-style patterns to exclude files.\n")
		fmt.Fprintf(os.Stderr, "  Default exclusions: .git, node_modules, vendor, hidden directories\n\n")
		fmt.Fprintf(os.Stderr, "Auto-splitting:\n")
		fmt.Fprintf(os.Stderr, "  Output automatically splits into multiple files if exceeding 100 MB.\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("lcb version %s\n", version)
		os.Exit(0)
	}

	// Resolve root directory to absolute path
	if rootDir == "." {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not get working directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Verify root directory exists
	info, err := os.Stat(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: directory does not exist: %s\n", rootDir)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: not a directory: %s\n", rootDir)
		os.Exit(1)
	}

	// Load .lcbignore file if it exists
	ignoreFilePath := filepath.Join(rootDir, ".lcbignore")
	matcher, err := ignore.New(ignoreFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not parse .lcbignore: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "scanning %s\n", rootDir)
	}

	// Walk the directory to find all markdown files
	files, err := walker.Walk(rootDir, matcher)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not scan directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no markdown files found")
		os.Exit(0)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "found %d markdown files\n", len(files))
		for _, f := range files {
			fmt.Fprintf(os.Stderr, "  %s\n", f.Path)
		}
	}

	// Bundle the files
	b := bundler.New(rootDir, outputPath, verbose)
	outputFiles, err := b.Bundle(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not bundle files: %v\n", err)
		os.Exit(1)
	}

	if len(outputFiles) == 1 {
		fmt.Fprintf(os.Stderr, "bundled %d files to %s\n", len(files), outputFiles[0])
	} else {
		fmt.Fprintf(os.Stderr, "bundled %d files to %d parts: %s\n", len(files), len(outputFiles), strings.Join(outputFiles, ", "))
	}
}
