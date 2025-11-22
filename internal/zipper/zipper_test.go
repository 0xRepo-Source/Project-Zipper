package zipper

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipWithProgress(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "src")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	fileA := filepath.Join(root, "a.txt")
	fileB := filepath.Join(root, "nested", "b.txt")

	if err := os.MkdirAll(filepath.Dir(fileB), 0o755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	if err := os.WriteFile(fileA, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to write a.txt: %v", err)
	}
	if err := os.WriteFile(fileB, []byte("world!"), 0o644); err != nil {
		t.Fatalf("failed to write b.txt: %v", err)
	}

	archive := filepath.Join(tmp, "archive.zip")
	var calls []struct {
		done  int64
		total int64
	}

	stats, err := ZipWithProgress(root, archive, func(done, total int64) {
		calls = append(calls, struct {
			done  int64
			total int64
		}{done: done, total: total})
	})
	if err != nil {
		t.Fatalf("ZipWithProgress returned error: %v", err)
	}

	expectedSize := int64(len("hello") + len("world!"))
	if stats.TotalBytes != expectedSize {
		t.Fatalf("expected total bytes %d, got %d", expectedSize, stats.TotalBytes)
	}
	if stats.FileCount != 2 {
		t.Fatalf("expected file count 2, got %d", stats.FileCount)
	}

	if len(calls) == 0 {
		t.Fatalf("expected progress callbacks, got none")
	}

	last := calls[len(calls)-1]
	if last.total != expectedSize {
		t.Fatalf("expected final total %d, got %d", expectedSize, last.total)
	}
	if last.done != expectedSize {
		t.Fatalf("expected final done %d, got %d", expectedSize, last.done)
	}

	zipFile, err := os.Open(archive)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	t.Cleanup(func() { _ = zipFile.Close() })

	info, err := zipFile.Stat()
	if err != nil {
		t.Fatalf("failed to stat archive: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected non-empty archive")
	}

	reader, err := zip.NewReader(zipFile, info.Size())
	if err != nil {
		t.Fatalf("failed to open archive reader: %v", err)
	}
	nonDir := 0
	for _, f := range reader.File {
		if fi := f.FileInfo(); fi != nil && !fi.IsDir() {
			nonDir++
		}
	}
	if nonDir != 2 {
		t.Fatalf("expected 2 file entries in archive, got %d", nonDir)
	}
}
