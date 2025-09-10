package generators

import (
	"context"
	"os"
	"path/filepath"

	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

// Generator defines the interface for code generators
type Generator interface {
	// SetConfig sets the configuration options for the generator
	SetConfig(config map[string]string)
	
	// Generate generates code for an entire module
	Generate(ctx context.Context, module *ast.Module, dest FS) error
}

// FS provides a filesystem abstraction that supports writing
// Compatible with fs.FS but adds write operations
type FS interface {
	// WriteFile writes data to a file, creating directories as needed
	WriteFile(name string, data []byte, perm os.FileMode) error
	
	// MkdirAll creates a directory and all necessary parents
	MkdirAll(path string, perm os.FileMode) error
	
	// Join joins path elements into a single path
	Join(elem ...string) string
}

// osFS implements FS using the os package for real filesystem operations
type osFS struct {
	root string
}

// NewOSFS creates a new filesystem rooted at the given directory
func NewOSFS(root string) FS {
	return &osFS{root: root}
}

// WriteFile implements FS.WriteFile
func (fs *osFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	fullPath := filepath.Join(fs.root, name)
	
	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	return os.WriteFile(fullPath, data, perm)
}

// MkdirAll implements FS.MkdirAll
func (fs *osFS) MkdirAll(path string, perm os.FileMode) error {
	fullPath := filepath.Join(fs.root, path)
	return os.MkdirAll(fullPath, perm)
}

// Join implements FS.Join
func (fs *osFS) Join(elem ...string) string {
	return filepath.Join(elem...)
}