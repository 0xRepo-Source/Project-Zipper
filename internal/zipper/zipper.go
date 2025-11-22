package zipper

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ProgressFunc reports the number of source bytes processed out of the total.
type ProgressFunc func(done, total int64)

// ArchiveStats describes the payload processed while creating an archive.
type ArchiveStats struct {
	TotalBytes int64
	FileCount  int
}

// Zip archives the contents of srcDir into zipPath using only the Go standard library.
func Zip(srcDir, zipPath string) error {
	_, err := ZipWithProgress(srcDir, zipPath, nil)
	return err
}

// ZipWithProgress is identical to Zip but reports progress via the callback.
func ZipWithProgress(srcDir, zipPath string, progress ProgressFunc) (stats ArchiveStats, err error) {
	stats, err = scanDirectory(srcDir)
	if err != nil {
		return stats, err
	}

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return stats, err
	}
	defer func() {
		if cerr := zipFile.Close(); err == nil {
			err = cerr
		}
	}()

	writer := zip.NewWriter(zipFile)
	defer func() {
		if cerr := writer.Close(); err == nil {
			err = cerr
		}
	}()

	done := int64(0)
	callProgress := func() {
		if progress != nil {
			progress(done, stats.TotalBytes)
		}
	}
	callProgress()

	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(rel)
		if d.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writerEntry, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		pw := progressWriter{
			w:        writerEntry,
			done:     &done,
			total:    stats.TotalBytes,
			progress: progress,
		}

		if _, err = io.Copy(&pw, file); err != nil {
			file.Close()
			return err
		}
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}

		return nil
	})
	if err != nil {
		return stats, err
	}

	callProgress()
	return stats, err
}

type progressWriter struct {
	w        io.Writer
	done     *int64
	total    int64
	progress ProgressFunc
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	if pw.progress != nil && n > 0 {
		*pw.done += int64(n)
		pw.progress(*pw.done, pw.total)
	}
	return n, err
}

func scanDirectory(root string) (ArchiveStats, error) {
	stats := ArchiveStats{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		stats.TotalBytes += info.Size()
		stats.FileCount++
		return nil
	})
	return stats, err
}
