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
		dirsFlag    string
		verbose     bool
		showVersion bool
	)

	flag.StringVar(&outputPath, "output", "context.md", "output file path")
	flag.StringVar(&dirsFlag, "dir", ".", "root directory(s) to scan (comma-separated for multiple)")
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
		fmt.Fprintf(os.Stderr, "  lcb                                   # Bundle current directory to context.md\n")
		fmt.Fprintf(os.Stderr, "  lcb --output=bundle.md                # Custom output filename\n")
		fmt.Fprintf(os.Stderr, "  lcb --dir=./docs                      # Bundle specific directory\n")
		fmt.Fprintf(os.Stderr, "  lcb --dir=./docs,./specs,./guides     # Bundle multiple directories\n")
		fmt.Fprintf(os.Stderr, "  lcb --verbose                         # Show detailed progress\n\n")
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

	// Parse comma-separated directories
	rawDirs := strings.Split(dirsFlag, ",")
	var dirs []string
	for _, d := range rawDirs {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		dirs = append(dirs, d)
	}

	if len(dirs) == 0 {
		fmt.Fprintln(os.Stderr, "error: no directories specified")
		os.Exit(1)
	}

	// Resolve all directories to absolute paths and validate
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: could not get working directory: %v\n", err)
		os.Exit(1)
	}

	var absDirs []string
	for _, d := range dirs {
		absDir := d
		if d == "." {
			absDir = cwd
		} else if !filepath.IsAbs(d) {
			absDir = filepath.Join(cwd, d)
		}

		info, err := os.Stat(absDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: directory does not exist: %s\n", d)
			os.Exit(1)
		}
		if !info.IsDir() {
			fmt.Fprintf(os.Stderr, "error: not a directory: %s\n", d)
			os.Exit(1)
		}

		absDirs = append(absDirs, absDir)
	}

	// Check for overlapping directories
	overlaps := walker.DetectOverlaps(absDirs)
	for _, pair := range overlaps {
		fmt.Fprintf(os.Stderr, "warning: directory %s contains %s - files may be deduplicated\n", pair[0], pair[1])
	}

	if verbose {
		if len(absDirs) == 1 {
			fmt.Fprintf(os.Stderr, "scanning %s\n", absDirs[0])
		} else {
			fmt.Fprintf(os.Stderr, "scanning %d directories:\n", len(absDirs))
			for _, d := range absDirs {
				fmt.Fprintf(os.Stderr, "  %s\n", d)
			}
		}
	}

	// Walk the directory(s) to find all markdown files
	var files []walker.FileInfo
	if len(absDirs) == 1 {
		// Single directory - use original Walk for backward compatibility
		ignoreFilePath := filepath.Join(absDirs[0], ".lcbignore")
		matcher, err := ignore.New(ignoreFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not parse .lcbignore: %v\n", err)
			os.Exit(1)
		}
		files, err = walker.Walk(absDirs[0], matcher)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not scan directory: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Multiple directories - use WalkMultiple with per-directory ignore
		matcherFunc := func(root string) (*ignore.Matcher, error) {
			ignoreFilePath := filepath.Join(root, ".lcbignore")
			return ignore.New(ignoreFilePath)
		}
		files, err = walker.WalkMultiple(absDirs, matcherFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not scan directories: %v\n", err)
			os.Exit(1)
		}
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
	b := bundler.New(absDirs, outputPath, verbose)
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
