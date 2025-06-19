package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"suppress-checker/pkg/interfaces"
)

// FileSystemScanner implements the Scanner interface for scanning the filesystem
type FileSystemScanner struct {
	supportedFiles []string
	excludePatterns []string
	includePatterns []string
}

// NewFileSystemScanner creates a new filesystem scanner
func NewFileSystemScanner(supportedFiles []string) *FileSystemScanner {
	if len(supportedFiles) == 0 {
		supportedFiles = []string{".tryvi-ignore", "suppressions.xml", ".suppress-ignore"}
	}
	
	return &FileSystemScanner{
		supportedFiles: supportedFiles,
	}
}

// NewTryviScanner creates a scanner specifically for .tryvi-ignore files
func NewTryviScanner() *FileSystemScanner {
	return NewFileSystemScanner([]string{".tryvi-ignore"})
}

// SetExcludePatterns sets patterns to exclude during scanning
func (s *FileSystemScanner) SetExcludePatterns(patterns []string) {
	s.excludePatterns = patterns
}

// SetIncludePatterns sets patterns to include during scanning
func (s *FileSystemScanner) SetIncludePatterns(patterns []string) {
	s.includePatterns = patterns
}

// Scan recursively searches for suppression files in the given directory
func (s *FileSystemScanner) Scan(ctx context.Context, rootDir string) ([]string, error) {
	var foundFiles []string
	
	// Clean and validate root directory
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", rootDir, err)
	}
	
	// Check if root directory exists
	if _, err := os.Stat(absRootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", absRootDir)
	}
	
	err = filepath.WalkDir(absRootDir, func(path string, d os.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if err != nil {
			// Log error but continue scanning
			return nil
		}
		
		// Skip directories
		if d.IsDir() {
			// Check if we should skip this directory
			if s.shouldExcludeDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		
		// Check if this file matches our supported files
		fileName := d.Name()
		if s.isSupportedFile(fileName) && s.shouldIncludeFile(path) && !s.shouldExcludeFile(path) {
			foundFiles = append(foundFiles, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", absRootDir, err)
	}
	
	return foundFiles, nil
}

// SupportedFileNames returns the list of file names this scanner looks for
func (s *FileSystemScanner) SupportedFileNames() []string {
	return s.supportedFiles
}

// isSupportedFile checks if the given filename is in our supported files list
func (s *FileSystemScanner) isSupportedFile(fileName string) bool {
	for _, supported := range s.supportedFiles {
		if fileName == supported {
			return true
		}
	}
	return false
}

// shouldExcludeDir checks if a directory should be excluded from scanning
func (s *FileSystemScanner) shouldExcludeDir(dirPath string) bool {
	// Common directories to exclude
	commonExcludes := []string{".git", ".svn", ".hg", "node_modules", ".vscode", ".idea"}
	
	dirName := filepath.Base(dirPath)
	for _, exclude := range commonExcludes {
		if dirName == exclude {
			return true
		}
	}
	
	// Check user-defined exclude patterns
	for _, pattern := range s.excludePatterns {
		if matched, _ := filepath.Match(pattern, dirPath); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, dirName); matched {
			return true
		}
	}
	
	return false
}

// shouldIncludeFile checks if a file should be included based on include patterns
func (s *FileSystemScanner) shouldIncludeFile(filePath string) bool {
	// If no include patterns specified, include all
	if len(s.includePatterns) == 0 {
		return true
	}
	
	for _, pattern := range s.includePatterns {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
	}
	
	return false
}

// shouldExcludeFile checks if a file should be excluded based on exclude patterns
func (s *FileSystemScanner) shouldExcludeFile(filePath string) bool {
	for _, pattern := range s.excludePatterns {
		if matched, _ := filepath.Match(pattern, filePath); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return true
		}
		// Check if the file is in an excluded directory path
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	
	return false
}

// Ensure FileSystemScanner implements the Scanner interface
var _ interfaces.Scanner = (*FileSystemScanner)(nil) 