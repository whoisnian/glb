package fs

import (
	"os"
	"path/filepath"
	"strings"
)

func Clean(rawPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rawPath, "~/") {
		return filepath.Join(homeDir, rawPath[2:]), nil
	}
	return filepath.Clean(rawPath), nil
}
