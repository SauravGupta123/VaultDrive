package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/SauravGupta123/FileSync/shared"
)

func syncPull(cfg *Config) error {
	fmt.Println("Starting two-way synchronization...")

	// Get server files
	resp, err := http.Get(cfg.ServerURL + "/list")
	if err != nil {
		return fmt.Errorf("failed to fetch file list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	var serverFiles []shared.FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&serverFiles); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Get local files
	localFiles, err := scanLocalFiles(cfg.WatchDir)
	if err != nil {
		return fmt.Errorf("failed to scan local files: %w", err)
	}

	// Create maps for easier comparison
	serverMap := make(map[string]shared.FileMetadata)
	localMap := make(map[string]shared.FileMetadata)

	for _, file := range serverFiles {
		serverMap[file.Name] = file
	}

	for _, file := range localFiles {
		localMap[file.Name] = file
	}

	// Process files
	fmt.Println("Comparing files...")

	// Download files that exist on server but not locally
	for name, serverFile := range serverMap {
		if localFile, exists := localMap[name]; !exists {
			// File exists on server but not locally - download it
			localPath := filepath.Join(cfg.WatchDir, name)
			fmt.Printf("[DOWNLOAD] %s\n", name)
			
			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
				log.Printf("[ERROR] Failed to create directory for %s: %v", name, err)
				continue
			}
			
			if err := downloadFile(cfg.ServerURL, name, localPath); err != nil {
				log.Printf("[ERROR] Failed to download %s: %v", name, err)
			}
		} else {
			// File exists on both - compare modification times
			if serverFile.ModifiedAt.After(localFile.ModifiedAt) {
				// Server version is newer - download it
				localPath := filepath.Join(cfg.WatchDir, name)
				fmt.Printf("[DOWNLOAD] %s (server version is newer)\n", name)
				
				// Backup local file before overwriting
				backupPath := localPath + ".backup." + time.Now().Format("20060102150405")
				if err := os.Rename(localPath, backupPath); err != nil {
					log.Printf("[ERROR] Failed to backup %s: %v", name, err)
				}
				
				if err := downloadFile(cfg.ServerURL, name, localPath); err != nil {
					log.Printf("[ERROR] Failed to download %s: %v", name, err)
					// Restore backup if download failed
					if err := os.Rename(backupPath, localPath); err != nil {
						log.Printf("[ERROR] Failed to restore backup for %s: %v", name, err)
					}
				} else {
					// Remove backup if download succeeded
					os.Remove(backupPath)
				}
			} else if localFile.ModifiedAt.After(serverFile.ModifiedAt) {
				// Local version is newer - upload it
				localPath := filepath.Join(cfg.WatchDir, name)
				fmt.Printf("[UPLOAD] %s (local version is newer)\n", name)
				
				// Backup server file before overwriting
				// (In a real implementation, we might want to implement server-side backup)
				
				if err := uploadFile(cfg.ServerURL, localPath); err != nil {
					log.Printf("[ERROR] Failed to upload %s: %v", name, err)
				}
			} else {
				// Files have the same modification time - check content (simplified approach)
				fmt.Printf("[SKIP] %s (already in sync)\n", name)
			}
		}
	}

	// Upload files that exist locally but not on server
	for name := range localMap {
		if _, exists := serverMap[name]; !exists {
			// File exists locally but not on server - upload it
			localPath := filepath.Join(cfg.WatchDir, name)
			fmt.Printf("[UPLOAD] %s\n", name)
			
			if err := uploadFile(cfg.ServerURL, localPath); err != nil {
				log.Printf("[ERROR] Failed to upload %s: %v", name, err)
			}
		}
	}

	fmt.Println("Synchronization completed.")
	return nil
}