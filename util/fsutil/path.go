package fsutil

import (
	"os"
	"path/filepath"
	"strings"
)

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
