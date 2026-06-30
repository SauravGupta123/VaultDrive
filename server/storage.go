package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// SaveFileWithChecksum saves a file and returns its SHA-256 checksum
func SaveFileWithChecksum(storageDir, filename string, file io.Reader) (string, error) {
	path := filepath.Join(storageDir, filename)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("failed to create directories for path %s: %w", path, err)
	}

	out, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer out.Close()

	// Calculate checksum while writing
	hash := sha256.New()
	multiWriter := io.MultiWriter(out, hash)

	_, err = io.Copy(multiWriter, file)
	if err != nil {
		// Delete partially written file on error
		os.Remove(path)
		return "", fmt.Errorf("failed to write file %s: %w", path, err)
	}

	checksum := fmt.Sprintf("%x", hash.Sum(nil))
	return checksum, nil
}

func DeleteFile(storageDir, filename string) error {
	// Sanitize filename to prevent directory traversal
	sanitizedFilename := filepath.Clean(filename)
	if strings.Contains(sanitizedFilename, "..") {
		return fmt.Errorf("invalid filename")
	}
	
	path := filepath.Join(storageDir, sanitizedFilename)
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
