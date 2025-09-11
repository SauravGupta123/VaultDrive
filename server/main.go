package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	cfg := LoadConfig()

	// Get absolute path for clearer logging
	absPath, err := filepath.Abs(cfg.StorageDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for storage directory: %v", err)
	}

	http.HandleFunc("/upload", uploadHandler(cfg))
	http.HandleFunc("/delete", deleteHandler(cfg))
	http.HandleFunc("/list", listHandler(cfg))
	http.HandleFunc("/download", downloadHandler(cfg))

	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("Server running on %s\n", addr)
	fmt.Printf("Storage directory: %s (absolute path: %s)\n", cfg.StorageDir, absPath)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
