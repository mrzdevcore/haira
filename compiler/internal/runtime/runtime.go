// Package runtime provides the embedded Haira runtime.
// The runtime (stdlib, agentic, UI, dependencies) is bundled as a tar.gz
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
	"strings"
	"sync"
)

//go:embed bundle.tar.gz
var bundleData []byte

// cached extraction results
var (
	once       sync.Once
	goFiles    map[string][]byte
	uiFiles    map[string][]byte
	goMod      []byte
	goSum      []byte
	extractErr error
)

// extract unpacks the tar.gz bundle once and caches the results.
func extract() {
	once.Do(func() {
		goFiles = make(map[string][]byte)
		uiFiles = make(map[string][]byte)

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
			case strings.HasPrefix(name, "haira/ui/"):
				// UI files: strip "haira/" prefix → "ui/dist/haira-ui.js"
				rel := strings.TrimPrefix(name, "haira/")
				uiFiles[rel] = data
			case strings.HasPrefix(name, "haira/") && strings.HasSuffix(name, ".go"):
				// Go source files: "haira/agent.go" → "agent.go"
				goName := filepath.Base(name)
				goFiles[goName] = data
			}
		}
	})
}

// GoFiles returns all runtime Go source files as a map of filename → content.
func GoFiles() map[string][]byte {
	extract()
	return goFiles
}

// UIFiles returns all UI files (HTML + dist/) as a map of relative path → content.
func UIFiles() map[string][]byte {
	extract()
	return uiFiles
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
