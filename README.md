# VaultDrive

A production-grade real-time file synchronization tool built with Go. Automatically syncs files between local directory and remote server with data integrity verification.

## Features

- 🚀 **Real-time sync** - Automatically uploads/deletes files as they change
- 💾 **Memory efficient** - Streams large files without loading into memory
- ✅ **Data integrity** - SHA-256 checksum verification on every upload
- 🔄 **Smart deduplication** - Prevents duplicate uploads from multiple file events
- ⚡ **Concurrent uploads** - 3 worker threads for parallel processing
- 🌐 **Web interface** - Browser-based file manager with preview
- 📦 **Two-way sync** - Pull command for bidirectional synchronization

## Quick Start

### Prerequisites

- Go 1.20 or later
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/SauravGupta123/FileSync.git
cd FileSync

# Build binaries
go build -o sync ./client
go build -o server/server ./server
```

### Usage

**1. Start the server:**

```bash
./server/server
# Server starts on http://localhost:9090
# Files stored in ./data/synced/
```

**2. Start the client (in another terminal):**

```bash
# Watch mode - automatically sync changes
./sync watch

# List files on server
./sync list

# Check sync status
./sync status

# Upload single file
./sync push myfile.txt

# Two-way sync
./sync pull
```

**3. Access web interface:**

Open http://localhost:9090/files in your browser

## Configuration

Edit `client/config.go` and `server/config.go` to customize:

**Client:**
- Server URL (default: `http://localhost:9090`)
- Watch directory (default: `./myfolder`)

**Server:**
- Port (default: `9090`)
- Storage directory (default: `./data/synced`)

## How It Works

### Intelligent Event Handling

When you save a file, your editor may trigger multiple file system events. VaultDrive uses **1-second debouncing** to coalesce these events into a single upload.

**Before:** 10 file events = 10 duplicate uploads  
**After:** 10 file events = 1 smart upload after changes complete

### Memory-Efficient Streaming

Traditional approach loads entire file into memory before uploading. VaultDrive streams files directly.

**Before:** 500MB file = 500MB memory usage (OOM risk)  
**After:** 500MB file = ~10MB memory usage (constant)

### Data Integrity

Every upload is verified with SHA-256 checksum. Corrupted uploads are automatically rejected.

```
Client: Upload file → Calculate SHA-256 → Send hash
Server: Receive file → Calculate SHA-256 → Verify match → Accept/Reject
```

## Examples

### Example 1: Auto-sync a directory

```bash
# Terminal 1: Start server
./server/server

# Terminal 2: Start watching
./sync watch

# Terminal 3: Make changes
cd myfolder
echo "Hello World" > test.txt
# File automatically uploads in ~1 second
```

### Example 2: Sync between machines

```bash
# Machine A: Upload files
cd myfolder
cp ~/documents/*.pdf .
# Files auto-upload

# Machine B: Download files
./sync pull
# Files downloaded to myfolder/
```

### Example 3: Web interface

1. Start server: `./server/server`
2. Visit: http://localhost:9090/files
3. Upload, download, delete, or preview files in browser

## Testing

### Test Event Deduplication

```bash
./server/server &
./sync watch &

# Rapidly modify a file
for i in {1..10}; do echo "line $i" >> myfolder/test.txt; sleep 0.05; done

# Check logs - should show only 1 upload despite 10 modifications
```

### Test Large File Handling

```bash
# Create 100MB test file
dd if=/dev/zero of=myfolder/large.bin bs=1M count=100

# Watch logs - memory usage stays low, upload succeeds with checksum
```

### Test Data Integrity

```bash
# Files uploaded show checksum in server logs:
# [UPLOAD] Saved file: test.txt (checksum: abc123...)

# Verify on server
shasum -a 256 data/synced/test.txt
```

## Architecture

```
┌─────────────┐         HTTP          ┌─────────────┐
│   Client    │◄─────────────────────►│   Server    │
│             │                        │             │
│ - fsnotify  │   POST /upload         │ - REST API  │
│ - debounce  │   (with checksum)      │ - Storage   │
│ - workers   │                        │ - Verify    │
│ - stream    │   GET /download        │ - Web UI    │
└─────────────┘   DELETE /delete       └─────────────┘
                  GET /list
```

**Client:**
- Watches local directory with `fsnotify`
- Debounces events (1s window)
- 3 worker goroutines upload concurrently
- Streams files with `io.Pipe()` (no memory buffering)
- Calculates SHA-256 during upload

**Server:**
- REST API for upload/download/delete/list
- Verifies checksums, rejects corrupted uploads
- Streams to disk with `io.MultiWriter()`
- Web interface on `/files`

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/upload` | Upload file with checksum |
| GET | `/download?name=file` | Download file |
| DELETE | `/delete?name=file` | Delete file |
| GET | `/list` | List all files (JSON) |
| GET | `/files` | Web file manager |
| GET | `/preview?name=file` | Preview file in browser |

## Tech Stack

**Language:** Go 1.20+

**Key Libraries:**
- `fsnotify` - File system event monitoring
- `cobra` - CLI interface
- `net/http` - HTTP client/server
- `crypto/sha256` - Checksum calculation

**Go Concepts:**
- Goroutines & channels (worker pool)
- `io.Pipe()` (streaming)
- `sync.Mutex` (concurrent access)
- `io.TeeReader` / `io.MultiWriter` (hash during I/O)
- Context-based shutdown

## Project Structure

```
VaultDrive/
├── client/
│   ├── main.go          # CLI commands
│   ├── watcher.go       # fsnotify + debouncing
│   ├── uploader.go      # Streaming upload
│   ├── sync.go          # Two-way sync
│   └── config.go        # Configuration
├── server/
│   ├── main.go          # HTTP server
│   ├── handler.go       # Request handlers
│   ├── storage.go       # File operations
│   ├── filemanager.go   # Web UI
│   └── templates/       # HTML templates
├── shared/
│   └── models.go        # Shared types
└── data/
    └── synced/          # Uploaded files
```

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open Pull Request

## License

MIT License - see LICENSE file for details

## Future Enhancements

- [ ] Chunked uploads with resume capability
- [ ] Retry queue with exponential backoff
- [ ] Conflict detection for simultaneous edits
- [ ] File compression for text files
- [ ] End-to-end encryption
- [ ] Multi-user authentication
- [ ] Progress tracking with cancellation
- [ ] Database for metadata tracking

## Troubleshooting

**Server won't start:**
```bash
# Check if port 9090 is in use
lsof -i :9090
# Kill existing process or change port in config
```

**Files not syncing:**
```bash
# Check client logs for errors
./sync watch
# Check server logs
./server/server
# Verify network connectivity
curl http://localhost:9090/list
```

**Large file upload fails:**
- Increase server timeout if needed
- Check available disk space
- Monitor memory usage

## Contact

For issues or questions, please open an issue on GitHub.
