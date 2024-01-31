package fileutils

import (
	"fmt"
	"os"
)

// IsDir checks if a path points to an existing directory
func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if fileInfo.Mode().IsDir() {
		return true, nil
	}
	return false, nil
}

// FileExists checks if a path points to an existing file
func FileExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err == nil {
		if fileInfo.Mode().IsRegular() {
			return true, nil
		}
		return false, fmt.Errorf("%s is not a file", path)
	}
	return false, err
}
