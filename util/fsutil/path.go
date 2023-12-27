package fsutil

import (
	"os"
	"path"
	"path/filepath"
)

// ExpandHomeDir expands prefix '~' in rawFilePath with HOME environment variable.
func ExpandHomeDir(rawFilePath string) (string, error) {
	if len(rawFilePath) == 0 || rawFilePath[0] != '~' || (len(rawFilePath) > 1 && rawFilePath[1] != '/' && rawFilePath[1] != '\\') {
		return filepath.Clean(rawFilePath), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || len(rawFilePath) == 1 {
		return homeDir, err
	}
	return filepath.Join(homeDir, rawFilePath[1:]), nil
}

// ResolveUrlPath resolves rawUrlPath within the baseFilePath.
func ResolveUrlPath(baseFilePath string, rawUrlPath string) string {
	if rawUrlPath == "" || rawUrlPath[0] != '/' {
		rawUrlPath = "/" + rawUrlPath
	}
	return filepath.Join(baseFilePath, filepath.FromSlash(path.Clean(rawUrlPath)))
}
