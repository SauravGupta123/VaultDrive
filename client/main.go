package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SauravGupta123/FileSync/shared"
	"github.com/spf13/cobra"
)

func main() {
	cfg := LoadConfig()

	// Ensure watch directory exists
	if _, err := os.Stat(cfg.WatchDir); os.IsNotExist(err) {
		log.Printf("[INFO] Creating directory: %s\n", cfg.WatchDir)
		os.MkdirAll(cfg.WatchDir, 0755)
	}

	var rootCmd = &cobra.Command{
		Use:   "sync",
		Short: "File synchronization tool",
		Long:  "A tool for synchronizing files between local directory and remote server",
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List files on the server",
		Run: func(cmd *cobra.Command, args []string) {
			listFiles(cfg)
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show synchronization status",
		Run: func(cmd *cobra.Command, args []string) {
			showStatus(cfg)
		},
	}

	var pushCmd = &cobra.Command{
		Use:   "push [file]",
		Short: "Upload a file to the server",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pushFile(cfg, args[0])
		},
	}

	var watchCmd = &cobra.Command{
		Use:   "watch",
		Short: "Watch local directory and sync files automatically",
		Run: func(cmd *cobra.Command, args []string) {
			watchFolder(cfg)
		},
	}

	rootCmd.AddCommand(listCmd, statusCmd, pushCmd, watchCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func listFiles(cfg *Config) {
	resp, err := http.Get(cfg.ServerURL + "/list")
	if err != nil {
		log.Fatalf("[ERROR] Failed to fetch file list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("[ERROR] Server returned status: %s", resp.Status)
	}

	var files []shared.FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		log.Fatalf("[ERROR] Failed to decode response: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No files found on server")
		return
	}

	fmt.Printf("%-30s %-10s %-20s\n", "NAME", "SIZE", "MODIFIED")
	fmt.Println(strings.Repeat("-", 65))
	for _, file := range files {
		fmt.Printf("%-30s %-10d %-20s\n", file.Name, file.Size, file.ModifiedAt.Format("2006-01-02 15:04:05"))
	}
}

func showStatus(cfg *Config) {
	// Get server files
	resp, err := http.Get(cfg.ServerURL + "/list")
	if err != nil {
		log.Fatalf("[ERROR] Failed to fetch file list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("[ERROR] Server returned status: %s", resp.Status)
	}

	var serverFiles []shared.FileMetadata
	if err := json.NewDecoder(resp.Body).Decode(&serverFiles); err != nil {
		log.Fatalf("[ERROR] Failed to decode response: %v", err)
	}

	// Get local files
	localFiles, err := scanLocalFiles(cfg.WatchDir)
	if err != nil {
		log.Fatalf("[ERROR] Failed to scan local files: %v", err)
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

	// Categorize files
	var localOnly, serverOnly, synced []shared.FileMetadata

	// Check for local-only files
	for name, file := range localMap {
		if _, exists := serverMap[name]; !exists {
			localOnly = append(localOnly, file)
		}
	}

	// Check for server-only files
	for name, file := range serverMap {
		if _, exists := localMap[name]; !exists {
			serverOnly = append(serverOnly, file)
		}
	}

	// Check for synced files
	for name, localFile := range localMap {
		if serverFile, exists := serverMap[name]; exists {
			// For simplicity, we'll consider files synced if they have the same name
			// In a more robust implementation, we might compare sizes or checksums
			if localFile.Size == serverFile.Size {
				synced = append(synced, localFile)
			}
		}
	}

	// Print results
	fmt.Println("=== SYNC STATUS ===")

	if len(localOnly) > 0 {
		fmt.Println("\nLocal only:")
		for _, file := range localOnly {
			fmt.Printf("  %s (%d bytes)\n", file.Name, file.Size)
		}
	}

	if len(serverOnly) > 0 {
		fmt.Println("\nServer only:")
		for _, file := range serverOnly {
			fmt.Printf("  %s (%d bytes)\n", file.Name, file.Size)
		}
	}

	if len(synced) > 0 {
		fmt.Println("\nSynced:")
		for _, file := range synced {
			fmt.Printf("  %s (%d bytes)\n", file.Name, file.Size)
		}
	}

	if len(localOnly) == 0 && len(serverOnly) == 0 && len(synced) == 0 {
		fmt.Println("No files found locally or on server")
	}
}

func pushFile(cfg *Config, filePath string) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Try relative to watch directory
		relPath := filepath.Join(cfg.WatchDir, filePath)
		if _, err := os.Stat(relPath); os.IsNotExist(err) {
			log.Fatalf("[ERROR] File does not exist: %s", filePath)
		}
		absPath = relPath
	}

	if err := uploadFile(cfg.ServerURL, absPath); err != nil {
		log.Fatalf("[ERROR] Failed to upload file: %v", err)
	}

	fmt.Printf("[SUCCESS] File uploaded: %s\n", filepath.Base(absPath))
}

func scanLocalFiles(dir string) ([]shared.FileMetadata, error) {
	var files []shared.FileMetadata

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path from watch directory
		relPath, err := filepath.Rel(dir, path)
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

	return files, err
}
