package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
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

	// Create a channel for file upload jobs
	jobs := make(chan string, 100)

	// Number of worker goroutines
	numWorkers := 3

	// WaitGroup to wait for all workers to finish on shutdown
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, cfg.ServerURL, &wg)
	}

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Printf("[WATCH] Watching folder: %s with %d workers\n", cfg.WatchDir, numWorkers)

	// Main event loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				close(jobs)
				wg.Wait()
				return
			}

			// Debug log
			if event.Op != fsnotify.Chmod {
				log.Printf("[EVENT] %s → %s\n", event.Op, event.Name)
			}

			switch {
			case event.Op&fsnotify.Create == fsnotify.Create,
				event.Op&fsnotify.Write == fsnotify.Write:
				// Send file path to jobs channel instead of uploading directly
				jobs <- event.Name

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				go func(file string) {
					if err := deleteFile(cfg.ServerURL, filepath.Base(file)); err != nil {
						log.Printf("[ERROR] Delete failed: %v\n", err)
					}
				}(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				close(jobs)
				wg.Wait()
				return
			}
			log.Printf("[ERROR] Watcher error: %v\n", err)

		case <-sigChan:
			log.Println("[INFO] Shutdown signal received, closing jobs channel...")
			close(jobs)
			wg.Wait()
			log.Println("[INFO] All workers finished, exiting...")
			return
		}
	}
}

// worker processes file upload jobs from the jobs channel
func worker(id int, jobs <-chan string, serverURL string, wg *sync.WaitGroup) {
	defer wg.Done()

	for filePath := range jobs {
		log.Printf("[Worker-%d] starting upload %s\n", id, filepath.Base(filePath))

		// Delay slightly to avoid partial writes
		time.Sleep(500 * time.Millisecond)

		if err := uploadFile(serverURL, filePath); err != nil {
			log.Printf("[ERROR] Upload failed: %v\n", err)
		} else {
			log.Printf("[Worker-%d] finished %s\n", id, filepath.Base(filePath))
		}
	}

	log.Printf("[Worker-%d] shutting down\n", id)
}
