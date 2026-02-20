// Package runtime provides the embedded minimal Haira runtime.
// The minimal runtime contains zero-dependency stdlib modules (io, http,
// string, math, json, conv, array, map, regex, time, env, fs, log)
// that are embedded into the compiler binary via //go:embed.
package runtime

import "embed"

//go:embed embedded/haira/*.go
var minimalFS embed.FS

// Files returns all embedded runtime Go source files as a map of filename → content.
func Files() map[string][]byte {
	entries, err := minimalFS.ReadDir("embedded/haira")
	if err != nil {
		return nil
	}
	files := make(map[string][]byte, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := minimalFS.ReadFile("embedded/haira/" + e.Name())
		if err != nil {
			continue
		}
		files[e.Name()] = data
	}
	return files
}
