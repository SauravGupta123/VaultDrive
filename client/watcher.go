package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func watchFolder(cfg *Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("[ERROR] Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(cfg.WatchDir)
	if err != nil {
		log.Fatalf("[ERROR] Failed to watch dir %s: %v", cfg.WatchDir, err)
	}

	log.Printf("[WATCH] Watching folder: %s\n", cfg.WatchDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Debug log
			log.Printf("[EVENT] %s → %s\n", event.Op, event.Name)

			switch {
			case event.Op&fsnotify.Create == fsnotify.Create,
				event.Op&fsnotify.Write == fsnotify.Write:
				// Delay slightly to avoid partial writes
				go func(file string) {
					time.Sleep(500 * time.Millisecond)
					if err := uploadFile(cfg.ServerURL, file); err != nil {
						log.Printf("[ERROR] Upload failed: %v\n", err)
					}
				}(event.Name)

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go func(file string) {
					if err := deleteFile(cfg.ServerURL, filepath.Base(file)); err != nil {
						log.Printf("[ERROR] Delete failed: %v\n", err)
					}
				}(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("[ERROR] Watcher error: %v\n", err)
		}
	}
}
