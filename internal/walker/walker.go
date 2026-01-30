package walker

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// Default directories to exclude from scanning.
var defaultExclusions = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

// FileInfo holds information about a discovered markdown file.
type FileInfo struct {
	Path  string // Relative path from root
	Depth int    // Directory depth (0 = root level)
}

// Walk recursively finds all .md files in the given root directory.
// Files are returned sorted by depth (shallower first), then alphabetically.
// Automatically skips .git, node_modules, vendor, and hidden directories.
func Walk(root string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Skip excluded and hidden directories
		if d.IsDir() {
			if defaultExclusions[name] {
				return fs.SkipDir
			}
			// Skip hidden directories (starting with .) except root
			if strings.HasPrefix(name, ".") && path != root {
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

		// Get relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Calculate depth
		depth := strings.Count(relPath, string(filepath.Separator))

		files = append(files, FileInfo{
			Path:  relPath,
			Depth: depth,
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
