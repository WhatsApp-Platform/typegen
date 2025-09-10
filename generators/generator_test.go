package generators

import (
	"testing"
)

func TestInMemoryFS_WriteFile(t *testing.T) {
	fs := NewInMemoryFS()
	
	// Test writing a simple file
	content := []byte("test content")
	err := fs.WriteFile("test.txt", content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	
	// Verify file exists
	if !fs.FileExists("test.txt") {
		t.Error("File should exist after WriteFile")
	}
	
	// Verify content
	retrieved, exists := fs.GetFile("test.txt")
	if !exists {
		t.Error("GetFile should return true for existing file")
	}
	
	if string(retrieved) != string(content) {
		t.Errorf("Content mismatch. Expected %q, got %q", string(content), string(retrieved))
	}
}

func TestInMemoryFS_WriteFileWithDirs(t *testing.T) {
	fs := NewInMemoryFS()
	
	// Test writing a file in nested directories
	content := []byte("nested content")
	err := fs.WriteFile("path/to/nested/file.txt", content, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	
	// Verify directories were created
	expectedDirs := []string{"path", "path/to", "path/to/nested"}
	for _, dir := range expectedDirs {
		if !fs.DirExists(dir) {
			t.Errorf("Directory %q should exist", dir)
		}
	}
	
	// Verify file exists
	if !fs.FileExists("path/to/nested/file.txt") {
		t.Error("File should exist in nested directory")
	}
}

func TestInMemoryFS_MkdirAll(t *testing.T) {
	fs := NewInMemoryFS()
	
	err := fs.MkdirAll("deep/nested/directory", 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	
	// Verify all directories were created
	expectedDirs := []string{"deep", "deep/nested", "deep/nested/directory"}
	for _, dir := range expectedDirs {
		if !fs.DirExists(dir) {
			t.Errorf("Directory %q should exist", dir)
		}
	}
}

func TestInMemoryFS_ListFiles(t *testing.T) {
	fs := NewInMemoryFS()
	
	// Write multiple files
	files := []string{"file1.txt", "dir/file2.txt", "dir/subdir/file3.txt"}
	for _, file := range files {
		err := fs.WriteFile(file, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("WriteFile failed for %s: %v", file, err)
		}
	}
	
	// List files
	listedFiles := fs.ListFiles()
	
	// Verify all files are listed (should be sorted)
	expected := []string{"dir/file2.txt", "dir/subdir/file3.txt", "file1.txt"}
	if len(listedFiles) != len(expected) {
		t.Fatalf("Expected %d files, got %d", len(expected), len(listedFiles))
	}
	
	for i, expectedFile := range expected {
		if listedFiles[i] != expectedFile {
			t.Errorf("File mismatch at index %d. Expected %q, got %q", i, expectedFile, listedFiles[i])
		}
	}
}

func TestInMemoryFS_Join(t *testing.T) {
	fs := NewInMemoryFS()
	
	result := fs.Join("path", "to", "file.txt")
	
	// This should behave like filepath.Join
	expected := "path/to/file.txt"
	if result != expected {
		t.Errorf("Join result mismatch. Expected %q, got %q", expected, result)
	}
}

func TestInMemoryFS_GetFileString(t *testing.T) {
	fs := NewInMemoryFS()
	
	content := "test string content"
	err := fs.WriteFile("test.txt", []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	
	retrieved, exists := fs.GetFileString("test.txt")
	if !exists {
		t.Error("GetFileString should return true for existing file")
	}
	
	if retrieved != content {
		t.Errorf("Content mismatch. Expected %q, got %q", content, retrieved)
	}
}

func TestInMemoryFS_NonExistentFile(t *testing.T) {
	fs := NewInMemoryFS()
	
	// Test non-existent file
	_, exists := fs.GetFile("nonexistent.txt")
	if exists {
		t.Error("GetFile should return false for non-existent file")
	}
	
	if fs.FileExists("nonexistent.txt") {
		t.Error("FileExists should return false for non-existent file")
	}
	
	if fs.Exists("nonexistent.txt") {
		t.Error("Exists should return false for non-existent file")
	}
}