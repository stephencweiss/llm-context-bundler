package walker

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/plaid/llm-context-bundler/internal/ignore"
)

// Default directories to exclude from scanning.
var defaultExclusions = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

// FileInfo holds information about a discovered markdown file.
type FileInfo struct {
	Path        string // Relative path from source directory root
	Depth       int    // Directory depth (0 = root level)
	SourceDir   string // Absolute path to the source directory
	SourceLabel string // Label derived from source directory name (e.g., "docs")
}

// Walk recursively finds all .md files in the given root directory.
// Files are returned sorted by depth (shallower first), then alphabetically.
// Automatically skips .git, node_modules, vendor, and hidden directories.
// If matcher is provided, also skips paths matching .lcbignore patterns.
func Walk(root string, matcher *ignore.Matcher) ([]FileInfo, error) {
	return WalkWithLabel(root, matcher, "")
}

// WalkWithLabel is like Walk but allows specifying a custom source label.
// If sourceLabel is empty, it defaults to the basename of root.
func WalkWithLabel(root string, matcher *ignore.Matcher, sourceLabel string) ([]FileInfo, error) {
	var files []FileInfo

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	if sourceLabel == "" {
		sourceLabel = DeriveLabel(root)
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Get relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip excluded and hidden directories
		if d.IsDir() {
			if defaultExclusions[name] {
				return fs.SkipDir
			}
			// Skip hidden directories (starting with .) except root
			if strings.HasPrefix(name, ".") && path != root {
				return fs.SkipDir
			}
			// Check ignore patterns for directories
			if matcher != nil && matcher.Match(relPath) {
				return fs.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			return nil
		}

		// Check ignore patterns for files
		if matcher != nil && matcher.Match(relPath) {
			return nil
		}

		// Calculate depth
		depth := strings.Count(relPath, string(filepath.Separator))

		files = append(files, FileInfo{
			Path:        relPath,
			Depth:       depth,
			SourceDir:   absRoot,
			SourceLabel: sourceLabel,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by depth first, then alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].Depth != files[j].Depth {
			return files[i].Depth < files[j].Depth
		}
		return files[i].Path < files[j].Path
	})

	return files, nil
}

// WalkMultiple walks multiple directories and returns combined results.
// Each file's SourceDir and SourceLabel are set appropriately.
// Files are sorted: by source directory order, then by depth, then alphabetically.
// Duplicate files (by absolute path) are automatically deduplicated.
func WalkMultiple(roots []string, matcherFunc func(root string) (*ignore.Matcher, error)) ([]FileInfo, error) {
	var allFiles []FileInfo
	seenPaths := make(map[string]bool) // Absolute path -> bool for dedup

	// Resolve labels upfront to handle collisions
	labels := ResolveLabels(roots)

	for _, root := range roots {
		matcher, err := matcherFunc(root)
		if err != nil {
			return nil, fmt.Errorf("error loading ignore for %s: %w", root, err)
		}

		files, err := WalkWithLabel(root, matcher, labels[root])
		if err != nil {
			return nil, fmt.Errorf("error walking %s: %w", root, err)
		}

		// Deduplicate based on absolute file path
		for _, f := range files {
			absPath := filepath.Join(f.SourceDir, f.Path)
			if !seenPaths[absPath] {
				seenPaths[absPath] = true
				allFiles = append(allFiles, f)
			}
		}
	}

	return allFiles, nil
}

// DetectOverlaps checks if any directories contain each other.
// Returns pairs of [parent, child] for overlapping directories.
func DetectOverlaps(dirs []string) [][]string {
	var overlaps [][]string
	absDirs := make([]string, len(dirs))

	for i, d := range dirs {
		abs, _ := filepath.Abs(d)
		absDirs[i] = abs
	}

	for i := 0; i < len(absDirs); i++ {
		for j := i + 1; j < len(absDirs); j++ {
			if strings.HasPrefix(absDirs[j]+string(filepath.Separator), absDirs[i]+string(filepath.Separator)) {
				overlaps = append(overlaps, []string{dirs[i], dirs[j]})
			} else if strings.HasPrefix(absDirs[i]+string(filepath.Separator), absDirs[j]+string(filepath.Separator)) {
				overlaps = append(overlaps, []string{dirs[j], dirs[i]})
			}
		}
	}
	return overlaps
}

// DeriveLabel creates a display label from a directory path.
// "./docs" -> "docs", "/absolute/path/to/specs" -> "specs"
func DeriveLabel(dirPath string) string {
	clean := filepath.Clean(dirPath)
	return filepath.Base(clean)
}

// ResolveLabels generates unique labels for multiple directories.
// If two directories have the same basename, parent context is added.
// e.g., "./project1/docs" and "./project2/docs" become "project1-docs" and "project2-docs"
func ResolveLabels(dirs []string) map[string]string {
	labels := make(map[string]string)
	baseCounts := make(map[string][]string) // basename -> list of dirs with that base

	// First pass: group by basename
	for _, dir := range dirs {
		base := DeriveLabel(dir)
		baseCounts[base] = append(baseCounts[base], dir)
	}

	// Second pass: resolve collisions
	for _, dir := range dirs {
		base := DeriveLabel(dir)
		if len(baseCounts[base]) == 1 {
			// No collision, use simple basename
			labels[dir] = base
		} else {
			// Collision: add parent directory context
			clean := filepath.Clean(dir)
			parent := filepath.Base(filepath.Dir(clean))
			if parent == "." || parent == "/" {
				// No meaningful parent, use full path
				labels[dir] = strings.ReplaceAll(clean, string(filepath.Separator), "-")
			} else {
				labels[dir] = parent + "-" + base
			}
		}
	}

	return labels
}
