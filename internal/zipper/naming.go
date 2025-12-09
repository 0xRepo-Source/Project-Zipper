package zipper

import (
	"fmt"
	"os"
	"path/filepath"
)

// NextArchiveName determines a unique zip filename for baseName within dir.
func NextArchiveName(dir, baseName string) (string, error) {
	if dir == "" {
		dir = "."
	}

	tryName := func(version int) string {
		if version == 0 {
			return filepath.Join(dir, fmt.Sprintf("%s.zip", baseName))
		}
		return filepath.Join(dir, fmt.Sprintf("%s-v%d.zip", baseName, version))
	}

	for version := 0; ; version++ {
		candidate := tryName(version)
		if _, err := os.Stat(candidate); err != nil {
			if os.IsNotExist(err) {
				return candidate, nil
			}
			return "", err
		}
	}
}

// NextGzipArchiveName determines a unique tar.gz filename for baseName within dir.
func NextGzipArchiveName(dir, baseName string) (string, error) {
	if dir == "" {
		dir = "."
	}

	tryName := func(version int) string {
		if version == 0 {
			return filepath.Join(dir, fmt.Sprintf("%s.tar.gz", baseName))
		}
		return filepath.Join(dir, fmt.Sprintf("%s-v%d.tar.gz", baseName, version))
	}

	for version := 0; ; version++ {
		candidate := tryName(version)
		if _, err := os.Stat(candidate); err != nil {
			if os.IsNotExist(err) {
				return candidate, nil
			}
			return "", err
		}
	}
}
