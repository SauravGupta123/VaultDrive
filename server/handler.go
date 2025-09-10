package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

		err := DeleteFile(cfg.StorageDir, filepath.Base(filename))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to delete file %s: %v", filename, err), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(APIResponse{Status: "success", Message: "file deleted"})
	}
}
