package runcmd

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

/* ------------------------------------------------------------------------- */
/* Test getLastRunTimestamp                                                  */
/* ------------------------------------------------------------------------- */

func TestGetLastRunTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".tempo-lastrun")

	// Case 1: File does not exist, should return 0
	if ts := getLastRunTimestamp(cacheFile); ts != 0 {
		t.Errorf("Expected timestamp 0 for non-existent file, got %d", ts)
	}

	// Case 2: Valid timestamp file
	expectedTimestamp := time.Now().Unix()
	if err := os.WriteFile(cacheFile, []byte(strconv.FormatInt(expectedTimestamp, 10)), 0644); err != nil {
		t.Fatalf("Failed to create test cache file: %v", err)
	}

	ts := getLastRunTimestamp(cacheFile)
	if ts == 0 {
		t.Errorf("Expected a valid timestamp, got 0")
	}

	// Case 3: File contains invalid data (not an integer)
	if err := os.WriteFile(cacheFile, []byte("invalid-timestamp"), 0644); err != nil {
		t.Fatalf("Failed to create invalid timestamp file: %v", err)
	}

	ts = getLastRunTimestamp(cacheFile)
	if ts != 0 {
		t.Errorf("Expected timestamp 0 for invalid data, got %d", ts)
	}

	// Case 4: File is empty
	if err := os.WriteFile(cacheFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty timestamp file: %v", err)
	}

	ts = getLastRunTimestamp(cacheFile)
	if ts != 0 {
		t.Errorf("Expected timestamp 0 for empty file, got %d", ts)
	}

	// Case 5: File cannot be read due to permissions
	if err := os.Chmod(cacheFile, 0000); err == nil { // Set to unreadable
		defer func() {
			// Restore permission after test
			if err := os.Chmod(cacheFile, 0644); err != nil {
				t.Logf("Warning: Failed to restore file permissions for %s: %v", cacheFile, err)
			}
		}()
		ts = getLastRunTimestamp(cacheFile)
		if ts != 0 {
			t.Errorf("Expected timestamp 0 for unreadable file, got %d", ts)
		}
	} else {
		t.Log("[SKIP] Skipping unreadable file test due to OS limitations")
	}
}

/* ------------------------------------------------------------------------- */
/* Test saveLastRunTimestamp                                                 */
/* ------------------------------------------------------------------------- */

func TestSaveLastRunTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	cacheFile := filepath.Join(tempDir, ".tempo-lastrun")

	// Case 1: Save the last run timestamp
	err := saveLastRunTimestamp(cacheFile)
	if err != nil {
		t.Fatalf("Failed to save last run timestamp: %v", err)
	}

	// Read the saved timestamp
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		t.Fatalf("Failed to read saved timestamp: %v", err)
	}

	if len(data) == 0 {
		t.Errorf("Expected a timestamp value in the file, but it is empty")
	}

	// Case 2: Attempt to save to a read-only directory
	readOnlyDir := filepath.Join(tempDir, "readonly")
	if err := os.Mkdir(readOnlyDir, 0555); err != nil { // Read-only directory
		t.Fatalf("Failed to create read-only directory: %v", err)
	}
	readonlyCacheFile := filepath.Join(readOnlyDir, ".tempo-lastrun")

	err = saveLastRunTimestamp(readonlyCacheFile)
	if err == nil {
		t.Errorf("Expected an error when writing to a read-only directory, but got none")
	}
}

/* ------------------------------------------------------------------------- */
/* Test getFileLastModifiedTime                                              */
/* ------------------------------------------------------------------------- */

func TestGetFileLastModifiedTime(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "testfile.txt")

	// Create test file
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get last modified time
	ts, err := getFileLastModifiedTime(testFile)
	if err != nil {
		t.Fatalf("Failed to get last modified time: %v", err)
	}

	if ts <= 0 {
		t.Errorf("Expected a valid last modified timestamp, got %d", ts)
	}

	// Case 1: Non-existent file should return an error
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	_, err = getFileLastModifiedTime(nonExistentFile)
	if err == nil {
		t.Errorf("Expected an error for non-existent file, but got none")
	}

}
