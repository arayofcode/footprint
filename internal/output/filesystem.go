package output

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type FileSystemWriter struct {
	OutputDir string
}

func NewFileSystemWriter(outputDir string) *FileSystemWriter {
	if outputDir == "" {
		outputDir = "dist"
	}
	return &FileSystemWriter{OutputDir: outputDir}
}

func (w *FileSystemWriter) Write(ctx context.Context, filename string, data []byte) error {
	if err := os.MkdirAll(w.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	path := filepath.Join(w.OutputDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", filename, err)
	}

	return nil
}
