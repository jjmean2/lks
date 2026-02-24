//go:build !windows

package main

import "os"

// IsExecutable checks if the given file info represents an executable file.
func IsExecutable(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}
