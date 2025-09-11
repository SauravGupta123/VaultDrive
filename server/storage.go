package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SauravGupta123/FileSync/shared"
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

func ListFiles(storageDir string) ([]shared.FileMetadata, error) {
	var files []shared.FileMetadata

	err := filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from storage directory
		relPath, err := filepath.Rel(storageDir, path)
		if err != nil {
			return err
		}

		// Create file metadata
		fileMeta := shared.FileMetadata{
			Name:       relPath,
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		}

		files = append(files, fileMeta)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}
