# File Synchronizer

A real-time file synchronization tool built with Go that automatically uploads files from a local directory to a remote server when they are created or modified, and deletes them remotely when removed locally.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Go Concepts Used](#go-concepts-used)
- [Architecture](#architecture)
- [Installation](#installation)
- [Usage](#usage)
- [CLI Commands](#cli-commands)
- [Configuration](#configuration)
- [Testing Concurrency](#testing-concurrency)
- [Building from Source](#building-from-source)
- [Project Structure](#project-structure)

## Introduction

File Synchronizer is a lightweight, efficient file synchronization tool that monitors a local directory for changes and automatically synchronizes files with a remote server. It uses file system notifications to detect changes in real-time and employs a worker pool pattern for concurrent file uploads, making it highly efficient for handling multiple files simultaneously.

The tool now includes a web-based file manager with HTML preview capabilities for text and PDF files, making it easier to browse and preview files stored on the server.

## Features

- **Real-time File Monitoring**: Automatically detects file creation, modification, and deletion events
- **Concurrent Uploads**: Uses a worker pool pattern with goroutines for parallel file processing
- **Graceful Shutdown**: Handles termination signals properly to ensure all uploads complete
- **Cross-platform Compatibility**: Works on Windows, macOS, and Linux
- **File Persistence**: Stores uploaded files on the server with proper directory structure
- **CLI Interface**: Command-line interface with multiple subcommands for different operations
- **File Listing**: List files stored on the server with metadata
- **Status Comparison**: Compare local and server files to see synchronization status
- **Single File Upload**: Upload individual files on demand
- **Two-way Synchronization**: Pull command for bidirectional file synchronization
- **Web-based File Manager**: Browser interface to manage files on the server
- **HTML Preview**: Preview text and PDF files directly in the browser
- **Download Files**: Download files from the server to your local machine
- **Logging**: Comprehensive logging for monitoring and debugging
- **Error Handling**: Robust error handling for network and file operations

## Technology Stack

- **Go (Golang)**: Main programming language
- **fsnotify**: File system notifications library
- **net/http**: Built-in HTTP client and server
- **mime/multipart**: Multipart form data handling
- **html/template**: HTML templating for web interface
- **Cobra**: CLI library for command-line interface

## Go Concepts Used

- **Goroutines**: Lightweight threads for concurrent file processing
- **Channels**: Communication mechanism between goroutines for job queuing
- **Worker Pool Pattern**: Efficiently manages concurrent file uploads with a fixed number of workers
- **WaitGroup**: Synchronization primitive to wait for all goroutines to complete
- **Context**: Graceful shutdown handling with signal notifications
- **Templates**: HTML templating for dynamic web content generation
- **Error Handling**: Proper error handling with descriptive messages
- **Concurrency Primitives**: Mutexes and other synchronization primitives for safe concurrent access

## Architecture

The application consists of two main components:

### Client
- Monitors a local directory for file system events using `fsnotify`
- Uses a worker pool pattern with goroutines for concurrent file uploads
- Communicates with the server via HTTP REST API
- Handles graceful shutdown on interrupt signals
- Provides CLI interface with multiple subcommands

### Server
- Receives file upload and delete requests via HTTP endpoints
- Stores files in a configurable directory
- Provides REST API endpoints for file operations
- Lists files with metadata (name, size, modification time)
- Includes web-based file manager with HTML preview capabilities

## Installation

### Prerequisites

- Go 1.24.1 or later
- Git (for cloning the repository)

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/SauravGupta123/FileSync.git
   cd FileSync
   ```

2. Build both client and server:
   ```bash
   go build -o  sync ./client
   go build -o server/server ./server
   ```
3. Setting alias to the terminal

 ```bash
     alias sync="/Users/path_to_sync/"
   ```

## Usage

### Starting the Server

```bash
# Basic usage (uses default port 9090 and storage directory ./data/synced)
./server/server

# With custom port and storage directory
./server/server -port=8080 -dir=/path/to/storage
```

### Starting the Client

The client now works as a CLI tool with multiple subcommands:

```bash
# Show help
sync --help

# Watch local directory for changes (previous behavior)
sync watch

# List files on server
sync list

# Show sync status
sync status

# Upload a single file
sync push <filename>

# Two-way synchronization
sync pull
```

## CLI Commands

### `sync list`

Lists all files stored on the server with their metadata.

```bash
sync list
```

Output example:
```
NAME                           SIZE       MODIFIED            
-----------------------------------------------------------------
document.pdf                   102400     2025-09-10 14:30:25
image.jpg                      204800     2025-09-10 14:32:10
notes.txt                      1024       2025-09-10 14:35:45
```

### `sync status`

Compares local files with server files and shows synchronization status.

```bash
sync status
```

Output example:
```
=== SYNC STATUS ===

Local only:
  local-file.txt (2048 bytes)

Server only:
  server-file.pdf (5120 bytes)

Synced:
  document.pdf (102400 bytes)
```

### `sync push <file>`

Uploads a single file to the server.

```bash
# Upload file from local watch directory
sync push document.pdf

# Upload file with absolute path
sync push /path/to/file.txt
```

### `sync pull`

Runs a one-time two-way synchronization between local and server directories. This command compares files on both sides and:
- Downloads files that exist on the server but not locally
- Uploads files that exist locally but not on the server
- For files that exist on both sides, compares modification times and synchronizes the newer version
- Creates backups before overwriting files

```bash
sync pull
```

### `sync watch`

Runs the file watcher that automatically synchronizes files when they change.

```bash
sync watch
```

## Configuration

### Client Configuration

The client configuration is hardcoded in `client/config.go`:
- **Server URL**: `http://localhost:9090`
- **Watch Directory**: `./client/myfolder`

### Server Configuration

The server configuration can be passed as command-line arguments:
- **Port**: `-port=9090` (default: 9090)
- **Storage Directory**: `-dir=./data/synced` (default: ./data/synced)

## Web Interface

The server includes a web-based file manager accessible at `http://localhost:9090/files` (or your configured port) that provides:

### File Manager Features
- Browse files and directories stored on the server
- Upload files directly through the web interface
- Download files from the server
- Delete files from the server
- Breadcrumb navigation for directory structure
- File previews for supported formats

### HTML Preview
- **Text Files**: Preview content of text-based files (txt, md, log, csv, json, xml, html, css, js, go, py, java, c, cpp)
- **PDF Files**: Embedded PDF viewer for direct preview
- **Images**: Thumbnail previews for JPG, JPEG, PNG, and GIF files
- File metadata display (size, modification time)
- Responsive design for various screen sizes

To access the web interface, start the server and navigate to `http://localhost:9090/files` in your browser.

## Testing Concurrency

To test the concurrent file upload feature:

1. Start the server:
   ```bash
   ./server/server
   ```

2. Start the client in watch mode:
   ```bash
   sync watch
   ```

3. Create multiple files simultaneously:
   ```bash
   # Create 10 test files at once
   for i in {1..10}; do echo "Test content $i" > client/myfolder/testfile$i.txt; done
   ```

4. Observe the logs to see multiple workers processing files concurrently:
   ```
   [Worker-1] starting upload testfile1.txt
   [Worker-2] starting upload testfile2.txt
   [Worker-3] starting upload testfile3.txt
   [Worker-1] finished testfile1.txt
   [Worker-2] finished testfile2.txt
   ```

## Go Packages Used

### Standard Library Packages
- **fmt**: Formatted I/O operations
- **net/http**: HTTP client and server implementations
- **os**: Platform-independent interface to operating system functionality
- **path/filepath**: File path manipulations
- **encoding/json**: JSON encoding and decoding
- **io**: Basic I/O interfaces and operations
- **html/template**: Data-driven templates for generating HTML
- **time**: Time and date operations
- **sync**: Synchronization primitives
- **crypto/sha256**: SHA-256 hash algorithm
- **mime/multipart**: Multipart form data handling

### Third-party Packages
- **github.com/fsnotify/fsnotify**: File system notifications
- **github.com/spf13/cobra**: CLI library for command-line applications

## Building from Source

To build the project from source:

1. Clone the repository:
   ```bash
   git clone https://github.com/SauravGupta123/FileSync.git
   cd FileSync
   ```

2. Build the client:
   ```bash
   go build -o sync ./client
   ```

3. Build the server:
   ```bash
   go build -o server/server ./server
   ```

4. Run the server:
   ```bash
   ./server/server
   ```

5. Run the client:
   ```bash
   ./sync --help
   ```

## Project Structure

```
FileSync/
├── client/
│   ├── config.go        # Client configuration
│   ├── main.go          # Client entry point with CLI commands
│   ├── uploader.go      # File upload/delete/download functions
│   ├── watcher.go       # File system watcher with worker pool
│   ├── sync.go          # Two-way synchronization functionality
│   └── myfolder/        # Default directory to watch for changes
├── server/
│   ├── config.go        # Server configuration
│   ├── main.go          # Server entry point
│   ├── handler.go       # HTTP request handlers
│   ├── storage.go       # File storage operations
│   ├── filemanager.go   # Web file manager and preview handlers
│   ├── templates/       # HTML templates for web interface
│   │   ├── files.html   # File manager template
│   │   └── preview.html # File preview template
│   └── data/
│       └── synced/      # Default directory for stored files
├── shared/
│   └── models.go        # Shared data structures
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md            # This file
```

## API Endpoints

### Server Endpoints

- `POST /upload` - Upload a file
- `DELETE /delete?name={filename}` - Delete a file
- `GET /list` - List files with metadata
- `GET /download?name={filename}` - Download a file
- `GET /files` - Web-based file manager interface
- `GET /preview?name={filename}` - HTML preview of text/PDF files

## Logging

Both client and server provide detailed logging:

### Client Logs
```
[CONFIG] Server: http://localhost:9090 | WatchDir: ./client/myfolder
[INFO] Creating directory: ./client/myfolder
[WATCH] Watching folder: ./client/myfolder with 3 workers
[EVENT] CREATE → /path/to/client/myfolder/test.txt
[Worker-1] starting upload test.txt
[Worker-1] finished test.txt
```

### Server Logs
```
Server running on :9090
Storage directory: ./data/synced
[UPLOAD] Received file: test.txt
[UPLOAD] Saved file: ./data/synced/test.txt
[DELETE] Deleted file: ./data/synced/test.txt
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.