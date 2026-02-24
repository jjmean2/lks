//go:build windows

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// IsExecutable checks if the given file info represents an executable file.
func IsExecutable(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	ext := strings.ToUpper(filepath.Ext(info.Name()))
	if ext == "" {
		return false
	}

	pathext := os.Getenv("PATHEXT")
	if pathext == "" {
		pathext = ".COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC"
	}

	for _, e := range strings.Split(pathext, ";") {
		if ext == strings.ToUpper(e) {
			return true
		}
	}

	return false
}
