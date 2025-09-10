package main

import (
	"log"
)

type Config struct {
	ServerURL string
	WatchDir  string
}

func LoadConfig() *Config {
	server := "http://localhost:9090"
	dir := "./myfolder"

	log.Printf("[CONFIG] Server: %s | WatchDir: %s\n", server, dir)

	return &Config{
		ServerURL: server,
		WatchDir:  dir,
	}
}
