package fsx

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func WriteFileAtomic(path string, content []byte, mode os.FileMode) error {
	parent := filepath.Dir(path)
	base := filepath.Base(path)

	tempFile, err := os.CreateTemp(parent, "."+base+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tempFile.Chmod(mode); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		if runtime.GOOS != "windows" {
			return fmt.Errorf("rename temp file: %w", err)
		}
		if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
			return fmt.Errorf("remove destination before rename: %w", removeErr)
		}
		if renameErr := os.Rename(tempPath, path); renameErr != nil {
			return fmt.Errorf("rename temp file after remove: %w", renameErr)
		}
	}
	cleanup = false

	// #nosec G304 -- parent directory path is derived from explicit caller-provided destination path.
	if dirHandle, err := os.Open(parent); err == nil {
		_ = dirHandle.Sync()
		_ = dirHandle.Close()
	}
	return nil
}
