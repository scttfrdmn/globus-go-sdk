// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transfer

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"
)

// FileIterator provides a memory-efficient way to iterate through files
// in a recursive directory structure without loading everything into memory
type FileIterator interface {
	// Next returns the next file or directory, or false if iteration is complete
	Next() (FileListItem, bool)
	
	// Error returns any error that occurred during iteration
	Error() error
	
	// Reset restarts the iteration from the beginning
	Reset() error
	
	// Close releases resources used by the iterator
	Close() error
}

// StreamingFileIterator implements FileIterator by streaming files from a directory
// and using a queue to iterate through subdirectories without loading everything
// into memory at once
type StreamingFileIterator struct {
	client       *Client
	endpointID   string
	rootPath     string
	queue        []string
	currentFiles []FileListItem
	fileIndex    int
	dirIndex     int
	recursive    bool
	showHidden   bool
	err          error
	mu           sync.Mutex
	listedDirs   map[string]bool
	maxDepth     int
	currentDepth int
	concurrency  int
	wg           sync.WaitGroup
	// Channels
	resultChan   chan FileListItem
	errorChan    chan error
	closeChan    chan struct{}
}

// NewStreamingFileIterator creates a new StreamingFileIterator
func NewStreamingFileIterator(
	ctx context.Context, 
	client *Client, 
	endpointID, 
	rootPath string, 
	options *StreamingIteratorOptions,
) (*StreamingFileIterator, error) {
	if options == nil {
		options = &StreamingIteratorOptions{
			Recursive:   true,
			ShowHidden:  true,
			MaxDepth:    -1, // No limit
			Concurrency: 4,
		}
	}
	
	iterator := &StreamingFileIterator{
		client:       client,
		endpointID:   endpointID,
		rootPath:     rootPath,
		queue:        []string{rootPath},
		listedDirs:   make(map[string]bool),
		recursive:    options.Recursive,
		showHidden:   options.ShowHidden,
		maxDepth:     options.MaxDepth,
		concurrency:  options.Concurrency,
		resultChan:   make(chan FileListItem, options.Concurrency*100),
		errorChan:    make(chan error, 1),
		closeChan:    make(chan struct{}),
	}
	
	// Start the initial crawl
	iterator.startCrawling(ctx)
	
	return iterator, nil
}

// StreamingIteratorOptions contains options for the streaming iterator
type StreamingIteratorOptions struct {
	// Recursive specifies whether to list directories recursively
	Recursive bool
	
	// ShowHidden specifies whether to include hidden files
	ShowHidden bool
	
	// MaxDepth is the maximum depth to recurse (-1 means no limit)
	MaxDepth int
	
	// Concurrency is the number of concurrent directory listings
	Concurrency int
}

// Next returns the next file or directory, or false if iteration is complete
func (s *StreamingFileIterator) Next() (FileListItem, bool) {
	select {
	case file, ok := <-s.resultChan:
		if !ok {
			return FileListItem{}, false
		}
		return file, true
	case err := <-s.errorChan:
		s.err = err
		return FileListItem{}, false
	case <-s.closeChan:
		return FileListItem{}, false
	}
}

// Error returns any error that occurred during iteration
func (s *StreamingFileIterator) Error() error {
	return s.err
}

// Reset restarts the iteration from the beginning
func (s *StreamingFileIterator) Reset() error {
	s.Close()
	
	// Clear channels
	s.resultChan = make(chan FileListItem, s.concurrency*100)
	s.errorChan = make(chan error, 1)
	s.closeChan = make(chan struct{})
	
	// Reset state
	s.queue = []string{s.rootPath}
	s.listedDirs = make(map[string]bool)
	s.currentFiles = nil
	s.fileIndex = 0
	s.dirIndex = 0
	s.currentDepth = 0
	s.err = nil
	
	// Start crawling again
	s.startCrawling(context.Background())
	
	return nil
}

// Close releases resources used by the iterator
func (s *StreamingFileIterator) Close() error {
	// Signal worker goroutines to stop
	close(s.closeChan)
	
	// Wait for all workers to finish
	s.wg.Wait()
	
	// Clean up channels
	drainChannel(s.resultChan)
	drainErrorChannel(s.errorChan)
	
	return nil
}

// startCrawling begins the process of crawling the directory structure
func (s *StreamingFileIterator) startCrawling(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(s.resultChan)
		defer close(s.errorChan)
		
		// Create a semaphore to limit concurrency
		sem := make(chan struct{}, s.concurrency)
		
		// Create a context that's canceled when Close is called
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		
		// Monitor close signal
		go func() {
			select {
			case <-s.closeChan:
				cancel()
			case <-ctx.Done():
				// Context was canceled elsewhere
			}
		}()
		
		// Start with the root directory
		for len(s.queue) > 0 {
			select {
			case <-ctx.Done():
				return
			case <-s.closeChan:
				return
			default:
				// Process the next directory
				s.mu.Lock()
				currentDir := s.queue[0]
				s.queue = s.queue[1:]
				s.mu.Unlock()
				
				// Skip if already listed
				if s.listedDirs[currentDir] {
					continue
				}
				
				// Mark as listed
				s.listedDirs[currentDir] = true
				
				// Acquire semaphore
				sem <- struct{}{}
				
				// Process directory in a goroutine
				s.wg.Add(1)
				go func(dir string, depth int) {
					defer s.wg.Done()
					defer func() { <-sem }()
					
					s.processDirectory(ctx, dir, depth)
				}(currentDir, s.calculateDepth(currentDir))
			}
		}
		
		// Wait for all directory listings to complete
		for i := 0; i < s.concurrency; i++ {
			sem <- struct{}{}
		}
	}()
}

// processDirectory processes a single directory
func (s *StreamingFileIterator) processDirectory(ctx context.Context, dir string, depth int) {
	// List files in the directory
	listOptions := &ListFileOptions{
		ShowHidden: s.showHidden,
	}
	
	// Debug print
	fmt.Printf("DEBUG: Listing directory: %s (depth %d)\n", dir, depth)
	
	var listing *FileList
	var err error
	
	// Just use the real client directly
	listing, err = s.client.ListFiles(ctx, s.endpointID, dir, listOptions)
	
	if err != nil {
		fmt.Printf("DEBUG: Error listing directory %s: %v\n", dir, err)
		select {
		case s.errorChan <- err:
		default:
			// Error channel is full
		}
		return
	}
	
	// Debug print
	fmt.Printf("DEBUG: Got %d items in directory %s\n", len(listing.Data), dir)
	
	// Process files
	for _, file := range listing.Data {
		fmt.Printf("DEBUG: Processing file: %s (type: %s)\n", file.Name, file.Type)
		
		select {
		case <-ctx.Done():
			return
		case <-s.closeChan:
			return
		case s.resultChan <- file:
			// Successfully sent file
			fmt.Printf("DEBUG: Sent file to channel: %s\n", file.Name)
		}
		
		// If it's a directory and we're recursive, add it to the queue
		if file.Type == "dir" && s.recursive && (s.maxDepth < 0 || depth < s.maxDepth) {
			dirPath := path.Join(dir, file.Name)
			fmt.Printf("DEBUG: Adding directory to queue: %s\n", dirPath)
			
			s.mu.Lock()
			if !s.listedDirs[dirPath] {
				s.queue = append(s.queue, dirPath)
				fmt.Printf("DEBUG: Queue length is now %d\n", len(s.queue))
			}
			s.mu.Unlock()
		}
	}
}

// calculateDepth calculates the depth of a path relative to the root path
func (s *StreamingFileIterator) calculateDepth(path string) int {
	// Simple implementation for depth calculation
	if path == s.rootPath {
		return 0
	}
	
	depth := 0
	rootLen := len(s.rootPath)
	
	for i := rootLen; i < len(path); i++ {
		if path[i] == '/' {
			depth++
		}
	}
	
	return depth
}

// drainChannel drains a channel to prevent goroutine leaks
func drainChannel(ch chan FileListItem) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

// drainErrorChannel drains an error channel to prevent goroutine leaks
func drainErrorChannel(ch chan error) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

// CollectFiles gathers files from an iterator into a slice
// Warning: This loads all files into memory and should only be used when
// the number of files is known to be reasonably small
func CollectFiles(iterator FileIterator) ([]FileListItem, error) {
	var files []FileListItem
	
	// Add a small delay to let the iterator process the directory queue
	time.Sleep(50 * time.Millisecond)
	
	for {
		file, ok := iterator.Next()
		if !ok {
			if err := iterator.Error(); err != nil {
				return nil, err
			}
			break
		}
		
		files = append(files, file)
	}
	
	return files, nil
}