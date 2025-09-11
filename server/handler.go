package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func uploadHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form (max 10MB file for demo)
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename := filepath.Base(header.Filename) // sanitize filename
		err = SaveFile(cfg.StorageDir, filename, file)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to save file %s: %v", filename, err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(APIResponse{Status: "success", Message: "file uploaded"})
	}
}

func deleteHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("name")
		if filename == "" {
			http.Error(w, "missing filename", http.StatusBadRequest)
			return
		}

		// Sanitize filename to prevent directory traversal
		sanitizedFilename := filepath.Base(filename)

		err := DeleteFile(cfg.StorageDir, sanitizedFilename)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to delete file %s: %v", sanitizedFilename, err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(APIResponse{Status: "success", Message: "file deleted"})
	}
}

func listHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := ListFiles(cfg.StorageDir)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to list files: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	}
}

func downloadHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := r.URL.Query().Get("name")
		if filename == "" {
			http.Error(w, "missing filename", http.StatusBadRequest)
			return
		}

		// Sanitize filename to prevent directory traversal
		sanitizedFilename := filepath.Base(filename)
		filePath := filepath.Join(cfg.StorageDir, sanitizedFilename)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}

		// Set headers for file download
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", sanitizedFilename))
		w.Header().Set("Content-Type", "application/octet-stream")

		// Open file
		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to open file: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Copy file to response
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to send file: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
