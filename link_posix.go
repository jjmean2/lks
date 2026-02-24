//go:build !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func getDestPath(destDir, filename string) string {
	return filepath.Join(destDir, filename)
}

func getLinkState(destDir, sourceDir, filename string) (LinkState, string) {
	destPath := getDestPath(destDir, filename)
	sourcePath := filepath.Join(sourceDir, filename)

	fileInfo, err := os.Lstat(destPath)
	if err != nil {
		return StateNone, ""
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(destPath)
		if err == nil {
			if !filepath.IsAbs(target) {
				target = filepath.Join(destDir, target)
			}
			absTarget, _ := filepath.Abs(target)
			absSource, _ := filepath.Abs(sourcePath)

			if absTarget == absSource {
				return StateLinked, ""
			}
			return StateInvalid, target
		}
	}
	return StateNone, ""
}

func checkLinkConflict(destPath string) error {
	fileInfo, err := os.Lstat(destPath)
	if err != nil {
		return nil
	}
	if fileInfo.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("skipped because destination exists and is not a symlink")
	}
	return nil
}

func createLink(sourcePath, destPath string) error {
	return os.Symlink(sourcePath, destPath)
}

func removeLink(destPath string) error {
	return os.Remove(destPath)
}
