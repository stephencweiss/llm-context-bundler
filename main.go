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

func main() {
	var (
		outputPath string
		rootDir    string
		verbose    bool
	)

	flag.StringVar(&outputPath, "output", "context.md", "output file path")
	flag.StringVar(&rootDir, "dir", ".", "root directory to scan")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose output")
	flag.Parse()

	// Resolve root directory to absolute path
	if rootDir == "." {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not get working directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Load .lcbignore file if it exists
	ignoreFilePath := filepath.Join(rootDir, ".lcbignore")
	matcher, err := ignore.New(ignoreFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not parse .lcbignore: %v\n", err)
		os.Exit(1)
	}

	// Walk the directory to find all markdown files
	files, err := walker.Walk(rootDir, matcher)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not walk directory: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no markdown files found")
		os.Exit(0)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "found %d markdown files\n", len(files))
	}

	// Bundle the files
	b := bundler.New(rootDir, outputPath, verbose)
	outputFiles, err := b.Bundle(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not bundle files: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "bundled %d files to %s\n", len(files), strings.Join(outputFiles, ", "))
}
