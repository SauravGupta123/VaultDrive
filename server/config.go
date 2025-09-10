package main

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Port       string
	StorageDir string
}

func LoadConfig() *Config {
	port := flag.String("port", "9090", "Port for server to run on")
	storageDir := flag.String("dir", "./data/synced", "Directory to store synced files")
	flag.Parse()

	// Validate that the storage directory is writable
	if err := validateStorageDir(*storageDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	return &Config{
		Port:       *port,
		StorageDir: *storageDir,
	}
}

func validateStorageDir(dir string) error {
	// Try to create the directory
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory %s: %w", dir, err)
	}

	// Try to write a test file to ensure the directory is writable
	testFile := dir + "/.write_test"
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("storage directory %s is not writable: %w", dir, err)
	}
	defer os.Remove(testFile) // Clean up test file
	file.Close()

	return nil
}
