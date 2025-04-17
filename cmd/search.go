package cmd

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

const DEFAULT_GO_CHANNELS int = 10

/*
SearchFilesParallel efficiently searches for all `.cdslck` files under the given rootDirectory.

It performs the search recursively, but processes each subdirectory in parallel using goroutines,
limited by a semaphore to avoid overwhelming system resources. The function uses a WaitGroup
to track active directory-processing goroutines and a channel to collect matching file paths.

This parallel approach significantly speeds up traversal of large directory trees while
maintaining controlled concurrency.
*/
func SearchFilesParallel(rootDirectory string) []string {
	// Use a channel to collect results from goroutines
	filesChan := make(chan string, 100) // Buffer of 100 files for performance, not a limitation
	// Semaphore to limit concurrency to 10 goroutines
	semaphore := make(chan struct{}, DEFAULT_GO_CHANNELS)

	// waitgroup to track all directory processing
	var wg sync.WaitGroup

	// Function to process a directory with semaphore limit
	var processDir func(dir string)
	processDir = func(dir string) {
		defer func() {
			<-semaphore // Release semaphore when done
			wg.Done()
		}()

		// Use WalkDir but only for the current directory level
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip on error
			}

			// Skip processing subdirectories in this walk
			if d.IsDir() && path != dir {
				// Add the subdirectory to be processed separately
				wg.Add(1)
				go func(subdir string) {
					// Wait for a slot in the worker pool
					semaphore <- struct{}{} // Acquire semaphore
					processDir(subdir)
				}(path)
				return filepath.SkipDir // Skip the directory in current walk
			}

			// Process files
			if !d.IsDir() && strings.HasSuffix(path, ".cdslck") {
				filesChan <- path
			}

			return nil
		})

	}
	// Start with root directory
	wg.Add(1)
	semaphore <- struct{}{} // Acquire semaphore for first directory
	go processDir(rootDirectory)

	// Wait for all directory processing to complete
	go func() {
		wg.Wait()
		close(filesChan)
	}()

	// Collect all results
	var files []string
	for file := range filesChan {
		files = append(files, file)
	}

	return files
}
