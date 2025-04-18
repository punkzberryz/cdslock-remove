package search

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSearchFilesParallel checks whether SearchFilesParallel correctly finds only .cdslck files
func TestSearchFilesParallel(t *testing.T) {
	// Create a temporary root directory
	rootDir := t.TempDir()

	// Define the files and their relative paths
	filesToCreate := []string{
		"file1.cdslck",
		"file2.cdslck",
		"file3.txt",
		"subdir1/file4.cdslck",
		"subdir1/file5.log",
		"subdir2/subsub/file6.cdslck",
	}

	// Expected .cdslck file full paths
	expected := map[string]bool{}

	// Create the files
	for _, relPath := range filesToCreate {
		fullPath := filepath.Join(rootDir, relPath)

		// Create subdirectories as needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Write a dummy file
		if err := os.WriteFile(fullPath, []byte("test data"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Add to expected if it's a .cdslck file
		if filepath.Ext(fullPath) == ".cdslck" {
			expected[fullPath] = true
		}
	}

	// Run the actual function
	found := SearchFilesParallel(rootDir)

	// Check that we found the correct number of files
	if len(found) != len(expected) {
		t.Errorf("Expected %d files, but got %d", len(expected), len(found))
	}

	// Check that each returned file is expected
	for _, path := range found {
		if _, ok := expected[path]; !ok {
			t.Errorf("Unexpected file found: %s", path)
		}
	}
}
