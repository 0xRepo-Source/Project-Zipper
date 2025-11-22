package zipper

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNextArchiveName(t *testing.T) {
	tmp := t.TempDir()

	path0, err := NextArchiveName(tmp, "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path0) != "project.zip" {
		t.Fatalf("expected project.zip, got %s", filepath.Base(path0))
	}

	if err := os.WriteFile(path0, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to create zip placeholder: %v", err)
	}

	path1, err := NextArchiveName(tmp, "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path1) != "project-v1.zip" {
		t.Fatalf("expected project-v1.zip, got %s", filepath.Base(path1))
	}

	if err := os.WriteFile(path1, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to create versioned zip placeholder: %v", err)
	}

	path2, err := NextArchiveName(tmp, "project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path2) != "project-v2.zip" {
		t.Fatalf("expected project-v2.zip, got %s", filepath.Base(path2))
	}
}
