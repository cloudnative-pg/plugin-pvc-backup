/*
Copyright The CloudNativePG Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fileutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// This implementation is based on https://github.com/nmrshll/go-cp/blob/master/cp.go

func replaceHomeFolder(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	var buffer bytes.Buffer
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	_, err = buffer.WriteString(usr.HomeDir)
	if err != nil {
		return "", err
	}
	_, err = buffer.WriteString(strings.TrimPrefix(path, "~"))
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// AbsolutePath converts a path (relative or absolute) into an absolute one.
// Supports '~' notation for $HOME directory of the current user.
func AbsolutePath(path string) (string, error) {
	homeReplaced, err := replaceHomeFolder(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(homeReplaced)
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherwise, attempt to create a hard link
// between the two files. If that fails, copy the file contents from src to dst.
// Creates any missing directories. Supports '~' notation for $HOME directory of the current user.
func CopyFile(src, dst string) error {
	srcAbs, err := AbsolutePath(src)
	if err != nil {
		return err
	}
	dstAbs, err := AbsolutePath(dst)
	if err != nil {
		return err
	}

	// open source file
	sfi, err := os.Stat(srcAbs)
	if err != nil {
		return err
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}

	// open dest file
	dfi, err := os.Stat(dstAbs)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err != nil {
		// file doesn't exist
		err := os.MkdirAll(filepath.Dir(dst), 0o750)
		if err != nil {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return err
		}
	}
	if err = os.Link(src, dst); err == nil {
		return err
	}
	return copyFileContents(src, dst)
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) error {
	// Open the source file for reading
	srcFile, err := os.Open(src) // nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	// Open the destination file for writing
	dstFile, err := os.Create(dst) // nolint:gosec
	if err != nil {
		return err
	}
	// Return any errors that result from closing the destination file
	// Will return nil if no errors occurred
	defer func() {
		cerr := dstFile.Close()
		if err == nil {
			err = cerr
		}
	}()

	// Copy the contents of the source file into the destination files
	const size = 1024 * 1024
	buf := make([]byte, size)
	if _, err = io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		return err
	}
	err = dstFile.Sync()
	return err
}
