package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileInfo represents a file or directory in the file manager
type FileInfo struct {
	Name     string
	Size     int64
	ModTime  time.Time
	IsDir    bool
	RelPath  string
	Depth    int
}

// FileManagerData holds data for the file manager template
type FileManagerData struct {
	Files          []FileInfo
	CurrentPath    string
	BasePath       string
	Breadcrumb     []string
	SuccessMessage string
	ErrorMessage   string
}

// PreviewData holds data for the file preview template
type PreviewData struct {
	Filename  string
	FilePath  string
	BasePath  string
	Size      int64
	ModTime   time.Time
	Content   string
}

// Custom template functions
var funcMap = template.FuncMap{
	"formatSize": func(size int64) string {
		if size == 0 {
			return "0 B"
		}
		units := []string{"B", "KB", "MB", "GB"}
		unitIndex := 0
		sizeFloat := float64(size)
		for sizeFloat >= 1024 && unitIndex < len(units)-1 {
			sizeFloat /= 1024
			unitIndex++
		}
		if unitIndex == 0 {
			return fmt.Sprintf("%d %s", int(sizeFloat), units[unitIndex])
		}
		return fmt.Sprintf("%.1f %s", sizeFloat, units[unitIndex])
	},
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"urlquery": func(s string) string {
		return url.QueryEscape(s)
	},
	"joinPath": func(base, folder string) string {
		return filepath.Join(base, folder)
	},
	"parentPath": func(path string) string {
		return filepath.Dir(path)
	},
	"isImage": func(filename string) bool {
		ext := strings.ToLower(filepath.Ext(filename))
		return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
	},
	"isTextFile": func(filename string) bool {
		ext := strings.ToLower(filepath.Ext(filename))
		textExts := []string{".txt", ".md", ".log", ".csv", ".json", ".xml", ".html", ".css", ".js"}
		for _, textExt := range textExts {
			if ext == textExt {
				return true
			}
		}
		return false
	},
	"isPdf": func(filename string) bool {
		ext := strings.ToLower(filepath.Ext(filename))
		return ext == ".pdf"
	},
}

// fileManagerHandler handles the /files endpoint
func fileManagerHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse template with custom functions
		tmpl, err := template.New("files.html").Funcs(funcMap).ParseFiles("templates/files.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the path from query parameter
		path := r.URL.Query().Get("path")
		if path == "" {
			path = "."
		}

		// Get success or error messages from query parameters
		successMessage := r.URL.Query().Get("success")
		errorMessage := r.URL.Query().Get("error")

		// Sanitize path to prevent directory traversal
		sanitizedPath := filepath.Clean(path)
		if strings.Contains(sanitizedPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Full path to the directory
		fullPath := filepath.Join(cfg.StorageDir, sanitizedPath)
		
		// Check if the path exists and is a directory
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "Path not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Error accessing path: %v", err), http.StatusInternalServerError)
			return
		}
		
		if !info.IsDir() {
			http.Error(w, "Path is not a directory", http.StatusBadRequest)
			return
		}

		// Read directory contents
		entries, err := ioutil.ReadDir(fullPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read directory: %v", err), http.StatusInternalServerError)
			return
		}

		// Convert to FileInfo slice
		var files []FileInfo
		for _, entry := range entries {
			relPath := filepath.Join(sanitizedPath, entry.Name())
			// Calculate depth for indentation
			depth := strings.Count(sanitizedPath, string(os.PathSeparator))
			if sanitizedPath == "." {
				depth = 0
			}
			
			fileInfo := FileInfo{
				Name:     entry.Name(),
				Size:     entry.Size(),
				ModTime:  entry.ModTime(),
				IsDir:    entry.IsDir(),
				RelPath:  relPath,
				Depth:    depth,
			}
			files = append(files, fileInfo)
		}

		// Sort files: directories first, then by name
		sort.Slice(files, func(i, j int) bool {
			if files[i].IsDir && !files[j].IsDir {
				return true
			}
			if !files[i].IsDir && files[j].IsDir {
				return false
			}
			return files[i].Name < files[j].Name
		})

		// Create breadcrumb
		var breadcrumb []string
		if sanitizedPath != "." {
			breadcrumb = strings.Split(sanitizedPath, string(os.PathSeparator))
			// Remove empty strings from the beginning
			if len(breadcrumb) > 0 && breadcrumb[0] == "" {
				breadcrumb = breadcrumb[1:]
			}
			if len(breadcrumb) > 0 && breadcrumb[0] == "." {
				breadcrumb = breadcrumb[1:]
			}
		}

		// Prepare data for template
		data := FileManagerData{
			Files:          files,
			CurrentPath:    sanitizedPath,
			BasePath:       sanitizedPath,
			Breadcrumb:     breadcrumb,
			SuccessMessage: successMessage,
			ErrorMessage:   errorMessage,
		}

		// Execute template
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// previewHandler handles the /preview endpoint for file previews
func previewHandler(cfg *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse template with custom functions
		tmpl, err := template.New("preview.html").Funcs(funcMap).ParseFiles("templates/preview.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the filename from query parameter
		filename := r.URL.Query().Get("name")
		if filename == "" {
			http.Error(w, "Missing filename", http.StatusBadRequest)
			return
		}

		// Sanitize filename to prevent directory traversal
		sanitizedFilename := filepath.Clean(filename)
		if strings.Contains(sanitizedFilename, "..") {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}

		// Full path to the file
		filePath := filepath.Join(cfg.StorageDir, sanitizedFilename)

		// Check if file exists
		info, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Error accessing file: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if it's a directory
		if info.IsDir() {
			http.Error(w, "Cannot preview directories", http.StatusBadRequest)
			return
		}

		// Read file content for text files (first 100 lines)
		var content string
		if isTextFile(sanitizedFilename) {
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusInternalServerError)
				return
			}
			
			// Limit content to first 100 lines or 10KB, whichever is smaller
			contentStr := string(data)
			lines := strings.Split(contentStr, "\n")
			if len(lines) > 100 {
				lines = lines[:100]
				contentStr = strings.Join(lines, "\n") + "\n... (truncated)"
			}
			if len(contentStr) > 10240 { // 10KB
				contentStr = contentStr[:10240] + "\n... (truncated)"
			}
			content = contentStr
		}

		// Get base path for breadcrumb
		basePath := filepath.Dir(sanitizedFilename)
		if basePath == "." || basePath == "/" {
			basePath = ""
		}

		// Prepare data for template
		data := PreviewData{
			Filename: filepath.Base(sanitizedFilename),
			FilePath: sanitizedFilename,
			BasePath: basePath,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			Content:  content,
		}

		// Execute template
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// Helper function to check if a file is a text file
func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExts := []string{".txt", ".md", ".log", ".csv", ".json", ".xml", ".html", ".css", ".js", ".go", ".py", ".java", ".c", ".cpp"}
	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}
	return false
}