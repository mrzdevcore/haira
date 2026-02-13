package haira

import "os"

// Env returns the value of an environment variable.
func Env(key string) string {
	return os.Getenv(key)
}
