package output

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemWriterWritesFile(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewFileSystemWriter(tempDir)

	data := []byte("hello world")
	if err := writer.Write(context.Background(), "report.json", data); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	path := filepath.Join(tempDir, "report.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file to exist, got %v", err)
	}
	if string(contents) != string(data) {
		t.Fatalf("expected %q, got %q", string(data), string(contents))
	}
}

func TestFileSystemWriterCreatesDefaultDir(t *testing.T) {
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousDir)
	})

	writer := NewFileSystemWriter("")
	data := []byte("summary")
	if err := writer.Write(context.Background(), "summary.md", data); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	path := filepath.Join(workingDir, "dist", "summary.md")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected file to exist, got %v", err)
	}
	if string(contents) != string(data) {
		t.Fatalf("expected %q, got %q", string(data), string(contents))
	}
}
