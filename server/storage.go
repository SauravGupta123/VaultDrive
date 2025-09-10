package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func SaveFile(storageDir, filename string, file io.Reader) error {
	path := filepath.Join(storageDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directories for path %s: %w", path, err)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

func DeleteFile(storageDir, filename string) error {
	path := filepath.Join(storageDir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", path)
	}
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}
