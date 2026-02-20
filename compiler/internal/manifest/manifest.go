// Package manifest parses package.haira project manifest files.
//
// Format:
//
//	name = "my-project"
//	version = "0.1.0"
//	description = "Optional description"
//	entry = "main.haira"
//
// Lines starting with # are comments. Blank lines are ignored.
// If entry is not specified, it defaults to "main.haira".
package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Filename is the name of the project manifest file.
const Filename = "package.haira"

// Package represents a parsed package.haira manifest.
type Package struct {
	Name        string
	Version     string
	Description string
	Entry       string // default: "main.haira"
}

// Find looks for a package.haira file in the given directory.
// Returns the full path if found, or "" if not found.
func Find(dir string) string {
	path := filepath.Join(dir, Filename)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

// Load reads and parses a package.haira file.
func Load(path string) (*Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	pkg := &Package{
		Entry: "main.haira",
	}

	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		idx := strings.Index(line, "=")
		if idx < 0 {
			return nil, fmt.Errorf("%s:%d: expected key = \"value\"", path, i+1)
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Strip quotes
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		switch key {
		case "name":
			pkg.Name = value
		case "version":
			pkg.Version = value
		case "description":
			pkg.Description = value
		case "entry":
			pkg.Entry = value
		default:
			return nil, fmt.Errorf("%s:%d: unknown key %q", path, i+1, key)
		}
	}

	if pkg.Name == "" {
		return nil, fmt.Errorf("%s: missing required field 'name'", path)
	}
	if pkg.Version == "" {
		return nil, fmt.Errorf("%s: missing required field 'version'", path)
	}

	return pkg, nil
}

// DefaultManifest returns the content for a new package.haira file.
func DefaultManifest(name string) string {
	return fmt.Sprintf("name = %q\nversion = \"0.1.0\"\nentry = \"main.haira\"\n", name)
}
