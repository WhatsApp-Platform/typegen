package generators

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// InMemoryFS implements FS interface for testing purposes
type InMemoryFS struct {
	files map[string][]byte
	dirs  map[string]bool
}

// NewInMemoryFS creates a new in-memory filesystem for testing
func NewInMemoryFS() *InMemoryFS {
	return &InMemoryFS{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

// WriteFile implements FS.WriteFile
func (fs *InMemoryFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	// Normalize path separators
	name = filepath.ToSlash(name)
	
	// Track directory creation
	dir := filepath.Dir(name)
	if dir != "." {
		fs.dirs[dir] = true
		// Create all parent directories
		parts := strings.Split(dir, "/")
		for i := range parts {
			parentDir := strings.Join(parts[:i+1], "/")
			fs.dirs[parentDir] = true
		}
	}
	
	// Store the file
	fs.files[name] = make([]byte, len(data))
	copy(fs.files[name], data)
	
	return nil
}

// MkdirAll implements FS.MkdirAll
func (fs *InMemoryFS) MkdirAll(path string, perm os.FileMode) error {
	// Normalize path separators
	path = filepath.ToSlash(path)
	
	fs.dirs[path] = true
	
	// Create all parent directories
	parts := strings.Split(path, "/")
	for i := range parts {
		parentDir := strings.Join(parts[:i+1], "/")
		fs.dirs[parentDir] = true
	}
	
	return nil
}

// Join implements FS.Join
func (fs *InMemoryFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// GetFile returns the content of a file for testing assertions
func (fs *InMemoryFS) GetFile(path string) ([]byte, bool) {
	path = filepath.ToSlash(path)
	content, exists := fs.files[path]
	if !exists {
		return nil, false
	}
	// Return a copy to prevent modification
	result := make([]byte, len(content))
	copy(result, content)
	return result, true
}

// GetFileString returns the content of a file as a string for testing assertions
func (fs *InMemoryFS) GetFileString(path string) (string, bool) {
	content, exists := fs.GetFile(path)
	if !exists {
		return "", false
	}
	return string(content), true
}

// ListFiles returns all file paths that have been written
func (fs *InMemoryFS) ListFiles() []string {
	var files []string
	for path := range fs.files {
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

// ListDirs returns all directory paths that have been created
func (fs *InMemoryFS) ListDirs() []string {
	var dirs []string
	for path := range fs.dirs {
		dirs = append(dirs, path)
	}
	sort.Strings(dirs)
	return dirs
}

// Exists checks if a file or directory exists
func (fs *InMemoryFS) Exists(path string) bool {
	path = filepath.ToSlash(path)
	_, fileExists := fs.files[path]
	_, dirExists := fs.dirs[path]
	return fileExists || dirExists
}

// FileExists checks if a specific file exists
func (fs *InMemoryFS) FileExists(path string) bool {
	path = filepath.ToSlash(path)
	_, exists := fs.files[path]
	return exists
}

// DirExists checks if a specific directory exists
func (fs *InMemoryFS) DirExists(path string) bool {
	path = filepath.ToSlash(path)
	_, exists := fs.dirs[path]
	return exists
}