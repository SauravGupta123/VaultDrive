package main

import (
	"log"
	"os"
)

func main() {
	cfg := LoadConfig()

	// Ensure watch directory exists
	if _, err := os.Stat(cfg.WatchDir); os.IsNotExist(err) {
		log.Printf("[INFO] Creating directory: %s\n", cfg.WatchDir)
		os.MkdirAll(cfg.WatchDir, 0755)
	}

	// Start watching
	watchFolder(cfg)
}
