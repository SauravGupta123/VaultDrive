package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func uploadFile(serverURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Calculate checksum while reading file
	hash := sha256.New()
	file.Seek(0, 0) // Reset to beginning
	
	// Use pipe to stream multipart form without loading into memory
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	// Start goroutine to write multipart form data
	go func() {
		defer pipeWriter.Close()
		defer writer.Close()

		// Add checksum field
		checksumWriter, err := writer.CreateFormField("checksum")
		if err != nil {
			pipeWriter.CloseWithError(fmt.Errorf("failed to create checksum field: %w", err))
			return
		}

		// Calculate checksum by reading file once
		teeReader := io.TeeReader(file, hash)
		tempBuf := new(bytes.Buffer)
		if _, err := io.Copy(tempBuf, teeReader); err != nil {
			pipeWriter.CloseWithError(fmt.Errorf("failed to read file for checksum: %w", err))
			return
		}
		checksum := fmt.Sprintf("%x", hash.Sum(nil))
		checksumWriter.Write([]byte(checksum))

		// Add file field
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			pipeWriter.CloseWithError(fmt.Errorf("failed to create form file: %w", err))
			return
		}

		// Stream file content from temp buffer
		if _, err := io.Copy(part, tempBuf); err != nil {
			pipeWriter.CloseWithError(fmt.Errorf("failed to copy file: %w", err))
			return
		}
	}()

	// Create HTTP request with streaming body
	req, err := http.NewRequest("POST", serverURL+"/upload", pipeReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	// Don't set Content-Length - let it be chunked transfer encoding
	// req.ContentLength = fileInfo.Size() // This was incorrect due to multipart overhead

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %s: %s", resp.Status, string(bodyBytes))
	}

	fmt.Printf("[UPLOAD] %s → success (size: %d bytes)\n", filepath.Base(filePath), fileInfo.Size())
	return nil
}

func deleteFile(serverURL, fileName string) error {
	url := fmt.Sprintf("%s/delete?name=%s", serverURL, fileName)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %s", resp.Status)
	}

	fmt.Printf("[DELETE] %s → success\n", fileName)
	return nil
}

func downloadFile(serverURL, fileName, localPath string) error {
	url := fmt.Sprintf("%s/download?name=%s", serverURL, fileName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create local file
	out, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	// Copy downloaded content to local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %w", err)
	}

	fmt.Printf("[DOWNLOAD] %s → success\n", fileName)
	return nil
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
