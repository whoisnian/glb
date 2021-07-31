package fsutil

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ResolveHomeDir resolve prefix '~' in path with environment variable.
func ResolveHomeDir(rawPath string) (string, error) {
	if strings.HasPrefix(rawPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, rawPath[2:]), nil
	}
	return filepath.Clean(rawPath), nil
}

// ResolveBase resolve rawPath within the basePath.
func ResolveBase(basePath string, rawPath string) string {
	if rawPath == "" || rawPath[0] != '/' {
		rawPath = "/" + rawPath
	}
	return filepath.Join(basePath, filepath.FromSlash(path.Clean(rawPath)))
}
