//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getDestPath(destDir, filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	return filepath.Join(destDir, base+".cmd")
}

func getLinkState(destDir, sourceDir, filename string) (LinkState, string) {
	destPath := getDestPath(destDir, filename)
	sourcePath := filepath.Join(sourceDir, filename)

	contentBytes, err := os.ReadFile(destPath)
	if err != nil {
		return StateNone, ""
	}

	content := string(contentBytes)
	absSource, _ := filepath.Abs(sourcePath)

	expectedTargetLine := fmt.Sprintf(`"%s" %%*`, absSource)
	if strings.Contains(content, expectedTargetLine) {
		return StateLinked, ""
	}

	// Try to extract the mismatching target
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "%*") {
			target := strings.TrimSpace(strings.TrimSuffix(line, "%*"))
			target = strings.Trim(target, `"`)
			return StateInvalid, target
		}
	}

	return StateNone, ""
}

func checkLinkConflict(destPath string) error {
	contentBytes, err := os.ReadFile(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	content := string(contentBytes)
	if !strings.HasPrefix(strings.TrimSpace(content), "@echo off") {
		return fmt.Errorf("skipped because destination exists and is not a recognized wrapper script")
	}
	return nil
}

func createLink(sourcePath, destPath string) error {
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}
	content := fmt.Sprintf("@echo off\r\n\"%s\" %%*\r\n", absSource)
	return os.WriteFile(destPath, []byte(content), 0755)
}

func removeLink(destPath string) error {
	return os.Remove(destPath)
}
