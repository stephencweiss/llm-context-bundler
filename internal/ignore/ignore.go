package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// Matcher checks if paths should be ignored based on patterns.
type Matcher struct {
	patterns []pattern
}

type pattern struct {
	glob        glob.Glob
	negation    bool
	matchByName bool // If true, match against basename only
}

// New creates a new Matcher from a .lcbignore file.
// Returns an empty matcher if the file doesn't exist.
func New(ignoreFilePath string) (*Matcher, error) {
	m := &Matcher{}

	file, err := os.Open(ignoreFilePath)
	if os.IsNotExist(err) {
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for negation
		negation := false
		if strings.HasPrefix(line, "!") {
			negation = true
			line = line[1:]
		}

		// Determine if this pattern should match by name only
		matchByName := !strings.Contains(line, "/")

		// Convert gitignore-style patterns to glob patterns
		globPattern := convertToGlob(line)

		g, err := glob.Compile(globPattern, filepath.Separator)
		if err != nil {
			// Skip invalid patterns
			continue
		}

		m.patterns = append(m.patterns, pattern{
			glob:        g,
			negation:    negation,
			matchByName: matchByName,
		})
	}

	return m, scanner.Err()
}

// Match returns true if the given path should be ignored.
func (m *Matcher) Match(path string) bool {
	ignored := false
	basename := filepath.Base(path)

	for _, p := range m.patterns {
		var matches bool
		if p.matchByName {
			matches = p.glob.Match(basename)
		} else {
			matches = p.glob.Match(path)
		}
		if matches {
			ignored = !p.negation
		}
	}

	return ignored
}

// convertToGlob converts gitignore-style patterns to glob patterns.
func convertToGlob(pattern string) string {
	// Handle directory-only patterns (ending with /)
	pattern = strings.TrimSuffix(pattern, "/")

	// Handle patterns starting with /
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
	}

	return pattern
}
