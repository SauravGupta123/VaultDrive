# File Synchronizer

A real-time file synchronization tool built with Go that automatically uploads files from a local directory to a remote server when they are created or modified, and deletes them remotely when removed locally.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Technology Stack](#technology-stack)
- [Architecture](#architecture)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Testing Concurrency](#testing-concurrency)
- [Building from Source](#building-from-source)
- [Project Structure](#project-structure)

## Introduction

File Synchronizer is a lightweight, efficient file synchronization tool that monitors a local directory for changes and automatically synchronizes files with a remote server. It uses file system notifications to detect changes in real-time and employs a worker pool pattern for concurrent file uploads, making it highly efficient for handling multiple files simultaneously.

## Features

- **Real-time File Monitoring**: Automatically detects file creation, modification, and deletion events
- **Concurrent Uploads**: Uses a worker pool pattern with goroutines for parallel file processing
- **Graceful Shutdown**: Handles termination signals properly to ensure all uploads complete
- **Cross-platform Compatibility**: Works on Windows, macOS, and Linux
- **File Persistence**: Stores uploaded files on the server with proper directory structure
- **Logging**: Comprehensive logging for monitoring and debugging
- **Error Handling**: Robust error handling for network and file operations

## Technology Stack

- **Go (Golang)**: Main programming language
- **fsnotify**: File system notifications library
- **net/http**: Built-in HTTP client and server
- **mime/multipart**: Multipart form data handling

## Architecture

The application consists of two main components:

### Client
- Monitors a local directory for file system events using `fsnotify`
- Uses a worker pool pattern with goroutines for concurrent file uploads
- Communicates with the server via HTTP REST API
- Handles graceful shutdown on interrupt signals

### Server
- Receives file upload and delete requests via HTTP endpoints
- Stores files in a configurable directory
- Provides REST API endpoints for file operations

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
   go build -o client/client ./client
   go build -o server/server ./server
   ```

3. Run the server:
   ```bash
   ./server/server
   ```

4. Run the client (in a new terminal):
   ```bash
   ./client/client
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

```bash
# Basic usage (uses default server URL http://localhost:9090 and watch directory ./myfolder)
./client/client
```

### Testing the Synchronization

1. Create or modify files in the client's watch directory (`./myfolder` by default)
2. Observe the logs showing file uploads
3. Check that files appear in the server's storage directory (`./data/synced` by default)
4. Delete files from the watch directory and observe them being removed from the server

## Configuration

### Client Configuration

The client configuration is hardcoded in `client/config.go`:
- **Server URL**: `http://localhost:9090`
- **Watch Directory**: `./myfolder`

### Server Configuration

The server configuration can be passed as command-line arguments:
- **Port**: `-port=9090` (default: 9090)
- **Storage Directory**: `-dir=./data/synced` (default: ./data/synced)

## Testing Concurrency

To test the concurrent file upload feature:

1. Start the server:
   ```bash
   ./server/server
   ```

2. Start the client:
   ```bash
   ./client/client
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

## Building from Source

### Prerequisites

Ensure you have Go installed (version 1.24.1 or later):
```bash
go version
```

### Building

1. Clone the repository:
   ```bash
   git clone https://github.com/SauravGupta123/FileSync.git
   cd FileSync
   ```

2. Initialize Go modules (if needed):
   ```bash
   go mod tidy
   ```

3. Build the client and server:
   ```bash
   go build -o client/client ./client
   go build -o server/server ./server
   ```

### Cross-compilation

To build for different platforms:

```bash
# For Windows
GOOS=windows GOARCH=amd64 go build -o client/client.exe ./client
GOOS=windows GOARCH=amd64 go build -o server/server.exe ./server

# For Linux
GOOS=linux GOARCH=amd64 go build -o client/client ./client
GOOS=linux GOARCH=amd64 go build -o server/server ./server

# For macOS
GOOS=darwin GOARCH=amd64 go build -o client/client ./client
GOOS=darwin GOARCH=amd64 go build -o server/server ./server
```

## Project Structure

```
FileSync/
├── client/
│   ├── config.go        # Client configuration
│   ├── main.go          # Client entry point
│   ├── uploader.go      # File upload/delete functions
│   ├── watcher.go       # File system watcher with worker pool
│   └── myfolder/        # Default directory to watch for changes
├── server/
│   ├── config.go        # Server configuration
│   ├── main.go          # Server entry point
│   ├── handler.go       # HTTP request handlers
│   ├── storage.go       # File storage operations
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

## Logging

Both client and server provide detailed logging:

### Client Logs
```
[CONFIG] Server: http://localhost:9090 | WatchDir: ./myfolder
[INFO] Creating directory: ./myfolder
[WATCH] Watching folder: ./myfolder with 3 workers
[EVENT] CREATE → /path/to/myfolder/test.txt
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