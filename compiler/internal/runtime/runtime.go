// Package runtime provides the embedded Haira runtime.
// The runtime (core, stdlib packages, UI, dependencies) is bundled as a tar.gz
// archive at build time and embedded into the compiler binary via //go:embed.
// At compile time, the archive is extracted to provide Go source files,
// UI assets, and module files to the temp build directory.
package runtime

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

//go:embed bundle.tar.gz
var bundleData []byte

// cached extraction results
var (
	once       sync.Once
	goFiles    map[string][]byte // key = "haira/agent.go" or "postgres/postgres.go"
	uiFiles    map[string][]byte
	cliFiles   map[string][]byte // CLI-only files (e.g., haira-ui.js for `haira webui`)
	goMod      []byte
	goSum      []byte
	extractErr error
)

// extract unpacks the tar.gz bundle once and caches the results.
func extract() {
	once.Do(func() {
		goFiles = make(map[string][]byte)
		uiFiles = make(map[string][]byte)
		cliFiles = make(map[string][]byte)

		gz, err := gzip.NewReader(bytes.NewReader(bundleData))
		if err != nil {
			extractErr = err
			return
		}
		defer gz.Close()

		tr := tar.NewReader(gz)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				extractErr = err
				return
			}
			if hdr.Typeflag != tar.TypeReg {
				continue
			}

			data, err := io.ReadAll(tr)
			if err != nil {
				extractErr = err
				return
			}

			name := filepath.ToSlash(filepath.Clean(hdr.Name))

			switch {
			case name == "go.mod":
				goMod = data
			case name == "go.sum":
				goSum = data
			case strings.HasPrefix(name, "cli/"):
				// CLI-only files: used by `haira webui`, not compiled programs
				rel := strings.TrimPrefix(name, "cli/")
				cliFiles[rel] = data
			case strings.HasPrefix(name, "haira/ui/"):
				// UI files: strip "haira/" prefix → "ui/loader.html"
				rel := strings.TrimPrefix(name, "haira/")
				uiFiles[rel] = data
			case strings.HasSuffix(name, ".go"):
				// Go source files: preserve package path
				// "haira/agent.go", "postgres/postgres.go", etc.
				goFiles[name] = data
			}
		}
	})
}

// GoFiles returns all runtime Go source files as a map of path → content.
// Keys include the package directory: "haira/agent.go", "postgres/postgres.go", etc.
func GoFiles() map[string][]byte {
	extract()
	return goFiles
}

// GoFilesForPackage returns Go source files for a specific package directory.
// For example, GoFilesForPackage("haira") returns {"agent.go": data, "io.go": data, ...}.
func GoFilesForPackage(pkg string) map[string][]byte {
	extract()
	result := make(map[string][]byte)
	prefix := pkg + "/"
	for path, data := range goFiles {
		if strings.HasPrefix(path, prefix) {
			filename := strings.TrimPrefix(path, prefix)
			// Skip files in subdirectories (e.g., haira/ui/ files)
			if !strings.Contains(filename, "/") {
				result[filename] = data
			}
		}
	}
	return result
}

// StdlibPackages returns the list of stdlib package directory names found in the bundle.
// These are packages other than "haira" that contain Go source files.
func StdlibPackages() []string {
	extract()
	seen := map[string]bool{}
	for path := range goFiles {
		dir := filepath.Dir(path)
		if dir != "haira" && dir != "." {
			seen[dir] = true
		}
	}
	var pkgs []string
	for pkg := range seen {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	return pkgs
}

// UIFiles returns UI files (e.g. observe.html) as a map of relative path → content.
// These are included in compiled .haira programs.
func UIFiles() map[string][]byte {
	extract()
	return uiFiles
}

// CLIFiles returns CLI-only files (e.g., haira-ui.js) as a map of filename → content.
// These are embedded in the haira CLI binary but NOT in compiled .haira programs.
func CLIFiles() map[string][]byte {
	extract()
	return cliFiles
}

// GoMod returns the runtime go.mod content.
func GoMod() []byte {
	extract()
	return goMod
}

// GoSum returns the runtime go.sum content.
func GoSum() []byte {
	extract()
	return goSum
}
