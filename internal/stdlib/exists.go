// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package stdlib

import (
	"io/fs"
	"os"
)

// IsDirExists returns true if the path exists and is a directory.
func IsDirExists(path string) (bool, error) {
	return isDirExists(os.Stat(path))
}

// IsFileExists returns true if the path exists and is a regular file.
func IsFileExists(path string) (bool, error) {
	return isFileExists(os.Stat(path))
}

// IsDirExistsFS returns true if the path exists and is a directory.
// Accepts filesystem as a parameter, so this will work with any filesystem.
func IsDirExistsFS(path string, filesystem fs.FS) (bool, error) {
	return isDirExists(fs.Stat(filesystem, path))
}

// IsFileExistsFS returns true if the path exists and is a regular file.
// Accepts filesystem as a parameter, so this will work with any filesystem.
func IsFileExistsFS(path string, filesystem fs.FS) (bool, error) {
	return isFileExists(fs.Stat(filesystem, path))
}

// isDirExists returns true if the item exists and is a directory.
func isDirExists(sb fs.FileInfo, err error) (bool, error) {
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return sb.IsDir(), nil
}

// isFileExists returns true if the path exists and is a regular file.
func isFileExists(sb fs.FileInfo, err error) (bool, error) {
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	} else if sb.IsDir() {
		return false, nil
	}
	return sb.Mode().IsRegular(), nil
}
